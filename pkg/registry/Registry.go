package registry

import (
	"fmt"
	"smr/pkg/container"
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

func (registry *Registry) Remove(group string, name string, project string) bool {
	if registry.Containers[group] == nil {
		return true
	} else {
		delete(registry.Containers[group], name)
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

func (registry *Registry) Name(baseName string, project string) (string, int) {
	index := registry.GenerateIndex(baseName, project)
	return fmt.Sprintf("%s-%s-%d", project, baseName, index), index
}

func (registry *Registry) NameReplicas(baseName string, project string, index int) (string, int) {
	return fmt.Sprintf("%s-%s-%d", project, baseName, index), index
}

func (registry *Registry) BackOffTracking(group string, name string) {
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]int)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] += 5
}

func (registry *Registry) BackOffReset(group string, name string) {
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]int)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] = 0
}
