package helpers

import "github.com/simplecontainer/smr/pkg/logger"

func LogIfError(err error) {
	if err != nil {
		logger.Log.Error(err.Error())
	}
}
