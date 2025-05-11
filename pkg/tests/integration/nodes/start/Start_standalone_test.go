//go:build integration
// +build integration

package start_test

import (
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
)

func TestStandaloneNode(t *testing.T) {
	flags.Parse()

	opts := node.DefaultNodeOptions("test", 1)
	opts.Image = flags.Image
	opts.Tag = flags.Tag
	if flags.BinaryDir != "" {
		opts.BinaryDir = flags.BinaryDir
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

	t.Logf("Test completed successfully")
}
