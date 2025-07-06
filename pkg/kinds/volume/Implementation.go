package volume

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/containers"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (volume *Volume) Start() error {
	volume.Started = true
	return nil
}
func (volume *Volume) GetShared() ishared.Shared {
	return volume.Shared
}

func (volume *Volume) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_VOLUME, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(volume.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if obj.ChangeDetected() {
		var container platforms.IContainer
		container, err = containers.NewEmpty(volume.Shared.Manager.Config.Platform)

		if err != nil {
			return common.Response(http.StatusInternalServerError, "", err, nil), err
		} else {
			err = container.CreateVolume(request.Definition.Definition.(*v1.VolumeDefinition))

			if err != nil {
				return common.Response(http.StatusInternalServerError, "", err, nil), err
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (volume *Volume) Replay(user *authentication.User) (iresponse.Response, error) {
	return iresponse.Response{}, nil
}

func (volume *Volume) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_VOLUME, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(volume.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}

func (volume *Volume) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_VOLUME, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(volume.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "", err, nil), err
	} else {
		var container platforms.IContainer
		container, err = containers.NewEmpty(volume.Shared.Manager.Config.Platform)

		err = container.DeleteVolume(fmt.Sprintf("%s-%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name), false)

		if err != nil {
			return common.Response(http.StatusInternalServerError, "", err, nil), err
		}

		return common.Response(http.StatusOK, "object deleted", nil, nil), nil
	}
}

func (volume *Volume) Event(event ievents.Event) error {
	return nil
}
