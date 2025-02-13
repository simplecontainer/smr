package implementation

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
	"strings"
	"time"
)

func New(definition *v1.GitopsDefinition) *Gitops {
	format := f.New(definition.GetPrefix(), "kind", static.KIND_GITOPS, definition.GetMeta().Group, definition.GetMeta().Name)
	logpath := fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1))

	duration, err := time.ParseDuration(definition.Spec.PoolingInterval)

	if err != nil {
		// If invalid fallback to default of 5 minutes
		duration = time.Second * 360
	}

	gitops := &Gitops{
		Git:             internal.NewGit(definition, logpath),
		LogPath:         logpath,
		DirectoryPath:   definition.Spec.DirectoryPath,
		PoolingInterval: duration,
		AutomaticSync:   definition.Spec.AutomaticSync,
		Context:         definition.Spec.Context,
		Commit: &object.Commit{
			Hash:         plumbing.Hash{},
			Author:       object.Signature{},
			Committer:    object.Signature{},
			MergeTag:     "",
			PGPSignature: "",
			Message:      "",
			TreeHash:     plumbing.Hash{},
			ParentHashes: nil,
			Encoding:     "",
		},
		Status: status.New(),
		Auth: &Auth{
			CertKeyRef:  definition.Spec.CertKeyRef,
			HttpAuthRef: definition.Spec.HttpAuthRef,
		},
		Definition: definition,
	}

	return gitops
}

func (gitops *Gitops) Sync(logger *zap.Logger, client *client.Http, user *authentication.User, definitionsOrdered []*common.Request) ([]*common.Request, error) {
	var requests = make([]*common.Request, 0)
	var err error

	for _, request := range definitionsOrdered {
		logger.Info("syncing object", zap.String("object", request.Definition.GetMeta().Name))

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

func (gitops *Gitops) Drift(client *client.Http, user *authentication.User, definitionsOrdered []*common.Request) (bool, error) {
	for _, request := range definitionsOrdered {
		request.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
		request.Definition.GetRuntime().SetNode(gitops.Definition.GetRuntime().GetNode())

		obj, err := request.Compare(client, user)

		if err != nil {
			if err.Error() == static.STATUS_RESPONSE_NOT_FOUND {
				return true, nil
			} else {
				return false, err
			}
		}

		if obj.ChangeDetected() {
			// we want to ignore meta runtime information since it doesn't affect change
			var changes []jsondiff.Operation

			for _, change := range obj.GetDiff() {
				if !strings.HasPrefix(change.Path, "/meta/runtime/") {
					changes = append(changes, change)
				}
			}

			if len(changes) > 0 {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, nil
		}
	}

	return true, nil
}

func (gitops *Gitops) ShouldSync() bool {
	return gitops.AutomaticSync || gitops.DoSync
}

func (gitops *Gitops) ToJson() ([]byte, error) {
	return json.Marshal(gitops)
}
