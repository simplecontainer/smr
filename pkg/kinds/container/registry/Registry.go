package registry

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"strconv"
	"strings"
)

func (registry *Registry) AddOrUpdate(group string, name string, containerAddr platforms.IContainer) {
	registry.ContainersLock.Lock()

	if registry.Containers[group] == nil {
		tmp := make(map[string]platforms.IContainer)
		tmp[name] = containerAddr

		registry.Containers[group] = tmp
	} else {
		registry.Containers[group][name] = containerAddr
	}

	registry.ContainersLock.Unlock()
}

func (registry *Registry) Sync(container platforms.IContainer) error {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_STATE, static.KIND_CONTAINER, container.GetGroup(), container.GetGeneratedName())
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	bytes, err := container.ToJson()

	if err != nil {
		return err
	}

	return obj.Wait(format, bytes)
}

func (registry *Registry) Remove(group string, name string) error {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	if registry.Containers[group] == nil {
		return errors.New("container not found")
	} else {
		delete(registry.Containers[group], name)

		if len(registry.Containers[group]) == 0 {
			delete(registry.Containers, group)
		}

		format := f.New(static.SMR_PREFIX, static.CATEGORY_STATE, static.KIND_CONTAINER, group, name)
		obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

		err := obj.Propose(format, nil)

		if err != nil {
			return err
		}

		return nil
	}
}

func (registry *Registry) FindLocal(group string, name string) platforms.IContainer {
	registry.ContainersLock.RLock()
	defer registry.ContainersLock.RUnlock()

	if registry.Containers[group] != nil {
		if registry.Containers[group][name] != nil {
			return registry.Containers[group][name]
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (registry *Registry) Find(group string, name string) platforms.IContainer {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_STATE, static.KIND_CONTAINER, group, name)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	registry.ContainersLock.RLock()

	if registry.Containers[group] != nil && registry.Containers[group][name] != nil {
		registry.ContainersLock.RUnlock()
		return registry.Containers[group][name]
	} else {
		registry.ContainersLock.RUnlock()

		obj.Find(format)

		if obj.Exists() {
			instance, err := platforms.NewGhost(obj.GetDefinition())

			if err != nil {
				logger.Log.Error(err.Error())
				return nil
			}

			return instance
		}

		return nil
	}
}

func (registry *Registry) FindGroup(group string) []platforms.IContainer {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_STATE, static.KIND_CONTAINER, group)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	var result []platforms.IContainer
	objs, _ := obj.FindMany(format)

	if len(objs) > 0 {
		for _, o := range objs {
			instance, err := platforms.NewGhost(o.GetDefinition())

			if err != nil {
				logger.Log.Error(err.Error())
				continue
			}

			result = append(result, instance)
		}
	}

	return result
}

func (registry *Registry) All() map[string]map[string]platforms.IContainer {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_STATE, static.KIND_CONTAINER)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	var result = make(map[string]map[string]platforms.IContainer)
	objs, _ := obj.FindMany(format)

	if len(objs) > 0 {
		for _, o := range objs {
			instance, err := platforms.NewGhost(o.GetDefinition())

			if err != nil {
				logger.Log.Error(err.Error())
				continue
			}

			if result[instance.GetGroup()] != nil {
				result[instance.GetGroup()][instance.GetGeneratedName()] = instance
			} else {
				tmp := make(map[string]platforms.IContainer)
				tmp[instance.GetGeneratedName()] = instance

				result[instance.GetGroup()] = tmp
			}
		}
	}

	return result
}

func (registry *Registry) Name(client *client.Http, group string, name string) (string, []uint64) {
	indexes := registry.GetIndexes(group, name)
	index := uint64(1)

	if len(indexes) > 0 {
		index = indexes[len(indexes)-1] + 1
	}

	return fmt.Sprintf("%s-%s-%d", group, name, index), indexes
}

func (registry *Registry) NameReplica(group string, name string, index uint64) string {
	return fmt.Sprintf("%s-%s-%d", group, name, index)
}

func (registry *Registry) BackOffTracking(group string, name string) {
	registry.ContainersLock.Lock()
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]uint64)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] += 1
	registry.ContainersLock.Unlock()
}

func (registry *Registry) BackOffReset(group string, name string) {
	registry.ContainersLock.Lock()
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]uint64)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] = 0
	registry.ContainersLock.Unlock()
}

func (registry *Registry) GetIndexes(group string, name string) []uint64 {
	registry.ContainersLock.RLock()
	containers := registry.Containers[group]
	registry.ContainersLock.RUnlock()

	var indexes = make([]uint64, 0)

	if len(containers) > 0 {
		for _, containerObj := range containers {
			if containerObj.GetName() == name {
				split := strings.Split(containerObj.GetGeneratedName(), "-")
				index, err := strconv.ParseUint(split[len(split)-1], 10, 64)

				if err != nil {
					logger.Log.Fatal("Failed to convert string to uint64 for index calculation")
				}

				indexes = append(indexes, index)
			}
		}
	}

	return indexes
}
