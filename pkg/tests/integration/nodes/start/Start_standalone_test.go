//go:build integration
// +build integration

package start_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simplecontainer/smr/pkg/tests/flags"
	"github.com/simplecontainer/smr/pkg/tests/node"
	"testing"
)

func TestStandaloneNode(t *testing.T) {
	gofakeit.Seed(0)

	opts := node.DefaultNodeOptions(gofakeit.Username(), 1)
	opts.Image = flags.Image
	opts.Tag = flags.Tag
	if flags.BinaryPath != "" {
		opts.BinaryDir = flags.BinaryPath
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
