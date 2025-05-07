package command

import (
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/spf13/cobra"
)

func (command Engine) GetName() string {
	return command.Name
}

func (command Engine) GetCondition(api *api.Api) bool {
	return command.Condition(api)
}

func (command Engine) GetFunctions() []func(*api.Api, []string) {
	return command.Functions
}

func (command Engine) GetDependsOn() []func(*api.Api, []string) {
	return command.DependsOn
}

func (command Engine) SetFlags(cmd *cobra.Command) {
	command.Flags(cmd)
}
