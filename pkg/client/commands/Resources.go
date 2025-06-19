package commands

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/resources"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/f"
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
			Parent:    "smrctl",
			Name:      "apply",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					pack, _, err := determineDefinitions(args[0], cli)
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if len(pack.Definitions) != 0 {
						for _, definition := range pack.Definitions {
							err = definition.ProposeApply(cli.Context.GetClient(), cli.Context.APIURL)
							if err != nil {
								helpers.PrintAndExit(err, 1)
							}

							fmt.Printf("object applied: %s\n", definition.Definition.GetKind())
						}
					} else {
						fmt.Println("specified file/url is not valid definition/pack")
					}
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Client{
			Parent:    "smrctl",
			Name:      "remove",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					pack, format, err := determineDefinitions(args[0], cli)
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if len(pack.Definitions) != 0 {
						for _, definition := range pack.Definitions {
							err = definition.ProposeRemove(cli.Context.GetClient(), cli.Context.APIURL)
							if err != nil {
								helpers.PrintAndExit(err, 1)
							}
							fmt.Printf("object applied: %s\n", definition.Definition.GetKind())
						}
					} else {
						err = resources.Delete(cli.Context, format.GetPrefix(), format.GetVersion(),
							format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())

						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Println("object proposed for deleting")
						}
					}
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Client{
			Parent:    "smrctl",
			Name:      "list",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := f.Build(args[0], cli.Group)
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var objects []json.RawMessage

					switch format.GetKind() {
					case static.KIND_GITOPS:
						objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(),
							static.CATEGORY_STATE, format.GetKind())
						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
						formaters.Gitops(objects)
					case static.KIND_CONTAINERS:
						objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(),
							static.CATEGORY_STATE, format.GetKind())
						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
						formaters.Container(objects)
					default:
						objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(),
							static.CATEGORY_KIND, format.GetKind())
						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
						formaters.Default(objects)
					}
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Client{
			Parent:    "smrctl",
			Name:      "get",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					action(cli, args, "get")
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Client{
			Parent:    "smrctl",
			Name:      "inspect",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					action(cli, args, "inspect")
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Client{
			Parent:    "smrctl",
			Name:      "edit",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					action(cli, args, "edit")
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
	)
}

func determineDefinitions(entity string, cli *client.Client) (*packer.Pack, iformat.Format, error) {
	var pack = packer.New()
	var format iformat.Format
	var err error

	u, err := url.ParseRequestURI(entity)

	if err != nil || !u.IsAbs() {
		// Not url - then it is file path
		var stat os.FileInfo
		stat, err = os.Stat(entity)

		if os.IsNotExist(err) {
			// File path doesn't exist - then it is format
			format, err = f.Build(entity, cli.Group)

			if err != nil {
				// It is not format - error out
				return nil, format, err
			}
		}

		if stat != nil && stat.IsDir() {
			// It is directory - check if it is pack
			kinds := relations.NewDefinitionRelationRegistry()
			kinds.InTree()

			pack, err = packer.Read(entity, kinds)
			if err != nil {
				return nil, format, err
			}
		} else if stat != nil {
			// It is file check if it is valid yaml definition

			var definitions []byte
			definitions, err = packer.ReadYAMLFile(entity)
			if err != nil {
				return nil, format, err
			}

			pack.Definitions, err = packer.Parse(definitions)
			if err != nil {
				return nil, format, err
			}
		}
	} else {
		// It is URL so download file and process it as yaml definition

		var definitions []byte
		definitions, err = packer.Download(u)
		if err != nil {
			return nil, format, err
		}

		pack.Definitions, err = packer.Parse(definitions)
		if err != nil {
			return nil, format, err
		}
	}

	return pack, format, nil
}

func action(cli *client.Client, args []string, action string) {
	format, err := f.Build(args[0], cli.Group)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var response json.RawMessage

	switch action {
	case "get":
		response, err = resources.Get(cli.Context, format.GetPrefix(), format.GetVersion(),
			format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())
	case "inspect":
		response, err = resources.Inspect(cli.Context, format.GetPrefix(), format.GetVersion(),
			format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())
	case "edit":
		response, err = resources.Edit(cli.Context, format.GetPrefix(), format.GetVersion(),
			format.GetCategory(), format.GetKind(), format.GetGroup(), format.GetName())
	}

	if err != nil {
		helpers.PrintAndExit(err, 1)
	} else {
		fmt.Println(string(response))
	}
}
