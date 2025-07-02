package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	TDTypes "github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	TDVolume "github.com/docker/docker/api/types/volume"
	IDClient "github.com/docker/docker/client"
	jsoniter "github.com/json-iterator/go"
	"github.com/mholt/archives"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

func New(name string, definition idefinitions.IDefinition) (*Docker, error) {
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

	var readinesses *internal.Readinesses
	readinesses, err = internal.NewReadinesses(definition.(*v1.ContainersDefinition).Spec.Readiness)

	if err != nil {
		return nil, err
	}

	container := &Docker{
		Name:           definition.(*v1.ContainersDefinition).Meta.Name,
		GeneratedName:  name,
		Labels:         internal.NewLabels(definition.(*v1.ContainersDefinition).Meta.Labels),
		Group:          definition.(*v1.ContainersDefinition).Meta.Group,
		Image:          definition.(*v1.ContainersDefinition).Spec.Image,
		Tag:            definition.(*v1.ContainersDefinition).Spec.Tag,
		RegistryAuth:   definition.(*v1.ContainersDefinition).Spec.RegistryAuth,
		Replicas:       definition.(*v1.ContainersDefinition).Spec.Replicas,
		Lock:           sync.RWMutex{},
		Env:            definition.(*v1.ContainersDefinition).Spec.Envs,
		Entrypoint:     definition.(*v1.ContainersDefinition).Spec.Entrypoint,
		Args:           definition.(*v1.ContainersDefinition).Spec.Args,
		Configuration:  smaps.NewFromMap(definition.(*v1.ContainersDefinition).Spec.Configuration),
		NetworkMode:    definition.(*v1.ContainersDefinition).Spec.NetworkMode,
		Networks:       internal.NewNetworks(definition.(*v1.ContainersDefinition).Spec.Networks),
		Ports:          internal.NewPorts(definition.(*v1.ContainersDefinition).Spec.Ports),
		Readiness:      readinesses,
		Resources:      internal.NewResources(definition.(*v1.ContainersDefinition).Spec.Resources),
		Configurations: internal.NewConfigurations(definition.(*v1.ContainersDefinition).Spec.Configurations),
		Volumes:        volumes,
		VolumeInternal: TDVolume.Volume{},
		Capabilities:   definition.(*v1.ContainersDefinition).Spec.Capabilities,
		User:           definition.(*v1.ContainersDefinition).Spec.User,
		GroupAdd:       definition.(*v1.ContainersDefinition).Spec.GroupAdd,
		Privileged:     definition.(*v1.ContainersDefinition).Spec.Privileged,
		definition:     definitionCopy, // Holds for local reference inside engine itself
	}

	return container, nil
}

func IsDaemonRunning() (string, error) {
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

	version, err := cli.ServerVersion(ctx)

	return version.Version, err
}

