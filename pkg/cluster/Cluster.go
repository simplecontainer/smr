package cluster

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/node"
)

func New() *Cluster {
	nodes := node.NewNodes()

	return &Cluster{
		Node:    node.NewNode(),
		Cluster: nodes,
	}
}

func Restore(config *configuration.Configuration) (*Cluster, error) {
	cluster := node.NewNodes()

	for _, c := range config.KVStore.Cluster {
		cluster.Add(c)
	}

	if len(cluster.Nodes) == 0 {
		return nil, errors.New("cluster is empty")
	}

	return &Cluster{
		Node: &node.Node{
			NodeID:   config.KVStore.Node,
			NodeName: config.NodeName,
			URL:      config.KVStore.URL,
		},
		Cluster: cluster,
	}, nil
}
