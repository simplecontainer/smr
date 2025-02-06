package replicas

type Replicas struct {
	NodeID  uint64
	Create  []uint64
	Destroy []uint64
	Cluster []uint64
}

type Distributed struct {
	Group    string
	Name     string
	Spread   string
	Replicas map[uint64]*Replicas
}
