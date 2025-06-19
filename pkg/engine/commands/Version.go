package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

func Version() {
	Commands = append(Commands,
		command.Engine{
			Parent:    "smr",
			Name:      "version",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(a iapi.Api, args []string) {
					fmt.Println(a.GetVersion().Node)
				},
			},
			DependsOn: []func(iapi.Api, []string){
				func(cli iapi.Api, args []string) {},
			},
			Flags: EmptyFlag,
		},
	)
}
