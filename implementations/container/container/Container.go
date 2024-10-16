package container

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	dockerNetwork "github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/simplecontainer/smr/implementations/container/container/internal"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/static"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func NewContainerFromDefinition(config *configuration.Configuration, name string, definition v1.ContainerDefinition) (*Container, error) {
	// Make deep copy of the definition, so we can preserve it for later usage
	definitionEncoded, err := json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	var definitionCopy v1.ContainerDefinition

	err = json.Unmarshal(definitionEncoded, &definitionCopy)

	if err != nil {
		return nil, err
	}

	var volumes *internal.Volumes
	volumes, err = internal.NewVolumes(definition.Spec.Container.Volumes, config)

	if err != nil {
		return nil, err
	}

	if definition.Spec.Container.Tag == "" {
		definition.Spec.Container.Tag = "latest"
	}

	container := &Container{
		Static: Static{
			Name:          definition.Meta.Name,
			GeneratedName: name,
			Labels:        definition.Meta.Labels,
			Group:         definition.Meta.Group,
			Image:         definition.Spec.Container.Image,
			Tag:           definition.Spec.Container.Tag,
			Replicas:      definition.Spec.Container.Replicas,
			Env:           definition.Spec.Container.Envs,
			Entrypoint:    definition.Spec.Container.Entrypoint,
			Args:          definition.Spec.Container.Args,
			NetworkMode:   definition.Spec.Container.NetworkMode,
			Networks:      internal.NewNetworks(definition.Spec.Container.Networks),
			Ports:         internal.NewPorts(definition.Spec.Container.Ports),
			Volumes:       volumes,
			Readiness:     internal.NewReadinesses(definition.Spec.Container.Readiness),
			Resources:     internal.NewResources(definition.Spec.Container.Resources),
			Capabilities:  definition.Spec.Container.Capabilities,
			Privileged:    definition.Spec.Container.Privileged,
			Definition:    definitionCopy,
		},
		Runtime: Runtime{
			Auth:          GetAuth(definition.Spec.Container.Image, config.Environment),
			Id:            "",
			Networks:      nil,
			State:         "",
			FoundRunning:  false,
			Configuration: definition.Spec.Container.Configuration,
			Owner:         Owner{},
		},
		Status: status.Status{
			State:      &status.StatusState{},
			LastUpdate: time.Now(),
		},
	}

	container.Runtime.Networks = container.Static.Networks
	container.Status.CreateGraph()

	if container.Runtime.Configuration == nil {
		container.Runtime.Configuration = make(map[string]string)
	}

	return container, nil
}

func Existing(name string) *Container {
	container := &Container{
		Static: Static{
			Name:                   name,
			GeneratedName:          name,
			GeneratedNameNoProject: name,
			Image:                  "image",
			Tag:                    "tag",
			Networks:               &internal.Networks{},
		},
		Runtime: Runtime{
			Id:            "",
			State:         "",
			FoundRunning:  false,
			Networks:      nil,
			Ready:         false,
			Configuration: make(map[string]string),
		},
		Status: status.Status{},
	}

	container.Runtime.Networks = internal.NewNetworks([]v1.ContainerNetwork{})
	container.Static.Name = name

	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	defer func(cli *dockerClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		panic(err)
	}

	if c := container.self(containers); c != nil {
		data, _ := cli.ContainerInspect(ctx, container.Runtime.Id)

		if c.State == "running" {
			for _, netw := range container.Static.Networks.Networks {
				if data.NetworkSettings.Networks[netw.Reference.Name] != nil {
					var netwInspect types.NetworkResource
					netwInspect, err = cli.NetworkInspect(ctx, data.NetworkSettings.Networks[netw.Reference.Name].NetworkID, types.NetworkInspectOptions{
						Scope:   "",
						Verbose: false,
					})

					NetworkId := data.NetworkSettings.Networks[netw.Reference.Name].NetworkID
					IpAddress := data.NetworkSettings.Networks[netw.Reference.Name].IPAddress
					NetwrName := ""

					if err == nil {
						NetworkId = netwInspect.ID
						NetwrName = netwInspect.Name
					}

					container.AddNetworkInfoTS(NetworkId, IpAddress, NetwrName)
				}
			}

			container.Runtime.Id = data.ID
			container.Runtime.State = data.State.Status
		}

		return container
	} else {
		return nil
	}
}

