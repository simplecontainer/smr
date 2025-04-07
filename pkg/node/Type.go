package node

import (
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type Nodes struct {
	Nodes []*Node
}

type Node struct {
	NodeID     uint64
	NodeName   string
	Version    string
	API        string
	URL        string
	ConfChange raftpb.ConfChange `json:"-"`
	State      State             `json:"-"`
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
