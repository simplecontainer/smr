//go:build integration
// +build integration

package start_test

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
	"time"
)

func TestClusterMode(t *testing.T) {
	gofakeit.Seed(0)

	leaderOpts := node.DefaultNodeOptions(gofakeit.Username(), 1)
	leaderOpts.Image = flags.Image
	leaderOpts.Tag = flags.Tag
	if flags.BinaryPath != "" {
		leaderOpts.BinaryPath = flags.BinaryPath
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

	followerOpts := node.DefaultNodeOptions(gofakeit.Username(), 2)
	followerOpts.Image = flags.Image
	followerOpts.Tag = flags.Tag
	followerOpts.Join = true
	followerOpts.Peer = fmt.Sprintf("https://%s:%d", leaderIP, leader.Ports.Control)
	if flags.BinaryPath != "" {
		followerOpts.BinaryPath = flags.BinaryPath
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
	time.Sleep(5 * time.Second)
}
