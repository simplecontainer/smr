package docker

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	IDClient "github.com/docker/docker/client"
)

func DockerInspect(DockerID string) (types.ContainerJSON, error) {
	ctx := context.Background()
	cli := &IDClient.Client{}

	var err error

	cli, err = IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
	if err != nil {
		return types.ContainerJSON{}, err
	}

	defer func(cli *IDClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	data, err := cli.ContainerInspect(ctx, DockerID)

	if err != nil {
		return types.ContainerJSON{}, err
	}

	return data, nil
}

func DockerGet(containerName string) (types.Container, error) {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer func(cli *IDClient.Client) {
		err := cli.Close()
		if err != nil {

		}
	}(cli)

	containers, err := cli.ContainerList(ctx, TDContainer.ListOptions{
		All: true,
	})

	if err != nil {
		return types.Container{}, err
	}

	for i, container := range containers {
		// Check if name or id of docker container passed and act accordingly
		if container.ID == containerName {
			return containers[i], nil
		} else {
			for _, name := range container.Names {
				if name == "/"+containerName {
					return containers[i], nil
				}
			}
		}
	}

	return types.Container{}, errors.New("container not found")
}