func (container *Docker) Run() error {
	c, _ := container.Get()

	if c == nil {
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

		container.Labels.Add("managed", "smr")
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
			User:         container.User,
			ExposedPorts: container.Ports.ToPortExposed(),
		}, &TDContainer.HostConfig{
			DNS:          container.Docker.DNS,
			Mounts:       container.Volumes.ToMounts(),
			PortBindings: container.Ports.ToPortMap(),
			GroupAdd:     container.GroupAdd,
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
func (container *Docker) PreRun(config *configuration.Configuration, client *clients.Http, user *authentication.User, runtime *types.Runtime) error {
	var DNS []string

	if config.Environment.Container.NodeIP != "" {
		DNS = []string{config.Environment.Container.NodeIP, "127.0.0.1"}
	} else {
		DNS = []string{"127.0.0.1"}
	}

	if len(container.definition.Spec.Dns) != 0 {
		DNS = append(DNS, container.definition.Spec.Dns...)
	}

	container.Docker.DNS = DNS

	runtime.ObjectDependencies = make([]f.Format, 0)

	err := container.PrepareConfiguration(config, client, user, runtime)

	if err != nil {
		return err
	}

	err = container.PrepareConfigurations(client, user, runtime)

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

	err = container.PrepareAuth(runtime)

	if err != nil {
		return err
	}

	for _, depends := range container.definition.Spec.Dependencies {
		runtime.ObjectDependencies = append(runtime.ObjectDependencies, f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_CONTAINERS, depends.Group, depends.Name))
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
		err = cli.ContainerStop(ctx, c.ID, TDContainer.StopOptions{
			Signal:  signal,
			Timeout: &duration,
		})

		if err != nil {
			if IDClient.IsErrNotFound(err) {
				return nil
			}
		}

		return err
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

	// Resource cleanup after delete is not that fast - hence rename before delete
	container.Rename(fmt.Sprintf("%s-delete-%d", container.GetGeneratedName(), time.Now().UnixMicro()))

	err = cli.ContainerRemove(ctx, container.DockerID, TDContainer.RemoveOptions{
		Force: true,
	})

	if err != nil {
		if IDClient.IsErrNotFound(err) {
			return nil
		}

		return err
	}

	err = container.Wait(string(TDContainer.WaitConditionRemoved))

	if err != nil {
		if IDClient.IsErrNotFound(err) {
			return nil
		}
	}

	return err
}
func (container *Docker) Wait(condition string) error {
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

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	WaitCondition := TDContainer.WaitCondition(condition)

	statusCh, errCh := cli.ContainerWait(ctx, container.DockerID, WaitCondition)
	select {
	case <-ctxTimeout.Done():
		return errors.New("timeout waiting for the condition")
	case err = <-errCh:
		return err
	case <-statusCh:
		return nil
	}
}
func (container *Docker) Clean() error {
	if c, _ := container.Get(); c != nil {
		state, err := container.GetState()

		if err != nil || state.State == "exited" || state.State == "created" {
			if !errors.Is(errors.New("container not found"), err) {
				return container.Delete()
			} else {
				return nil
			}
		}

		if err = container.Stop(static.SIGTERM); err != nil {
			return err
		}

		return container.Delete()
	} else {
		return nil
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

		err = cli.ContainerRename(ctx, c.ID, newName)

		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.New("container is not found")
	}
}

func (container *Docker) Exec(ctx context.Context, command []string, interactive bool) (string, *bufio.Reader, net.Conn, error) {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

		if err != nil {
			return "", nil, nil, err
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}(cli)

		config := TDContainer.ExecOptions{
			AttachStderr: true,
			AttachStdout: true,
			AttachStdin:  interactive,
			Tty:          interactive,
			Cmd:          command,
		}

		exec, err := cli.ContainerExecCreate(ctx, container.GeneratedName, config)
		if err != nil {
			return "", nil, nil, err
		}

		resp, err := cli.ContainerExecAttach(ctx, exec.ID, TDTypes.ExecStartCheck{})
		if err != nil {
			return "", nil, nil, err
		}

		return exec.ID, resp.Reader, resp.Conn, nil
	} else {
		return "", nil, nil, errors.New("container is not running")
	}
}
func (container *Docker) ExecInspect(ID string) (bool, int, error) {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		return false, 1, err
	}

	res, err := cli.ContainerExecInspect(ctx, ID)

	if err != nil {
		return false, 1, err
	}

	return res.Running, res.ExitCode, nil
}

func (container *Docker) Logs(ctx context.Context, follow bool) (io.ReadCloser, error) {
	if c, _ := container.Get(); c != nil && (c.State == "running" || c.State == "exited") {
		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}

		logs, err := cli.ContainerLogs(ctx, container.DockerID, TDContainer.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: false,
			Follow:     follow,
		})

		if err != nil {
			return nil, err
		}

		return logs, nil
	} else {
		return nil, errors.New("container is not found")
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

func (container *Docker) InitContainer(definition *v1.ContainersInternal, config *configuration.Configuration, client *clients.Http, user *authentication.User, runtime *types.Runtime) error {
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

func (container *Docker) Usage() (*TDContainer.StatsResponse, error) {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}

		defer func(cli *IDClient.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		statsReader, err := cli.ContainerStats(ctx, container.DockerID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get container stats: %w", err)
		}
		defer statsReader.Body.Close()

		statsData, err := io.ReadAll(statsReader.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read stats data: %w", err)
		}

		var stats *TDContainer.StatsResponse
		if err := json.Unmarshal(statsData, stats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stats JSON: %w", err)
		}

		return stats, nil
	} else {
		return nil, errors.New("container is not running")
	}
}

func (container *Docker) ToJSON() ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(container)
}
