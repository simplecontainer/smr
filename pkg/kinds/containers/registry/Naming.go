package registry

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/logger"
	"strconv"
	"strings"
)

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

func (registry *Registry) GetIndexes(group string, name string) []uint64 {
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

	return indexes
}
