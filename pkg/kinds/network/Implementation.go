package network

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/network/implementation"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
	"time"
)

func (network *Network) Start() error {
	network.Started = true
	return nil
}

func (network *Network) GetShared() interface{} {
	return network.Shared
}

func (network *Network) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_NETWORK, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(network.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if request.Definition.GetRuntime().Node != network.Shared.Manager.Config.KVStore.Node {
		return common.Response(http.StatusOK, "networks are local scoped", err, nil), nil
	}

	var networkObj *implementation.Network

	if obj.Exists() {
		networkObj = implementation.New(obj.GetDefinitionByte())
	} else {
		networkObj = implementation.New(definition)
	}

	members, found, err := networkObj.Find()

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	}

	if found {
		if len(members) > 0 {
			return common.Response(http.StatusBadRequest, "", errors.New("disconnect all container from network and try again"), nil), err
		} else {
			err = networkObj.Remove()

			if err != nil {
				return common.Response(http.StatusInternalServerError, "", err, nil), err
			}

			for {
				select {
				case <-time.After(5 * time.Second):
					return common.Response(http.StatusInternalServerError, "", errors.New("network didn't delete properly"), nil), err
				case <-time.Tick(500 * time.Millisecond):
					err = networkObj.Create()

					if err == nil {
						return common.Response(http.StatusOK, "object applied", nil, nil), nil
					}
				}
			}
		}
	} else {
		err = networkObj.Create()
	}

	if err != nil {
		return common.Response(http.StatusInternalServerError, "internal error", err, nil), err
	}

	events.DispatchGroup([]events.Event{
		events.NewKindEvent(events.EVENT_CHANGED, request.Definition, nil),
		events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
	}, network.Shared, request.Definition.GetRuntime().GetNode())

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (network *Network) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_NETWORK, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(network.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}
func (network *Network) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_NETWORK, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(network.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if request.Definition.GetRuntime().Node != network.Shared.Manager.Config.KVStore.Node {
		return common.Response(http.StatusOK, "networks are local scoped", err, nil), nil
	}

	var networkObj *implementation.Network

	if obj.Exists() {
		networkObj = implementation.New(obj.GetDefinitionByte())
	} else {
		networkObj = implementation.New(definition)
	}

	members, found, err := networkObj.Find()

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	}

	if found {
		if len(members) > 0 {
			return common.Response(http.StatusBadRequest, "", errors.New("disconnect all container from network and try again"), nil), err
		} else {
			err = networkObj.Remove()
		}
	}

	if err != nil {
		return common.Response(http.StatusInternalServerError, "internal error", err, nil), err
	}

	events.DispatchGroup([]events.Event{
		events.NewKindEvent(events.EVENT_DELETED, request.Definition, nil),
		events.NewKindEvent(events.EVENT_INSPECT, request.Definition, nil),
	}, network.Shared, request.Definition.GetRuntime().GetNode())

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (network *Network) Event(event ievents.Event) error {
	return nil
}
