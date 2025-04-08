package node

import (
	"encoding/json"
	"github.com/docker/docker/client"
	"io"
)

func NewNodes() *Nodes {
	return &Nodes{}
}

func (nodes *Nodes) NewNode(nodeName string, url string, API string) *Node {
	n := NewNode()

	n.NodeID = nodes.generateID(uint64(0))
	n.NodeName = nodeName
	n.API = API
	n.URL = url

	return n
}

func (nodes *Nodes) NewNodeRequest(body io.ReadCloser, id uint64) (*Node, error) {
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

	n := NewNode()

	n.NodeID = nodes.generateID(id)
	n.NodeName = request["nodeName"]
	n.API = request["API"]
	n.URL = request["node"]

	_, err = client.ParseHostURL(n.URL)

	if err != nil {
		return nil, err
	}

	return n, nil
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
		if n.NodeID == node.NodeID {
			nodes.Nodes = append(nodes.Nodes[:i], nodes.Nodes[i+1:]...)
		}
	}
}

func (nodes *Nodes) Find(node *Node) *Node {
	if node == nil {
		return nil
	}

	for _, n := range nodes.Nodes {
		if n.URL == node.URL {
			return n
		}
	}

	return nil
}

func (nodes *Nodes) FindById(id uint64) *Node {
	if id == 0 {
		return nil
	}

	for _, n := range nodes.Nodes {
		if n.NodeID == id {
			return n
		}
	}

	return nil
}

func (nodes *Nodes) generateID(currentNodeId uint64) uint64 {
	maxID := uint64(0)

	for _, node := range nodes.Nodes {
		if node.NodeID > maxID {
			maxID = node.NodeID
		}
	}

	if currentNodeId > maxID {
		maxID = currentNodeId
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
