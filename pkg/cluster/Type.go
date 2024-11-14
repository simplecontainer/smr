package cluster

import (
	"github.com/simplecontainer/smr/pkg/raft"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Cluster struct {
	Node       *Node
	Cluster    []string
	EtcdClient *clientv3.Client
	KVStore    *raft.KVStore
	RaftNode   *raft.RaftNode
}

type Node struct {
	NodeID uint64
	URL    string
}
