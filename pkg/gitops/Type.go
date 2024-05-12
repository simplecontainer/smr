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
	CertKey          *certkey.CertKey
	HttpAuth         *httpauth.HttpAuth
	CertKeyRef       v1.CertKeyRef
	HttpAuthRef      v1.HttpauthRef
	GitopsQueue      chan Event
	Ctx              context.Context
	Ticker           *time.Ticker
	LastSyncedCommit plumbing.Hash
}

const KILL string = "kill"
const STOP string = "stop"

type Event struct {
	Event string
}
