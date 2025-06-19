package commands

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/spf13/cobra"
)

var (
	EmptyCondition = func(*client.Client) bool { return true }
	EmptyFunction  = func(cli *client.Client, args []string) {}
	EmptyDepend    = []func(*client.Client, []string){EmptyFunction}
	EmptyFlag      = func(cmd *cobra.Command) {}
)
