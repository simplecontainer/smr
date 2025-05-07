package ishared

import (
	"github.com/simplecontainer/smr/pkg/cluster"
)

type Shared interface {
	GetCluster() *cluster.Cluster
	Drain()
	IsDrained() bool
}
