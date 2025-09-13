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

func Gitops() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("gitops").Args(cobra.NoArgs).Function(cmdPsGitops).BuildWithValidation(),
		command.NewBuilder().Parent("gitops").Name("repositories").Args(cobra.MaximumNArgs(1)).Function(cmdRepositories).Flags(cmdPsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("gitops").Name("definitions").Args(cobra.MinimumNArgs(1)).Function(cmdDefinitions).Flags(cmdPsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("gitops").Name("commit").Function(cmdCommit).Flags(cmdCommitFlags).Args(cobra.ExactArgs(3)).BuildWithValidation(),
		command.NewBuilder().Parent("gitops").Name("sync").Args(cobra.ExactArgs(1)).Function(cmdSync).BuildWithValidation(),
		command.NewBuilder().Parent("gitops").Name("refresh").Args(cobra.ExactArgs(1)).Function(cmdRefresh).BuildWithValidation(),
	)
}

func cmdRepositories(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build("gitops", cli.Group)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var objects []json.RawMessage
	objects, err = resources.ListState(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind())
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	formaters.Repositories(objects)
}

func cmdDefinitions(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var object json.RawMessage
	object, err = resources.Inspect(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind(), format.GetGroup(), format.GetName())
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var objects []json.RawMessage
	objects = append(objects, object)

	formaters.Definitions(objects)
}
