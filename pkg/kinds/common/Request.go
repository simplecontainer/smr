package common

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
	"strings"
)

func NewRequest(kind string) (*Request, error) {
	request := &Request{
		Definition: definitions.New(kind),
		DeleteC:    make(chan ievents.Event),
	}

	if request.Definition.Definition == nil {
		return nil, errors.New(fmt.Sprintf("kind is not defined as definition %s", kind))
	}

	return request, nil
}

func NewRequestFromJson(kind string, data []byte) (*Request, error) {
	request := &Request{
		Definition: definitions.New(kind),
	}

	if request.Definition.Definition == nil {
		return nil, errors.New(fmt.Sprintf("kind is not defined as definition %s", kind))
	}

	err := request.Definition.Definition.FromJson(data)

	if err != nil {
		return nil, err
	}

	return request, nil
}

func (request *Request) Apply(client *clients.Http, user *authentication.User) (iobjects.ObjectInterface, error) {
	return request.Action("apply", client, user)
}

func (request *Request) State(client *clients.Http, user *authentication.User) (iobjects.ObjectInterface, error) {
	return request.Action("state", client, user)
}

func (request *Request) Compare(client *clients.Http, user *authentication.User) (iobjects.ObjectInterface, error) {
	return request.Action("compare", client, user)
}

func (request *Request) Remove(client *clients.Http, user *authentication.User) (iobjects.ObjectInterface, error) {
	return request.Action("remove", client, user)
}

func (request *Request) Action(action string, client *clients.Http, user *authentication.User) (iobjects.ObjectInterface, error) {
	valid, err := request.Definition.Validate()

	if !valid {
		return nil, err
	}

	format := f.New(request.Definition.GetPrefix(), static.CATEGORY_KIND, request.Definition.GetKind(), request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
	obj := objects.New(client.Get(user.Username), user)

	switch action {
	case "apply":
		_, err = request.Definition.Apply(format, obj)
		break
	case "state":
		_, err = request.Definition.State(format, obj)
		break
	case "compare":
		_, err = request.Definition.Changed(format, obj)
		break
	case "remove":
		_, err = request.Definition.Delete(format, obj)
		break
	}

	return obj, err
}

func (request *Request) AttemptApply(client *http.Client, API string) error {
	return request.Send("attempt/apply", http.MethodPost, client, API)
}

func (request *Request) AttemptState(client *http.Client, API string) error {
	return request.Send("attempt/state", http.MethodPost, client, API)
}

func (request *Request) AttemptRemove(client *http.Client, API string) error {
	return request.Send("attempt/remove", http.MethodDelete, client, API)
}

// This is non-blocking and async method
func (request *Request) ProposeApply(client *http.Client, API string) error {
	return request.Send("propose/apply", http.MethodPost, client, API)
}

// This is non-blocking and async method
func (request *Request) ProposeRemove(client *http.Client, API string) error {
	return request.Send("propose/remove", http.MethodDelete, client, API)
}

// This is blocking and sync method
func (request *Request) ProposeState(client *http.Client, API string) error {
	return request.Send("propose/state", http.MethodPost, client, API)
}

func (request *Request) Send(action string, method string, client *http.Client, API string) error {
	bytes, err := request.Definition.ToJSON()

	if err != nil {
		return err
	}

	response := network.Send(client, fmt.Sprintf("%s/api/v1/%s", API, action), method, bytes)

	if !response.Success {
		if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
			err = errors.New(response.ErrorExplanation)
		}
	}

	return err
}
