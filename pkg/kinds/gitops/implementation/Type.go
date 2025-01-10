package implementation

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"time"
)

type Gitops struct {
	RepoURL         string
	Revision        string
	DirectoryPath   string
	PoolingInterval string
	LastPoll        time.Time
	ForcePoll       bool
	AutomaticSync   bool
	ManualSync      bool
	Commit          *object.Commit
	Path            string
	Status          *status.Status
	Auth            *Auth
	AuthResolved    transport.AuthMethod `json:"-"`
	API             string
	Context         string
	Definition      *v1.GitopsDefinition
}

type Auth struct {
	CertKeyRef  v1.GitopsCertKeyRef
	HttpAuthRef v1.GitopsHttpauthRef
}

type FileKind struct {
	File string
	Kind string
}
