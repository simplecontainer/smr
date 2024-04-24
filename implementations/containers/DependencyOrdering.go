package main

import (
	"smr/pkg/container"
	"smr/pkg/registry"
	"sort"
)

func (implementation *Implementation) orderByDependencies(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	sort.Sort(container.ByDepenendecies(order))

	return order
}
