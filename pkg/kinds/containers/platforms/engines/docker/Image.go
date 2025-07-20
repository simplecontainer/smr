package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	TDImage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/image"
	"io"
	"os"
	"strings"
)

func (container *Docker) PullImage(ctx context.Context, cli *client.Client) error {
	if container.CheckIfImagePresent(ctx, cli) != nil {
		container.ImageState.SetStatus(image.StatusPulling)

		pullOpts := TDImage.PullOptions{
			All:           false,
			RegistryAuth:  "",
			PrivilegeFunc: nil,
			Platform:      "",
		}

		var err error

		if container.RegistryAuth != "" {
			pullOpts, err = container.GetDockerAuth()

			if err != nil {
				container.ImageState.SetStatus(image.StatusFailed)
				return err
			}
		}

		reader, err := cli.ImagePull(ctx, container.Image+":"+container.Tag, pullOpts)
		if err != nil {
			container.ImageState.SetStatus(image.StatusFailed)
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
				container.ImageState.SetStatus(image.StatusFailed)
				return err
			}
		}

		defer func(reader io.ReadCloser) {
			err = reader.Close()
			if err != nil {
				return
			}
		}(reader)

		container.ImageState.SetStatus(image.StatusPulled)
		return nil
	} else {
		container.ImageState.SetStatus(image.StatusPulled)
		return nil
	}
}

func (container *Docker) CheckIfImagePresent(ctx context.Context, cli *client.Client) error {
	images, err := cli.ImageList(ctx, TDImage.ListOptions{
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

func (container *Docker) GetDockerAuth() (TDImage.PullOptions, error) {
	auth, err := decodeAuth(container.RegistryAuth)

	if err != nil {
		return TDImage.PullOptions{}, err
	}

	var encoded string
	encoded, err = registry.EncodeAuthConfig(auth)

	if err != nil {
		return TDImage.PullOptions{}, err
	}

	return TDImage.PullOptions{
		RegistryAuth: encoded,
	}, nil
}

func decodeAuth(authStr string) (registry.AuthConfig, error) {
	if authStr == "" {
		return registry.AuthConfig{}, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(authStr)
	if err != nil {
		return registry.AuthConfig{}, err
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return registry.AuthConfig{}, fmt.Errorf("invalid auth format")
	}

	return registry.AuthConfig{
		Username: parts[0],
		Password: parts[1],
	}, nil
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
