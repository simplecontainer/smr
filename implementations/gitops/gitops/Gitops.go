package gitops

import (
	"github.com/go-git/go-git/v5/plumbing"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

func New(gitopsObj *v1.GitopsDefinition) *Gitops {
	return &Gitops{
		RepoURL:          gitopsObj.Spec.RepoURL,
		Revision:         gitopsObj.Spec.Revision,
		DirectoryPath:    gitopsObj.Spec.DirectoryPath,
		PoolingInterval:  gitopsObj.Spec.PoolingInterval,
		AutomaticSync:    gitopsObj.Spec.AutomaticSync,
		InSync:           false,
		API:              gitopsObj.Spec.API,
		Context:          gitopsObj.Spec.Context,
		CertKeyRef:       gitopsObj.Spec.CertKeyRef,
		HttpAuthRef:      gitopsObj.Spec.HttpAuthRef,
		LastSyncedCommit: plumbing.Hash{},
		CertKey:          nil,
		HttpAuth:         nil,
		Definition:       gitopsObj,
	}
}
