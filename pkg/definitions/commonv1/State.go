package commonv1

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/wI2L/jsondiff"
	"sync"
	"time"
)

type State struct {
	Gitops  *Gitops       `json:"gitops,omitempty" yaml:"gitops,omitempty"`
	Options []*Opts       `json:"options,omitempty" yaml:"options,omitempty"`
	Lock    *sync.RWMutex `json:"-" yaml:"-"`
}

type Gitops struct {
	Synced   bool                 `json:"synced,omitempty" yaml:"synced,omitempty"`
	Drifted  bool                 `json:"drifted,omitempty" yaml:"drifted,omitempty"`
	Missing  bool                 `json:"missing,omitempty" yaml:"missing,omitempty"`
	NotOwner bool                 `json:"notOwner,omitempty" yaml:"notOwner,omitempty"`
	Error    bool                 `json:"error,omitempty" yaml:"error,omitempty"`
	State    string               `json:"state,omitempty" yaml:"state,omitempty"`
	Messages []Message            `json:"messages,omitempty" yaml:"messages,omitempty"`
	Commit   plumbing.Hash        `json:"commit,omitempty" yaml:"commit,omitempty"`
	Changes  []jsondiff.Operation `json:"changes,omitempty" yaml:"changes,omitempty"`
	LastSync time.Time            `json:"lastSync,omitempty" yaml:"lastSync,omitempty"`
}

type Message struct {
	Category  string    `json:"category,omitempty" yaml:"category,omitempty"`
	Message   string    `json:"message,omitempty" yaml:"message,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
}

func NewState() *State {
	return &State{
		Gitops:  &Gitops{},
		Options: make([]*Opts, 0),
		Lock:    &sync.RWMutex{},
	}
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

func (g *Gitops) AddMessage(category string, message string) {
	msg := Message{
		Category:  category,
		Message:   message,
		Timestamp: time.Now(),
	}

	if len(g.Messages) > 0 {
		last := g.Messages[len(g.Messages)-1]

		if last.Category == msg.Category && last.Message == msg.Message {
			g.Messages[len(g.Messages)-1] = msg
			return
		}
	}

	g.Messages = append(g.Messages, msg)
}
func (g *Gitops) AddError(err error) {
	msg := Message{
		Category:  "error",
		Message:   err.Error(),
		Timestamp: time.Now(),
	}

	if len(g.Messages) > 0 {
		last := g.Messages[len(g.Messages)-1]

		if last.Message == msg.Message {
			g.Messages[len(g.Messages)-1] = msg
			return
		}
	}

	g.Messages = append(g.Messages, msg)
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
	state.Lock.Lock()
	defer state.Lock.Unlock()

	state.clearOptUnsafe(name)
	state.Options = append(state.Options, &Opts{name, value})
}

func (state *State) GetOpt(name string) *Opts {
	state.Lock.RLock()
	defer state.Lock.RUnlock()

	if state.Options != nil {
		for _, v := range state.Options {
			if v.Name == name {
				return v
			}
		}
	}

	return &Opts{}
}

func (state *State) ClearOpt(name string) {
	state.Lock.Lock()
	defer state.Lock.Unlock()

	state.clearOptUnsafe(name)
}

func (state *State) clearOptUnsafe(name string) {
	if state.Options != nil {
		for k, v := range state.Options {
			if v.Name == name {
				state.Options = helpers.RemoveElement(state.Options, k)
			}
		}
	}
}
