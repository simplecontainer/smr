package upgrade

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/control/generic"
	"github.com/simplecontainer/smr/pkg/control/registry"
)

type Command struct {
	*generic.GenericCommand
}

func init() {
	registry.RegisterCommandType("upgrade", func() icontrol.Command {
		return &Command{
			GenericCommand: &generic.GenericCommand{},
		}
	})
}

func NewUpgradeCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: generic.NewCommand("upgrade", options),
	}
}

func (c *Command) Node(api iapi.Api, params map[string]string) error {
	fmt.Println("Executing upgrade on node")
	return nil
}

func (c *Command) Agent(api iapi.Api, params map[string]string) error {
	fmt.Println("Executing upgrade on agent")
	return nil
}
