package implementation

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
)

func (gitops *Gitops) Drift(client *client.Http, user *authentication.User, definitionsOrdered []*common.Request) (bool, error) {
	for _, request := range definitionsOrdered {
		request.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
		_, err := request.Compare(client, user)

		if err != nil && errors.Is(err, errors.New("object changed")) {
			return true, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
