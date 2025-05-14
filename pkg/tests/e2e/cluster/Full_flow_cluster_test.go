//go:build e2e
// +build e2e

package cluster

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
)

func TestClusterMode(t *testing.T) {
	nm := node.NewNodeManager()
	nm.SetupTestCleanup(t)

	leaderOpts := node.DefaultNodeOptions("leader", 1)
	leaderOpts.Image = flags.Image
	leaderOpts.Tag = flags.Tag
	if flags.BinaryPath != "" {
		leaderOpts.BinaryPath = flags.BinaryPath
	}

	leader, err := nm.CreateAndStartNodeWithOptions(
		t,
		func(t *testing.T, options interface{}) (node.NodeCleaner, error) {
			nodeOpts, ok := options.(node.Options)
			if !ok {
				return nil, fmt.Errorf("invalid options type")
			}
			return node.New(t, nodeOpts)
		},
		leaderOpts,
		func(nodeTmp node.NodeCleaner, t *testing.T) error {
			n, ok := nodeTmp.(*node.Node)
			if !ok {
				return fmt.Errorf("invalid node type")
			}
			t.Logf("starting standalone node with image %s:%s", flags.Image, flags.Tag)

			return n.Start(t)
		},
	)

	if nm.HandleError(t, err, "failed to create or start node") {
		t.FailNow()
	}

	leader = leader.(*node.Node)

	followerOpts := node.DefaultNodeOptions("follower", 2)
	followerOpts.Image = flags.Image
	followerOpts.Tag = flags.Tag
	followerOpts.Join = true
	followerOpts.Peer = fmt.Sprintf("https://%s:%d", leader.GetIP(), leader.GetPorts().Control)
	if flags.BinaryPath != "" {
		followerOpts.BinaryPath = flags.BinaryPath
	}

	follower, err := nm.CreateAndStartNodeWithOptions(
		t,
		func(t *testing.T, options interface{}) (node.NodeCleaner, error) {
			nodeOpts, ok := options.(node.Options)
			if !ok {
				return nil, fmt.Errorf("invalid options type")
			}
			return node.New(t, nodeOpts)
		},
		followerOpts,
		func(nodeTmp node.NodeCleaner, t *testing.T) error {
			n, ok := nodeTmp.(*node.Node)
			if !ok {
				return fmt.Errorf("invalid node type")
			}
			t.Logf("starting standalone node with image %s:%s", flags.Image, flags.Tag)

			err = n.Import(t, leader.GetContext())

			return n.Start(t)
		},
	)

	if nm.HandleError(t, err, "failed to create or start node") {
		t.FailNow()
	}

	fmt.Println(leader.GetIP())
	fmt.Println(follower.GetIP())

	t.Logf("test finished")
}
