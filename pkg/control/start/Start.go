package start

import (
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/control/generic"
	"github.com/simplecontainer/smr/pkg/control/registry"
	"github.com/simplecontainer/smr/pkg/logger"
)

type Command struct {
	*generic.GenericCommand
}

func init() {
	registry.RegisterCommandType("start", func() icontrol.Command {
		return &Command{
			GenericCommand: &generic.GenericCommand{},
		}
	})
}

func NewStartCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: generic.NewCommand("start", options),
	}
}

func (c *Command) Node(api iapi.Api, params map[string]string) error {
	return nil
}

func (c *Command) Agent(api iapi.Api, params map[string]string) error {
	logger.Log.Info("cluster started with success")
	return nil
}
