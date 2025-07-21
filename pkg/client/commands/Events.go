package commands

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

func Events() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("events").Function(cmdEvents).Flags(cmdEventsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("commit").Function(cmdCommit).Flags(cmdCommitFlags).Args(cobra.ExactArgs(3)).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("sync").Args(cobra.ExactArgs(1)).Function(cmdSync).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("refresh").Args(cobra.ExactArgs(1)).Function(cmdRefresh).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("restart").Args(cobra.ExactArgs(1)).Function(cmdRestart).BuildWithValidation(),
	)
}

func cmdCommit(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	commit := implementation.NewCommit()

	err = commit.Parse(args[1], args[2])
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	bytes, err := commit.ToJson()
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	event := events.New(events.EVENT_COMMIT, static.KIND_GITOPS, static.SMR_PREFIX, static.KIND_GITOPS, format.GetGroup(), format.GetName(), bytes)

	bytes, err = event.ToJSON()

	Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)
}
func cmdCommitFlags(cmd *cobra.Command) {
	cmd.Flags().String("node", "simplecontainer-node-1", "Node")
	cmd.Flags().String("wait", "", "Wait for specific event")
	cmd.Flags().String("resource", "", "Specify resource you want to track")
}

func cmdEvents(api iapi.Api, cli *client.Client, args []string) {
	ctx, cancel := context.WithCancel(context.Background())

	err := cli.Events(ctx, cancel, viper.GetString("wait"), viper.GetString("resource"), cli.Tracker)

	if err != nil {
		return
	}
}
func cmdEventsFlags(cmd *cobra.Command) {
	cmd.Flags().String("node", "simplecontainer-node-1", "Node")
	cmd.Flags().String("wait", "", "Wait for specific event")
	cmd.Flags().String("resource", "", "Specify resource you want to track")
}

func cmdSync(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	event := events.New(events.EVENT_SYNC, static.KIND_GITOPS, static.SMR_PREFIX, static.KIND_GITOPS, format.GetGroup(), format.GetName(), nil)

	var bytes []byte
	bytes, err = event.ToJSON()

	Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)
}

func cmdRefresh(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	event := events.New(events.EVENT_REFRESH, static.KIND_GITOPS, static.SMR_PREFIX, static.KIND_GITOPS, format.GetGroup(), format.GetName(), nil)

	var bytes []byte
	bytes, err = event.ToJSON()

	Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)
}

func cmdRestart(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	event := events.New(events.EVENT_RESTART, static.KIND_CONTAINERS, static.SMR_PREFIX, static.KIND_CONTAINERS, format.GetGroup(), format.GetName(), nil)

	var bytes []byte
	bytes, err = event.ToJSON()

	Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)
}

func Event(context *contexts.ClientContext, prefix string, version string, category string, kind string, group string, name string, data []byte) {
	response := network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/kind/propose/%s/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group, name), http.MethodPost, data)

	if response.Success {
		fmt.Println(response.Explanation)
	} else {
		fmt.Println(response.ErrorExplanation)
	}
}
