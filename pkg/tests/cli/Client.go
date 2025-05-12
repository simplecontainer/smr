package cli

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/tests/engine"
	"github.com/simplecontainer/smr/pkg/tests/helpers"
	"os"
	"testing"
)

type Cli struct {
	Smrctl     *engine.Engine
	BinaryPath string
	Root       string
}

type Options struct {
	BinaryPath string
}

func DefaultCliOptions() Options {
	return Options{}
}

func New(t *testing.T, opts Options) (*Cli, error) {
	cli := &Cli{}

	root := helpers.GetProjectRoot(t)
	if err := os.Chdir(root); err != nil {
		return nil, fmt.Errorf("failed to change directory to project root: %w", err)
	}

	cli.Root = root

	if opts.BinaryPath != "" {
		cli.BinaryPath = opts.BinaryPath
	} else {
		panic("no binary provided")
	}

	cliOptions := engine.DefaultEngineOptions()

	cli.Smrctl = engine.NewEngineWithOptions(fmt.Sprintf("%s", cli.BinaryPath), cliOptions)

	return cli, nil
}
