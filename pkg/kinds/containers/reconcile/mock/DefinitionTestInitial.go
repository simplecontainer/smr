package mock

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

func DefinitionTestInitial(name string, platform string) *v1.ContainersDefinition {
	config := NewConfig(platform)
	container := &v1.ContainersDefinition{
		Meta: commonv1.Meta{
			Name:   name,
			Group:  "internal",
			Labels: nil,
		},
		Spec: v1.ContainersInternal{
			Image: "dummy",
			Tag:   "latest",
			Envs: []string{
				fmt.Sprintf("LOG_LEVEL=%s", "info"),
			},
			Entrypoint: []string{"/bin/bash"},
			Args:       []string{"-c", "'echo test'"},
			Ports: []v1.ContainersPort{
				{
					Container: "1443",
					Host:      "1443",
				},
				{
					Container: "2379",
					Host:      "2379",
				},
				{
					Container: "9212",
					Host:      "9212",
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
					HostPath:   fmt.Sprintf("%s/.%s", config.Environment.Home, name),
					MountPoint: "/home/node/smr",
				},
				{
					Name:       "ssh",
					Type:       "bind",
					HostPath:   fmt.Sprintf("%s/.ssh", config.Environment.Home),
					MountPoint: "/home/node/.ssh",
				},
				{
					Name:       "tmp",
					Type:       "bind",
					HostPath:   "/tmp",
					MountPoint: "/tmp",
				},
			},
			Replicas: 1,
			Dns:      []string{"127.0.0.1"},
		},
	}

	return container
}
