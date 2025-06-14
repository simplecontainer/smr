package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/spf13/cobra"
)

func Version() {
	Commands = append(Commands,
		command.Client{
			Parent: "smrctl",
			Name:   "version",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					fmt.Println(cli.Version.Version)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
	)
}
