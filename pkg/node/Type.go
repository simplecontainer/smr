package node

type Nodes struct {
	Nodes []*Node
}

type Node struct {
	NodeID   uint64
	NodeName string
	API      string
	URL      string
}
