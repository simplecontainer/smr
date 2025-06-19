package command

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/icommand"
	"github.com/spf13/cobra"
)

type Builder struct {
	parent    string
	name      string
	flags     func(cmd *cobra.Command)
	args      func(*cobra.Command, []string) error
	condition func(iapi.Api, *client.Client) bool
	command   func(iapi.Api, *client.Client, []string)
	dependsOn []func(iapi.Api, *client.Client, []string)
}

func NewBuilder() *Builder {
	return &Builder{
		args:      cobra.NoArgs,
		flags:     EmptyFlag,
		condition: EmptyCondition,
		dependsOn: EmptyDepend,
		command:   EmptyFunction,
	}
}

func (cb *Builder) Parent(parent string) *Builder {
	cb.parent = parent
	return cb
}

func (cb *Builder) Name(name string) *Builder {
	cb.name = name
	return cb
}

func (cb *Builder) Flags(flags func(cmd *cobra.Command)) *Builder {
	cb.flags = flags
	return cb
}

func (cb *Builder) Args(args func(*cobra.Command, []string) error) *Builder {
	cb.args = args
	return cb
}

func (cb *Builder) Function(fn func(iapi.Api, *client.Client, []string)) *Builder {
	cb.command = fn
	return cb
}

func (cb *Builder) Condition(fn func(iapi.Api, *client.Client) bool) *Builder {
	cb.condition = fn
	return cb
}

func (cb *Builder) DependsOn(fns ...func(iapi.Api, *client.Client, []string)) *Builder {
	cb.dependsOn = append(cb.dependsOn, fns...)
	return cb
}

func (cb *Builder) Build() icommand.Command {
	return Command{
		Parent:    cb.parent,
		Name:      cb.name,
		Args:      cb.args,
		Flags:     cb.flags,
		Command:   cb.command,
		Condition: cb.condition,
		DependsOn: cb.dependsOn,
	}
}

func (cb *Builder) Validate() error {
	if cb.name == "" {
		return fmt.Errorf("command name is required")
	}
	if cb.parent == "" {
		return fmt.Errorf("command parent is required")
	}
	return nil
}

func (cb *Builder) BuildWithValidation() icommand.Command {
	if err := cb.Validate(); err != nil {
		panic(err)
	}

	return cb.Build()
}
