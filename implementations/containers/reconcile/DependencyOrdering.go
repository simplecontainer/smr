package reconcile

import (
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/registry"
	"sort"
)

func orderByDependencies(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	sort.Sort(container.ByDepenendecies(order))

	return order
}
