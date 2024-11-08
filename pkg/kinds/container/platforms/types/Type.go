package types

import "github.com/simplecontainer/smr/pkg/f"

const EVENT_NETWORK_CONNECT = "conn"
const EVENT_NETWORK_DISCONNECT = "disscon"
const EVENT_START = "start"
const EVENT_STOP = "start"
const EVENT_KILL = "start"
const EVENT_DIE = "start"

type Runtime struct {
	Id                 string
	State              string
	FoundRunning       bool
	FirstObserved      bool
	Ready              bool
	Configuration      map[string]string
	Owner              Owner
	ObjectDependencies []*f.Format
}

type Owner struct {
	Kind            string
	GroupIdentifier string
}

type ExecResult struct {
	Stdout string
	Stderr string
	Exit   int
}

type Events struct {
	Kind    string
	Group   string
	Name    string
	Message string
}
