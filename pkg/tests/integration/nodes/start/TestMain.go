package start

import (
	"flag"
	"os"
	"testing"

	_ "github.com/simplecontainer/smr/pkg/tests/flags"
)

func TestMain(m *testing.M) {
	flag.Parse()

	code := m.Run()

	os.Exit(code)
}
