package cluster

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Cluster struct {
	Node       *node.Node
	Cluster    *node.Nodes
	Client     *client.Http
	EtcdClient *clientv3.Client
	KVStore    *raft.KVStore
	RaftNode   *raft.RaftNode
	Started    bool
}
