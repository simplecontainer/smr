package implementation

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/packer"
	"time"
)

type Gitops struct {
	Git             *internal.Git
	Node            *node.Node
	LogPath         string
	DirectoryPath   string
	PoolingInterval time.Duration
	LastPoll        time.Time
	ForceClone      bool
	AutomaticSync   bool
	ForceSync       bool
	Commit          *object.Commit
	Status          *status.Status
	Auth            *Auth
	Context         string
	Definition      *v1.GitopsDefinition
	Pack            *packer.Pack
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
