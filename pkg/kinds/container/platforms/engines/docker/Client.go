package docker

import (
	"context"
	TDTypes "github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	TDMount "github.com/docker/docker/api/types/mount"
	IDClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func (container *Docker) RunRaw() (*TDTypes.Container, error) {
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
	DNS = append(DNS, container.Definition.Spec.Container.Dns...)

	resp, err = cli.ContainerCreate(ctx, &TDContainer.Config{
		Hostname:     container.GeneratedName,
		Labels:       container.GenerateLabels(),
		Image:        container.Image + ":" + container.Tag,
		Env:          container.Env,
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

	return container.Get()
}
