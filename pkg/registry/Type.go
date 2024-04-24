package registry

import "smr/pkg/container"

type Registry struct {
	Containers     map[string]map[string]*container.Container
	Indexes        map[string][]int
	BackOffTracker map[string]map[string]int
}
