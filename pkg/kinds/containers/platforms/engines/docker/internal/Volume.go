package internal

import (
	"errors"
	"fmt"
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
	Type          mount.Type
	ContainerName string
	Name          string
	HostPath      string
	MountPoint    string
	ReadOnly      bool
	SubPath       string
	Prepared      bool
}

func NewVolumes(name string, volumes []v1.ContainersVolume) (*Volumes, error) {
	volumesObj := &Volumes{
		Volumes: make([]*Volume, 0),
	}

	for _, v := range volumes {
		err := volumesObj.Add(name, v)

		if err != nil {
			return volumesObj, err
		}
	}

	volumesObj.Lock = sync.RWMutex{}

	return volumesObj, nil
}

func NewVolume(name string, volume v1.ContainersVolume) *Volume {
	return &Volume{
		ContainerName: name,
		Type:          mount.Type(volume.Type),
		Name:          volume.Name,
		HostPath:      volume.HostPath,
		MountPoint:    volume.MountPoint,
	}
}

func (volumes *Volumes) Add(name string, volume v1.ContainersVolume) error {
	for _, v := range volumes.Volumes {
		if v.MountPoint == volume.MountPoint {
			fmt.Println(volumes.Volumes)
			return errors.New(fmt.Sprintf("mountpoints need to be unique: %s", volume.MountPoint))
		}
	}

	volumes.Volumes = append(volumes.Volumes, NewVolume(name, volume))
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
		m := v.ToMount()

		if m != nil {
			mounts = append(mounts, *m)
		}

		m = nil
	}

	return mounts
}

func (vol *Volume) ToMount() *mount.Mount {
	switch vol.Type {
	case mount.TypeBind:
		return &mount.Mount{
			Type:     vol.Type,
			Source:   vol.HostPath,
			Target:   vol.MountPoint,
			ReadOnly: vol.ReadOnly,
		}
	case mount.TypeVolume:
		return &mount.Mount{
			Type:     vol.Type,
			Source:   vol.Name,
			Target:   vol.MountPoint,
			ReadOnly: vol.ReadOnly,
			VolumeOptions: &mount.VolumeOptions{
				Subpath: vol.SubPath,
			},
		}
	default:
		return nil
	}
}
