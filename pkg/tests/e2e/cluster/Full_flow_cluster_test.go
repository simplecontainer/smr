//go:build e2e
// +build e2e

package cluster

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/tests/cli"
	"github.com/simplecontainer/smr/pkg/tests/engine"
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

	cliopts := cli.DefaultCliOptions()
	if flags.BinaryPathCli != "" {
		cliopts.BinaryPath = flags.BinaryPathCli
	}

	cli, err := cli.New(t, cliopts)
	if nm.HandleError(t, err, "failed to create CLI") {
		t.FailNow()
	}

	nm.RunCommand(t, func() error {
		return follower.GetSmr().Run(t, engine.NewStringCmd("agent drain"))
	}, "remove container")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("context import %s -y", follower.GetContext()))
	}, "context import")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s", events.EVENT_DRAIN_SUCCESS))
	}, "wait for container deleted")

	nm.RunCommand(t, func() error {
		return leader.GetSmr().Run(t, engine.NewStringCmd("agent drain"))
	}, "remove container")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("context import %s -y", leader.GetContext()))
	}, "context import")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s", events.EVENT_DRAIN_SUCCESS))
	}, "wait for container deleted")

	t.Logf("test finished")
}
