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

type ControlStatus string

const (
	StatusNotStarted ControlStatus = "not_started"
	StatusInProgress ControlStatus = "in_progress"
	StatusFailed     ControlStatus = "failed"
	StatusSuccess    ControlStatus = "success"
)

type Health struct {
	Cluster        bool
	Etcd           bool
	Running        bool
	MemoryPressure bool
	CPUPressure    bool
}

type Control struct {
	Starting   ControlStatus
	Upgrading  ControlStatus
	Draining   ControlStatus
	Recovering ControlStatus
}

type State struct {
	Health  Health
	Control Control
}
