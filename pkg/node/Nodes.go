package node

import (
	"fmt"
	"sort"
)

func NewNodes() *Nodes {
	return &Nodes{}
}

func (nodes *Nodes) NewNode(nodeName string, url string, API string) *Node {
	n := NewNode()

	n.NodeID = nodes.GenerateID()
	n.NodeName = nodeName
	n.API = API
	n.URL = url

	return n
}

func (nodes *Nodes) Add(node *Node) {
	if node == nil {
		return
	}

	for _, n := range nodes.Nodes {
		if n.NodeName == node.NodeName {
			return
		}
	}

	nodes.Nodes = append(nodes.Nodes, node)

	sort.Slice(nodes.Nodes, func(i, j int) bool {
		return nodes.Nodes[i].NodeID < nodes.Nodes[j].NodeID
	})
}

func (nodes *Nodes) AddOrUpdate(node *Node) {
	if node == nil {
		return
	}

	for i, n := range nodes.Nodes {
		if n.NodeName == node.NodeName {
			nodes.Nodes[i] = node
			return
		}
	}

	nodes.Nodes = append(nodes.Nodes, node)

	sort.Slice(nodes.Nodes, func(i, j int) bool {
		return nodes.Nodes[i].NodeID < nodes.Nodes[j].NodeID
	})
}

func (nodes *Nodes) Remove(node *Node) {
	if node == nil {
		return
	}

	for i, n := range nodes.Nodes {
		if n.NodeID == node.NodeID {
			fmt.Println("Removed node", node)
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

func (nodes *Nodes) GenerateID() uint64 {
	usedIDs := make(map[uint64]struct{})

	for _, node := range nodes.Nodes {
		usedIDs[node.NodeID] = struct{}{}
	}

	for i := uint64(1); ; i++ {
		if _, exists := usedIDs[i]; !exists {
			return i
		}
	}
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
