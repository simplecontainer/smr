package implementation

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"time"
)

type Gitops struct {
	Git             *internal.Git
	LogPath         string
	DirectoryPath   string
	PoolingInterval time.Duration
	LastPoll        time.Time
	ForcePoll       bool
	AutomaticSync   bool
	DoSync          bool
	Commit          *object.Commit
	Status          *status.Status
	Auth            *Auth
	AuthResolved    transport.AuthMethod `json:"-"`
	Context         string
	Definition      *v1.GitopsDefinition
	Definitions     []common.Request
	Ghost           bool
}

type Auth struct {
	CertKeyRef  v1.GitopsCertKeyRef
	HttpAuthRef v1.GitopsHttpauthRef
}

type FileKind struct {
	File string
	Kind string
}
