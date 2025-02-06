package implementation

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
)

func (gitops *Gitops) Sync(logger *zap.Logger, client *client.Http, user *authentication.User, definitionsOrdered []*common.Request) ([]*common.Request, error) {
	var requests = make([]*common.Request, 0)
	var err error

	for _, request := range definitionsOrdered {
		logger.Debug("syncing object", zap.String("object", request.Definition.GetMeta().Name))

		request.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
		request.Definition.GetRuntime().SetNode(gitops.Definition.GetRuntime().GetNode())

		err = request.ProposeApply(client.Clients[user.Username].Http, client.Clients[user.Username].API)

		if err != nil {
			return nil, err
		}

		logger.Debug("object synced", zap.String("object", request.Definition.GetMeta().Name))
	}

	return requests, err
}
