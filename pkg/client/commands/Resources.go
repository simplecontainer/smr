package commands

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/resources"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/formaters"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

func Resources() {
	Commands = append(Commands,
		command.Client{
			Parent: "smrctl",
			Name:   "apply",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					entity := args[0]

					u, err := url.ParseRequestURI(entity)

					var pack = packer.New()

					if err != nil || !u.IsAbs() {
						var stat os.FileInfo
						stat, err = os.Stat(entity)

						if os.IsNotExist(err) {
							helpers.PrintAndExit(err, 1)
						}

						if stat.IsDir() {
							kinds := relations.NewDefinitionRelationRegistry()
							kinds.InTree()

							pack, err = packer.Read(entity, kinds)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}
						} else {
							var definitions []byte
							definitions, err = packer.ReadYAMLFile(entity)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}

							pack.Definitions, err = packer.Parse(definitions)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}
						}
					} else {
						var definitions []byte
						definitions, err = packer.Download(u)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}

						pack.Definitions, err = packer.Parse(definitions)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
					}

					if len(pack.Definitions) != 0 {
						for _, definition := range pack.Definitions {
							err = definition.ProposeApply(cli.Context.GetClient(), cli.Context.APIURL)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}

							fmt.Println(fmt.Sprintf("object applied: %s", definition.Definition.GetKind()))
						}
					} else {
						fmt.Println("specified file/url is not valid definition/pack")
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "remove",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					entity := args[0]

					u, err := url.ParseRequestURI(entity)

					var pack = packer.New()
					var format iformat.Format

					if err != nil || !u.IsAbs() {
						var stat os.FileInfo
						stat, err = os.Stat(entity)

						if os.IsNotExist(err) {
							format, err = helpers.BuildFormat(entity, cli.Group)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}
						}

						if stat.IsDir() {
							kinds := relations.NewDefinitionRelationRegistry()
							kinds.InTree()

							pack, err = packer.Read(entity, kinds)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}
						} else {
							var definitions []byte
							definitions, err = packer.ReadYAMLFile(entity)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}

							pack.Definitions, err = packer.Parse(definitions)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}
						}
					} else {
						var definitions []byte
						definitions, err = packer.Download(u)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}

						pack.Definitions, err = packer.Parse(definitions)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
					}

					if len(pack.Definitions) != 0 {
						for _, definition := range pack.Definitions {
							err = definition.ProposeRemove(cli.Context.GetClient(), cli.Context.APIURL)

							if err != nil {
								helpers.PrintAndExit(err, 1)
							}

							fmt.Println(fmt.Sprintf("object applied: %s", definition.Definition.GetKind()))
						}
					} else {
						err = resources.Delete(cli.Context, format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())

						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Println("object proposed for deleting")
						}
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "list",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[0], cli.Group)

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
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "get",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[1], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var response json.RawMessage
					response, err = resources.Get(cli.Context, format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())

					if err != nil {
						helpers.PrintAndExit(err, 1)
					} else {
						fmt.Println(string(response))
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "inspect",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[1], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var response json.RawMessage
					response, err = resources.Inspect(cli.Context, format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())

					if err != nil {
						helpers.PrintAndExit(err, 1)
					} else {
						fmt.Println(string(response))
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "edit",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[1], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var response json.RawMessage
					response, err = resources.Edit(cli.Context, format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())

					if err != nil {
						helpers.PrintAndExit(err, 1)
					} else {
						fmt.Println(string(response))
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {},
		},
	)
}
