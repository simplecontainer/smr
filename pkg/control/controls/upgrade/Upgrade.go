package upgrade

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/control/controls"
)

type Command struct {
	*controls.GenericCommand
}

func NewUpgradeCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: controls.NewCommand("upgrade", options),
	}
}

func (c *Command) Node(api *api.Api, params map[string]string) error {
	fmt.Println("Executing upgrade on node")
	return nil
}

func (c *Command) Agent(cli *client.Client, params map[string]string) error {
	fmt.Println("Executing upgrade on agent")
	return nil
}
