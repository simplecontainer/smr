package status

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hmdsefi/gograph"
	"time"
)

type Status struct {
	State            *StatusState                `json:"state"`
	PreviousState    *StatusState                `json:"previousState"`
	StateMachine     gograph.Graph[*StatusState] `json:"-"`
	Reconciling      bool
	PendingDelete    bool
	InSync           bool
	LastSyncedCommit plumbing.Hash
	LastUpdate       time.Time
}

type StatusState struct {
	State    string `json:"state"`
	category int8
}

const CATEGORY_PRERUN = 0
const CATEGORY_WHILERUN = 1
const CATEGORY_POSTRUN = 2
const CATEGORY_END = 3

const CREATED string = "created"
const SYNCING string = "syncing"
const BACKOFF string = "backoff"
const CLONING_GIT string = "cloning"
const CLONED_GIT string = "cloned"
const INVALID_GIT string = "gitinvalid"
const INVALID_DEFINITIONS string = "definitionsinvalid"
const INSYNC string = "insync"
const DRIFTED string = "drifted"
const INSPECTING string = "inspecting"
const PENDING_DELETE string = "pending_delete"
const ANOTHER_OWNER string = "not_owner"
