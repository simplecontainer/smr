package node

import (
	"github.com/simplecontainer/smr/pkg/version"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type Nodes struct {
	Nodes []*Node
}

type Node struct {
	NodeID     uint64
	NodeName   string
	API        string
	URL        string
	ConfChange raftpb.ConfChange `yaml:"-" json:"-"`
	State      State
	Version    *version.Version
}

type State struct {
	Health  Health
	Control Control
}

type Health struct {
	Cluster        bool
	Etcd           bool
	Running        bool
	MemoryPressure bool
	CPUPressure    bool
}

type Control struct {
	Upgrading  bool
	Draining   bool
	Recovering bool
}
