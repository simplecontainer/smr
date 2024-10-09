package gitops

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
)

func (gitops *Gitops) Sync(client *client.Http, user *authentication.User, definitionsOrdered []map[string]string) error {
	for _, fileInfo := range definitionsOrdered {
		fileName := fileInfo["name"]

		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, fileName))

		response := gitops.sendRequest(client, user, "https://localhost:1443/api/v1/apply", definition)

		if !response.Success {
			return errors.New(response.ErrorExplanation)
		}
	}

	return nil
}
