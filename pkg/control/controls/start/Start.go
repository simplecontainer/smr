package start

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/control/controls"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Command struct {
	*controls.GenericCommand
}

func NewStartCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: controls.NewCommand("start", options),
	}
}

func (c *Command) Node(mgr *manager.Manager, params map[string]string) error {
	fmt.Println("Executing start on node")
	return nil
}

func (c *Command) Agent(cli *client.Client, params map[string]string) error {
	fmt.Println("Executing start on agent")
	return nil
}
