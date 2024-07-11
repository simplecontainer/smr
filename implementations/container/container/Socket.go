package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func IsDockerRunning() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	_, err = cli.ContainerList(ctx, types.ContainerListOptions{})

	if err != nil {
		panic(err)
	}
}
