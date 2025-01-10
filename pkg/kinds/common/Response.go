package common

import (
	"github.com/simplecontainer/smr/pkg/contracts"
	"net/http"
)

func Response(status int, explanation string, err error) contracts.Response {
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	return contracts.Response{
		HttpStatus:       status,
		Explanation:      explanation,
		ErrorExplanation: errString,
		Success:          status == http.StatusOK,
	}
}
