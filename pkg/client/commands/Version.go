package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
)

func Version() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("version").Function(cmdVersion).BuildWithValidation(),
	)
}

func cmdVersion(api iapi.Api, cli *client.Client, args []string) {
	fmt.Println(cli.Version.Version)
}
