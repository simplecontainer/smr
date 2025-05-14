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
		Replay:  false,
	}
}

func (cluster *Cluster) Peers() *node.Nodes {
	var peers = node.NewNodes()

	for _, n := range cluster.Cluster.Nodes {
		if n.NodeID != cluster.Node.NodeID {
			peers.Add(n)
		}
	}

	if len(peers.Nodes) == 0 {
		peers.Add(cluster.Node)
	}

	return peers
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
			NodeID:   config.KVStore.Node.NodeID,
			NodeName: config.NodeName,
			URL:      config.KVStore.URL,
			API:      config.KVStore.API,
		},
		Cluster: cluster,
		Replay:  config.KVStore.Replay,
		Join:    config.KVStore.Join,
	}, nil
}
