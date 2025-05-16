package factory

import (
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/control/drain"
	"github.com/simplecontainer/smr/pkg/control/restart"
	"github.com/simplecontainer/smr/pkg/control/start"
	"github.com/simplecontainer/smr/pkg/control/upgrade"
)

func NewCommand(cmd string, opts map[string]string) icontrol.Command {
	switch cmd {
	case "drain":
		return drain.NewDrainCommand(opts)
	case "start":
		return start.NewStartCommand(opts)
	case "restart":
		return restart.NewRestartCommand(opts)
	case "upgrade":
		return upgrade.NewUpgradeCommand(opts)
	}

	panic("unknown control")
	return nil
}
