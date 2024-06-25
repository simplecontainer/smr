package gitops

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/qdnqn/smr/pkg/certkey"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpauth"
	"time"
)

type RepositoryWatcher struct {
	Repositories map[string]*Gitops
}

type Gitops struct {
	RepoURL          string
	Revision         string
	DirectoryPath    string
	PoolingInterval  string
	AutomaticSync    bool
	CertKeyRef       v1.CertKeyRef
	HttpAuthRef      v1.HttpauthRef
	LastSyncedCommit plumbing.Hash
	InSync           bool
	Definition       v1.Gitops
	CertKey          *certkey.CertKey   `json:"-"`
	HttpAuth         *httpauth.HttpAuth `json:"-"`
	GitopsQueue      chan Event         `json:"-"`
	Ctx              context.Context    `json:"-"`
	Ticker           *time.Ticker       `json:"-"`
}

const RESTART string = "restart"
const KILL string = "kill"
const STOP string = "stop"

type Event struct {
	Event string
}
