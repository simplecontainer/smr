package restart

import (
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/control/generic"
	"github.com/simplecontainer/smr/pkg/control/registry"
	"github.com/simplecontainer/smr/pkg/control/start"
	"github.com/simplecontainer/smr/pkg/engine/agent"
	"github.com/simplecontainer/smr/pkg/engine/node"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/spf13/viper"
	"time"
)

type Command struct {
	*generic.GenericCommand
}

func init() {
	registry.RegisterCommandType("restart", func() icontrol.Command {
		return &Command{
			GenericCommand: &generic.GenericCommand{},
		}
	})
}

func NewRestartCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: generic.NewCommand("restart", options),
	}
}

func (c *Command) Node(api iapi.Api, params map[string]string) error {
	return nil
}

func (c *Command) Agent(api iapi.Api, params map[string]string) error {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		return err
	}

	node.Clean()

	viper.Set("y", true)
	node.Start("/opt/smr/smr", "start")

	parsed, err := helpers.EnforceHTTPS(viper.GetString("raft"))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	b := control.NewCommandBatch()
	b.AddCommand(start.NewStartCommand(map[string]string{
		"raft":    parsed.String(),
		"cidr":    conf.Flannel.CIDR,
		"backend": conf.Flannel.Backend,
	}))

	time.Sleep(10 * time.Second)
	agent.Start(b)

	return nil
}
