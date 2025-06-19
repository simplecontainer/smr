package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
)

func Version() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smr").Name("version").Function(cmdVersion).BuildWithValidation(),
	)
}

func cmdVersion(a iapi.Api, cli *client.Client, args []string) {
	fmt.Println(a.GetVersion().Node)
}
