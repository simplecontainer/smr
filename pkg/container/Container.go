package container

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io/ioutil"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"smr/pkg/runtime"
	"strconv"
	"strings"
	"time"
)

func NewContainerFromDefinition(runtime runtime.Runtime, name string, definition definitions.Definition) *Container {
	// Make deep copy of the definition so we can preseve it for deep equals later
	definitionEncoded, err := json.Marshal(definition)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}

	var definitionCopy definitions.Definition

	err = json.Unmarshal(definitionEncoded, &definitionCopy)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}

	return &Container{
		Static: Static{
			Name:                   definition.Meta.Name,
			GeneratedName:          name,
			GeneratedNameNoProject: strings.Replace(name, fmt.Sprintf("%s-", runtime.PROJECT), "", 1),
			Group:                  definition.Meta.Group,
			Image:                  definition.Spec.Container.Image,
			Tag:                    definition.Spec.Container.Tag,
			Replicas:               definition.Spec.Container.Replicas,
			Networks:               definition.Spec.Container.Networks,
			Env:                    definition.Spec.Container.Envs,
			MappingFiles:           definition.Spec.Container.FileMounts,
			MappingPorts:           definition.Spec.Container.Ports,
			ExposedPorts:           convertPortMappingsToExposedPorts(definition.Spec.Container.Ports),
			Definition:             definitionCopy,
		},
		Runtime: Runtime{
			Auth:          GetAuth(definition.Spec.Container.Image, runtime),
			Id:            "",
			Networks:      map[string]Network{},
			State:         "",
			FoundRunning:  false,
			FirstObserved: true,
			Configuration: definition.Spec.Container.Configuration,
			Resources:     mapAnyToResources(definition.Spec.Container.Resources),
		},
		Status: Status{
			BackOffRestart: false,
			Healthy:        true,
			Ready:          true,
			Running:        false,
		},
	}
}

func Existing(name string) *Container {
	container := &Container{
		Static: Static{
			Name:                   name,
			GeneratedName:          name,
			GeneratedNameNoProject: name,
			Image:                  "image",
			Tag:                    "tag",
			Networks:               []string{"network"},
		},
		Runtime: Runtime{
			Id:            "",
			State:         "",
			FoundRunning:  false,
			FirstObserved: true,
			Networks:      map[string]Network{},
			Ready:         false,
			Configuration: make(map[string]any),
		},
		Status: Status{
			BackOffRestart: false,
			Healthy:        true,
			Ready:          true,
		},
	}

	container.Static.Name = name

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		panic(err)
	}

	if c := container.self(containers); c != nil {
		data, _ := cli.ContainerInspect(ctx, container.Runtime.Id)

		if c != nil && c.State == "running" {
			container.Runtime.FoundRunning = true

			for _, netw := range container.Static.Networks {
				if data.NetworkSettings.Networks[netw] != nil {
					container.Runtime.Networks[netw] = Network{
						NetworkId: data.NetworkSettings.Networks[netw].NetworkID,
						IP:        data.NetworkSettings.Networks[netw].IPAddress,
					}
				}
			}

			container.Runtime.Id = data.ID
			container.Runtime.State = data.State.Status
		}

		container.Runtime.FirstObserved = false

		return container
	} else {
		return nil
	}
}

func GetContainers() []types.Container {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})

	if err != nil {
		panic(err)
	}

	containersFiltered := make([]types.Container, 0, 0)

	for _, container := range containers {
		if container.Labels["managed"] == "smr" && container.State == "running" {
			containersFiltered = append(containersFiltered, container)
		}
	}

	return containersFiltered
}

func GetContainersAllStates() []types.Container {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		panic(err)
	}

	containersFiltered := make([]types.Container, 0, 0)

	for _, container := range containers {
		if container.Labels["managed"] == "ghostmgr" {
			containersFiltered = append(containersFiltered, container)
		}
	}

	return containersFiltered
}

