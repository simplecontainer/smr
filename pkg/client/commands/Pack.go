package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/distribution/reference"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/packer/ocicredentials"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

func Pack() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smrctl").Name("pack").Args(cobra.NoArgs).Function(cmdPack).Flags(cmdPackFlags).BuildWithValidation(),
		command.NewBuilder().Parent("pack").Name("init").Args(cobra.ExactArgs(1)).Function(cmdPackInit).Flags(cmdPackInitFlags).BuildWithValidation(),
		command.NewBuilder().Parent("pack").Name("login").Args(cobra.ExactArgs(0)).Function(cmdPackLogin).Flags(cmdPackLoginFlags).BuildWithValidation(),
		command.NewBuilder().Parent("pack").Name("logout").Args(cobra.ExactArgs(0)).Function(cmdPackLogout).Flags(cmdPackLogoutFlags).BuildWithValidation(),
		command.NewBuilder().Parent("pack").Name("push").Args(cobra.ExactArgs(1)).Function(cmdPackPush).Flags(cmdPackPushFlags).BuildWithValidation(),
		command.NewBuilder().Parent("pack").Name("pull").Args(cobra.ExactArgs(1)).Function(cmdPackPull).Flags(cmdPackPullFlags).BuildWithValidation(),
	)
}

func cmdPack(api iapi.Api, cli *client.Client, args []string) {
	fmt.Println("to get started run: smrctl pack init")
}

func cmdPackFlags(cmd *cobra.Command) {}

func cmdPackInit(api iapi.Api, cli *client.Client, args []string) {
	err := packer.Init(args[0])
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("pack created")
}
func cmdPackInitFlags(cmd *cobra.Command) {}

func cmdPackLogin(api iapi.Api, cli *client.Client, args []string) {
	registry := viper.GetString("registry")

	err := cli.Credentials.Load()
	if cli.Credentials.Credentials[registry].IsAnonymous() || err != nil {
		err = cli.Credentials.Input(registry)
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	}

	credentials, exists := cli.Credentials.Find(registry)
	if !exists {
		err = cli.Credentials.Input(registry)
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	}

	OCIClient, err := packer.NewClient(credentials)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = OCIClient.TestLogin(context.Background())
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = cli.Credentials.Save()
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("login successfully")
}
func cmdPackLoginFlags(cmd *cobra.Command) {
	cmd.Flags().String("registry", ocicredentials.DefaultRegistry, "Registry for packs")
}

func cmdPackLogout(api iapi.Api, cli *client.Client, args []string) {
	registry := viper.GetString("registry")

	err := cli.Credentials.Load()
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	_, exists := cli.Credentials.Find(registry)
	if !exists {
		helpers.PrintAndExit(errors.New("registry not found"), 1)
	}

	delete(cli.Credentials.Credentials, registry)

	err = cli.Credentials.Save()
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("logout successfully")
}
func cmdPackLogoutFlags(cmd *cobra.Command) {
	cmd.Flags().String("registry", ocicredentials.DefaultRegistry, "Registry for packs")
}

func cmdPackPush(api iapi.Api, cli *client.Client, args []string) {
	registry := viper.GetString("registry")
	ref, err := reference.Parse(args[0])
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	named, ok := ref.(reference.Named)
	if !ok {
		helpers.PrintAndExit(err, 1)
	}

	repository := named.Name()

	err = cli.Credentials.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = cli.Credentials.Input(registry)
			if err != nil {
				helpers.PrintAndExit(err, 1)
			}
		}
	}

	credentials, exists := cli.Credentials.Find(registry)
	if !exists {
		err = cli.Credentials.Input(registry)
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	}

	OCIClient, err := packer.NewClient(credentials)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ocipath := filepath.Join("/", "tmp", "simplecontainer", repository)

	if viper.GetString("sign") != "" {
		cli.Signer.PrivateKeyPath = viper.GetString("sign")
		cli.Signer.SignerName = viper.GetString("author")
		cli.Signer.SignerEmail = viper.GetString("email")
	}

	pack, err := OCIClient.CreatePackage(context.Background(), cli.Signer, repository, ocipath)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = OCIClient.UploadPackage(context.Background(), ocipath, repository, pack.Version)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("pack pushed successfully")
}
func cmdPackPushFlags(cmd *cobra.Command) {
	cmd.Flags().String("registry", ocicredentials.DefaultRegistry, "Registry for packs")
	cmd.Flags().String("sign", "", "Path to the private key for signing")
	cmd.Flags().String("author", "", "Signing author")
	cmd.Flags().String("email", "", "Signing author email")
}

func cmdPackPull(api iapi.Api, cli *client.Client, args []string) {
	registry := viper.GetString("registry")

	ref, err := reference.Parse(args[0])
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	named, ok := ref.(reference.Named)
	if !ok {
		helpers.PrintAndExit(err, 1)
	}

	repository := named.Name()

	tagged, ok := ref.(reference.Tagged)
	if !ok {
		helpers.PrintAndExit(err, 1)
	}

	tag := tagged.Tag()

	err = cli.Credentials.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cli.Credentials.Default(registry)
			err = cli.Credentials.Save()
			if err != nil {
				helpers.PrintAndExit(err, 1)
			}
		}
	}

	credentials, exists := cli.Credentials.Find(registry)
	if !exists {
		err = cli.Credentials.Input(registry)
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	}

	OCIClient, err := packer.NewClient(credentials)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	pack, err := OCIClient.DownloadPackage(context.Background(), repository, tag, filepath.Join(".", repository))
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("pack pulled successfully: ", pack.Name, pack.Version)
}
func cmdPackPullFlags(cmd *cobra.Command) {
	cmd.Flags().String("registry", ocicredentials.DefaultRegistry, "Registry for packs")
}
