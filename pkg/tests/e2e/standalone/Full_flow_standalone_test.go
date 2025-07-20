//go:build e2e
// +build e2e

package standalone_test

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/tests/cli"
	"github.com/simplecontainer/smr/pkg/tests/engine"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
	"time"
)

func TestStandaloneMode(t *testing.T) {
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
		return cli.Smrctl.Run(t, engine.NewStringCmd("apply %s/%s/tests/minimal/definitions/Containers.yaml",
			cli.Root, flags.ExamplesDir))
	}, "apply tests/minimal/definitions/Containers.yaml")

	cli.Smrctl.RunAndCapture(t, engine.NewStringCmd("events --wait %s --resource simplecontainer.io/v1/kind/containers/example/example-busybox-1",
		status.READY))

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("ps"))
	}, "ps command")

	go func() {
		nm.RunCommand(t, func() error {
			return concreteNode.GetSmr().Run(t, engine.NewStringCmd("agent restart"))
		}, "agent restart")
	}()

	nm.RunCommand(t, func() error {
		for {
			cli.Smrctl.SetFailOnError(false)
			err = cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s", events.EVENT_CLUSTER_STARTED))

			if err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}

		return nil
	}, "events wait for replay")

	output, err := cli.Smrctl.RunAndCapture(t, engine.NewStringCmd("get containers/example/busybox"))

	if output == "resource not found" || err != nil {
		t.Fail()
	}

	nm.RunCommand(t, func() error {
		return concreteNode.GetSmr().Run(t, engine.NewStringCmd("agent drain"))
	}, "remove container")

	nm.RunCommand(t, func() error {
		return cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s", events.EVENT_DRAIN_SUCCESS))
	}, "wait for container deleted")

	t.Logf("test finished")
}
