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
	AuthType         *AuthType
	CertKeyRef       v1.GitopsCertKeyRef
	HttpAuthRef      v1.GitopsHttpauthRef
	LastSyncedCommit plumbing.Hash
	CertKey          *certkey.CertKey   `json:"-"`
	HttpAuth         *httpauth.HttpAuth `json:"-"`
	Definition       *v1.GitopsDefinition
}

type AuthType struct {
	AuthType string
}

const httpAuthType string = "http"
const sshAuthType string = "ssh"

type Event struct {
	Event string
}
