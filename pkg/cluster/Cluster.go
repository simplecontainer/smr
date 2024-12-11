package cluster

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/node"
)

func New() *Cluster {
	nodes := node.NewNodes()

	return &Cluster{
		Node:       node.NewNode(),
		Cluster:    nodes,
		EtcdClient: NewEtcdClient(),
	}
}

func Restore(config *configuration.Configuration) (*Cluster, error) {
	cluster := node.NewNodes()

	for i, c := range config.KVStore.Cluster {
		if c.URL == "" {
			cluster.Nodes = append(cluster.Nodes[:i], cluster.Nodes[i+1:]...)
		}
	}

	if len(cluster.Nodes) == 0 {
		return nil, errors.New("cluster is empty")
	}

	return &Cluster{
		Node: &node.Node{
			NodeID: config.KVStore.Node,
			URL:    config.KVStore.URL,
		},
		Cluster: cluster,
	}, nil
}
