package node

import "encoding/json"

func NewNode() *Node {
	return &Node{
		NodeID:   0,
		NodeName: "",
		API:      "",
		URL:      "",
	}
}

func (node *Node) ToJson() ([]byte, error) {
	return json.Marshal(node)
}
