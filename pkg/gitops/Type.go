package gitops

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing"
	"smr/pkg/certkey"
	"smr/pkg/definitions"
	"smr/pkg/httpauth"
	"time"
)

type RepositoryWatcher struct {
	Repositories map[string]*Gitops
}

type Gitops struct {
	RepoURL          string `json:"repoURL"`
	Revision         string `json:"revision"`
	DirectoryPath    string `json:"directoryPath"`
	PoolingInterval  string `json:"poolingInterval"`
	CertKey          *certkey.CertKey
	HttpAuth         *httpauth.HttpAuth
	CertKeyRef       definitions.CertKeyRef
	HttpAuthRef      definitions.HttpauthRef
	AutomaticSync    bool
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
