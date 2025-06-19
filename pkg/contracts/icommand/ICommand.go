package icommand

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

type Command interface {
	GetParent() string
	GetName() string
	GetFlags() func(command *cobra.Command)
	SetFlags(cmd *cobra.Command)
	GetArgs() func(*cobra.Command, []string) error
	GetCondition(api iapi.Api, cli *client.Client) bool
	GetCommand() func(iapi.Api, *client.Client, []string)
	GetDependsOn() []func(iapi.Api, *client.Client, []string)
}
