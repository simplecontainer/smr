package node

type Nodes struct {
	Nodes []*Node
}

type Node struct {
	NodeID uint64
	URL    string
}
