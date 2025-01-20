package cluster

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
)

type Cluster struct {
	Node     *node.Node
	Cluster  *node.Nodes
	Client   *client.Http
	NodeConf chan node.Node
	KVStore  *raft.KVStore
	RaftNode *raft.RaftNode
	Started  bool
}
