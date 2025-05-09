package start_test

import (
	"flag"
	"github.com/simplecontainer/smr/pkg/tests/integration/node"
	"testing"
)

var image string
var tag string

func init() {
	flag.StringVar(&image, "image", "smr", "Smr image")
	flag.StringVar(&tag, "tag", "latest", "Smr tag")
}

func TestStart(t *testing.T) {
	testingNode1 := node.New("testing-node", 1, image, tag)
	err := testingNode1.Start(t)

	if err != nil {
		testingNode1.Clean(t)
	}

	testingNode2 := node.New("testing-node", 1, image, tag)
	err = testingNode2.Start(t)

	if err != nil {
		testingNode2.Clean(t)
	}
}
