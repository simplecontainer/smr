package types

import (
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/smaps"
)

const EVENT_NETWORK_CONNECT = "conn"
const EVENT_NETWORK_DISCONNECT = "disscon"
const EVENT_START = "start"
const EVENT_STOP = "start"
const EVENT_KILL = "start"
const EVENT_DIE = "start"

type Runtime struct {
	Id                 string
	State              string
	Ready              bool
	Configuration      *smaps.Smap
	ObjectDependencies []*f.Format
	NodeIP             string
	Agent              string
}

type ExecResult struct {
	Stdout string
	Stderr string
	Exit   int
}
