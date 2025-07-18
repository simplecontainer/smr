package status

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hmdsefi/gograph"
	"time"
)

type Status struct {
	State            *StatusState                `json:"state"`
	Pending          *Pending                    `json:"pending"`
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

type Pending struct {
	Pending string
}

const PENDING_SYNC = "pending_sync"
const PENDING_DELETE = "delete"

const CATEGORY_PRERUN = 0
const CATEGORY_WHILERUN = 1
const CATEGORY_POSTRUN = 2
const CATEGORY_END = 3

const CREATED string = "created"
const SYNCING string = "syncing"
const SYNCING_STATE string = "syncing_state"
const BACKOFF string = "backoff"
const CLONING_GIT string = "cloning"
const COMMIT_GIT string = "pushing_changes"
const CLONED_GIT string = "cloned"
const INVALID_GIT string = "gitinvalid"
const INVALID_DEFINITIONS string = "definitionsinvalid"
const INSYNC string = "insync"
const DRIFTED string = "drifted"
const INSPECTING string = "inspecting"
const DELETE string = "pending_delete"
const ANOTHER_OWNER string = "not_owner"
