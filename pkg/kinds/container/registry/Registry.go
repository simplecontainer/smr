package registry

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/logger"
	"sort"
	"strconv"
	"strings"
)

func (registry *Registry) AddOrUpdate(group string, name string, project string, containerAddr platforms.IContainer) {
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

func (registry *Registry) Name(group string, name string, project string) (string, int) {
	index := registry.GenerateIndex(group, project)
	return fmt.Sprintf("%s-%s-%d", group, name, index), index
}

func (registry *Registry) NameReplicas(group string, name string, project string, index int) (string, int) {
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

func (registry *Registry) GenerateIndex(name string, project string) int {
	var indexes []int = registry.GetIndexes(name, project)
	var index int = 0

	if len(indexes) > 0 {
		sort.Ints(indexes)
		index = indexes[len(indexes)-1] + 1
	}

	if index < 0 {
		index = 0
	}

	return index
}

func (registry *Registry) GetIndexes(name string, project string) []int {
	containers := registry.Containers[name]

	var indexes = make([]int, 0)
	name = fmt.Sprintf("%s-%s", project, name)

	if len(containers) > 0 {
		for _, containerObj := range containers {
			split := strings.Split(containerObj.GetGeneratedName(), "-")
			index, err := strconv.Atoi(split[len(split)-1])

			if err != nil {
				logger.Log.Fatal("Failed to convert string to int for index calculation")
			}

			indexes = append(indexes, index)
		}
	}

	if len(indexes) == 0 {
		indexes = append(indexes, 0)
	}

	return indexes
}
