package internal

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	TDContainer "github.com/docker/docker/api/types/container"
	TDNetwork "github.com/docker/docker/api/types/network"
	IDClient "github.com/docker/docker/client"
)

func Inspect(DockerID string) (types.ContainerJSON, error) {
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

func InspectNetwork(NetworkID string) (TDNetwork.Inspect, error) {
	ctx := context.Background()

	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		return TDNetwork.Inspect{}, err
	}

	return cli.NetworkInspect(ctx, NetworkID, types.NetworkInspectOptions{
		Scope:   "",
		Verbose: false,
	})
}

func Get(name string) (types.Container, error) {
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
		if container.ID == name {
			return containers[i], nil
		} else {
			for _, n := range container.Names {
				if n == "/"+name {
					return containers[i], nil
				}
			}
		}
	}

	return types.Container{}, errors.New("container not found")
}

func GetNetwork(name string) (TDNetwork.Summary, error) {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	defer cli.Close()

	var networks []TDNetwork.Summary
	networks, err = cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, c := range networks {
		if c.Name == name {
			return c, nil
		}
	}

	return TDNetwork.Summary{}, errors.New("network not found")
}
