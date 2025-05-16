package command

import (
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "smr",
		Short: "SMR CLI",
	}
}

func (command Engine) GetName() string {
	return command.Name
}

func (command Engine) GetCondition(api iapi.Api) bool {
	return command.Condition(api)
}

func (command Engine) GetFunctions() []func(iapi.Api, []string) {
	return command.Functions
}

func (command Engine) GetDependsOn() []func(iapi.Api, []string) {
	return command.DependsOn
}

func (command Engine) SetFlags(cmd *cobra.Command) {
	command.Flags(cmd)
}
