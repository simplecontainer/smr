package commands

import (
	"encoding/json"
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

func Containers() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("containers").Args(cobra.NoArgs).Function(cmdPs).Flags(cmdPsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("containers").Name("images").Args(cobra.MaximumNArgs(1)).Function(cmdImages).Flags(cmdPsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("containers").Name("restart").Args(cobra.ExactArgs(1)).Function(cmdRestart).BuildWithValidation(),
	)
}

func cmdImages(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build("containers", cli.Group)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var objects []json.RawMessage
	objects, err = resources.ListState(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind())
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	formaters.Images(objects)
}