func (container *Container) SetOwner(owner string) {
	if owner != "" {
		splitted := strings.SplitN(owner, ".", 2)

		if len(splitted) == 2 {
			container.Runtime.Owner.Kind = splitted[0]
			container.Runtime.Owner.GroupIdentifier = splitted[1]
		}
	}
}

func (container *Container) Run(environment *configuration.Environment, client *client.Http, dnsCache *dns.Records, user *authentication.User) (*types.Container, error) {
	c, _ := container.Get()

	if c == nil {
		return container.run(c, environment, client, dnsCache, user)
	}

	return c, nil
}

func (container *Container) run(c *types.Container, environment *configuration.Environment, client *client.Http, dnsCache *dns.Records, user *authentication.User) (*types.Container, error) {
	ctx := context.Background()
	cli := &dockerClient.Client{}

	var err error

	cli, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	defer func(cli *dockerClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	err = container.PullImage(ctx, cli)

	if err != nil {
		return nil, err
	}

	resp := dockerContainer.ContainerCreateCreatedBody{}

	var unpackedEnvs []string
	unpackedEnvs, err = UnpackSecretsEnvs(client, user, container.Static.Env)

	if err != nil {
		return nil, err
	}

	var exposedPorts nat.PortSet
	exposedPorts, err = container.Static.Ports.ToPortExposed()

	if err != nil {
		return nil, err
	}

	var portBindings nat.PortMap
	portBindings, err = container.Static.Ports.ToPortMap()

	if err != nil {
		return nil, err
	}

	var mounts []mount.Mount
	mounts, err = container.Static.Volumes.ToMounts()

	if err != nil {
		return nil, err
	}

	resp, err = cli.ContainerCreate(ctx, &dockerContainer.Config{
		Hostname:     container.Static.GeneratedName,
		Labels:       container.GenerateLabels(),
		Image:        container.Static.Image + ":" + container.Static.Tag,
		Env:          unpackedEnvs,
		Entrypoint:   container.Static.Entrypoint,
		Cmd:          container.Static.Args,
		Tty:          false,
		ExposedPorts: exposedPorts,
	}, &dockerContainer.HostConfig{
		DNS: []string{
			environment.AGENTIP,
		},
		Mounts:       mounts,
		PortBindings: portBindings,
		NetworkMode:  dockerContainer.NetworkMode(container.Static.NetworkMode),
		Privileged:   container.Static.Privileged,
		CapAdd:       container.Static.Capabilities,
	}, container.GetNetwork(), nil, container.Static.GeneratedName)

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
		var netwInspect types.NetworkResource

		netwInspect, err = cli.NetworkInspect(ctx, dnetw.NetworkID, types.NetworkInspectOptions{
			Scope:   "",
			Verbose: false,
		})

		NetworkId := dnetw.NetworkID
		IpAddress := dnetw.IPAddress
		NetwrName := ""

		if err == nil {
			NetworkId = netwInspect.ID
			NetwrName = netwInspect.Name
		}

		container.AddNetworkInfoTS(NetworkId, IpAddress, NetwrName)
	}

	if container.Static.NetworkMode != "host" {
		err = container.SolveAgentNetworking()

		if err != nil {
			return nil, err
		}

		container.UpdateDns(dnsCache)

		if err != nil {
			return nil, err
		}
	}

	return container.Get()
}

