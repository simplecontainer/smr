package replicas

import (
	"github.com/r3labs/diff/v3"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
)

type Replicas struct {
	NodeID          uint64
	Definition      *v1.ContainerDefinition
	Shared          *shared.Shared
	Distributed     *Distributed
	CreateScoped    []platforms.IContainer
	DeleteScoped    []platforms.IContainer
	Agent           string
	ChangeLog       diff.Changelog
	Spread          v1.ContainerSpread
	ExistingIndexes []uint64
}

type Distributed struct {
	Group    string
	Name     string
	Spread   string
	Replicas map[uint64]*ScopedReplicas
}

type ScopedReplicas struct {
	Create   []R
	Remove   []R
	Existing []R
	Numbers  Numbers
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
