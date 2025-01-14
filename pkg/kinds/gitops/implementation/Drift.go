package implementation

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (gitops *Gitops) Drift(client *client.Http, user *authentication.User, definitionsOrdered []FileKind) (bool, error) {
	for _, file := range definitionsOrdered {
		definition, err := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, file.File))

		if err != nil {
			return true, err
		}

		request, err := common.NewRequest(file.Kind)
		request.Definition.FromJson(definition)
		request.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)

		var bytes []byte
		bytes, err = request.Definition.ToJson()

		response := network.Send(client.Clients[user.Username].Http, "https://localhost:1443/api/v1/compare", http.MethodPost, bytes)

		switch response.HttpStatus {
		case http.StatusTeapot:
			return false, nil
			break
		case http.StatusBadRequest:
			return true, errors.New(response.ErrorExplanation)
			break
		}
	}

	return true, nil
}
