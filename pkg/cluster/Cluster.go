package cluster

import (
	"encoding/json"
	"errors"
	"github.com/simplecontainer/smr/pkg/configuration"
	"io"
	"strconv"
	"strings"
)

func New(body io.ReadCloser) (*Cluster, error) {
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

	cluster := strings.Split(request["cluster"], ",")

	for i, c := range cluster {
		if c == "" {
			cluster = append(cluster[:i], cluster[i+1:]...)
		}
	}

	if len(cluster) == 0 {
		return nil, errors.New("cluster is empty")
	}

	return &Cluster{
		Node:       node,
		Cluster:    strings.Split(request["cluster"], ","),
		EtcdClient: NewEtcdClient(),
	}, nil
}

func Restore(config *configuration.Configuration) (*Cluster, error) {
	cluster := make([]string, 0)
	for i, c := range config.KVStore.Cluster {
		if c == "" {
			cluster = append(cluster[:i], cluster[i+1:]...)
		}
	}

	if len(cluster) == 0 {
		return nil, errors.New("cluster is empty")
	}

	return &Cluster{
		Node: &Node{
			NodeID: config.KVStore.Node,
			URL:    config.KVStore.URL,
		},
		Cluster: cluster,
	}, nil
}

func (c *Cluster) Start(body io.ReadCloser) {}

func (c *Cluster) Add(node *Node) {
	for _, url := range c.Cluster {
		if url == node.URL {
			return
		}
	}

	c.Cluster = append(c.Cluster, node.URL)
}
func (c *Cluster) Remove(node *Node) {
	for i, url := range c.Cluster {
		if url == node.URL {
			c.Cluster = append(c.Cluster[:i], c.Cluster[i+1:]...)
		}
	}
}
