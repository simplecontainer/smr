package command

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/spf13/cobra"
)

func (command Client) GetName() string {
	return command.Name
}

func (command Client) GetCondition(c *client.Client) bool {
	return command.Condition(c)
}

func (command Client) GetFunctions() []func(*client.Client, []string) {
	return command.Functions
}

func (command Client) GetDependsOn() []func(*client.Client, []string) {
	return command.DependsOn
}

func (command Client) SetFlags(cmd *cobra.Command) {
	command.Flags(cmd)
}
