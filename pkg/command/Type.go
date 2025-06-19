package command

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

var (
	EmptyCondition = func(iapi.Api, *client.Client) bool { return true }
	EmptyFunction  = func(iapi.Api, *client.Client, []string) {}
	EmptyDepend    = []func(iapi.Api, *client.Client, []string){EmptyFunction}
	EmptyFlag      = func(cmd *cobra.Command) {}
)

type Command struct {
	Parent    string
	Name      string
	Flag      string
	Args      func(*cobra.Command, []string) error
	Condition func(iapi.Api, *client.Client) bool
	Command   func(iapi.Api, *client.Client, []string)
	DependsOn []func(iapi.Api, *client.Client, []string)
	Flags     func(command *cobra.Command)
}
