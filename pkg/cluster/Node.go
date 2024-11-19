package cluster

import (
	"encoding/json"
	"io"
	"strconv"
)

func NewNode(nodeID string, URL string) (*Node, error) {
	node := &Node{}
	number, err := strconv.ParseUint(nodeID, 10, 64)

	if err != nil {
		return nil, err
	}

	node.NodeID = number
	node.URL = URL

	return node, nil
}

func NewNodeRequest(body io.ReadCloser) (*Node, error) {
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

	node := &Node{}
	number, err := strconv.ParseUint(request["node"], 10, 64)

	if err != nil {
		return nil, err
	}

	node.NodeID = number
	node.URL = request["url"]

	return node, nil
}
