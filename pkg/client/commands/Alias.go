package commands

import (
	"encoding/json"
	"errors"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/resources"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/formaters"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
)

func Alias() {
	Commands = append(Commands,
		command.Client{
			Parent: "smrctl",
			Name:   "ps",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.MaximumNArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					if len(args) == 0 {
						args = append(args, "containers")
					}

					switch args[0] {
					case static.KIND_CONTAINERS:
						break
					case static.KIND_GITOPS:
						break
					default:
						helpers.PrintAndExit(errors.New("ps works only for containers and gitops resources"), 1)
					}

					format, err := f.Build(args[0], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var objects []json.RawMessage

					switch format.GetKind() {
					case static.KIND_GITOPS:
						objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind())
						formaters.Gitops(objects)
						break
					case static.KIND_CONTAINERS:
						objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_STATE, format.GetKind())
						formaters.Container(objects)
						break
					default:
						objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(), static.CATEGORY_KIND, format.GetKind())
						formaters.Default(objects)
						break
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("output", "full", "output format: full, short")
			},
		},
	)
}
