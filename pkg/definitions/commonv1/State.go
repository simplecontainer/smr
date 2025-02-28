package commonv1

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/wI2L/jsondiff"
	"time"
)

type State struct {
	Gitops  Gitops
	Options []*Opts
}

type Gitops struct {
	Synced           bool
	Drifted          bool
	Missing          bool
	NotOwner         bool
	Error            bool
	State            string
	ErrorDescription string
	Commit           plumbing.Hash
	Changes          []jsondiff.Operation
	LastSync         time.Time
}

func (g *Gitops) Clean() {
	g.Synced = false
	g.Drifted = false
	g.Missing = false
	g.NotOwner = false
}

func (g *Gitops) Set(state string, value bool) {
	g.Clean()

	switch state {
	case GITOPS_SYNCED:
		g.Synced = value
		break
	case GITOPS_DRIFTED:
		g.Drifted = value
		break
	case GITOPS_MISSING:
		g.Missing = value
		break
	case GITOPS_NOTOWNER:
		g.NotOwner = value
		break
	case GITOPS_ERROR:
		g.Error = value
		break
	}
}

func (g *Gitops) SetError(err error) {
	g.ErrorDescription = err.Error()
}

var GITOPS_SYNCED = "synced"
var GITOPS_DRIFTED = "drifted"
var GITOPS_MISSING = "missing"
var GITOPS_NOTOWNER = "notowner"
var GITOPS_ERROR = "error"

type Opts struct {
	Name  string
	Value string
}

func (opt *Opts) IsEmpty() bool {
	return opt.Name == "" && opt.Value == ""
}

func (state *State) AddOpt(name string, value string) {
	state.Options = append(state.Options, &Opts{name, value})
}

func (state *State) GetOpt(name string) *Opts {
	if state.Options != nil {
		for _, v := range state.Options {
			if v.Name == name {
				return v
			}
		}
	}

	return &Opts{}
}
