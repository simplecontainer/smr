package node

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
)

func Clean() {
	container, err := Container()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = container.Clean()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("node container removed")
}
