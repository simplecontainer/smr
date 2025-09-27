package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/spf13/cobra"
)

func Pack() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("pack").Args(cobra.NoArgs).Function(cmdPack).Flags(cmdPackFlags).BuildWithValidation(),
		command.NewBuilder().Parent("pack").Name("init").Args(cobra.ExactArgs(1)).Function(cmdPackInit).Flags(cmdPackInitFlags).BuildWithValidation(),
	)
}

func cmdPack(api iapi.Api, cli *client.Client, args []string) {
	fmt.Println("to get started run: smrctl pack init")
}

func cmdPackFlags(cmd *cobra.Command) {}

func cmdPackInit(api iapi.Api, cli *client.Client, args []string) {
	err := packer.Init(args[0])
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("pack created")
}

func cmdPackInitFlags(cmd *cobra.Command) {}
