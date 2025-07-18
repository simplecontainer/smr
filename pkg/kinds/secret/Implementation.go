package secret

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

func (secret *Secret) Start() error {
	secret.Started = true
	return nil
}

func (secret *Secret) GetShared() ishared.Shared {
	return secret.Shared
}

func (secret *Secret) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_SECRET, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(secret.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		events.DispatchGroup([]events.Event{
			events.NewKindEvent(events.EVENT_CHANGED, request.Definition, nil),
			events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
		}, secret.Shared, request.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}
}
func (secret *Secret) Replay(user *authentication.User) (iresponse.Response, error) {
	return iresponse.Response{}, nil
}
func (secret *Secret) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_SECRET, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(secret.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}

func (secret *Secret) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_SECRET, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(secret.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	} else {
		events.DispatchGroup([]events.Event{
			events.NewKindEvent(events.EVENT_DELETED, request.Definition, nil),
			events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
		}, secret.Shared, request.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, "object deleted", nil, nil), nil
	}
}

func (secret *Secret) Event(event ievents.Event) error {
	switch event.GetType() {
	case events.EVENT_CHANGE:
		err := secret.Shared.Manager.KindsRegistry[static.KIND_CONTAINERS].Event(event)
		if err != nil {
			return err
		}

		return nil
	default:
		return nil
	}
}
