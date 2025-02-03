package internal

import (
	"errors"
	"github.com/docker/docker/api/types/mount"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
	"sync"
)

type Volumes struct {
	Volumes []*Volume
	Lock    sync.RWMutex
}

type Volume struct {
	Type       mount.Type
	Name       string
	HostPath   string
	MountPoint string
}

func NewVolumes(volumes []v1.ContainerVolume) (*Volumes, error) {
	volumesObj := &Volumes{
		Volumes: make([]*Volume, 0),
	}

	for _, v := range volumes {
		err := volumesObj.Add(v)

		if err != nil {
			return volumesObj, err
		}
	}

	volumesObj.Lock = sync.RWMutex{}

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
	tmpSlice := make([]*Volume, 0)

	for _, v := range volumes.Volumes {
		if v.Type == "resource" {
			err := os.Remove(v.HostPath)

			if err != nil {
				return err
			}
		} else {
			tmpSlice = append(tmpSlice, v)
		}
	}

	volumes.Volumes = tmpSlice
	return nil
}

func (volumes *Volumes) ToMounts() []mount.Mount {
	mounts := make([]mount.Mount, 0)

	for _, v := range volumes.Volumes {
		switch v.Type {
		case mount.TypeBind:
			mounts = append(mounts, mount.Mount{
				Type:   v.Type,
				Source: v.HostPath,
				Target: v.MountPoint,
			})
			break
		case mount.TypeVolume:
			mounts = append(mounts, mount.Mount{
				Type:   v.Type,
				Source: v.Name,
				Target: v.MountPoint,
			})
			break
		case "resource":
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: v.HostPath,
				Target: v.MountPoint,
			})
			break
		}
	}

	return mounts
}
