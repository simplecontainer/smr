package main

import (
	"github.com/qdnqn/smr/implementations/container/container"
	"github.com/qdnqn/smr/pkg/registry"
)

func (implementation *Implementation) findContainers(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	return order
}
