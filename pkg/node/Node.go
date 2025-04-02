package node

import (
	"encoding/json"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func NewNode() *Node {
	return &Node{
		NodeID:   0,
		NodeName: "",
		API:      "",
		URL:      "",
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
			}
		}
	}

	return nil
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

func (node *Node) ToJson() ([]byte, error) {
	return json.Marshal(node)
}
