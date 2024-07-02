package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"os"
)

func (container *Container) CopyFromContainer(pathContainer string, pathHost string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	reader, _, err := cli.CopyFromContainer(ctx, container.Runtime.Id, pathContainer)

	if err != nil {
		panic(err)
	}

	container.Exports.path = pathHost

	writer, err := os.Create(container.Exports.path)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	defer reader.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		panic(err)
	}

	return err
}
func (container *Container) CopyToContainer(reader io.Reader, pathContainer string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	err = cli.CopyToContainer(ctx, container.Runtime.Id, pathContainer, reader, types.CopyToContainerOptions{})

	if err != nil {
		panic(err)
	}

	return err
}
