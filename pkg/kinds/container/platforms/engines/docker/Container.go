package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	TDTypes "github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	TDMount "github.com/docker/docker/api/types/mount"
	dockerNetwork "github.com/docker/docker/api/types/network"
	IDClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/secrets"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"io/ioutil"
	"strconv"
	"time"
)

func New(name string, config *configuration.Configuration, definition *v1.ContainerDefinition) (*Docker, error) {
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

	container := &Docker{
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
		Configuration: definition.Spec.Container.Configuration,
		NetworkMode:   definition.Spec.Container.NetworkMode,
		Networks:      internal.NewNetworks(definition.Spec.Container.Networks),
		Ports:         internal.NewPorts(definition.Spec.Container.Ports),
		Readiness:     internal.NewReadinesses(definition.Spec.Container.Readiness),
		Resources:     internal.NewResources(definition.Spec.Container.Resources),
		Volumes:       volumes,
		Capabilities:  definition.Spec.Container.Capabilities,
		Privileged:    definition.Spec.Container.Privileged,
		Definition:    definitionCopy,
	}

	return container, nil
}

func IsDaemonRunning() {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	defer func(cli *IDClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	_, err = cli.ContainerList(ctx, TDContainer.ListOptions{})

	if err != nil {
		panic(err)
	}
}

func (container *Docker) Start() error {
	if c, _ := container.Get(); c != nil && c.State == "exited" {
		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		return cli.ContainerStart(ctx, container.DockerID, TDContainer.StartOptions{})
	} else {
		return errors.New("container is nil or already started")
	}
}
func (container *Docker) Stop(signal string) error {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		duration := 10
		return cli.ContainerStop(ctx, c.ID, TDContainer.StopOptions{
			Signal:  signal,
			Timeout: &duration,
		})
	} else {
		return errors.New("container is nil when trying to stop it")
	}
}
func (container *Docker) Kill(signal string) error {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		return cli.ContainerKill(ctx, c.ID, signal)
	} else {
		return errors.New("container is nil or not running when trying to kill it")
	}
}
func (container *Docker) Restart() error {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		duration := 10
		return cli.ContainerRestart(ctx, container.DockerID, TDContainer.StopOptions{
			Signal:  "SIGTERM",
			Timeout: &duration,
		})
	} else {
		return errors.New("container is nil or not running when trying to restart it")
	}
}
func (container *Docker) Delete() error {
	if c, _ := container.Get(); c != nil && c.State != "running" {
		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		err = cli.ContainerRemove(ctx, container.DockerID, TDContainer.RemoveOptions{
			Force: true,
		})

		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.New("cannot delete container that is running")
	}
}
func (container *Docker) Rename(newName string) error {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	defer func(cli *IDClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	container.GeneratedName = newName
	err = cli.ContainerRename(ctx, container.DockerID, newName)

	if err != nil {
		panic(err)
	}

	return err
}
func (container *Docker) Exec(command []string) types.ExecResult {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		var execResult types.ExecResult

		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		config := TDTypes.ExecConfig{
			AttachStderr: true,
			AttachStdout: true,
			Cmd:          command,
		}

		exec, err := cli.ContainerExecCreate(ctx, container.GeneratedName, config)
		if err != nil {
			panic(err)
		}

		resp, err := cli.ContainerExecAttach(context.Background(), exec.ID, TDTypes.ExecStartCheck{})
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
		return types.ExecResult{}
	}
}

func (container *Docker) Get() (*TDTypes.Container, error) {
	dockerContainer, err := DockerGet(container.GeneratedName)

	if err != nil {
		return nil, err
	}

	container.DockerID = dockerContainer.ID
	container.DockerState = dockerContainer.State

	if dockerContainer.State == "running" {
		err = container.SyncNetworkInformation()

		if err != nil {
			return nil, err
		}
	}

	return &dockerContainer, nil
}
func (container *Docker) Run(config *configuration.Configuration, client *client.Http, dnsCache *dns.Records, user *authentication.User) (*TDTypes.Container, error) {
	c, _ := container.Get()

	if c == nil || c.State == "exited" {
		ctx := context.Background()
		cli := &IDClient.Client{}

		var err error

		cli, err = IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		err = container.PullImage(ctx, cli)

		if err != nil {
			return nil, err
		}

		resp := TDContainer.CreateResponse{}

		var unpackedEnvs []string
		unpackedEnvs, err = secrets.UnpackSecretsEnvs(client, user, container.Env)

		if err != nil {
			return nil, err
		}

		var exposedPorts nat.PortSet
		exposedPorts, err = container.Ports.ToPortExposed()

		if err != nil {
			return nil, err
		}

		var portBindings nat.PortMap
		portBindings, err = container.Ports.ToPortMap()

		if err != nil {
			return nil, err
		}

		var mounts []TDMount.Mount
		mounts, err = container.Volumes.ToMounts()

		if err != nil {
			return nil, err
		}

		DNS := []string{}

		if len(container.Definition.Spec.Container.Dns) == 0 {
			DNS = append(DNS, []string{config.Environment.AGENTIP, "127.0.0.1"}...)
		} else {
			DNS = append(DNS, container.Definition.Spec.Container.Dns...)
		}

		resp, err = cli.ContainerCreate(ctx, &TDContainer.Config{
			Hostname:     container.GeneratedName,
			Labels:       container.GenerateLabels(),
			Image:        container.Image + ":" + container.Tag,
			Env:          unpackedEnvs,
			Entrypoint:   container.Entrypoint,
			Cmd:          container.Args,
			Tty:          false,
			ExposedPorts: exposedPorts,
		}, &TDContainer.HostConfig{
			DNS:          DNS,
			Mounts:       mounts,
			PortBindings: portBindings,
			NetworkMode:  TDContainer.NetworkMode(container.NetworkMode),
			Privileged:   container.Privileged,
			CapAdd:       container.Capabilities,
		}, container.BuildNetwork(), nil, container.GeneratedName)

		if err != nil {
			return nil, err
		}

		container.DockerID = resp.ID

		if err = cli.ContainerStart(ctx, resp.ID, TDContainer.StartOptions{}); err != nil {
			return nil, err
		}

		err = container.SyncNetworkInformation()

		if err != nil {
			return nil, err
		}

		if container.NetworkMode != "host" {
			err = container.AttachToNetworks(config.Node)

			if err != nil {
				return nil, err
			}

			container.UpdateDns(dnsCache)
		}

		return container.Get()
	}

	return c, nil
}

