package node

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type ClusterOptions struct {
	BaseName      string
	Image         string
	Tag           string
	LeaderCount   int
	FollowerCount int
	BinaryPath    string
	StartTimeout  time.Duration
}

func DefaultClusterOptions() ClusterOptions {
	return ClusterOptions{
		BaseName:      "node",
		Image:         "default-image",
		Tag:           "latest",
		LeaderCount:   1,
		FollowerCount: 2,
		StartTimeout:  60 * time.Second,
	}
}

type Cluster struct {
	Options   ClusterOptions
	Leaders   []*Node
	Followers []*Node

	mutex sync.RWMutex
}

func NewCluster(opts ClusterOptions) *Cluster {
	if opts.StartTimeout == 0 {
		opts.StartTimeout = 60 * time.Second
	}

	return &Cluster{
		Options:   opts,
		Leaders:   make([]*Node, 0, opts.LeaderCount),
		Followers: make([]*Node, 0, opts.FollowerCount),
	}
}

func (c *Cluster) Setup(t *testing.T) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	t.Logf("[CLUSTER] Setting up cluster with %d leaders and %d followers",
		c.Options.LeaderCount, c.Options.FollowerCount)

	for i := 0; i < c.Options.LeaderCount; i++ {
		nodeOpts := Options{
			Name:       fmt.Sprintf("%s-leader", c.Options.BaseName),
			Index:      i + 1,
			Image:      c.Options.Image,
			Tag:        c.Options.Tag,
			Join:       i > 0, // First node doesn't join, others do
			BinaryPath: c.Options.BinaryPath,
		}

		// If this is not the first leader, set peer to first leader
		if i > 0 && len(c.Leaders) > 0 {
			nodeOpts.Peer = c.Leaders[0].GetIP()
		}

		node, err := New(t, nodeOpts)
		if err != nil {
			return fmt.Errorf("failed to create leader node %d: %w", i+1, err)
		}

		if err := node.Start(t); err != nil {
			return fmt.Errorf("failed to start leader node %d: %w", i+1, err)
		}

		c.Leaders = append(c.Leaders, node)

		// Wait a bit between leader startups
		if i < c.Options.LeaderCount-1 {
			time.Sleep(5 * time.Second)
		}
	}

	if len(c.Leaders) == 0 {
		return fmt.Errorf("no leader nodes created")
	}

	// Now create and start follower nodes
	firstLeaderIP := c.Leaders[0].GetIP()
	firstLeaderCtx := c.Leaders[0].GetContext()

	for i := 0; i < c.Options.FollowerCount; i++ {
		nodeOpts := Options{
			Name:       fmt.Sprintf("%s-follower", c.Options.BaseName),
			Index:      i + 1,
			Image:      c.Options.Image,
			Tag:        c.Options.Tag,
			Join:       true,
			Peer:       firstLeaderIP,
			BinaryPath: c.Options.BinaryPath,
		}

		node, err := New(t, nodeOpts)
		if err != nil {
			return fmt.Errorf("failed to create follower node %d: %w", i+1, err)
		}

		if err := node.Start(t); err != nil {
			return fmt.Errorf("failed to start follower node %d: %w", i+1, err)
		}

		// Import leader context
		if err := node.Import(t, firstLeaderCtx); err != nil {
			return fmt.Errorf("failed to import context to follower %d: %w", i+1, err)
		}

		c.Followers = append(c.Followers, node)

		// Wait a bit between follower startups
		if i < c.Options.FollowerCount-1 {
			time.Sleep(3 * time.Second)
		}
	}

	return c.waitForClusterReady(t, c.Options.StartTimeout)
}

// waitForClusterReady waits for the entire cluster to be ready
func (c *Cluster) waitForClusterReady(t *testing.T, timeout time.Duration) error {
	if len(c.Leaders) == 0 {
		return fmt.Errorf("no leader nodes in cluster")
	}

	// Use the first leader to check cluster status
	leader := c.Leaders[0]

	expectedNodes := c.Options.LeaderCount + c.Options.FollowerCount
	start := time.Now()

	t.Logf("[CLUSTER] Waiting for all %d nodes to join the cluster", expectedNodes)

	for {
		if time.Since(start) > timeout {
			return fmt.Errorf("timed out waiting for cluster to be ready after %v", timeout)
		}

		output, err := leader.RunCommand(t, "agent cluster members")
		if err != nil {
			t.Logf("[CLUSTER] Error getting cluster members: %v, retrying...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Count active nodes in output (this is a simplified check, adjust based on actual output format)
		// In a real implementation, you might parse JSON output or count specific strings
		nodeCount := countActiveNodes(output)

		if nodeCount >= expectedNodes {
			t.Logf("[CLUSTER] All %d nodes have joined the cluster", expectedNodes)
			return nil
		}

		t.Logf("[CLUSTER] Found %d/%d nodes in cluster, waiting...", nodeCount, expectedNodes)
		time.Sleep(2 * time.Second)
	}
}

// countActiveNodes is a placeholder for actual output parsing logic
// In a real implementation, parse the actual output format correctly
func countActiveNodes(output string) int {
	// This is a simplistic implementation - replace with actual parsing logic
	// based on the format of your output

	// For example, if each node is on a separate line:
	// return len(strings.Split(output, "\n"))

	// This is a placeholder that returns 1
	return 1
}

// GetAllNodes returns all nodes in the cluster
func (c *Cluster) GetAllNodes() []*Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	allNodes := make([]*Node, 0, len(c.Leaders)+len(c.Followers))
	allNodes = append(allNodes, c.Leaders...)
	allNodes = append(allNodes, c.Followers...)

	return allNodes
}

func (c *Cluster) GetLeader() *Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.Leaders) > 0 {
		return c.Leaders[0]
	}
	return nil
}

// Clean stops and cleans all nodes in the cluster
func (c *Cluster) Clean(t *testing.T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	t.Log("[CLUSTER] Cleaning up all nodes")

	// Clean followers first, then leaders (reverse order of creation)
	for _, node := range c.Followers {
		node.Clean(t)
	}

	// Clean leaders in reverse order
	for i := len(c.Leaders) - 1; i >= 0; i-- {
		c.Leaders[i].Clean(t)
	}

	c.Leaders = nil
	c.Followers = nil
}

// RunCommandOnAll runs a command on all nodes and returns a map of outputs
func (c *Cluster) RunCommandOnAll(t *testing.T, command string) map[string]string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	results := make(map[string]string)
	allNodes := c.GetAllNodes()

	for _, node := range allNodes {
		output, err := node.RunCommand(t, command)
		if err != nil {
			t.Logf("[CLUSTER] Error running command on %s: %v", node.Name, err)
			results[node.Name] = fmt.Sprintf("ERROR: %v", err)
		} else {
			results[node.Name] = output
		}
	}

	return results
}
