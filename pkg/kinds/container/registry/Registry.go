package registry

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/logger"
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

func (registry *Registry) Remove(group string, name string) bool {
	registry.ContainersLock.Lock()
	if registry.Containers[group] == nil {
		registry.ContainersLock.Unlock()
		return true
	} else {
		delete(registry.Containers[group], name)

		if len(registry.Containers[group]) == 0 {
			delete(registry.Containers, group)
		}

		registry.ContainersLock.Unlock()
		return true
	}
}

func (registry *Registry) Find(group string, name string) platforms.IContainer {
	registry.ContainersLock.RLock()
	if registry.Containers[group] != nil {
		if registry.Containers[group][name] != nil {
			registry.ContainersLock.RUnlock()
			return registry.Containers[group][name]
		} else {
			registry.ContainersLock.RUnlock()
			return nil
		}
	} else {
		registry.ContainersLock.RUnlock()
		return nil
	}
}

func (registry *Registry) Name(client *client.Http, group string, name string) (string, []int) {
	indexes := registry.GetIndexes(group, name)
	index := 1

	if len(indexes) > 0 {
		index = indexes[len(indexes)-1] + 1
	}

	return fmt.Sprintf("%s-%s-%d", group, name, index), indexes
}

func (registry *Registry) NameReplicas(group string, name string, index int) (string, int) {
	return fmt.Sprintf("%s-%s-%d", group, name, index), index
}

func (registry *Registry) BackOffTracking(group string, name string) {
	registry.ContainersLock.Lock()
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]int)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] += 1
	registry.ContainersLock.Unlock()
}

func (registry *Registry) BackOffReset(group string, name string) {
	registry.ContainersLock.Lock()
	if registry.BackOffTracker[group] == nil {
		tmp := make(map[string]int)
		tmp[name] = 0

		registry.BackOffTracker[group] = tmp
	}

	registry.BackOffTracker[group][name] = 0
	registry.ContainersLock.Unlock()
}

func (registry *Registry) GetIndexes(group string, name string) []int {
	containers := registry.Containers[group]

	var indexes = make([]int, 0)

	if len(containers) > 0 {
		for _, containerObj := range containers {
			if containerObj.GetName() == name {
				split := strings.Split(containerObj.GetGeneratedName(), "-")
				index, err := strconv.Atoi(split[len(split)-1])

				if err != nil {
					logger.Log.Fatal("Failed to convert string to int for index calculation")
				}

				indexes = append(indexes, index)
			}
		}
	}

	return indexes
}
