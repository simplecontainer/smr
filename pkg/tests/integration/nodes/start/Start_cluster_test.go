package start_test

import (
	"testing"
	"time"

	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
)

func TestCluster(t *testing.T) {
	timeout := time.Duration(flags.Timeout) * time.Second

	leaderOpts := node.DefaultNodeOptions("leader", 1)
	leaderOpts.Image = flags.Image
	leaderOpts.Tag = flags.Tag
	if flags.BinaryDir != "" {
		leaderOpts.BinaryDir = flags.BinaryDir
	}

	leader, err := node.New(t, leaderOpts)
	if err != nil {
		t.Fatalf("Failed to create leader node: %v", err)
	}

	if flags.Cleanup {
		defer leader.Clean(t)
	}

	t.Logf("Starting leader node with image %s:%s", flags.Image, flags.Tag)
	if err := leader.Start(t); err != nil {
		t.Fatalf("Failed to start leader node: %v", err)
	}

	leaderCtx := leader.GetContext()
	leaderIP := leader.GetIP()

	followerOpts := node.DefaultNodeOptions("follower", 2)
	followerOpts.Image = flags.Image
	followerOpts.Tag = flags.Tag
	followerOpts.Join = true
	followerOpts.Peer = leaderIP
	if flags.BinaryDir != "" {
		followerOpts.BinaryDir = flags.BinaryDir
	}

	follower, err := node.New(t, followerOpts)
	if err != nil {
		t.Fatalf("Failed to create follower node: %v", err)
	}

	if flags.Cleanup {
		defer follower.Clean(t)
	}

	t.Logf("Starting follower node with image %s:%s", flags.Image, flags.Tag)
	if err := follower.Start(t); err != nil {
		t.Fatalf("Failed to start follower node: %v", err)
	}

	t.Logf("Importing leader context to follower")
	if err := follower.Import(t, leaderCtx); err != nil {
		t.Fatalf("Failed to import context: %v", err)
	}

	t.Logf("Waiting for follower to join cluster")
	if err := follower.WaitForEvent(t, "cluster.joined", timeout); err != nil {
		t.Fatalf("Follower failed to join cluster: %v", err)
	}

	t.Logf("Getting cluster status from leader")
	output, err := leader.RunCommand(t, "agent cluster status")
	if err != nil {
		t.Fatalf("Failed to get cluster status: %v", err)
	}

	t.Logf("Cluster status: %s", output)
	t.Logf("Test completed successfully")
}
