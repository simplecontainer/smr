package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
	"time"
)

func Context() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("context").Function(cmdContext).BuildWithValidation(),
		command.NewBuilder().Parent("context").Name("switch").Args(cobra.MaximumNArgs(1)).Function(cmdContextSwitch).BuildWithValidation(),

		command.NewBuilder().Parent("context").Name("export").Args(cobra.MaximumNArgs(1)).Function(cmdContextExport).Flags(cmdContextExportFlags).BuildWithValidation(),
		command.NewBuilder().Parent("export").Name("active").Args(cobra.MaximumNArgs(1)).Function(cmdContextExportActive).Flags(cmdContextExportActiveFlags).BuildWithValidation(),
		command.NewBuilder().Parent("context").Name("import").Args(cobra.MaximumNArgs(2)).Function(cmdContextImport).Flags(cmdContextImportFlags).BuildWithValidation(),
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

	encrypted, key, err := cli.Manager.ExportContext(name, viper.GetString("api"))
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if viper.GetBool("upload") {
		response, err := cli.Manager.Upload(viper.GetString("token"), viper.GetString("registry"), fmt.Sprintf("%s %s", encrypted, key))

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println(response)
	} else {
		fmt.Println(encrypted, key)
	}
}
func cmdContextExportFlags(cmd *cobra.Command) {
	cmd.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")
	cmd.Flags().String("registry", "https://app.simplecontainer.io", "Registry for context sharing")
	cmd.Flags().String("token", "", "Token for authentication and authorization")
	cmd.Flags().Bool("upload", false, "Upload context export on the specified registry")
}

func cmdContextExportActive(api iapi.Api, cli *client.Client, args []string) {
	name := ""

	if len(args) == 1 {
		name = args[0]
	}

	c, err := cli.Manager.GetActive()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	encrypted, key, err := cli.Manager.ExportContext(name, c.APIURL)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if viper.GetBool("upload") {
		response, err := cli.Manager.Upload(viper.GetString("token"), viper.GetString("registry"), fmt.Sprintf("%s %s", encrypted, key))

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println(response)
	} else {
		fmt.Println(encrypted, key)
	}
}

func cmdContextExportActiveFlags(cmd *cobra.Command) {
	cmd.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")
	cmd.Flags().String("registry", "https://app.simplecontainer.io", "Registry for context sharing")
	cmd.Flags().String("token", "", "Token for authentication and authorization")
	cmd.Flags().Bool("upload", false, "Upload context export on the specified registry")
}

func cmdContextImport(api iapi.Api, cli *client.Client, args []string) {
	type Context struct {
		ID        *uuid.UUID `json:"id"`
		TokenID   *uuid.UUID `json:"token_id"`
		ContextID *uuid.UUID `json:"context_id"`
		Context   string     `json:"context"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
	}

	type DownloadResponse struct {
		Contexts []Context `json:"members"`
	}

	var err error
	var contexts *DownloadResponse = &DownloadResponse{
		Contexts: make([]Context, 0),
	}

	if viper.GetBool("download") {
		response, err := cli.Manager.Download(viper.GetString("token"), viper.GetString("registry"))

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		err = json.Unmarshal([]byte(response), &contexts)

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	} else {
		if len(args) == 2 {
			contexts.Contexts = append(contexts.Contexts, Context{
				Context: fmt.Sprintf("%s %s", args[0], args[1]),
			})
		} else {
			helpers.PrintAndExit(errors.New("please provide valid context for import"), 1)
		}
	}

	for _, ctx := range contexts.Contexts {
		split := strings.Split(ctx.Context, " ")
		cli.Context, err = cli.Manager.ImportContext(split[0], split[1])
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
}
func cmdContextImportFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
	cmd.Flags().String("registry", "https://app.simplecontainer.io", "Registry for context sharing")
	cmd.Flags().String("token", "", "Token for authentication and authorization")
	cmd.Flags().Bool("download", false, "Download context export from the specified registry")
}