func (container *Container) SolveAgentNetworking() error {
	var dockerContainer *types.Container
	var err error

	dockerContainer, err = container.Get()

	if err != nil {
		return err
	}

	ctx := context.Background()
	cli := &dockerClient.Client{}

	cli, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())

	if err != nil {
		return err
	}

	agent := Existing("smr-agent")

	var agentContainer *types.Container
	agentContainer, err = agent.Get()

	networks := container.GetNetworkInfoTS()

	for _, network := range agentContainer.NetworkSettings.Networks {
		EndpointSettings := &dockerNetwork.EndpointSettings{
			NetworkID: network.NetworkID,
		}

		if networks.Find(network.NetworkID) != nil {
			continue
		}

		err = cli.NetworkConnect(ctx, network.NetworkID, dockerContainer.ID, EndpointSettings)

		if err != nil {
			return err
		}
	}

	if agentContainer != nil {
		networks = container.GetNetworkInfoTS()

		for _, network := range networks.Networks {
			err = network.FindNetworkAlias(static.SMR_ENDPOINT_NAME)

			if err == nil {
				err = network.FindNetworkAlias(container.Static.GeneratedName)

				if err == nil {
					err = network.Connect(agent.Runtime.Id)

					if err != nil && err.Error() != fmt.Sprintf("Error response from daemon: endpoint with name %s already exists in network %s", static.SMR_ENDPOINT_NAME, network.Reference.Name) {
						return err
					}
				}
			}
		}
	} else {
		return errors.New("agent is not running")
	}

	return nil
}

func (container *Container) UpdateDns(dnsCache *dns.Records) {
	networks := container.GetNetworkInfoTS()

	for _, network := range networks.Networks {
		dnsCache.AddARecord(container.GetDomain(network.Reference.Name), network.Docker.IP)
		dnsCache.AddARecord(container.GetHeadlessDomain(network.Reference.Name), network.Docker.IP)
	}
}

func (container *Container) Get() (*types.Container, error) {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	if c := container.self(containers); c != nil {
		data, _ := cli.ContainerInspect(ctx, container.Runtime.Id)

		if c.State == "running" {
			for _, dnetw := range data.NetworkSettings.Networks {
				netwInspect, err := cli.NetworkInspect(ctx, dnetw.NetworkID, types.NetworkInspectOptions{
					Scope:   "",
					Verbose: false,
				})

				NetworkId := dnetw.NetworkID
				IpAddress := dnetw.IPAddress
				NetwrName := ""

				if err == nil {
					NetworkId = netwInspect.ID
					NetwrName = netwInspect.Name
				}

				container.AddNetworkInfoTS(NetworkId, IpAddress, NetwrName)
			}

			container.Runtime.Id = data.ID
			container.Runtime.State = data.State.Status
		}

		return c, nil
	} else {
		return nil, errors.New("container not found")
	}
}

func (container *Container) GetFromId(runtimeId string) *types.Container {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
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

	if c := container.selfId(containers, runtimeId); c != nil {
		return c
	} else {
		return nil
	}
}

func (container *Container) Start() bool {
	if c, _ := container.Get(); c != nil && c.State == "exited" {
		ctx := context.Background()
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
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
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		duration := time.Second * 10
		err = cli.ContainerStop(ctx, container.Runtime.Id, &duration)

		if err != nil {
			return false
		}

		container.Runtime.State = "stoppped"

		return true
	} else {
		return false
	}
}

func (container *Container) Restart() bool {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
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

func (container *Container) Delete() error {
	if c, _ := container.Get(); c != nil && c.State != "running" {
		ctx := context.Background()
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		err = cli.ContainerRemove(ctx, container.Runtime.Id, types.ContainerRemoveOptions{
			Force: true,
		})

		if err != nil {
			return err
		}

		container.Runtime.State = "deleted"

		return nil
	} else {
		return errors.New("cannot delete container that is running")
	}
}

func (container *Container) Rename(newName string) error {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
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
	if c, _ := container.Get(); c != nil && c.State == "running" {
		var execResult ExecResult

		ctx := context.Background()
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
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
		case err = <-outputDone:
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

	if len(container.Static.Labels) > 0 {
		container.Static.Labels["managed"] = "smr"
		container.Static.Labels["group"] = container.Static.Group
		container.Static.Labels["name"] = container.Static.GeneratedName
		container.Static.Labels["last-update"] = strconv.FormatInt(now.Unix(), 10)
	} else {
		tmp := map[string]string{
			"managed":     "smr",
			"group":       container.Static.Group,
			"name":        container.Static.GeneratedName,
			"last-update": strconv.FormatInt(now.Unix(), 10),
		}

		container.Static.Labels = tmp
	}

	return container.Static.Labels
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

func (container *Container) selfId(containers []types.Container, runtimeId string) *types.Container {
	for i, c := range containers {
		if c.ID == runtimeId {
			return &containers[i]
		}
	}

	return nil
}
