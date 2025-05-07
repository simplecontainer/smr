package helpers

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/logger"
	"os"
)

func LogIfError(err error) {
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func PrintAndExit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("nil err passed to print")
	}

	os.Exit(code)
}
