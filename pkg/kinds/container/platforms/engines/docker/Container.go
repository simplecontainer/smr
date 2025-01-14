package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	TDTypes "github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	TDMount "github.com/docker/docker/api/types/mount"
	dockerNetwork "github.com/docker/docker/api/types/network"
	IDClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	jsoniter "github.com/json-iterator/go"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/secrets"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

func New(name string, config *configuration.Configuration, definition contracts.IDefinition) (*Docker, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
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
	volumes, err = internal.NewVolumes(definition.(*v1.ContainerDefinition).Spec.Container.Volumes, config)

	if err != nil {
		return nil, err
	}

	if definition.(*v1.ContainerDefinition).Spec.Container.Tag == "" {
		definition.(*v1.ContainerDefinition).Spec.Container.Tag = "latest"
	}

	container := &Docker{
		Name:          definition.(*v1.ContainerDefinition).Meta.Name,
		GeneratedName: name,
		Labels:        smaps.NewFromMap(definition.(*v1.ContainerDefinition).Meta.Labels),
		Group:         definition.(*v1.ContainerDefinition).Meta.Group,
		Image:         definition.(*v1.ContainerDefinition).Spec.Container.Image,
		Tag:           definition.(*v1.ContainerDefinition).Spec.Container.Tag,
		Replicas:      definition.(*v1.ContainerDefinition).Spec.Container.Replicas,
		Lock:          sync.RWMutex{},
		Env:           definition.(*v1.ContainerDefinition).Spec.Container.Envs,
		Entrypoint:    definition.(*v1.ContainerDefinition).Spec.Container.Entrypoint,
		Args:          definition.(*v1.ContainerDefinition).Spec.Container.Args,
		Configuration: smaps.NewFromMap(definition.(*v1.ContainerDefinition).Spec.Container.Configuration),
		NetworkMode:   definition.(*v1.ContainerDefinition).Spec.Container.NetworkMode,
		Networks:      internal.NewNetworks(definition.(*v1.ContainerDefinition).Spec.Container.Networks),
		Ports:         internal.NewPorts(definition.(*v1.ContainerDefinition).Spec.Container.Ports),
		Readiness:     internal.NewReadinesses(definition.(*v1.ContainerDefinition).Spec.Container.Readiness),
		Resources:     internal.NewResources(definition.(*v1.ContainerDefinition).Spec.Container.Resources),
		Volumes:       volumes,
		Capabilities:  definition.(*v1.ContainerDefinition).Spec.Container.Capabilities,
		Privileged:    definition.(*v1.ContainerDefinition).Spec.Container.Privileged,
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
func (container *Docker) Exec(command []string) (types.ExecResult, error) {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		var result types.ExecResult

		ctx := context.Background()
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return types.ExecResult{}, err
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				logger.Log.Error(err.Error())
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
				return result, nil
			}
			break

		case <-ctx.Done():
			return result, nil
		}

		stdout, err := ioutil.ReadAll(&stdoutBuffer)
		if err != nil {
			return result, nil
		}

		stderr, err := ioutil.ReadAll(&stderrBuffer)
		if err != nil {
			return result, nil
		}

		res, err := cli.ContainerExecInspect(ctx, exec.ID)
		if err != nil {
			return result, nil
		}

		result.Exit = res.ExitCode
		result.Stdout = string(stdout)
		result.Stderr = string(stderr)

		return result, nil
	} else {
		return types.ExecResult{}, errors.New("container is not running")
	}
}

func (container *Docker) Logs(follow bool) (io.ReadCloser, error) {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}

		logs, err := cli.ContainerLogs(context.Background(), container.DockerID, TDContainer.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: false,
			Follow:     follow,
			Tail:       "30",
		})

		if err != nil {
			return nil, err
		}

		return logs, nil
	} else {
		return nil, errors.New("container is not running")
	}
}

func (container *Docker) GetContainerState() (string, error) {
	dockerContainer, err := DockerGet(container.GeneratedName)

	if err != nil {
		return "", err
	}

	container.DockerID = dockerContainer.ID
	container.DockerState = dockerContainer.State

	return dockerContainer.State, nil
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

			//container.UpdateDns(dnsCache)
		}

		return container.Get()
	}

	return c, nil
}

func (container *Docker) Prepare(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	runtime.ObjectDependencies = make([]f.Format, 0)

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
	if agentContainerName == "" {
		return errors.New("agent container name is empty")
	}

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
func (container *Docker) UpdateDns(dnsCache *dns.Records, networkId string) {
	networks := container.GetNetworkInfoTS()

	for _, network := range networks.Networks {
		if network.Docker.NetworkId == networkId {
			dnsCache.AddARecord(container.GetDomain(network.Reference.Name), network.Docker.IP)
			dnsCache.AddARecord(container.GetHeadlessDomain(network.Reference.Name), network.Docker.IP)

			return
		}
	}
}

func (container *Docker) RemoveDns(dnsCache *dns.Records, networkId string) {
	networks := container.GetNetworkInfoTS()

	for _, network := range networks.Networks {
		if network.Docker.NetworkId == networkId {
			dnsCache.RemoveARecord(container.GetDomain(network.Reference.Name), network.Docker.IP)
			dnsCache.RemoveARecord(container.GetHeadlessDomain(network.Reference.Name), network.Docker.IP)

			return
		}
	}
}

func (container *Docker) GenerateLabels() map[string]string {
	now := time.Now()

	if container.Labels.Members > 0 {
		container.Labels.Add("managed", "smr")
		container.Labels.Add("group", container.Group)
		container.Labels.Add("name", container.GeneratedName)
	} else {
		container.Labels = smaps.NewFromMap(map[string]string{
			"managed":     "smr",
			"group":       container.Group,
			"name":        container.GeneratedName,
			"last-update": strconv.FormatInt(now.Unix(), 10),
		})
	}

	labels := make(map[string]string)

	container.Labels.Map.Range(func(key, value any) bool {
		labels[key.(string)] = value.(string)
		return true
	})

	return labels
}

func (container *Docker) ToJson() ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(container)
}
