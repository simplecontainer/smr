package registry

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/logger"
	"strconv"
	"strings"
)

func (registry *Registry) Name(client *clients.Http, prefix string, group string, name string) (string, []uint64, error) {
	indexes, err := registry.GetIndexes(prefix, group, name)
	index := uint64(1)

	if err != nil {
		return fmt.Sprintf("%s-%s-%d", group, name, index), indexes, err
	}

	if len(indexes) > 0 {
		index = indexes[len(indexes)-1] + 1
	}

	return fmt.Sprintf("%s-%s-%d", group, name, index), indexes, nil
}

func (registry *Registry) NameReplica(group string, name string, index uint64) string {
	return fmt.Sprintf("%s-%s-%d", group, name, index)
}

func (registry *Registry) GetIndexes(prefix string, group string, name string) ([]uint64, error) {
	var indexes = make([]uint64, 0)

	containers := registry.FindGroup(prefix, group)

	for _, containerObj := range containers {
		if containerObj.GetName() == strings.TrimSpace(name) && containerObj.GetGroup() == strings.TrimSpace(group) {
			split := strings.Split(containerObj.GetGeneratedName(), "-")
			index, err := strconv.ParseUint(split[len(split)-1], 10, 64)

			if err != nil {
				return nil, err
			}

			indexes = append(indexes, index)
		}
	}

	return indexes, nil
}

func (registry *Registry) GetIndexesLocal(prefix string, group string, name string) ([]uint64, error) {
	registry.ContainersLock.RLock()
	defer registry.ContainersLock.RUnlock()

	var indexes = make([]uint64, 0)

	for _, containerObj := range registry.Containers {
		if containerObj.GetName() == name {
			split := strings.Split(containerObj.GetGeneratedName(), "-")
			index, err := strconv.ParseUint(split[len(split)-1], 10, 64)

			if err != nil {
				logger.Log.Fatal("Failed to convert string to uint64 for index calculation")
			}

			indexes = append(indexes, index)
		}
	}

	return indexes, nil
}
