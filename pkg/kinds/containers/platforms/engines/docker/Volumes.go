package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/volume"
	IDClient "github.com/docker/docker/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

func (container *Docker) CreateVolume(definition *v1.VolumeDefinition) error {
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

	_, err = cli.VolumeCreate(ctx, volume.CreateOptions{
		Driver:     definition.Spec.Driver,
		DriverOpts: definition.Spec.DriverOpts,
		Labels:     definition.GetMeta().Labels,
		Name:       fmt.Sprintf("%s-%s", definition.GetMeta().Group, definition.GetMeta().Name),
	})

	return err
}

func (container *Docker) DeleteVolume(id string, force bool) error {
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

	return cli.VolumeRemove(ctx, id, force)
}
