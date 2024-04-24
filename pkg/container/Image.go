package container

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"smr/pkg/logger"
)

func (container *Container) PullImage(ctx context.Context, cli *client.Client) error {
	if !container.CheckIfImagePresent(ctx, cli) {
		logger.Log.Info(fmt.Sprintf("Pulling the image %s:%s", container.Static.Image, container.Static.Tag))

		reader, err := cli.ImagePull(ctx, container.Static.Image+":"+container.Static.Tag, container.GetDockerAuth())
		if err != nil {
			return err
		}

		type ErrorMessage struct {
			Error string
		}

		var errorMessage error
		buffIOReader := bufio.NewReader(reader)

		for {
			streamBytes, err := buffIOReader.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			json.Unmarshal(streamBytes, &errorMessage)
			if errorMessage != nil {
				return errorMessage
			}
		}

		defer reader.Close()
		logger.Log.Info(fmt.Sprintf("Pulled the image %s:%s", container.Static.Image, container.Static.Tag))

		return nil
	} else {
		logger.Log.Info(fmt.Sprintf("Image %s:%s already present", container.Static.Image, container.Static.Tag))
		return nil
	}
}

func (container *Container) CheckIfImagePresent(ctx context.Context, cli *client.Client) bool {
	images, err := cli.ImageList(ctx, types.ImageListOptions{
		All: true,
	})

	if err != nil {
		logger.Log.Fatal("Failed to list container images")
	}

	searchingFor := fmt.Sprintf("%s:%s", container.Static.Image, container.Static.Tag)

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == searchingFor {
				return true
			}
		}
	}

	return false
}

func (container *Container) GetDockerAuth() types.ImagePullOptions {
	return types.ImagePullOptions{
		RegistryAuth: container.Runtime.Auth,
	}
}
