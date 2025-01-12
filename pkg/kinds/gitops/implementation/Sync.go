package implementation

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"go.uber.org/zap"
	"strings"
)

func (gitops *Gitops) Sync(logger *zap.Logger, client *client.Http, user *authentication.User, definitionsOrdered []map[string]string) error {
	for _, fileInfo := range definitionsOrdered {
		fileName := fileInfo["name"]

		logger.Info("syncing object", zap.String("object", fileName))

		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, fileName))

		response := gitops.sendRequest(client, user, "https://localhost:1443/api/v1/apply", definition)

		if !response.Success {
			if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
				return errors.New(response.ErrorExplanation)
			} else {
				logger.Info(fmt.Sprintf(response.ErrorExplanation))
			}
		}
	}

	return nil
}
