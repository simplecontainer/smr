package commands

import (
	"encoding/json"
	"errors"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/resources"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/formaters"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
)

func Alias() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("ps").Args(cobra.MaximumNArgs(1)).Function(cmdPs).Flags(cmdPsFlags).BuildWithValidation(),
	)
}

func cmdPs(api iapi.Api, cli *client.Client, args []string) {
	if len(args) == 0 {
		args = append(args, "containers")
	}

	switch args[0] {
	case static.KIND_CONTAINERS:
		break
	case static.KIND_GITOPS:
		break
	default:
		helpers.PrintAndExit(errors.New("ps works only for containers and gitops resources"), 1)
	}

	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var objects []json.RawMessage

	switch format.GetKind() {
	case static.KIND_GITOPS:
		objects, err = resources.ListState(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind())

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		formaters.Gitops(objects)
		break
	case static.KIND_CONTAINERS:
		objects, err = resources.ListState(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind())

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		formaters.Container(objects)
		break
	default:
		helpers.PrintAndExit(errors.New("ps works only for containers and gitops resources"), 1)
		break
	}
}

func cmdPsFlags(cmd *cobra.Command) {
	cmd.Flags().String("output", "full", "output format: full, short")
}
