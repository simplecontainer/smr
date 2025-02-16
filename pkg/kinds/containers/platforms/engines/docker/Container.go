package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	TDTypes "github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	TDVolume "github.com/docker/docker/api/types/volume"
	IDClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	jsoniter "github.com/json-iterator/go"
	"github.com/mholt/archives"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/smaps"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

func New(name string, definition contracts.IDefinition) (*Docker, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	definitionEncoded, err := json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	var definitionCopy v1.ContainersDefinition

	err = json.Unmarshal(definitionEncoded, &definitionCopy)

	if err != nil {
		return nil, err
	}

	var volumes *internal.Volumes
	volumes, err = internal.NewVolumes(name, definition.(*v1.ContainersDefinition).Spec.Volumes)

	if err != nil {
		return nil, err
	}

	if definition.(*v1.ContainersDefinition).Spec.Tag == "" {
		definition.(*v1.ContainersDefinition).Spec.Tag = "latest"
	}

	container := &Docker{
		Name:           definition.(*v1.ContainersDefinition).Meta.Name,
		GeneratedName:  name,
		Labels:         internal.NewLabels(definition.(*v1.ContainersDefinition).Meta.Labels),
		Group:          definition.(*v1.ContainersDefinition).Meta.Group,
		Image:          definition.(*v1.ContainersDefinition).Spec.Image,
		Tag:            definition.(*v1.ContainersDefinition).Spec.Tag,
		Replicas:       definition.(*v1.ContainersDefinition).Spec.Replicas,
		Lock:           sync.RWMutex{},
		Env:            definition.(*v1.ContainersDefinition).Spec.Envs,
		Entrypoint:     definition.(*v1.ContainersDefinition).Spec.Entrypoint,
		Args:           definition.(*v1.ContainersDefinition).Spec.Args,
		Configuration:  smaps.NewFromMap(definition.(*v1.ContainersDefinition).Spec.Configuration),
		NetworkMode:    definition.(*v1.ContainersDefinition).Spec.NetworkMode,
		Networks:       internal.NewNetworks(definition.(*v1.ContainersDefinition).Spec.Networks),
		Ports:          internal.NewPorts(definition.(*v1.ContainersDefinition).Spec.Ports),
		Readiness:      internal.NewReadinesses(definition.(*v1.ContainersDefinition).Spec.Readiness),
		Resources:      internal.NewResources(definition.(*v1.ContainersDefinition).Spec.Resources),
		Volumes:        volumes,
		VolumeInternal: TDVolume.Volume{},
		Capabilities:   definition.(*v1.ContainersDefinition).Spec.Capabilities,
		Privileged:     definition.(*v1.ContainersDefinition).Spec.Privileged,
		Definition:     definitionCopy,
	}

	return container, nil
}

func IsDaemonRunning() error {
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

	return err
}

func (container *Docker) Run() error {
	c, _ := container.Get()

	if c == nil || c.State == "exited" {
		ctx := context.Background()
		cli := &IDClient.Client{}

		var err error

		cli, err = IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		err = container.PullImage(ctx, cli)

		if err != nil {
			return err
		}

		container.Labels.Add("group", container.Group)
		container.Labels.Add("name", container.GeneratedName)
		container.Labels.Add("created", strconv.FormatInt(time.Now().Unix(), 10))

		resp := TDContainer.CreateResponse{}

		resp, err = cli.ContainerCreate(ctx, &TDContainer.Config{
			Hostname:     container.GeneratedName,
			Labels:       container.Labels.ToMap(),
			Image:        container.Image + ":" + container.Tag,
			Env:          container.Env,
			Entrypoint:   container.Entrypoint,
			Cmd:          container.Args,
			Tty:          false,
			ExposedPorts: container.Ports.ToPortExposed(),
		}, &TDContainer.HostConfig{
			DNS:          container.Docker.DNS,
			Mounts:       container.Volumes.ToMounts(),
			PortBindings: container.Ports.ToPortMap(),
			NetworkMode:  TDContainer.NetworkMode(container.NetworkMode),
			Privileged:   container.Privileged,
			CapAdd:       container.Capabilities,
		}, container.BuildNetwork(), nil, container.GeneratedName)

		if err != nil {
			return err
		}

		container.DockerID = resp.ID
		defer container.Volumes.RemoveResources()
		err = container.MountResources()

		if err != nil {
			return err
		}

		if err = cli.ContainerStart(ctx, resp.ID, TDContainer.StartOptions{}); err != nil {
			return err
		}

		err = container.SyncNetwork()

		if err != nil {
			return err
		}

		_, err = container.Get()

		return nil
	} else {
		return errors.New("container is already running")
	}
}
func (container *Docker) PreRun(config *configuration.Configuration, client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	var DNS []string

	if config.Environment.NodeIP != "" {
		DNS = []string{config.Environment.NodeIP, "127.0.0.1"}
	} else {
		DNS = []string{"127.0.0.1"}
	}

	if len(container.Definition.Spec.Dns) != 0 {
		DNS = append(DNS, container.Definition.Spec.Dns...)
	}

	container.Docker.DNS = DNS

	runtime.ObjectDependencies = make([]f.Format, 0)

	err := container.PrepareConfiguration(client, user, runtime)

	if err != nil {
		return err
	}

	err = container.PrepareResources(client, user, runtime)

	if err != nil {
		return err
	}

	err = container.PrepareLabels(runtime)

	if err != nil {
		return err
	}

	err = container.PrepareEnvs(runtime)

	if err != nil {
		return err
	}

	err = container.PrepareReadiness(runtime)

	if err != nil {
		return err
	}

	return err
}
func (container *Docker) PostRun(config *configuration.Configuration, dnsCache *dns.Records) error {
	if container.NetworkMode != "host" {
		return container.UpdateDns(dnsCache)
	}

	return nil
}

