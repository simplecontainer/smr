package common

import (
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"net/http"
)

func BuildFromRequest(kind string, bytes []byte) (*Request, iresponse.Response) {
	request, err := NewRequest(kind)

	if err != nil {
		return nil, Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	if err = request.Definition.FromJson(bytes); err != nil {
		return nil, Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	valid, err := request.Definition.Validate()

	if !valid {
		return nil, Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	return request, iresponse.Response{}
}
