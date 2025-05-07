package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var Commands []command.Engine

func PreloadCommands() {

}

func Run(api *api.Api, c *cobra.Command) {
	c.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
	})

	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Printf("error: %s\n\n", err)
		_ = c.Usage()
		return nil
	})

	c.SetArgs(os.Args[1:])

	c.Run = func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Printf("unknown command: %s\n", strings.Join(args, " "))
		}
		_ = cmd.Usage()
	}

	for _, cmd := range Commands {
		cobraCmd := &cobra.Command{
			Use:   cmd.Name,
			Short: fmt.Sprintf("%s %s", cmd.Parent, cmd.Name),
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return nil
				}

				for _, sub := range cmd.Commands() {
					if sub.Name() == args[0] {
						return nil // valid subcommand
					}
				}

				return fmt.Errorf("unknown subcommand: %s", args[0])
			},
			PreRunE: func(c *cobra.Command, args []string) error {
				if !cmd.Condition(api) {
					return fmt.Errorf("condition failed for command %s", c.Use)
				}

				for _, dep := range cmd.DependsOn {
					dep(api, args)
				}

				return nil
			},
			Run: func(c *cobra.Command, args []string) {
				for _, fn := range cmd.Functions {
					fn(api, args)
				}
			},
		}

		cmd.SetFlags(cobraCmd)
		BindCommandFlags(cobraCmd)

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

	_ = c.Execute()
}

func SetupGlobalFlags(rootCmd *cobra.Command) {
	// Global flags
	rootCmd.PersistentFlags().String("home", "/home/node", "Root directory for all actions - keep default inside container")
	rootCmd.PersistentFlags().String("log", "info", "Log level: debug, info, warn, error, dpanic, panic, fatal")

	// Bind global flags to viper
	viper.BindPFlag("home", rootCmd.PersistentFlags().Lookup("home"))
	viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
}

// BindCommandFlags helps bind all flags from a command to viper
func BindCommandFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// Bind each flag to viper
		if err := viper.BindPFlag(flag.Name, flag); err != nil {
			fmt.Printf("warning: failed to bind flag '%s': %s\n", flag.Name, err)
		}
	})
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
