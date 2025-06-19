package commands

import (
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

var (
	EmptyCondition = func(iapi.Api) bool { return true }
	EmptyFunction  = func(api iapi.Api, args []string) {}
	EmptyDepend    = []func(iapi.Api, []string){EmptyFunction}
	EmptyFlag      = func(cmd *cobra.Command) {}
)
