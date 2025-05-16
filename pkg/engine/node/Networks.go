package node

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/spf13/viper"
)

func Networks() {
	container, err := Container()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	_, err = container.GetState()
	err = container.SyncNetwork()

	networks := container.GetNetwork()

	val, ok := networks[viper.GetString("network")]

	if ok {
		fmt.Println(val)
	}
}
