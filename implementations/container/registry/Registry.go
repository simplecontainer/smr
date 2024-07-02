package registry

import (
	"fmt"
	"github.com/simplecontainer/smr/implementations/container/container"
)

func (registry *Registry) AddOrUpdate(group string, name string, project string, containerAddr *container.Container) {
	if registry.Containers[group] == nil {
		tmp := make(map[string]*container.Container)
		tmp[name] = containerAddr

		registry.Containers[group] = tmp
	} else {
		registry.Containers[group][name] = containerAddr
	}
}

func (registry *Registry) Remove(group string, name string) bool {
	if registry.Containers[group] == nil {
		return true
	} else {
		delete(registry.Containers[group], name)

		if len(registry.Containers[group]) == 0 {
			delete(registry.Containers, group)
		}

		return true
	}
}

func (registry *Registry) Find(group string, name string) *container.Container {
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

func (registry *Registry) Name(group string, name string, project string) (string, int) {
	index := registry.GenerateIndex(group, project)
	return fmt.Sprintf("%s-%s-%s-%d", project, group, name, index), index
}

func (registry *Registry) NameReplicas(group string, name string, project string, index int) (string, int) {
	return fmt.Sprintf("%s-%s-%s-%d", project, group, name, index), index
}

func (registry *Registry) BackOffTracking(group string, name string) {
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]int)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] += 1
}

func (registry *Registry) BackOffReset(group string, name string) {
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]int)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] = 0
}
