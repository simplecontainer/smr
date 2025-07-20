package channels

import (
	"github.com/simplecontainer/smr/pkg/node"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func NewCluster() *Cluster {
	return &Cluster{
		Propose:       make(chan string),
		Insync:        make(chan bool),
		ConfChange:    make(chan raftpb.ConfChange),
		NodeUpdate:    make(chan node.Node),
		NodeFinalizer: make(chan node.Node),
	}
}
