package gitops

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/simplecontainer/smr/implementations/gitops/status"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"path"
	"time"
)

func New(gitopsObj *v1.GitopsDefinition) *Gitops {
	gitops := &Gitops{
		RepoURL:         gitopsObj.Spec.RepoURL,
		Revision:        gitopsObj.Spec.Revision,
		DirectoryPath:   gitopsObj.Spec.DirectoryPath,
		PoolingInterval: gitopsObj.Spec.PoolingInterval,
		AutomaticSync:   gitopsObj.Spec.AutomaticSync,
		API:             gitopsObj.Spec.API,
		Context:         gitopsObj.Spec.Context,
		Commit:          nil,
		Path:            fmt.Sprintf("/tmp/%s", path.Base(gitopsObj.Spec.RepoURL)),
		Status: &status.Status{
			State:            &status.StatusState{},
			LastUpdate:       time.Now(),
			Reconciling:      false,
			InSync:           false,
			LastSyncedCommit: plumbing.Hash{},
		},
		Auth: &Auth{
			CertKeyRef:  gitopsObj.Spec.CertKeyRef,
			HttpAuthRef: gitopsObj.Spec.HttpAuthRef,
		},
		AuthInternal: &AuthInternal{
			CertKey:  nil,
			HttpAuth: nil,
		},
		Definition: gitopsObj,
	}

	gitops.Status.CreateGraph()

	return gitops
}
