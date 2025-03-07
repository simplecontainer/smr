package node

import (
	"encoding/json"
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

func (node *Node) ToJson() ([]byte, error) {
	return json.Marshal(node)
}
