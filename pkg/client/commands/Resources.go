package commands

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/resources"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/formaters"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"strings"
)

func Resources() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("apply").Args(cobra.ExactArgs(1)).Function(cmdApply).Flags(cmdTemplateFlags).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("remove").Args(cobra.ExactArgs(1)).Function(cmdRemove).Flags(cmdTemplateFlags).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("template").Args(cobra.ExactArgs(1)).Function(cmdTemplate).Flags(cmdTemplateFlags).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("list").Args(cobra.ExactArgs(1)).Function(cmdList).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("get").Args(cobra.ExactArgs(1)).Function(cmdGet).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("inspect").Args(cobra.ExactArgs(1)).Function(cmdInspect).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("edit").Args(cobra.ExactArgs(1)).Function(cmdEdit).BuildWithValidation(),
	)
}

var set []string

func cmdApply(api iapi.Api, cli *client.Client, args []string) {
	pack, _, err := determineDefinitions(args[0], set, cli)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if len(pack.Definitions) != 0 {
		for _, definition := range pack.Definitions {
			err = definition.Definition.ProposeApply(cli.Context.GetHTTPClient(), cli.Context.APIURL)
			if err != nil {
				helpers.PrintAndExit(err, 1)
			}

			fmt.Printf("object proposed for apply: %s\n", definition.Definition.Definition.GetKind())
		}
	} else {
		fmt.Println("specified file/url is not valid definition/pack")
	}
}

func cmdRemove(api iapi.Api, cli *client.Client, args []string) {
	pack, format, err := determineDefinitions(args[0], set, cli)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if len(pack.Definitions) != 0 {
		for _, definition := range pack.Definitions {
			err = definition.Definition.ProposeRemove(cli.Context.GetHTTPClient(), cli.Context.APIURL)
			if err != nil {
				helpers.PrintAndExit(err, 1)
			}
			fmt.Printf("object proposed for deleting: %s\n", definition.Definition.Definition.GetKind())
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
}

func cmdTemplate(api iapi.Api, cli *client.Client, args []string) {
	pack, _, err := determineDefinitions(args[0], set, cli)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if len(pack.Definitions) != 0 {
		for _, definition := range pack.Definitions {
			definition.Definition.Definition.SetState(nil)
			definition.Definition.Definition.SetRuntime(nil)

			bytes, err := definition.Definition.Definition.ToJSON()

			if err != nil {
				helpers.PrintAndExit(err, 1)
			}

			var data map[string]interface{}

			if err := json.Unmarshal(bytes, &data); err != nil {
				helpers.PrintAndExit(err, 1)
			}

			yamlData, err := yaml.Marshal(data)
			if err != nil {
				helpers.PrintAndExit(err, 1)
			}

			fmt.Println(strings.TrimSpace(string(yamlData)))
			fmt.Println("---")
		}
	} else {
		fmt.Println("specified path is not valid pack")
	}
}

func cmdTemplateFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&set, "set", []string{}, "")
}

func cmdList(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var objects []json.RawMessage

	switch format.GetKind() {
	default:
		objects, err = resources.ListKind(cli.Context, format.GetPrefix(), format.GetVersion(),
			static.CATEGORY_KIND, format.GetKind())
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
		formaters.Default(objects)
	}
}

func cmdGet(api iapi.Api, cli *client.Client, args []string) {
	action(cli, args, "get")
}

func cmdInspect(api iapi.Api, cli *client.Client, args []string) {
	action(cli, args, "inspect")
}

func cmdEdit(api iapi.Api, cli *client.Client, args []string) {
	action(cli, args, "edit")
}

func determineDefinitions(entity string, set []string, cli *client.Client) (*packer.Pack, iformat.Format, error) {
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

			pack, err = packer.Read(entity, set, kinds)
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

			pack.Definitions, err = packer.Parse(entity, definitions, nil, nil)
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

		pack.Definitions, err = packer.Parse(u.String(), definitions, nil, nil)
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
		response, err = resources.Get(cli.Context, "kind", format.GetPrefix(), format.GetVersion(),
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
