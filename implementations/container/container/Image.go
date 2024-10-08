package container

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/logger"
	"io"
	"strings"
)

func (container *Container) PullImage(ctx context.Context, cli *client.Client) error {
	if !container.CheckIfImagePresent(ctx, cli) {
		logger.Log.Info(fmt.Sprintf("Pulling the image %s:%s", container.Static.Image, container.Static.Tag))
		container.Status.PulledImage = status.PULLING_IMAGE

		reader, err := cli.ImagePull(ctx, container.Static.Image+":"+container.Static.Tag, container.GetDockerAuth())
		if err != nil {
			return err
		}

		type ErrorMessage struct {
			Error string
		}

		var errorMessage error
		buffIOReader := bufio.NewReader(reader)

		var streamBytes []byte

		for {
			streamBytes, err = buffIOReader.ReadBytes('\n')
			if err == io.EOF {
				break
			}

			if errorMessage != nil {
				container.Status.PulledImage = status.PULLED_FAILED
				return errorMessage
			}

			err = json.Unmarshal(streamBytes, &errorMessage)

			if err != nil {
				container.Status.PulledImage = status.PULLED_FAILED
				return err
			}
		}

		defer func(reader io.ReadCloser) {
			err = reader.Close()
			if err != nil {
				return
			}
		}(reader)

		logger.Log.Info(fmt.Sprintf("pulled the image %s:%s", container.Static.Image, container.Static.Tag))
		container.Status.PulledImage = status.PULLING_IMAGE

		return nil
	} else {
		logger.Log.Info(fmt.Sprintf("image %s:%s already present", container.Static.Image, container.Static.Tag))
		return nil
	}
}

func (container *Container) CheckIfImagePresent(ctx context.Context, cli *client.Client) bool {
	images, err := cli.ImageList(ctx, types.ImageListOptions{
		All: true,
	})

	if err != nil {
		logger.Log.Fatal("failed to list container images")
	}

	searchingFor := fmt.Sprintf("%s:%s", container.Static.Image, container.Static.Tag)

	for _, image := range images {
		for _, tag := range image.RepoTags {
			registryTo, imageTo := splitReposSearchTerm(tag)
			registryFrom, imageFrom := splitReposSearchTerm(searchingFor)

			if registryTo == registryFrom && imageTo == imageFrom {
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

func splitReposSearchTerm(reposName string) (string, string) {
	nameParts := strings.SplitN(reposName, "/", 2)
	var indexName, remoteName string
	if len(nameParts) == 1 || (!strings.Contains(nameParts[0], ".") &&
		!strings.Contains(nameParts[0], ":") && nameParts[0] != "localhost") {
		// This is a Docker Index repos (ex: samalba/hipache or ubuntu)
		// 'docker.io'
		indexName = "docker.io"
		remoteName = reposName
	} else {
		indexName = nameParts[0]
		remoteName = nameParts[1]
	}
	return indexName, remoteName
}
