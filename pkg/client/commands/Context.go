package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
	"time"
)

func Context() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("context").Function(cmdContext).Build(),
		command.NewBuilder().Parent("context").Name("switch").Args(cobra.MaximumNArgs(1)).Function(cmdContext).Build(),

		command.NewBuilder().Parent("context").Name("export").Args(cobra.MaximumNArgs(1)).Function(cmdContextImport).Flags(cmdContextImportFlags).Build(),
		command.NewBuilder().Parent("context").Name("import").Args(cobra.ExactArgs(2)).Function(cmdContextExport).Flags(cmdContextExportFlags).Build(),
	)
}

func cmdContext(api iapi.Api, cli *client.Client, args []string) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	activeCtx, err := client.LoadActive(client.DefaultConfig(environment.ClientDirectory))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println(activeCtx.Name)
}

func cmdContextSwitch(api iapi.Api, cli *client.Client, args []string) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	contexts, err := cli.Manager.ListContexts()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if len(args) > 0 {
		var activeCtx *client.ClientContext
		activeCtx, err = client.LoadByName(args[0], client.DefaultConfig(environment.ClientDirectory))

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		err = cli.Manager.SetActive(activeCtx.Name)

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println(fmt.Sprintf("active context is %s", activeCtx.Name))
	} else {
		prompt := promptui.Select{
			Label: "Select a context",
			Items: contexts,
		}

		var result string
		_, result, err = prompt.Run()

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		err = cli.Manager.SetActive(result)

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println(fmt.Sprintf("active context is %s", result))
	}
}

func cmdContextExport(api iapi.Api, cli *client.Client, args []string) {
	name := ""

	if len(args) == 1 {
		name = args[0]
	}

	encrypted, key, err := cli.Manager.ExportContext(name)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println(encrypted, key)
}
func cmdContextExportFlags(cmd *cobra.Command) {
	cmd.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")
}

func cmdContextImport(api iapi.Api, cli *client.Client, args []string) {
	var err error
	cli.Context, err = cli.Manager.ImportContext(args[0], args[1])
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if cli.Context.APIURL == "" {
		helpers.PrintAndExit(errors.New("imported context has no API URL"), 1)
	}

	connCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = cli.Context.Connect(connCtx, true); err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if err = cli.Context.Save(); err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if err = cli.Manager.SetActive(cli.Context.Name); err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Printf("context '%s' successfully imported and set as active\n", cli.Context.Name)
}
func cmdContextImportFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
}
