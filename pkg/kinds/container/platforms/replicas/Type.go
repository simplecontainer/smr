package replicas

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Replicas struct {
	Group           string
	GeneratedIndex  int
	ExistingIndexes []int
	Replicas        int
	Changed         bool
	Spread          v1.ContainerSpread
	NodeID          uint64
}

type DistributedReplicas struct {
	Group    string
	Name     string
	Replicas map[uint64]*ScopedReplicas
}

type ScopedReplicas struct {
	Create  []R
	Remove  []R
	Numbers Numbers
}

type Numbers struct {
	Create   []int
	Destroy  []int
	Existing []int
}

type R struct {
	Group string
	Name  string
}
