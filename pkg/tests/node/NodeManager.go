package node

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
)

type NodeCleaner interface {
	Clean(*testing.T) error
	GetPorts() Ports
	GetIP() string
	GetContext() string
	Import(t *testing.T, context string) error
}

type NodeManager struct {
	nodes     []NodeCleaner
	mutex     sync.Mutex
	setupDone bool
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes:     make([]NodeCleaner, 0),
		setupDone: false,
	}
}

func (nm *NodeManager) Add(t *testing.T, node NodeCleaner) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.nodes = append(nm.nodes, node)

	if !nm.setupDone {
		nm.setupSignalHandlers(t)
		nm.setupDone = true
	}
}

func (nm *NodeManager) CreateNode(t *testing.T, creatorFn func(*testing.T) (NodeCleaner, error)) (NodeCleaner, error) {
	node, err := creatorFn(t)
	if err != nil {
		return nil, err
	}

	nm.Add(t, node)
	return node, nil
}

func (nm *NodeManager) CreateNodeWithOptions(t *testing.T, creatorFn func(*testing.T, interface{}) (NodeCleaner, error), opts interface{}) (NodeCleaner, error) {
	node, err := creatorFn(t, opts)
	if err != nil {
		return nil, err
	}

	nm.Add(t, node)
	return node, nil
}

func (nm *NodeManager) CreateAndStartNode(t *testing.T,
	creatorFn func(*testing.T) (NodeCleaner, error),
	startFn func(NodeCleaner, *testing.T) error) (NodeCleaner, error) {

	node, err := nm.CreateNode(t, creatorFn)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	if err := startFn(node, t); err != nil {

		if cleanErr := node.Clean(t); cleanErr != nil {
			t.Logf("Error cleaning node after start failure: %v", cleanErr)
		}

		return nil, fmt.Errorf("failed to start node: %w", err)
	}

	return node, nil
}

func (nm *NodeManager) CreateAndStartNodeWithOptions(t *testing.T,
	creatorFn func(*testing.T, interface{}) (NodeCleaner, error),
	opts interface{},
	startFn func(NodeCleaner, *testing.T) error) (NodeCleaner, error) {

	node, err := nm.CreateNodeWithOptions(t, creatorFn, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	if err := startFn(node, t); err != nil {
		if cleanErr := node.Clean(t); cleanErr != nil {
			t.Logf("Error cleaning node after start failure: %v", cleanErr)
		}

		return nil, fmt.Errorf("failed to start node: %w", err)
	}

	return node, nil
}

func (nm *NodeManager) setupSignalHandlers(t *testing.T) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("Received termination signal, cleaning up nodes")
		nm.CleanAll(t)
		os.Exit(1)
	}()
}

func (nm *NodeManager) CleanAll(t *testing.T) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	for _, node := range nm.nodes {
		if err := node.Clean(t); err != nil {
			fmt.Printf("Error cleaning up node: %v\n", err)
		}
	}

	nm.nodes = nm.nodes[:0]
}

func (nm *NodeManager) HandleError(t *testing.T, err error, message string) bool {
	if err != nil {
		t.Logf("%s: %v", message, err)
		nm.CleanAll(t)
		return true
	}
	return false
}

func (nm *NodeManager) RunCommand(t *testing.T, runFn func() error, cmdDescription string) {
	err := runFn()
	if nm.HandleError(t, err, fmt.Sprintf("Command failed: %s", cmdDescription)) {
		t.Fatalf("Failed to execute: %s", cmdDescription)
	}
}

func (nm *NodeManager) SetupTestCleanup(t *testing.T) {
	t.Cleanup(func() {
		t.Log("Test finished, cleaning up nodes")
		nm.CleanAll(t)
	})
}
