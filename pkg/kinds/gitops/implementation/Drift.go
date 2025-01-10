package implementation

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"net/http"
)

func (gitops *Gitops) Drift(client *client.Http, user *authentication.User, definitionsOrdered []FileKind) (bool, error) {
	for _, file := range definitionsOrdered {
		fileName := file.File

		definition, err := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, fileName))

		if err != nil {
			return true, err
		}

		response := gitops.sendRequest(client, user, "https://localhost:1443/api/v1/compare", definition)

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
