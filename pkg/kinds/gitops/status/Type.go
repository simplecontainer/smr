package status

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hmdsefi/gograph"
	"time"
)

type Status struct {
	State                  *StatusState `json:"state"`
	PreviousState          *StatusState `json:"previousState"`
	LastReadiness          bool
	LastReadinessTimestamp time.Time
	StateMachine           gograph.Graph[*StatusState] `json:"-"`
	Reconciling            bool
	InSync                 bool
	LastSyncedCommit       plumbing.Hash
	LastUpdate             time.Time
}

type StatusState struct {
	State    string `json:"state"`
	category int8   `json:"category"`
}

const CATEGORY_PRERUN = 0
const CATEGORY_WHILERUN = 1
const CATEGORY_POSTRUN = 2
const CATEGORY_END = 3

const STATUS_CREATED string = "created"
const STATUS_SYNCING string = "syncing"
const STATUS_BACKOFF string = "backoff"
const STATUS_CLONING_GIT string = "cloning"
const STATUS_CLONED_GIT string = "cloned"
const STATUS_INVALID_GIT string = "gitinvalid"
const STATUS_INVALID_DEFINITIONS string = "definitionsinvalid"
const STATUS_INSYNC string = "insync"
const STATUS_DRIFTED string = "drifted"
const STATUS_INSPECTING string = "inspecting"
const STATUS_PENDING_DELETE string = "pending_delete"
