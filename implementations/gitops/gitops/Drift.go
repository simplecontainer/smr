package gitops

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"net/http"
)

func (gitops *Gitops) Drift(client *client.Http, user *authentication.User, definitionsOrdered []map[string]string) (bool, error) {
	for _, fileInfo := range definitionsOrdered {
		fileName := fileInfo["name"]

		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, fileName))

		response := gitops.sendRequest(client, user, "https://localhost:1443/api/v1/apply", definition)

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
