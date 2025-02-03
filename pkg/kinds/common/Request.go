package common

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
	"strings"
)

func NewRequest(kind string) (*Request, error) {
	request := &Request{
		Definition: definitions.New(kind),
	}

	if request.Definition.Definition == nil {
		return nil, errors.New(fmt.Sprintf("kind is not defined as definition %s", kind))
	}

	return request, nil
}

func (request *Request) Apply(client *http.Client, API string) error {
	return request.Send("apply", http.MethodPost, client, API)
}

func (request *Request) Remove(client *http.Client, API string) error {
	return request.Send("delete", http.MethodDelete, client, API)
}

func (request *Request) Send(action string, method string, client *http.Client, API string) error {
	bytes, err := request.Definition.ToJson()

	if err != nil {
		return err
	}

	response := network.Send(client, fmt.Sprintf("https://%s/api/v1/definition/%s/%s", API, action, request.Definition.GetKind()), method, bytes)

	if !response.Success {
		if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
			err = errors.New(response.ErrorExplanation)
		}
	}

	return err
}
