package node

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/version"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func NewNode() *Node {
	return &Node{
		NodeID:   0,
		NodeName: "",
		API:      "",
		URL:      "",
		State:    NewState(),
		Version:  version.New("", ""),
	}
}

func NewNodeDefinition(cluster []*Node, nodeId uint64) *Node {
	for _, n := range cluster {
		if n.NodeID == nodeId {
			return &Node{
				NodeID:   n.NodeID,
				NodeName: n.NodeName,
				API:      n.API,
				URL:      n.URL,
				State:    n.State,
				Version:  n.Version,
			}
		}
	}

	return nil
}

func (node *Node) Accepting() bool {
	// If any of this status is true - node should not accept new objects
	return !(node.State.Control.Draining || node.State.Control.Upgrading)
}

func (node *Node) SetDrain(drain bool) {
	node.State.Control.Draining = drain
}

func (node *Node) SetUpgrade(upgrade bool) {
	node.State.Control.Upgrading = upgrade
}

func (node *Node) Parse(change raftpb.ConfChange) error {
	node.NodeID = change.NodeID
	node.ConfChange = change

	switch change.Type {
	case raftpb.ConfChangeAddNode:
		return json.Unmarshal(change.Context, node)
	default:
		return nil
	}
}

func (node *Node) ToJSON() ([]byte, error) {
	return json.Marshal(node)
}
