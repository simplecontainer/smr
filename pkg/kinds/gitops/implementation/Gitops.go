package implementation

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"path"
	"time"
)

func New(definition *v1.GitopsDefinition) *Gitops {
	fmt.Println(definition)
	fmt.Println(definition.Spec.DirectoryPath)

	gitops := &Gitops{
		RepoURL:         definition.Spec.RepoURL,
		Revision:        definition.Spec.Revision,
		DirectoryPath:   definition.Spec.DirectoryPath,
		PoolingInterval: definition.Spec.PoolingInterval,
		AutomaticSync:   definition.Spec.AutomaticSync,
		API:             definition.Spec.API,
		Context:         definition.Spec.Context,
		Commit:          nil,
		Path:            fmt.Sprintf("/tmp/%s", path.Base(definition.Spec.RepoURL)),
		Status: &status.Status{
			State:            &status.StatusState{},
			LastUpdate:       time.Now(),
			Reconciling:      false,
			InSync:           false,
			LastSyncedCommit: plumbing.Hash{},
		},
		Auth: &Auth{
			CertKeyRef:  definition.Spec.CertKeyRef,
			HttpAuthRef: definition.Spec.HttpAuthRef,
		},
		Definition: definition,
	}

	gitops.Status.CreateGraph()

	if gitops.PoolingInterval == "" {
		gitops.PoolingInterval = "360s"
	}

	return gitops
}
