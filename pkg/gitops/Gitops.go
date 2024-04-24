package gitops

import "smr/pkg/definitions"

func NewWatcher(gitops definitions.Gitops) *Gitops {

	return &Gitops{
		RepoURL:         gitops.Spec.RepoURL,
		Revision:        gitops.Spec.Revision,
		DirectoryPath:   gitops.Spec.DirectoryPath,
		PoolingInterval: gitops.Spec.PoolingInterval,
		CertKeyRef:      gitops.Spec.CertKeyRef,
		HttpAuthRef:     gitops.Spec.HttpAuthRef,
		CertKey:         nil,
		HttpAuth:        nil,
	}
}

func (gitops *Gitops) RunWatcher() {

}
