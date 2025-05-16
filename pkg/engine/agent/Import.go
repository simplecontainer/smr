package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"path/filepath"
	"time"
)

func Import(encrypted string, key string) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	cli := client.New(nil, environment.NodeDirectory)

	var err error
	cli.Context, err = cli.Manager.ImportContext(encrypted, key)
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

	if err := cli.Context.ImportCertificates(context.Background(), filepath.Join(environment.NodeDirectory, static.SSHDIR)); err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Printf("context '%s' successfully imported and set as active\n", cli.Context.Name)
}
