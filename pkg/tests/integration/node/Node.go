package node

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/tests/helpers"
	"github.com/simplecontainer/smr/pkg/tests/smr"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type Node struct {
	Name  string
	Index int
	Image string
	Tag   string

	Engine *smr.Engine
	Agent  *smr.Engine
}

func New(name string, index int, image string, tag string) *Node {
	name = fmt.Sprintf("%s-%d", name, index)

	return &Node{
		Name:  name,
		Index: index,
		Image: image,
		Tag:   tag,
	}
}

func (n *Node) Start(t *testing.T) error {
	root := helpers.GetProjectRoot(t)

	err := os.Chdir(root)
	if err != nil {
		return err
	}

	var absolutePathBinary string
	absolutePathBinary, err = filepath.Abs(filepath.Join(root, "smr-linux-amd64/smr"))
	if err != nil {
		return err
	}

	n.Engine = smr.NewEngine(absolutePathBinary)
	n.Agent = smr.NewEngine("sudo")

	t.Log("[TEST]: creating node")
	n.Engine.Run(t, fmt.Sprintf("node create --node %s --image %s --tag %s --port.control 0.0.0.0:%d --port.etcd %d --port.overlay 0.0.0.0:%d",
		n.Name, n.Image, n.Tag, 1442+n.Index, 2377+n.Index, 9211+n.Index))

	t.Log("[TEST]: starting node")
	n.Engine.Run(t, fmt.Sprintf("node start --node %s -y", n.Name))

	IP := strings.TrimSpace(n.Engine.RunAndCapture(t, fmt.Sprintf("node networks --node %s --network bridge", n.Name)))

	if IP == "" {
		return errors.New("[TEST] failed to get IP address of node container")
	}

	t.Log("[TEST]: starting agent and listening to events")
	n.Agent.RunBackground(t, fmt.Sprintf("%s agent start --node %s --raft https://%s:%d", absolutePathBinary, n.Name, IP, 9211+n.Index))

	done := make(chan bool)
	timeout := make(chan bool)

	go func() {
		t.Log("[TEST]: waiting for cluster started event")
		n.Agent.Run(t, fmt.Sprintf("%s agent events --node %s --wait %s", absolutePathBinary, n.Name, events.EVENT_CLUSTER_STARTED))
		t.Log("[TEST]: cluster started event received")
		done <- true
	}()

	go func() {
		time.Sleep(60 * time.Second)
		timeout <- true
	}()

	time.Sleep(10 * time.Second)

	select {
	case <-done:
		t.Log("[TEST]: node started and running successfully")
		return nil
	case <-timeout:
		return errors.New("test timed out waiting for cluster to start")
	}
}

func (n *Node) Clean(t *testing.T) {
	n.Engine.Run(t, fmt.Sprintf("node clean --node %s", n.Name))
	n.Agent.Stop(t)
}
