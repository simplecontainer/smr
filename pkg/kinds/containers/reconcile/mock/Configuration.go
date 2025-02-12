package mock

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/node"
)

func NewConfig(platform string) *configuration.Configuration {
	return &configuration.Configuration{
		Platform: platform,
		NodeName: "node-1",
		HostPort: configuration.HostPort{
			Host: "0.0.0.0",
			Port: "1443",
		},
		KVStore: &configuration.KVStore{
			Cluster:     []*node.Node{node.NewNode()},
			Node:        1,
			URL:         "172.0.0.2",
			JoinCluster: false,
		},
		Certificates: nil,
		Environment: &configuration.Environment{
			Home:          "/home/node/",
			NodeIP:        "172.0.0.2",
			NodeDirectory: "/home/node/.node-1/",
		},
	}
}
