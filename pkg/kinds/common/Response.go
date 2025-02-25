package common

import (
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"net/http"
)

func Response(status int, explanation string, err error, data []byte) iresponse.Response {
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	return iresponse.Response{
		HttpStatus:       status,
		Explanation:      explanation,
		ErrorExplanation: errString,
		Success:          status == http.StatusOK,
		Data:             data,
	}
}
