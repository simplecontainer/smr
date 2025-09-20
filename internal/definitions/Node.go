package definitions

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/spf13/viper"
	"unicode"
)

func Node(name string, config *configuration.Configuration, entrypoint []string, args []string) (*v1.ContainersDefinition, error) {
	// Prevent etcd port to contain anything except numbers
	for _, r := range config.Ports.Etcd {
		if !unicode.IsDigit(r) {
			return nil, errors.New("etcd port can only contain numbers")
		}
	}

	container := &v1.ContainersDefinition{
		Meta: &commonv1.Meta{
			Name:   name,
			Group:  "internal",
			Labels: nil,
		},
		Spec: &v1.ContainersInternal{
			Image: config.NodeImage,
			Tag:   config.NodeTag,
			Envs: []string{
				fmt.Sprintf("LOG_LEVEL=%s", viper.GetString("log")),
				fmt.Sprintf("HOME=/home/node"),
			},
			Entrypoint: entrypoint,
			Args:       args,
			User:       config.Environment.Host.User,
			GroupAdd:   config.Environment.Host.Groups,
			Ports: []v1.ContainersPort{
				{
					Container: "1443",
					Host:      config.Ports.Control,
				},
				{
					Container: "9212",
					Host:      config.Ports.Overlay,
				},
				{
					Container: "2379",
					Host:      fmt.Sprintf("127.0.0.1:%s", config.Ports.Etcd), // Always keep it like this
				},
			},
			Volumes: []v1.ContainersVolume{
				{
					Name:       "docker-socket",
					Type:       "bind",
					HostPath:   "/var/run/docker.sock",
					MountPoint: "/var/run/docker.sock",
				},
				{
					Name:       "smr",
					Type:       "bind",
					HostPath:   config.Environment.Host.NodeDirectory,
					MountPoint: "/home/node",
				},
				{
					Name:       "ssh",
					Type:       "bind",
					HostPath:   fmt.Sprintf("%s/.ssh", config.Environment.Host.NodeDirectory),
					MountPoint: "/home/node/.ssh",
				},
			},
			Replicas: 1,
			Dns:      []string{"127.0.0.1"},
		},
	}

	return container, nil
}
