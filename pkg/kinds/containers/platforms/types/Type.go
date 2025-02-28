package types

import (
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/smaps"
)

const EVENT_NETWORK_CONNECT = "conn"
const EVENT_NETWORK_DISCONNECT = "disscon"
const EVENT_START = "start"
const EVENT_STOP = "stop"
const EVENT_KILL = "kill"
const EVENT_DIE = "die"

type Runtime struct {
	Id                 string
	State              string
	Ready              bool
	Configuration      *smaps.Smap
	ObjectDependencies []f.Format
	Node               *node.Node
	NodeName           string
}

type ExecResult struct {
	Stdout string
	Stderr string
	Exit   int
}