func (container *Docker) UpdateDns(dnsCache *dns.Records) error {
	if dnsCache != nil {
		for _, network := range container.Networks.Networks {
			err := dnsCache.Propose(container.GetDomain(network.Reference.Name), network.Docker.IP, dns.AddRecord)

			if err != nil {
				return err
			}
		}

		return nil
	} else {
		return errors.New("dns cache is nil")
	}
}
func (container *Docker) RemoveDns(dnsCache *dns.Records, networkId string) error {
	if dnsCache != nil {
		for _, network := range container.Networks.Networks {
			if network.Docker.NetworkId == networkId {
				err := dnsCache.Propose(container.GetDomain(network.Reference.Name), network.Docker.IP, dns.RemoveRecord)

				if err != nil {
					return err
				}
			}
		}

		return nil
	} else {
		return errors.New("dns cache is nil")
	}
}

func (container *Docker) Start() error {
	if c, _ := container.Get(); c != nil && c.State == "exited" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return err
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
		return nil
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
	if c, _ := container.Get(); c != nil {
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
		return cli.ContainerRename(ctx, c.ID, newName)
	} else {
		return errors.New("container is not found")
	}
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

func (container *Docker) MountResources() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		return err
	}

	var files []archives.FileInfo

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}

	for _, vol := range container.Volumes.Volumes {
		if vol.Type == "resource" {
			files, err = archives.FilesFromDisk(ctx, nil, map[string]string{
				vol.HostPath: vol.MountPoint,
			})

			var buf bytes.Buffer
			err = format.Archive(ctx, &buf, files)

			if err != nil {
				return err
			}

			err = cli.CopyToContainer(ctx, container.DockerID, "/", &buf, TDContainer.CopyToContainerOptions{})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (container *Docker) InitContainer(definition v1.ContainersInternal, config *configuration.Configuration, client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		return errors.New(fmt.Sprintf("init container: %s", err.Error()))
	}

	container.Init, err = New(fmt.Sprintf("%s-init", container.GetGeneratedName()), &v1.ContainersDefinition{
		Kind:   container.GetDefinition().GetKind(),
		Prefix: container.GetDefinition().GetPrefix(),
		Meta:   container.GetDefinition().GetMeta(),
		Spec:   definition,
		State:  nil,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("init container: %s", err.Error()))
	}

	container.Init.Stop("SIGTERM")
	container.Init.Kill("SIGTERM")
	container.Init.Delete()

	err = container.Init.PreRun(config, client, user, runtime)

	if err != nil {
		return errors.New(fmt.Sprintf("init container: %s", err.Error()))
	}

	err = container.Init.Run()

	if err != nil {
		return errors.New(fmt.Sprintf("init container: %s", err.Error()))
	}

	statusCh, errCh := cli.ContainerWait(context.Background(), container.Init.DockerID, TDContainer.WaitConditionNotRunning)

	select {
	case <-statusCh:
		break
	case err = <-errCh:
		return errors.New(fmt.Sprintf("init container: %s", err.Error()))
	}

	var inspect TDTypes.ContainerJSON
	inspect, err = internal.Inspect(container.Init.DockerID)

	if err != nil {
		return errors.New(fmt.Sprintf("init container: %s", err.Error()))
	}

	if inspect.State.ExitCode != 0 {
		return errors.New("init container: init container exited with non-zero")
	}

	return nil
}

func (container *Docker) ToJson() ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(container)
}
