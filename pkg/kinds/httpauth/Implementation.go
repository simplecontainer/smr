package httpauth

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (httpauth *Httpauth) Start() error {
	httpauth.Started = true
	return nil
}
func (httpauth *Httpauth) GetShared() interface{} {
	return httpauth.Shared
}

func (httpauth *Httpauth) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_HTTPAUTH, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(httpauth.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		events.DispatchGroup([]events.Event{
			events.NewKindEvent(events.EVENT_CHANGED, request.Definition, nil),
			events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
		}, httpauth.Shared, request.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}
}
func (httpauth *Httpauth) Replay(user *authentication.User) (iresponse.Response, error) {
	return iresponse.Response{}, nil
}
func (httpauth *Httpauth) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_HTTPAUTH, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(httpauth.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}
func (httpauth *Httpauth) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_HTTPAUTH, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(httpauth.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	} else {
		events.DispatchGroup([]events.Event{
			events.NewKindEvent(events.EVENT_DELETED, request.Definition, nil),
			events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
		}, httpauth.Shared, request.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}

func (httpauth *Httpauth) Event(event ievents.Event) error {
	return nil
}
