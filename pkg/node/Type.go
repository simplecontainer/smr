package node

type Nodes struct {
	Nodes []*Node
}

type Node struct {
	NodeID   uint64
	NodeName string
	API      string
	URL      string
	State    State `json:"-"`
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
	Draining   bool
	Recovering bool
}
