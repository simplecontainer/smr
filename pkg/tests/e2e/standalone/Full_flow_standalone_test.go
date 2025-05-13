//go:build e2e
// +build e2e

package standalone_test

import (
	"context"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/tests/cli"
	"github.com/simplecontainer/smr/pkg/tests/engine"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestStandaloneNodeRestart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var nodes = make([]*node.Node, 0)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		cancel()

		for _, n := range nodes {
			n.Clean(t)
		}

		os.Exit(1)
	}()

	opts := node.DefaultNodeOptions("test", 1)
	opts.Image = flags.Image
	opts.Tag = flags.Tag

	if flags.BinaryPath != "" {
		opts.BinaryPath = flags.BinaryPath
	}

	n, err := node.New(t, opts)
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	nodes = append(nodes, n)
	defer n.Clean(t)

	t.Logf("starting standalone node with image %s:%s", flags.Image, flags.Tag)
	if err := n.Start(t); err != nil {
		t.Fatalf("failed to start node: %v", err)
	}

	cliopts := cli.DefaultCliOptions()
	if flags.BinaryPathCli != "" {
		cliopts.BinaryPath = flags.BinaryPathCli
	}

	cli, err := cli.New(t, cliopts)

	if err != nil {
		t.Fatalf("failed to create CLI: %v", err)
	}

	cli.Smrctl.Run(t, engine.NewStringCmd("context import %s -y", n.Context))

	cli.Smrctl.Run(t, engine.NewStringCmd("apply %s/%s/tests/minimal/definitions/Containers.yaml", cli.Root, flags.ExamplesDir))
	cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s --resource simplecontainer.io/v1/kind/containers/example/example-busybox-1", status.READY))
	cli.Smrctl.Run(t, engine.NewStringCmd("ps"))

	go func() {
		n.GetSmr().Run(t, engine.NewStringCmd("agent restart"))
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			cli.Smrctl.SetFailOnError(false)
			err = cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s", events.EVENT_CLUSTER_REPLAYED))

			if err == nil {
				break loop
			}

			time.Sleep(5 * time.Second)
		}
	}

	output, err := cli.Smrctl.RunAndCapture(t, engine.NewStringCmd("get containers/example/busybox"))

	if output == "resource not found" || err != nil {
		t.Fail()
	}

	n.GetSmr().Run(t, engine.NewStringCmd("agent drain"))
	cli.Smrctl.Run(t, engine.NewStringCmd("events --wait %s", events.EVENT_DRAIN_SUCCESS))

	n.Clean(t)

	t.Logf("test finished")
}
