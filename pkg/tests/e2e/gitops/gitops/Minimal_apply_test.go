//go:build e2e
// +build e2e

package gitops

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/tests/cli"
	"github.com/simplecontainer/smr/pkg/tests/engine"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
)

func TestStandaloneNodeMinimalContainer(t *testing.T) {
	nm := node.NewNodeManager()
	nm.SetupTestCleanup(t)

	opts := node.DefaultNodeOptions("test", 1)
	opts.Image = flags.Image
	opts.Tag = flags.Tag
	if flags.BinaryPath != "" {
		opts.BinaryPath = flags.BinaryPath
	}

	n, err := nm.CreateAndStartNodeWithOptions(
		t,
		func(t *testing.T, options interface{}) (node.NodeCleaner, error) {
			nodeOpts, ok := options.(node.Options)
			if !ok {
				return nil, fmt.Errorf("invalid options type")
			}
			return node.New(t, nodeOpts)
		},
		opts,
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

	// Type assertion to get concrete node type
	concreteNode, ok := n.(*node.Node)
	if !ok {
		t.Fatalf("invalid node type returned")
	}

	cliopts := cli.DefaultCliOptions()
	if flags.BinaryPathCli != "" {
		cliopts.BinaryPath = flags.BinaryPathCli
	}

	cli, err := cli.New(t, cliopts)
	if nm.HandleError(t, err, "failed to create CLI") {
		t.FailNow()
	}

	// Run commands with automatic error handling and cleanup
	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("context import %s -y", concreteNode.Context))
	}, "context import")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("apply %s/%s/tests/gitops-apps/definitions/gitops-plain.yaml",
			cli.Root, flags.ExamplesDir))
	}, "apply gitops-plain.yaml")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("sync gitops/examples/plain-manual"))
	}, "sync gitops example")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s --resource simplecontainer.io/v1/kind/containers/example/example-busybox-1",
			status.READY))
	}, "wait for container ready")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("ps"))
	}, "ps command")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("remove simplecontainer.io/v1/kind/containers/example/busybox"))
	}, "remove container")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s --resource simplecontainer.io/v1/kind/containers/example/example-busybox-1",
			events.EVENT_DELETED))
	}, "wait for container deleted")

	t.Logf("test finished")
}
