//go:build integration
// +build integration

package start_test

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
)

func TestClusterMode(t *testing.T) {
	flags.Parse()

	leaderOpts := node.DefaultNodeOptions("leader", 1)
	leaderOpts.Image = flags.Image
	leaderOpts.Tag = flags.Tag
	if flags.BinaryDir != "" {
		leaderOpts.BinaryDir = flags.BinaryDir
	}

	leader, err := node.New(t, leaderOpts)
	if err != nil {
		t.Fatalf("failed to create leader node: %v", err)
	}

	defer leader.Clean(t)

	t.Logf("starting leader node with image %s:%s", flags.Image, flags.Tag)
	if err := leader.Start(t); err != nil {
		t.Fatalf("failed to start leader node: %v", err)
	}

	leaderCtx := leader.GetContext()
	leaderIP := leader.GetIP()

	followerOpts := node.DefaultNodeOptions("follower", 2)
	followerOpts.Image = flags.Image
	followerOpts.Tag = flags.Tag
	followerOpts.Join = true
	followerOpts.Peer = fmt.Sprintf("https://%s:%d", leaderIP, leader.Ports.Control)
	if flags.BinaryDir != "" {
		followerOpts.BinaryDir = flags.BinaryDir
	}

	follower, err := node.New(t, followerOpts)
	if err != nil {
		t.Fatalf("failed to create follower node: %v", err)
	}

	defer follower.Clean(t)

	t.Logf("importing leader context to follower")
	if err := follower.Import(t, leaderCtx); err != nil {
		t.Fatalf("failed to import context: %v", err)
	}

	t.Logf("starting follower node with image %s:%s", flags.Image, flags.Tag)
	if err := follower.Start(t); err != nil {
		t.Fatalf("failed to start follower node: %v", err)
	}

	t.Logf("test completed successfully")
}
