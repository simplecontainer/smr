package node

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/static"
)

func Clean() {
	container, err := Container()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	defer func() {
		err = container.Delete()

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println("node deleted")
	}()

	err = container.Stop(static.SIGTERM)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = container.Wait("removed")

	fmt.Println("node stopped")
}
