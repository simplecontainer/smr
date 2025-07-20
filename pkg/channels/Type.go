package channels

import (
	"github.com/simplecontainer/smr/pkg/node"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type Cluster struct {
	Propose       chan string
	Insync        chan bool
	ConfChange    chan raftpb.ConfChange
	NodeUpdate    chan node.Node
	NodeFinalizer chan node.Node
}
