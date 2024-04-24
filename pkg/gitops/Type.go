package gitops

import (
	"smr/pkg/certkey"
	"smr/pkg/definitions"
	"smr/pkg/httpauth"
	"time"
)

type RepositoryWatchers struct {
	Repositories map[string]*Gitops
}

type Gitops struct {
	RepoURL         string        `json:"repoURL"`
	Revision        string        `json:"repoURL"`
	DirectoryPath   string        `json:"repoURL"`
	PoolingInterval time.Duration `json:"repoURL"`
	CertKey         *certkey.CertKey
	HttpAuth        *httpauth.HttpAuth
	CertKeyRef      definitions.CertKeyRef
	HttpAuthRef     definitions.HttpauthRef
}
