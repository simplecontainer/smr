package registry

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(client *client.Http, user *authentication.User) platforms.Registry {
	return &Registry{
		Containers:     make(map[string]platforms.IContainer),
		BackOffTracker: make(map[string]map[string]uint64),
		Client:         client,
		User:           user,
	}
}

func (registry *Registry) AddOrUpdate(group string, name string, containerAddr platforms.IContainer) {
	registry.ContainersLock.Lock()
	registry.Containers[common.GroupIdentifier(group, name)] = containerAddr
	registry.ContainersLock.Unlock()
}

func (registry *Registry) Remove(prefix string, group string, name string) error {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	if registry.Containers[common.GroupIdentifier(group, name)] == nil {
		return errors.New("container not found")
	} else {
		delete(registry.Containers, common.GroupIdentifier(group, name))

		format := f.New(prefix, static.CATEGORY_STATE, static.KIND_CONTAINERS, group, name)
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

	value, ok := registry.Containers[common.GroupIdentifier(group, name)]

	if ok {
		return value
	} else {
		return nil
	}
}

func (registry *Registry) Find(prefix string, group string, name string) platforms.IContainer {
	format := f.New(prefix, static.CATEGORY_STATE, static.KIND_CONTAINERS, group, name)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	registry.ContainersLock.RLock()
	value, ok := registry.Containers[common.GroupIdentifier(group, name)]
	registry.ContainersLock.RUnlock()

	if ok {
		return value
	} else {
		obj.Find(format)

		if obj.Exists() {
			instance, err := platforms.NewGhost(obj.GetDefinition())

			if err != nil {
				logger.Log.Error(err.Error())
				return nil
			}

			return instance
		} else {
			return nil
		}
	}
}

func (registry *Registry) FindGroup(prefix string, group string) []platforms.IContainer {
	format := f.New(prefix, static.CATEGORY_STATE, static.KIND_CONTAINERS, group)
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

func (registry *Registry) Sync(group string, name string) error {
	registry.ContainersLock.RLock()
	container, ok := registry.Containers[common.GroupIdentifier(group, name)]
	registry.ContainersLock.RUnlock()

	if ok {
		format := f.New(container.GetDefinition().GetPrefix(), static.CATEGORY_STATE, static.KIND_CONTAINERS, container.GetGroup(), container.GetGeneratedName())
		obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

		bytes, err := container.ToJson()

		if err != nil {
			return err
		}

		return obj.Wait(format, bytes)
	} else {
		return errors.New("container not found on this node")
	}
}

func (registry *Registry) BackOff(group string, name string) error {
	registry.ContainersLock.Lock()
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]uint64)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] += 1

	defer registry.ContainersLock.Unlock()

	if registry.BackOffTracker[group][name] > 5 {
		return errors.New("container is in backoff reset loop")
	}

	return nil
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
