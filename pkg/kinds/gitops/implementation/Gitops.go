package implementation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
	"os"
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
		Gitops: &GitopsInternal{
			Git:             internal.NewGit(definition, logpath),
			LogPath:         logpath,
			PatchQueue:      NewQueueTS(),
			DirectoryPath:   definition.Spec.DirectoryPath,
			PoolingInterval: duration,
			AutomaticSync:   definition.Spec.AutomaticSync,
			Context:         definition.Spec.Context,
			Pack:            packer.New(),
			Node:            node.NewNodeDefinition(config.KVStore.Cluster, config.KVStore.Node.NodeID),
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
			definition: definition,
		},
	}

	return gitops
}

func (gitops *Gitops) Commit(logger *zap.Logger, client *clients.Http, user *authentication.User, commit *Commit) error {
	clone := definitions.New(commit.Format.GetKind())

	if clone.Definition == nil {
		return errors.New(fmt.Sprintf("kind is not defined as definition: %s", commit.Format.GetKind()))
	}

	clone.Definition.GetMeta().SetName(commit.Format.GetName())
	clone.Definition.GetMeta().SetGroup(commit.Format.GetGroup())

	for _, def := range gitops.Gitops.Pack.Definitions {
		if def.Definition.Definition.IsOf(def.Definition.Definition) {
			bytes, err := def.Definition.Definition.ToJSON()

			if err != nil {
				return err
			}

			err = clone.FromJson(bytes)
			if err != nil {
				return err
			}

			err = clone.PatchJSON(commit.Patch)
			if err != nil {
				return err
			}

			clone.SetState(nil)
			clone.SetRuntime(nil)

			bytes, err = clone.ToYAML()
			if err != nil {
				return err
			}

			err = os.WriteFile(fmt.Sprintf("%s/%s/definitions/%s", gitops.Gitops.Git.Directory, gitops.Gitops.DirectoryPath, def.File), bytes, 0777)
			if err != nil {
				return err
			}

			err = gitops.Gitops.Git.CommitFiles("gitops bot update", []string{fmt.Sprintf("%s/definitions/%s", gitops.Gitops.DirectoryPath, def.File)})
			if err != nil {
				return err
			}

			return gitops.Gitops.Git.Push()
		}
	}

	return errors.New("definition not found")
}

func (gitops *Gitops) Sync(logger *zap.Logger, client *clients.Http, user *authentication.User) ([]*common.Request, []error) {
	var requests = make([]*common.Request, 0)
	var errs = make([]error, 0)

	for k, request := range gitops.Gitops.Pack.Definitions {
		logger.Info("syncing object", zap.String("object", request.Definition.Definition.GetMeta().Name))

		request.Definition.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.GetDefinition().GetMeta().Group, gitops.GetDefinition().GetMeta().Name)

		action := request.Definition.Definition.GetState().GetOpt("action").Value
		request.Definition.Definition.GetState().ClearOpt("action")

		switch action {
		default:
			request.Definition.ProposeApply(client.Clients[user.Username].Http, client.Clients[user.Username].API)
			logger.Info("object proposed for apply", zap.String("object", request.Definition.Definition.GetMeta().Name))
			break
		case static.STATE_KIND:
			request.Definition.ProposeState(client.Clients[user.Username].Http, client.Clients[user.Username].API)
			logger.Info("object proposed for apply", zap.String("object", request.Definition.Definition.GetMeta().Name))
			break
		case static.REMOVE_KIND:
			request.Definition.ProposeRemove(client.Clients[user.Username].Http, client.Clients[user.Username].API)
			logger.Info("object proposed for remove", zap.String("object", request.Definition.Definition.GetMeta().Name))

			gitops.Gitops.Pack.Definitions = helpers.RemoveElement(gitops.Gitops.Pack.Definitions, k)
			break
		}
	}

	return requests, errs
}

