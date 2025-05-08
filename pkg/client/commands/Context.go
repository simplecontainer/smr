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
	"github.com/spf13/viper"
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
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "context",
			Name:   "connect",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())

					parsed, err := helpers.EnforceHTTPS(viper.GetString("api"))

					manager := client.NewManager(client.DefaultConfig(environment.ClientDirectory))
					ctx, err := manager.CreateContext(viper.GetString("name"), parsed.String(), []byte(viper.GetString("bundle")))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					err = ctx.Connect(context.Background(), viper.GetBool("retry"))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					err = ctx.Save()

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					err = manager.SetActiveContext(viper.GetString("name"))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("name", "localhost", "Context name")
				cmd.Flags().String("api", "https://localhost:1443", "Node control endpoint")
				cmd.Flags().String("bundle", "", "Path to .pem bundle")
				cmd.Flags().Bool("retry", false, "Retry connect on fail with backoff")

				viper.BindPFlag("name", cmd.Flags().Lookup("name"))
				viper.BindPFlag("api", cmd.Flags().Lookup("api"))
				viper.BindPFlag("bundle", cmd.Flags().Lookup("bundle"))
				viper.BindPFlag("retry", cmd.Flags().Lookup("retry"))
			},
		},
		command.Client{
			Parent: "context",
			Name:   "export",
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

					var encrypted, key string
					encrypted, key, err = activeCtx.Export()

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Println(key, encrypted)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {

				},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")

				viper.BindPFlag("api", cmd.Flags().Lookup("api"))
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

				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())

					importedCtx, err := client.Import(client.DefaultConfig(environment.ClientDirectory), args[0], args[1])

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if importedCtx.APIURL == "" {
						helpers.PrintAndExit(errors.New("imported context has no API URL"), 1)
					}

					connCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if err := importedCtx.Connect(connCtx, true); err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if importedCtx.Name == "" {
						importedCtx.WithName(fmt.Sprintf("imported-%d", time.Now().Unix()))
					}

					if err := importedCtx.Save(); err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Printf("context '%s' successfully imported and set as active\n", importedCtx.Name)
				},
			},
			Flags: func(cmd *cobra.Command) {

			},
		},
	)
}
