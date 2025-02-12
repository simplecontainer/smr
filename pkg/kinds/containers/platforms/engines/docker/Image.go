package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"io"
	"os"
	"strings"
)

func (container *Docker) PullImage(ctx context.Context, cli *client.Client) error {
	if container.CheckIfImagePresent(ctx, cli) != nil {
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

		return nil
	} else {
		return nil
	}
}

func (container *Docker) CheckIfImagePresent(ctx context.Context, cli *client.Client) error {
	images, err := cli.ImageList(ctx, image.ListOptions{
		All: true,
	})

	if err != nil {
		return err
	}

	searchingFor := fmt.Sprintf("%s:%s", container.Image, container.Tag)

	for _, image := range images {
		for _, tag := range image.RepoTags {
			registryTo, imageTo := splitReposSearchTerm(tag)
			registryFrom, imageFrom := splitReposSearchTerm(searchingFor)

			if registryTo == registryFrom && imageTo == imageFrom {
				return nil
			}
		}
	}

	return errors.New("image not present")
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

		indexName = "quay.io"
		remoteName = reposName
	} else {
		indexName = nameParts[0]
		remoteName = nameParts[1]
	}
	return indexName, remoteName
}

func GetAuth(image string, environment *configuration.Environment) string {
	dockerConfig := fmt.Sprintf("%s/%s", environment.Home, ".docker/config.json")
	if _, err := os.Stat(dockerConfig); err == nil {
		body, err := os.ReadFile(dockerConfig)
		if err != nil {
			panic("Unable to read docker auth file")
		}

		config := map[string]map[string]map[string]string{}
		err = json.Unmarshal(body, &config)

		if err != nil {
			panic(err)
		}

		for registry, auth := range config["auths"] {
			if strings.Contains(image, registry) {
				return auth["auth"]
			}
		}

		return ""
	} else {
		return ""
	}
}