func (gitops *Gitops) SyncState(logger *zap.Logger, client *clients.Http, user *authentication.User) ([]*common.Request, []error) {
	var requests = make([]*common.Request, 0)
	var errs = make([]error, 0)

	for _, request := range gitops.Gitops.Pack.Definitions {
		logger.Info("syncing object", zap.String("object", request.Definition.Definition.GetMeta().Name))

		request.Definition.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.GetDefinition().GetMeta().Group, gitops.GetDefinition().GetMeta().Name)

		action := request.Definition.Definition.GetState().GetOpt("action").Value
		request.Definition.Definition.GetState().ClearOpt("action")

		switch action {
		case static.STATE_KIND:
			request.Definition.ProposeState(client.Clients[user.Username].Http, client.Clients[user.Username].API)
			logger.Info("object proposed for apply", zap.String("object", request.Definition.Definition.GetMeta().Name))
			break
		}
	}

	return requests, errs
}

func (gitops *Gitops) Drift(client *clients.Http, user *authentication.User) (bool, []error) {
	var flagDrift bool
	var flagError bool
	var errs = make([]error, 0)

	for _, request := range gitops.Gitops.Pack.Definitions {
		if !request.Definition.Definition.GetState().GetOpt("action").IsEmpty() {
			if request.Definition.Definition.GetState().GetOpt("action").Value == static.REMOVE_KIND {
				continue
			}
		}

		request.Definition.Definition.GetRuntime().SetOwner(static.KIND_GITOPS, gitops.GetDefinition().GetMeta().Group, gitops.GetDefinition().GetMeta().Name)
		request.Definition.Definition.GetRuntime().SetNode(gitops.GetDefinition().GetRuntime().GetNode())
		request.Definition.Definition.GetRuntime().SetNodeName(gitops.GetDefinition().GetRuntime().GetNodeName())

		obj, err := request.Definition.Compare(client, user)

		if err != nil {
			if err.Error() == static.RESPONSE_NOT_FOUND {
				request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_MISSING, true)
				flagDrift = true
			} else {
				request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_ERROR, true)
				request.Definition.Definition.GetState().Gitops.AddError(err)

				errs = append(errs, err)

				flagError = true
			}

			request.Definition.ProposeState(client.Clients[user.Username].Http, client.Clients[user.Username].API)
		}

		if obj.ChangeDetected() {
			// we want to ignore meta runtime information since it doesn't affect change
			var changes []jsondiff.Operation

			for _, change := range obj.GetDiff() {
				if strings.HasPrefix(change.Path, "/meta/runtime/owner") {
					c := definitions.New(request.Definition.Definition.GetKind())

					err = c.FromJson(obj.GetDefinitionByte())

					if err != nil {
						errs = append(errs, err)
						continue
					}

					if !c.GetRuntime().GetOwner().IsEqual(request.Definition.Definition.GetRuntime().GetOwner()) {
						// Take ownership if no owner is defined: gitops > empty owner
						if !c.GetRuntime().GetOwner().IsEmpty() {
							request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_NOTOWNER, true)

							err = errors.New(fmt.Sprintf("owner of the object is %s", request.Definition.Definition.GetRuntime().GetOwner()))
							request.Definition.Definition.GetState().Gitops.AddError(err)
							errs = append(errs, err)
						}
					}
				} else {
					if strings.HasPrefix(change.Path, "/meta/runtime/") || strings.HasPrefix(change.Path, "/state/") {
						continue
					} else {
						fmt.Println("change", change)
						changes = append(changes, change)
					}
				}
			}

			if request.Definition.Definition.GetState().Gitops.NotOwner {
				flagError = true
				err = errors.New(fmt.Sprintf("owner of the object is %s", request.Definition.Definition.GetRuntime().GetOwner()))
				request.Definition.Definition.GetState().Gitops.AddError(err)
			} else {
				if len(changes) > 0 {
					request.Definition.Definition.GetState().Gitops.AddMessage("warning", "object is drifted")
					request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_DRIFTED, true)
					request.Definition.Definition.GetState().Gitops.Changes = changes

					flagDrift = true
				} else {
					request.Definition.Definition.GetState().Gitops.AddMessage("success", "object synced successfully")
					request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_SYNCED, true)
				}

				request.Definition.Definition.GetState().AddOpt("action", static.STATE_KIND)
				request.Definition.ProposeState(client.Clients[user.Username].Http, client.Clients[user.Username].API)
			}
		} else {
			if obj.Exists() {
				c := definitions.New(request.Definition.Definition.GetKind())

				err = c.FromJson(obj.GetDefinitionByte())

				if err != nil {
					errs = append(errs, err)
					continue
				}

				if c.GetRuntime().GetOwner().IsEqual(request.Definition.Definition.GetRuntime().GetOwner()) {
					request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_SYNCED, true)
					request.Definition.Definition.GetState().Gitops.Commit = gitops.Gitops.Commit.Hash
					request.Definition.Definition.GetState().Gitops.AddMessage("success", "object synced successfully")

					if request.Definition.Definition.GetState().Gitops.LastSync.IsZero() {
						request.Definition.Definition.GetState().Gitops.LastSync = time.Now()
					}

					request.Definition.Definition.GetState().AddOpt("action", static.STATE_KIND)
					request.Definition.ProposeState(client.Clients[user.Username].Http, client.Clients[user.Username].API)
				}
			} else {
				request.Definition.Definition.GetState().Gitops.AddMessage("neutral", "object is not found on the cluster")
				request.Definition.Definition.GetState().Gitops.Set(commonv1.GITOPS_MISSING, true)
				request.Definition.ProposeState(client.Clients[user.Username].Http, client.Clients[user.Username].API)
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

func (gitops *Gitops) Update(pack *packer.Pack) error {
	var err error

	for _, req := range pack.Definitions {
		for k, definition := range gitops.Gitops.Pack.Definitions {
			if definition.Definition.Definition.IsOf(req.Definition.Definition) {
				err = req.Definition.Definition.Patch(gitops.Gitops.Pack.Definitions[k].Definition.Definition)
				req.Definition.Definition.SetState(gitops.Gitops.Pack.Definitions[k].Definition.Definition.GetState())

				if err != nil {
					definition.Definition.Definition.GetState().Gitops.AddError(err)
				} else {
					if !definition.Definition.Definition.IsOf(req.Definition.Definition) {
						gitops.Gitops.Pack.Definitions[k].Definition.Definition.GetState().AddOpt("action", "remove")

						req.Definition.Definition.GetState().AddOpt("action", "apply")
						gitops.Gitops.Pack.Definitions = append(gitops.Gitops.Pack.Definitions, req)
					} else {
						req.Definition.Definition.GetState().AddOpt("action", "apply")
						gitops.Gitops.Pack.Definitions[k] = req
					}
				}
			}
		}

		if req.Definition.Definition.GetState().GetOpt("action").IsEmpty() {
			req.Definition.Definition.GetState().AddOpt("action", "apply")
			gitops.Gitops.Pack.Definitions = append(gitops.Gitops.Pack.Definitions, req)
		}
	}

	for k, definition := range gitops.Gitops.Pack.Definitions {
		missing := true

		for _, req := range pack.Definitions {
			if definition.Definition.Definition.IsOf(req.Definition.Definition) {
				missing = false
			}
		}

		if missing {
			gitops.Gitops.Pack.Definitions[k].Definition.Definition.GetState().AddOpt("action", "remove")
		}
	}

	return err
}

func (gitops *Gitops) ShouldSync() bool {
	return gitops.GetAutoSync() || gitops.GetForceSync()
}

func (gitops *Gitops) ToJSON() ([]byte, error) {
	return json.Marshal(gitops)
}
