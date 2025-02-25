package node

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
)

func NewNode() *Node {
	return &Node{
		NodeID:   0,
		NodeName: "",
		API:      "",
		URL:      "",
	}
}

func NewNodeDefinition(runtime *commonv1.Runtime, cluster []*Node) *Node {
	for _, n := range cluster {
		if n.NodeID == runtime.GetNode() {
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
