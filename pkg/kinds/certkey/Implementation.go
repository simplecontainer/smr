package certkey

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (certkey *Certkey) Start() error {
	certkey.Started = true
	return nil
}
func (certkey *Certkey) GetShared() interface{} {
	return certkey.Shared
}

func (certkey *Certkey) Apply(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CERTKEY, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(certkey.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}
}

func (certkey *Certkey) Compare(user *authentication.User, definition []byte) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CERTKEY, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(certkey.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	} else {
		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}

func (certkey *Certkey) Delete(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CERTKEY, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(certkey.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "object deleted", nil, nil), nil
	}
}

func (certkey *Certkey) Event(event contracts.Event) error {
	return nil
}
