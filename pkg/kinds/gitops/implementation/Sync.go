package implementation

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"strings"
)

func (gitops *Gitops) Sync(logger *zap.Logger, client *client.Http, user *authentication.User, definitionsOrdered []FileKind) error {
	for _, file := range definitionsOrdered {
		name := file.File

		logger.Info("syncing object", zap.String("object", name))

		definition, err := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, name))

		if err != nil {
			return err
		}

		request, err := common.NewRequest(file.Kind)
		request.Definition.FromJson(definition)
		request.Definition.SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)

		var bytes []byte
		bytes, err = request.Definition.ToJson()

		if err != nil {
			return err
		}

		response := gitops.sendRequest(client, user, fmt.Sprintf("https://localhost:1443/api/v1/apply/%s", request.Definition.GetKind()), bytes)

		if !response.Success {
			if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
				return errors.New(response.ErrorExplanation)
			} else {
				logger.Info(fmt.Sprintf(response.ErrorExplanation))
			}
		}

		logger.Info("object synced", zap.String("object", name))
	}

	return nil
}
