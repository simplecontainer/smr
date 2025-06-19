package command

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "smr",
		Short: "SMR CLI",
	}
}

func (command Command) GetName() string {
	return command.Name
}

func (command Command) GetParent() string { return command.Parent }

func (command Command) SetFlags(cmd *cobra.Command) {
	command.Flags(cmd)
}

func (command Command) GetFlags() func(command *cobra.Command) { return command.Flags }

func (command Command) GetArgs() func(*cobra.Command, []string) error {
	return command.Args
}

func (command Command) GetCondition(api iapi.Api, cli *client.Client) bool {
	return command.Condition(api, cli)
}

func (command Command) GetCommand() func(iapi.Api, *client.Client, []string) {
	return command.Command
}

func (command Command) GetDependsOn() []func(iapi.Api, *client.Client, []string) {
	return command.DependsOn
}
