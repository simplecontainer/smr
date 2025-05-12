package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/tests/engine"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/helpers"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/simplecontainer/smr/pkg/events/events"
)

type Node struct {
	Name       string
	Index      int
	Image      string
	Tag        string
	Join       bool
	Peer       string
	BinaryPath string
	Cleaned    bool

	IP      string
	Context string

	Ports Ports

	smr     *engine.Engine
	sudoSmr *engine.Engine

	mutex sync.Mutex
}

type Ports struct {
	Control int
	Etcd    int
	Overlay int
}

type Options struct {
	Name       string
	Index      int
	Image      string
	Tag        string
	Join       bool
	Peer       string
	BinaryPath string
}

func DefaultNodeOptions(name string, index int) Options {
	return Options{
		Name:  name,
		Index: index,
		Image: "default-image",
		Tag:   "latest",
		Join:  false,
		Peer:  "",
	}
}

func New(t *testing.T, opts Options) (*Node, error) {
	if opts.Name == "" {
		return nil, errors.New("node name cannot be empty")
	}

	nodeName := fmt.Sprintf("%s-%d", opts.Name, opts.Index)

	node := &Node{
		Name:  nodeName,
		Index: opts.Index,
		Image: opts.Image,
		Tag:   opts.Tag,
		Join:  opts.Join,
		Peer:  opts.Peer,
		Ports: Ports{
			Control: 1442 + opts.Index,
			Etcd:    2378 + opts.Index,
			Overlay: 9211 + opts.Index,
		},
	}

	root := helpers.GetProjectRoot(t)
	if err := os.Chdir(root); err != nil {
		return nil, fmt.Errorf("failed to change directory to project root: %w", err)
	}

	if opts.BinaryPath != "" {
		node.BinaryPath = filepath.Join(root, opts.BinaryPath)
	} else {
		panic("no binary provided")
	}

	engineOptions := engine.DefaultEngineOptions()
	engineOptions.Suffix = fmt.Sprintf("--home %s", flags.Home)

	node.smr = engine.NewEngineWithOptions(fmt.Sprintf("%s", node.BinaryPath), engineOptions)
	node.sudoSmr = engine.NewEngineWithOptions(fmt.Sprintf("sudo %s", node.BinaryPath), engineOptions)

	return node, nil
}

func (n *Node) Start(t *testing.T) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	t.Logf("[NODE] Starting node %s", n.Name)

	createCmd := n.buildCreateCommand()
	if err := n.smr.RunString(t, createCmd); err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	startCmd := fmt.Sprintf("node start --node %s -y", n.Name)
	t.Logf("[NODE] Starting node container with: %s", startCmd)
	if err := n.smr.RunString(t, startCmd); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	networkCmd := fmt.Sprintf("node networks --node %s --network bridge", n.Name)
	output, err := n.smr.RunAndCaptureString(t, networkCmd)
	if err != nil {
		return fmt.Errorf("failed to get node networks: %w", err)
	}

	n.IP = strings.TrimSpace(output)
	if n.IP == "" {
		return errors.New("failed to get IP address of node container")
	}
	t.Logf("[NODE] Node IP: %s", n.IP)

	contextCmd := fmt.Sprintf("agent export --node %s --api %s:%d", n.Name, n.IP, n.Ports.Control)
	output, err = n.sudoSmr.RunAndCaptureString(t, contextCmd)
	if err != nil {
		return fmt.Errorf("failed to export agent context: %w", err)
	}
	n.Context = strings.TrimSpace(output)

	agentCmd := fmt.Sprintf("agent start --node %s --raft https://%s:%d",
		n.Name, n.IP, n.Ports.Overlay)
	if err := n.sudoSmr.RunBackgroundString(t, agentCmd); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return n.WaitForEvent(t, events.EVENT_CLUSTER_READY, 60*time.Second)
}

func (n *Node) buildCreateCommand() string {
	baseCmd := fmt.Sprintf(
		"node create --node %s --image %s --tag %s --port.control 0.0.0.0:%d --port.etcd %d --port.overlay 0.0.0.0:%d",
		n.Name, n.Image, n.Tag, n.Ports.Control, n.Ports.Etcd, n.Ports.Overlay)

	if n.Join {
		return fmt.Sprintf("%s --join --peer %s", baseCmd, n.Peer)
	}
	return baseCmd
}

func (n *Node) WaitForEvent(t *testing.T, eventName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	errCh := make(chan error, 1)

	go func() {
		cmd := fmt.Sprintf("agent events --node %s --wait %s", n.Name, eventName)
		if err := n.sudoSmr.RunString(t, cmd); err != nil {
			errCh <- fmt.Errorf("error waiting for event %s: %w", eventName, err)
			return
		}
		close(done)
	}()

	select {
	case <-done:
		t.Logf("[NODE] Event %s occurred on node %s", eventName, n.Name)
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for event %s on node %s after %v", eventName, n.Name, timeout)
	}
}

func (n *Node) Import(t *testing.T, context string) error {
	if context == "" {
		return errors.New("cannot import empty context")
	}

	t.Logf("[NODE] Importing agent context for node %s", n.Name)
	cmd := fmt.Sprintf("agent import --node %s -y %s", n.Name, context)
	if err := n.sudoSmr.RunString(t, cmd); err != nil {
		return fmt.Errorf("failed to import agent context: %w", err)
	}
	return nil
}

func (n *Node) Clean(t *testing.T) {
	if !n.Cleaned {
		n.mutex.Lock()
		defer n.mutex.Unlock()

		n.Cleaned = true

		t.Logf("[NODE] Cleaning up node %s", n.Name)

		if err := n.sudoSmr.Stop(t); err != nil {
			t.Logf("[NODE] Error stopping agent for node %s: %v", n.Name, err)
		}

		cmd := fmt.Sprintf("node clean --node %s", n.Name)
		if err := n.smr.RunString(t, cmd); err != nil {
			t.Logf("[NODE] Error cleaning node %s: %v", n.Name, err)
		}

		home, err := os.UserHomeDir()

		if err != nil {
			t.Logf("[NODE] Error cleanin node directory %s: %v", n.Name, err)
			return
		}

		err = os.RemoveAll(fmt.Sprintf("%s/nodes/%s", home, n.Name))

		if err != nil {
			t.Logf("[NODE] Error cleanin node directory %s: %v", n.Name, err)
		}

		time.Sleep(5 * time.Second)
	}
}

func (n *Node) RunCommand(t *testing.T, command string) (string, error) {
	t.Logf("[NODE] Running command on node %s: %s", n.Name, command)
	output, err := n.sudoSmr.RunAndCaptureString(t, command)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}
	return output, nil
}

func (n *Node) GetContext() string {
	return n.Context
}

func (n *Node) GetIP() string {
	return n.IP
}

func (n *Node) GetSmr() *engine.Engine {
	return n.smr
}

func (n *Node) GetSudoSmr() *engine.Engine {
	return n.sudoSmr
}

func (n *Node) SetNodePorts(control, etcd, overlay int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.Ports.Control = control
	n.Ports.Etcd = etcd
	n.Ports.Overlay = overlay
}