func (container *Container) Run(runtime runtime.Runtime, Badger *badger.DB) (*types.Container, error) {
	c := container.Get()

	if c == nil {
		err := container.CreateNetwork()

		if err != nil {
			return nil, err
		}

		ctx := context.Background()
		cli := &client.Client{}

		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}
		defer cli.Close()

		err = container.PullImage(ctx, cli)

		if err != nil {
			return nil, err
		}

		resp := dockerContainer.ContainerCreateCreatedBody{}

		resp, err = cli.ContainerCreate(ctx, &dockerContainer.Config{
			Labels:       container.GenerateLabels(),
			Image:        container.Static.Image + ":" + container.Static.Tag,
			Env:          container.Static.Env,
			Tty:          false,
			ExposedPorts: container.exposedPorts(),
		}, &dockerContainer.HostConfig{
			Mounts:       container.mappingToMounts(runtime),
			PortBindings: container.portMappings(),
		}, container.GenerateNetwork(), nil, container.Static.GeneratedName)

		if err != nil {
			return nil, err
		}

		if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return nil, err
		}

		data, err := cli.ContainerInspect(ctx, resp.ID)

		if err != nil {
			return nil, err
		}

		for _, dnetw := range data.NetworkSettings.Networks {
			container.Runtime.Networks[dnetw.NetworkID] = Network{
				dnetw.NetworkID,
				dnetw.IPAddress,
			}
		}

		agent := Existing("smr-agent")
		base := fmt.Sprintf("%s.%s.%s", "runtime", container.Static.Group, container.Static.GeneratedName)

		if agent != nil {
			for _, nid := range container.Runtime.Networks {
				logger.Log.Info(fmt.Sprintf("trying to attach smr-agent to the network %s", nid.NetworkId))

				if !container.AgentConnectToTheSameNetwork(agent.Runtime.Id, nid.NetworkId) {
					container.Stop()
					container.Delete()
					return c, errors.New("failed to attach agent to the same network as the container, stopping container and cleaning up")
				} else {
					logger.Log.Info(fmt.Sprintf("smr-agent attached to the network %s", nid.NetworkId))
					database.Put(Badger, fmt.Sprintf("%s.%s", base, "ip"), container.Runtime.Networks[nid.NetworkId].IP)
				}
			}

			container.Runtime.FoundRunning = false

			database.Put(Badger, fmt.Sprintf("%s.%s", base, "foundrunning"), strconv.FormatBool(container.Runtime.FoundRunning))

			return container.Get(), nil
		} else {
			logger.Log.Error(fmt.Sprintf("smr-agent not found"))
			container.Stop()
			container.Delete()
			return nil, errors.New("failed to find smr-agent container and cleaning up everything")
		}
	} else {
		return c, nil
	}
}
func (container *Container) Get() *types.Container {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		panic(err)
	}

	if c := container.self(containers); c != nil {
		data, _ := cli.ContainerInspect(ctx, container.Runtime.Id)

		if c != nil && c.State == "running" {
			if container.Runtime.FirstObserved {
				container.Runtime.FoundRunning = true
			}

			for _, dnetw := range data.NetworkSettings.Networks {
				container.Runtime.Networks[dnetw.NetworkID] = Network{
					dnetw.NetworkID,
					dnetw.IPAddress,
				}
			}

			container.Runtime.Id = data.ID
			container.Runtime.State = data.State.Status
		}

		container.Runtime.FirstObserved = false

		return c
	} else {
		container.Runtime.FirstObserved = false
		return nil
	}
}

func (container *Container) Start() bool {
	if c := container.Get(); c != nil && c.State == "exited" {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		err = cli.ContainerStart(ctx, container.Runtime.Id, types.ContainerStartOptions{})

		if err != nil {
			return false
		}

		return true
	} else {
		return false
	}
}
func (container *Container) Stop() bool {
	if c := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		duration := time.Second * 10
		err = cli.ContainerStop(ctx, container.Runtime.Id, &duration)

		if err != nil {
			return false
		}

		return true
	} else {
		return false
	}
}
func (container *Container) Restart() bool {
	if c := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		duration := time.Second * 10
		err = cli.ContainerRestart(ctx, container.Runtime.Id, &duration)

		if err != nil {
			return false
		}

		return true
	} else {
		return false
	}
}
func (container *Container) Delete() bool {
	if c := container.Get(); c != nil && c.State != "running" {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		err = cli.ContainerRemove(ctx, container.Runtime.Id, types.ContainerRemoveOptions{
			Force: true,
		})

		if err != nil {
			return false
		}

		return true
	} else {
		return false
	}
}
func (container *Container) Rename(newName string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	container.Static.GeneratedName = newName
	err = cli.ContainerRename(ctx, container.Runtime.Id, newName)

	if err != nil {
		panic(err)
	}

	return err
}

func (container *Container) Exec(command []string) ExecResult {
	if c := container.Get(); c != nil && c.State == "running" {
		var execResult ExecResult

		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		config := types.ExecConfig{
			AttachStderr: true,
			AttachStdout: true,
			Cmd:          command,
		}

		exec, err := cli.ContainerExecCreate(ctx, container.Static.GeneratedName, config)
		if err != nil {
			panic(err)
		}

		resp, err := cli.ContainerExecAttach(context.Background(), exec.ID, types.ExecStartCheck{})
		if err != nil {
			panic(err)
		}

		var stdoutBuffer, stderrBuffer bytes.Buffer
		outputDone := make(chan error)

		go func() {
			_, err = stdcopy.StdCopy(&stdoutBuffer, &stderrBuffer, resp.Reader)
			outputDone <- err
		}()

		select {
		case err := <-outputDone:
			if err != nil {
				return execResult
			}
			break

		case <-ctx.Done():
			return execResult
		}

		stdout, err := ioutil.ReadAll(&stdoutBuffer)
		if err != nil {
			return execResult
		}
		stderr, err := ioutil.ReadAll(&stderrBuffer)
		if err != nil {
			return execResult
		}

		res, err := cli.ContainerExecInspect(ctx, exec.ID)
		if err != nil {
			return execResult
		}

		execResult.Exit = res.ExitCode
		execResult.Stdout = string(stdout)
		execResult.Stderr = string(stderr)

		return execResult
	} else {
		return ExecResult{}
	}
}

func (container *Container) GenerateLabels() map[string]string {
	now := time.Now()
	return map[string]string{
		"managed":     "smr",
		"group":       container.Static.Group,
		"last-update": string(now.Unix()),
	}
}

func (container *Container) self(containers []types.Container) *types.Container {
	for i, c := range containers {
		for _, name := range c.Names {
			if name == "/"+container.Static.GeneratedName {
				container.Runtime.Id = c.ID
				container.Runtime.State = c.State

				return &containers[i]
			}
		}
	}

	return nil
}
