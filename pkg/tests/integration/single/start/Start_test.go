package start

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/tests/smr"
	"testing"
)

func TestStart(t *testing.T) {
	engine := smr.NewEngine("./smr-linux-amd64/smr")
	engine.Create(t)
	engine.Start(t)

	fmt.Println("Started the daemon")
}
