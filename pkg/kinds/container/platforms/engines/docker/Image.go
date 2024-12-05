package docker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/logger"
	"io"
	"strings"
)

func (container *Docker) PullImage(ctx context.Context, cli *client.Client) error {
	if !container.CheckIfImagePresent(ctx, cli) {
		logger.Log.Info(fmt.Sprintf("Pulling the image %s:%s", container.Image, container.Tag))

		reader, err := cli.ImagePull(ctx, container.Image+":"+container.Tag, container.GetDockerAuth())
		if err != nil {
			return err
		}

		type ErrorMessage struct {
			Error string
		}

		buffIOReader := bufio.NewReader(reader)

		for {
			_, err = buffIOReader.ReadBytes('\n')
			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}
		}

		defer func(reader io.ReadCloser) {
			err = reader.Close()
			if err != nil {
				return
			}
		}(reader)

		logger.Log.Info(fmt.Sprintf("pulled the image %s:%s", container.Image, container.Tag))

		return nil
	} else {
		logger.Log.Info(fmt.Sprintf("image %s:%s already present", container.Image, container.Tag))
		return nil
	}
}

func (container *Docker) CheckIfImagePresent(ctx context.Context, cli *client.Client) bool {
	images, err := cli.ImageList(ctx, image.ListOptions{
		All: true,
	})

	if err != nil {
		logger.Log.Fatal("failed to list container images")
	}

	searchingFor := fmt.Sprintf("%s:%s", container.Image, container.Tag)

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

func (container *Docker) GetDockerAuth() image.PullOptions {
	return image.PullOptions{
		RegistryAuth: container.Auth,
	}
}

func splitReposSearchTerm(reposName string) (string, string) {
	nameParts := strings.SplitN(reposName, "/", 2)
	var indexName, remoteName string
	if len(nameParts) == 1 || (!strings.Contains(nameParts[0], ".") &&
		!strings.Contains(nameParts[0], ":") && nameParts[0] != "localhost") {

		indexName = "docker.io"
		remoteName = reposName
	} else {
		indexName = nameParts[0]
		remoteName = nameParts[1]
	}
	return indexName, remoteName
}
