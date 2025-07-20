package custom

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (custom *Custom) Start() error {
	custom.Started = true
	return nil
}
func (custom *Custom) GetShared() ishared.Shared {
	return custom.Shared
}

func (custom *Custom) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CUSTOM, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(custom.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		events.DispatchGroup([]events.Event{
			events.NewKindEvent(events.EVENT_CHANGED, request.Definition, nil),
			events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
		}, custom.Shared, request.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}
}
func (custom *Custom) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CUSTOM, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(custom.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}
func (custom *Custom) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CUSTOM, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(custom.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	} else {
		events.DispatchGroup([]events.Event{
			events.NewKindEvent(events.EVENT_DELETED, request.Definition, nil),
			events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
		}, custom.Shared, request.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}

func (custom *Custom) Event(event ievents.Event) error {
	return nil
}
