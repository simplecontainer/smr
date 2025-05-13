//go:build e2e
// +build e2e

package bootstrap_test

import (
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
)

func TestStandaloneNode(t *testing.T) {
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

	defer n.Clean(t)

	t.Logf("starting standalone node with image %s:%s", flags.Image, flags.Tag)
	if err := n.Start(t); err != nil {
		t.Fatalf("failed to start node: %v", err)
	}

	n.Clean(t)

	t.Logf("test finished")
}
