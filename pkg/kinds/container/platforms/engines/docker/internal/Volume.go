package internal

import (
	"errors"
	"github.com/docker/docker/api/types/mount"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
	"strings"
)

type Volumes struct {
	Volumes     []*Volume
	HomeDir     string
	HostHomeDir string
}

type Volume struct {
	Type       mount.Type
	Name       string
	HostPath   string
	MountPoint string
}

func NewVolumes(volumes []v1.ContainerVolume, config *configuration.Configuration) (*Volumes, error) {
	volumesObj := &Volumes{
		Volumes: make([]*Volume, 0),
	}

	for _, v := range volumes {
		err := volumesObj.Add(v)

		if err != nil {
			return volumesObj, err
		}
	}

	volumesObj.HomeDir = config.Environment.HOMEDIR
	volumesObj.HostHomeDir = config.HostHome

	return volumesObj, nil
}

func NewVolume(volume v1.ContainerVolume) *Volume {
	return &Volume{
		Type:       mount.Type(volume.Type),
		Name:       volume.Name,
		HostPath:   volume.HostPath,
		MountPoint: volume.MountPoint,
	}
}

func (volumes *Volumes) Add(volume v1.ContainerVolume) error {
	for _, v := range volumes.Volumes {
		if v.MountPoint == volume.MountPoint {
			return errors.New("mountpoints need to be unique")
		}
	}

	volumes.Volumes = append(volumes.Volumes, NewVolume(volume))
	return nil
}

func (volumes *Volumes) RemoveResources() error {
	for k, v := range volumes.Volumes {
		if v.Type == "resource" {
			err := os.Remove(v.HostPath)

			if err != nil {
				return err
			}

			volumes.Volumes = append(volumes.Volumes[:k], volumes.Volumes[k+1:]...)
		}
	}

	return nil
}

func (volumes *Volumes) ToMounts() ([]mount.Mount, error) {
	mounts := make([]mount.Mount, 0)

	for _, v := range volumes.Volumes {
		switch v.Type {
		case mount.TypeBind:
			mounts = append(mounts, mount.Mount{
				Type:   v.Type,
				Source: strings.Replace(v.HostPath, "~", volumes.HostHomeDir, 1),
				Target: strings.Replace(v.MountPoint, "~", volumes.HomeDir, 1),
			})
			break
		case mount.TypeVolume:
			mounts = append(mounts, mount.Mount{
				Type:   v.Type,
				Source: v.Name,
				Target: strings.Replace(v.MountPoint, "~", volumes.HomeDir, 1),
			})
			break
		case "resource":
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: strings.Replace(v.HostPath, "~", volumes.HostHomeDir, 1),
				Target: strings.Replace(v.MountPoint, "~", volumes.HomeDir, 1),
			})
			break
		default:
			return mounts, errors.New("only supported types are: bind and volume")
		}
	}

	return mounts, nil
}
