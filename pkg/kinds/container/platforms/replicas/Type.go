package replicas

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Replicas struct {
	Group           string
	GeneratedIndex  uint64
	ExistingIndexes []uint64
	Replicas        uint64
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
	Create   []uint64
	Destroy  []uint64
	Existing []uint64
}

type R struct {
	Group string
	Name  string
}
