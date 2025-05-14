package cluster

import (
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
)

type Cluster struct {
	Node          *node.Node
	Cluster       *node.Nodes
	InSync        chan bool
	NodeConf      chan node.Node
	NodeFinalizer chan node.Node
	KVStore       *raft.KVStore
	RaftLeader    uint64
	RaftNode      *raft.RaftNode
	Started       bool
	Replay        bool
	Join          bool
}
