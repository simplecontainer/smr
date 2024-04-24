package registry

import (
	"smr/pkg/container"
	"smr/pkg/objects"
)

type Registry struct {
	Containers     map[string]map[string]*container.Container
	Indexes        map[string][]int
	BackOffTracker map[string]map[string]int
	Object         map[string]objects.Object
}
