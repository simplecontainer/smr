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
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func (gitops *Gitops) Sync(logger *zap.Logger, client *client.Http, user *authentication.User, definitionsOrdered []FileKind) ([]*common.Request, error) {
	var requests = make([]*common.Request, 0)
	var err error

	for _, file := range definitionsOrdered {
		logger.Debug("syncing object", zap.String("object", file.File))

		var definition []byte
		definition, err = definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, file.File))

		if err != nil {
			return requests, err
		}

		request, err := common.NewRequest(file.Kind)
		request.Definition.FromJson(definition)
		request.Definition.SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)

		requests = append(requests, request)

		var bytes []byte
		bytes, err = request.Definition.ToJson()

		if err != nil {
			return requests, err
		}

		response := network.Send(client.Clients[user.Username].Http, fmt.Sprintf("https://localhost:1443/api/v1/apply/%s", request.Definition.GetKind()), http.MethodPost, bytes)

		if !response.Success {
			if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
				err = errors.New(response.ErrorExplanation)
			} else {
				logger.Info(fmt.Sprintf(response.ErrorExplanation))
			}
		}

		logger.Debug("object synced", zap.String("object", file.File))
	}

	return requests, err
}
