package commands

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

func Events() {
	Commands = append(Commands,
		command.Client{
			Parent: "smrctl",
			Name:   "events",
			Condition: func(cli *client.Client) bool {
				return true
			},
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					ctx, cancel := context.WithCancel(context.Background())

					err := cli.Events(ctx, cancel, viper.GetString("wait"), viper.GetString("resource"), cli.Tracker)

					if err != nil {
						return
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("wait", "", "Wait for specific event")
				cmd.Flags().String("resource", "", "Specify resource you want to track")
			},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "sync",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := f.Build(args[0], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					event := events.New(events.EVENT_SYNC, static.KIND_GITOPS, static.SMR_PREFIX, static.KIND_GITOPS, format.GetGroup(), format.GetName(), nil)

					var bytes []byte
					bytes, err = event.ToJSON()

					Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)

				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "refresh",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := f.Build(args[0], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					event := events.New(events.EVENT_REFRESH, static.KIND_GITOPS, static.SMR_PREFIX, static.KIND_GITOPS, format.GetGroup(), format.GetName(), nil)

					var bytes []byte
					bytes, err = event.ToJSON()

					Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "restart",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := f.Build(args[0], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					event := events.New(events.EVENT_RESTART, static.KIND_CONTAINERS, static.SMR_PREFIX, static.KIND_CONTAINERS, format.GetGroup(), format.GetName(), nil)

					var bytes []byte
					bytes, err = event.ToJSON()

					Event(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_EVENT, format.GetKind(), format.GetGroup(), format.GetName(), bytes)

				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
	)
}

func Event(context *client.ClientContext, prefix string, version string, category string, kind string, group string, name string, data []byte) {
	response := network.Send(context.GetClient(), fmt.Sprintf("%s/api/v1/kind/propose/%s/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group, name), http.MethodPost, data)

	if response.Success {
		fmt.Println(response.Explanation)
	} else {
		fmt.Println(response.ErrorExplanation)
	}
}
