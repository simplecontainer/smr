package node

import (
	"encoding/json"
	"github.com/docker/docker/client"
	"io"
)

func NewNodes() *Nodes {
	return &Nodes{}
}

func NewNode() *Node {
	return &Node{
		NodeID: 0,
		URL:    "",
	}
}

func (nodes *Nodes) NewNode(url string) *Node {
	return &Node{
		NodeID: nodes.generateID(),
		URL:    url,
	}
}

func (nodes *Nodes) NewNodeRequest(body io.ReadCloser) (*Node, error) {
	var data []byte
	var err error

	data, err = io.ReadAll(body)

	if err != nil {
		return nil, err
	}

	var request map[string]string
	err = json.Unmarshal(data, &request)

	if err != nil {
		return nil, err
	}

	node := &Node{
		NodeID: nodes.generateID(),
		URL:    request["node"],
	}

	_, err = client.ParseHostURL(node.URL)

	if err != nil {
		return nil, err
	}

	return node, nil
}

func (nodes *Nodes) Add(node *Node) {
	if node == nil {
		return
	}

	for _, n := range nodes.Nodes {
		if n.URL == node.URL {
			return
		}
	}

	nodes.Nodes = append(nodes.Nodes, node)
}

func (nodes *Nodes) Remove(node *Node) {
	if node == nil {
		return
	}

	for i, n := range nodes.Nodes {
		if n.URL == node.URL {
			nodes.Nodes = append(nodes.Nodes[:i], nodes.Nodes[i+1:]...)
		}
	}
}

func (nodes *Nodes) generateID() uint64 {
	maxID := uint64(0)

	for _, node := range nodes.Nodes {
		if node.NodeID > maxID {
			maxID = node.NodeID
		}
	}

	return maxID + 1
}

func (nodes *Nodes) ToString() []string {
	toStringSlice := make([]string, 0)

	for _, n := range nodes.Nodes {
		if n.URL != "" {
			toStringSlice = append(toStringSlice, n.URL)
		}
	}

	return toStringSlice
}
