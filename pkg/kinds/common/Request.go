package common

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
	"strings"
)

func NewRequest(kind string) (*Request, error) {
	request := &Request{
		Definition: definitions.New(kind),
	}

	if request.Definition == nil {
		return nil, errors.New(fmt.Sprintf("kind is not defined as definition %s", kind))
	}

	return request, nil
}

func (request *Request) Load() {

}

func (request *Request) Apply(client *client.Http, user *authentication.User) error {
	bytes, err := request.Definition.ToJson()

	if err != nil {
		return err
	}

	response := network.Send(client.Clients[user.Username].Http, fmt.Sprintf("https://%s/api/v1/apply/%s", client.Clients[user.Username].API, request.Definition.GetKind()), http.MethodPost, bytes)

	if !response.Success {
		if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
			err = errors.New(response.ErrorExplanation)
		}
	}

	return err
}

func (request *Request) Delete(client *client.Http, user *authentication.User) error {
	bytes, err := request.Definition.ToJson()

	if err != nil {
		return err
	}

	response := network.Send(client.Clients[user.Username].Http, fmt.Sprintf("https://%s/api/v1/delete/%s", client.Clients[user.Username].API, request.Definition.GetKind()), http.MethodPost, bytes)

	if !response.Success {
		if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
			err = errors.New(response.ErrorExplanation)
		}
	}

	return err
}