func (container *Docker) Prepare(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	err := container.PrepareNetwork(client, user, runtime)

	if err != nil {
		return err
	}

	err = container.PrepareConfiguration(client, user, runtime)

	if err != nil {
		return err
	}

	err = container.PrepareResources(client, user, runtime)

	if err != nil {
		return err
	}

	container.PrepareLabels(runtime)
	container.PrepareEnvs(runtime)
	container.PrepareReadiness(runtime)

	return nil
}

func (container *Docker) AttachToNetworks(agentContainerName string) error {
	var dockerContainer *TDTypes.Container
	var err error

	dockerContainer, err = container.Get()

	if err != nil {
		return err
	}

	ctx := context.Background()
	cli := &IDClient.Client{}

	cli, err = IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		return err
	}

	logger.Log.Debug("trying to find agent container", zap.String("agent", agentContainerName))

	var agent TDTypes.Container
	agent, err = DockerGet(agentContainerName)

	if err != nil {
		return errors.New("failed to find agent container")
	}

	networks := container.GetNetworkInfoTS()

	for _, network := range agent.NetworkSettings.Networks {
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

	networks = container.GetNetworkInfoTS()

	for _, network := range networks.Networks {
		err = network.FindNetworkAlias(static.SMR_ENDPOINT_NAME)

		if err == nil {
			err = network.FindNetworkAlias(container.GeneratedName)

			if err == nil {
				err = network.Connect(agent.ID)

				if err != nil && err.Error() != fmt.Sprintf("Error response from daemon: endpoint with name %s already exists in network %s", static.SMR_ENDPOINT_NAME, network.Reference.Name) {
					return err
				}
			}
		}
	}

	return nil
}
func (container *Docker) UpdateDns(dnsCache *dns.Records) {
	/*
		for _, n := range containerObj.Networks.Networks {
			for _, ip := range shared.DnsCache.FindDeleteQueue(containerObj.GetDomain(n.Reference.Name)) {
				shared.DnsCache.RemoveARecord(containerObj.GetDomain(n.Reference.Name), ip)
				shared.DnsCache.RemoveARecord(containerObj.GetHeadlessDomain(n.Reference.Name), ip)

				obj := objects.New(shared.Client.Get("root"), &authentication.User{
					Username: "root",
					Domain:   "localhost",
				})

				obj.Remove(f.NewFromString(fmt.Sprintf("network.%s.%s.dns", containerObj.Static.Group, containerObj.Static.GeneratedName)))
			}

			shared.DnsCache.ResetDeleteQueue(containerObj.GetDomain(n.Reference.Name))
		}
	*/

	networks := container.GetNetworkInfoTS()

	for _, network := range networks.Networks {
		dnsCache.AddARecord(container.GetDomain(network.Reference.Name), network.Docker.IP)
		dnsCache.AddARecord(container.GetHeadlessDomain(network.Reference.Name), network.Docker.IP)
	}
}
func (container *Docker) GenerateLabels() map[string]string {
	now := time.Now()

	if len(container.Labels) > 0 {
		container.Labels["managed"] = "smr"
		container.Labels["group"] = container.Group
		container.Labels["name"] = container.GeneratedName
		container.Labels["last-update"] = strconv.FormatInt(now.Unix(), 10)
	} else {
		labels := map[string]string{
			"managed":     "smr",
			"group":       container.Group,
			"name":        container.GeneratedName,
			"last-update": strconv.FormatInt(now.Unix(), 10),
		}

		container.Labels = labels
	}

	return container.Labels
}
