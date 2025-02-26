package implementation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
	"strings"
	"time"
)

func New(definition *v1.GitopsDefinition, config *configuration.Configuration) *Gitops {
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
		Node:            node.NewNodeDefinition(definition.GetRuntime(), config.KVStore.Cluster),
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

func (gitops *Gitops) Sync(logger *zap.Logger, client *client.Http, user *authentication.User) ([]*common.Request, []error) {
	var requests = make([]*common.Request, 0)
	var errs = make([]error, 0)

	for _, request := range gitops.Definitions {
		logger.Info("syncing object", zap.String("object", request.Definition.GetMeta().Name))

		request.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
		request.Definition.GetRuntime().SetNode(gitops.Definition.GetRuntime().GetNode())

		err := request.ProposeApply(client.Clients[user.Username].Http, client.Clients[user.Username].API)

		if err != nil {
			errs = append(errs, err)

			request.Definition.GetState().Gitops.Set(commonv1.GITOPS_DRIFTED, true)
			request.Definition.GetState().Gitops.SetError(err)
		} else {
			request.Definition.GetState().Gitops.Set(commonv1.GITOPS_SYNCED, true)
			request.Definition.GetState().Gitops.Commit = gitops.Commit.Hash
			request.Definition.GetState().Gitops.LastSync = time.Now()

			logger.Debug("object synced", zap.String("object", request.Definition.GetMeta().Name))
		}
	}

	return requests, errs
}

func (gitops *Gitops) Drift(client *client.Http, user *authentication.User) (bool, []error) {
	var flagDrift bool
	var flagError bool
	var errs = make([]error, 0)

	for _, request := range gitops.Definitions {
		request.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
		request.Definition.GetRuntime().SetNode(gitops.Definition.GetRuntime().GetNode())

		obj, err := request.Compare(client, user)

		if err != nil {
			if err.Error() == static.STATUS_RESPONSE_NOT_FOUND {
				request.Definition.GetState().Gitops.Set(commonv1.GITOPS_MISSING, true)
				flagDrift = true
			} else {
				request.Definition.GetState().Gitops.Set(commonv1.GITOPS_ERROR, true)
				request.Definition.GetState().Gitops.SetError(err)

				errs = append(errs, err)

				flagError = true
			}
		}

		if obj.ChangeDetected() {
			// we want to ignore meta runtime information since it doesn't affect change
			var changes []jsondiff.Operation

			for _, change := range obj.GetDiff() {
				if strings.HasPrefix(change.Path, "/meta/runtime/owner") {
					c := definitions.New(request.Definition.GetKind())

					err = c.FromJson(obj.GetDefinitionByte())

					if err != nil {
						errs = append(errs, err)
						continue
					}

					if !c.GetRuntime().GetOwner().IsEqual(request.Definition.GetRuntime().GetOwner()) {
						request.Definition.GetState().Gitops.Set(commonv1.GITOPS_NOTOWNER, true)

						err = errors.New(fmt.Sprintf("owner of the object is %s", request.Definition.GetRuntime().GetOwner()))
						request.Definition.GetState().Gitops.SetError(err)
						errs = append(errs, err)
					}
				} else {
					if strings.HasPrefix(change.Path, "/meta/runtime/") || strings.HasPrefix(change.Path, "/state/") {
						continue
					} else {
						changes = append(changes, change)
					}
				}
			}

			if request.Definition.GetState().Gitops.NotOwner {
				flagError = true
				request.Definition.GetState().Gitops.SetError(errors.New("someone else is owner of the object"))
			} else {
				if len(changes) > 0 {
					request.Definition.GetState().Gitops.Set(commonv1.GITOPS_DRIFTED, true)
					request.Definition.GetState().Gitops.Changes = changes

					flagDrift = true
				}
			}
		}
	}

	if flagError {
		return false, errs
	}

	if flagDrift {
		return true, nil
	}

	return false, nil
}

func (gitops *Gitops) Update(reqs []*common.Request) error {
	for _, req := range reqs {
		update := false

		for k, definition := range gitops.Definitions {
			if definition.Definition.IsOf(req.Definition) {
				err := req.Definition.Patch(gitops.Definitions[k].Definition)

				if err != nil {
					return err
				}

				gitops.Definitions[k] = req
				update = true
			}
		}

		if !update {
			gitops.Definitions = append(gitops.Definitions, req)
		}
	}

	return nil
}

func (gitops *Gitops) ShouldSync() bool {
	return gitops.AutomaticSync || gitops.DoSync
}

func (gitops *Gitops) ToJson() ([]byte, error) {
	return json.Marshal(gitops)
}
