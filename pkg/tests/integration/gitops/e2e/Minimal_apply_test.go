//go:build integration
// +build integration

package e2e

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/tests/cli"
	"github.com/simplecontainer/smr/pkg/tests/engine"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
	"time"
)

func TestStandaloneNodeMinimalContainer(t *testing.T) {
	opts := node.DefaultNodeOptions("test", 1)
	opts.Image = flags.Image
	opts.Tag = flags.Tag
	if flags.BinaryPath != "" {
		opts.BinaryPath = flags.BinaryPath
	}

	n, err := node.New(t, opts)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	defer n.Clean(t)

	t.Logf("Starting standalone node with image %s:%s", flags.Image, flags.Tag)
	if err := n.Start(t); err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}

	cliopts := cli.DefaultCliOptions()
	if flags.BinaryPathCli != "" {
		cliopts.BinaryPath = flags.BinaryPathCli
	}

	cli, err := cli.New(t, cliopts)

	if err != nil {
		t.Fatalf("Failed to create CLI: %v", err)
	}

	cli.Smrctl.Run(t, engine.NewStringCmd("context import %s -y", n.Context))

	cli.Smrctl.Run(t, engine.NewStringCmd("apply %s/%s/tests/gitops-apps/definitions/gitops-plain.yaml", cli.Root, flags.ExamplesDir))
	cli.Smrctl.Run(t, engine.NewStringCmd("sync  gitops/examples/plain-manual"))
	cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s --resource simplecontainer.io/v1/kind/containers/example/example-busybox-1", status.READY))
	cli.Smrctl.Run(t, engine.NewStringCmd("ps"))

	time.Sleep(5 * time.Second)

	cli.Smrctl.Run(t, engine.NewStringCmd("remove simplecontainer.io/v1/kind/containers/example/busybox"))
	cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s --resource simplecontainer.io/v1/kind/containers/example/busybox", events.EVENT_DELETED))

	t.Logf("Test finished")
}
