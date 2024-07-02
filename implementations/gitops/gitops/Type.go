package gitops

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/simplecontainer/smr/implementations/gitops/certkey"
	"github.com/simplecontainer/smr/implementations/gitops/httpauth"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Gitops struct {
	RepoURL          string
	Revision         string
	DirectoryPath    string
	PoolingInterval  string
	AutomaticSync    bool
	InSync           bool
	CertKeyRef       v1.CertKeyRef
	HttpAuthRef      v1.HttpauthRef
	LastSyncedCommit plumbing.Hash
	CertKey          *certkey.CertKey   `json:"-"`
	HttpAuth         *httpauth.HttpAuth `json:"-"`
	Definition       *v1.Gitops
}

type Event struct {
	Event string
}
