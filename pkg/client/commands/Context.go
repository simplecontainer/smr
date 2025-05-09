package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/spf13/cobra"
	"time"
)

func Context() {
	Commands = append(Commands,
		command.Client{
			Parent: "smrctl",
			Name:   "context",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					activeCtx, err := client.LoadActive(client.DefaultConfig(environment.ClientDirectory))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Println(activeCtx.Name)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
			},
		},
		command.Client{
			Parent: "context",
			Name:   "export",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.MaximumNArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					name := ""

					if len(args) == 1 {
						name = args[0]
					}

					encrypted, key, err := cli.Manager.ExportContext(name)
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Println(encrypted, key)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {

				},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")
			},
		},
		command.Client{
			Parent: "context",
			Name:   "import",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(2),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					var err error
					cli.Context, err = cli.Manager.ImportContext(args[0], args[1])
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if cli.Context.APIURL == "" {
						helpers.PrintAndExit(errors.New("imported context has no API URL"), 1)
					}

					connCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if err = cli.Context.Connect(connCtx, true); err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if err = cli.Context.Save(); err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if err = cli.Manager.SetActive(cli.Context.Name); err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Printf("context '%s' successfully imported and set as active\n", cli.Context.Name)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
				},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
			},
		},
	)
}
