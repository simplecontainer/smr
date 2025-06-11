package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

var Commands []command.Client

func PreloadCommands() {
	Context()
	Resources()
	Version()
	Alias()
	Streams()
	Events()
}

func Run(cli *client.Client, c *cobra.Command) {
	c.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
	})

	SetupGlobalFlags(c)

	for _, cmd := range Commands {
		cobraCmd := &cobra.Command{
			Use:   cmd.Name,
			Short: fmt.Sprintf("%s %s", cmd.Parent, cmd.Name),
			Args:  cmd.Args,
			PreRunE: func(c *cobra.Command, args []string) error {
				var err error
				cli.Context, err = client.LoadActive(client.DefaultConfig(configuration.NewEnvironment(configuration.WithHostConfig()).ClientDirectory))

				if err != nil {
					switch c.Name() {
					case "import", "version":
						break
					default:
						fmt.Println("no active context found - try using smr context switch")
						os.Exit(1)
					}
				}

				if !cmd.Condition(cli) {
					return fmt.Errorf("condition failed for command %s", c.Use)
				}

				for _, dep := range cmd.DependsOn {
					dep(cli, args)
				}

				return nil
			},
			Run: func(c *cobra.Command, args []string) {
				c.Flags().VisitAll(func(flag *pflag.Flag) {
					if err := viper.BindPFlag(flag.Name, flag); err != nil {
						fmt.Printf("warning: failed to bind flag '%s': %s\n", flag.Name, err)
						os.Exit(1)
					}
				})

				for _, fn := range cmd.Functions {
					fn(cli, args)
				}
			},
		}

		cmd.SetFlags(cobraCmd)

		if cmd.Parent == "smr" || cmd.Parent == "" {
			c.AddCommand(cobraCmd)
		} else {
			parent := findCommand(c, cmd.Parent)

			if parent != nil {
				parent.AddCommand(cobraCmd)
			} else {
				fmt.Printf("warning: parent command '%s' not found for '%s'\n", cmd.Parent, cmd.Name)
			}
		}
	}

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}

func SetupGlobalFlags(rootCmd *cobra.Command) {
	// Global flags
	rootCmd.PersistentFlags().String("home", helpers.GetRealHome(), "Root directory for all actions - keep default inside container")
	rootCmd.PersistentFlags().String("log", "info", "Log level: debug, info, warn, error, dpanic, panic, fatal")

	// Bind global flags to viper
	viper.BindPFlag("home", rootCmd.PersistentFlags().Lookup("home"))
	viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))

}

func findCommand(cmd *cobra.Command, name string) *cobra.Command {
	if cmd.Use == name {
		return cmd
	}
	for _, c := range cmd.Commands() {
		if result := findCommand(c, name); result != nil {
			return result
		}
	}
	return nil
}
