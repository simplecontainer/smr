package status

import (
	"github.com/hmdsefi/gograph"
	"sync"
	"time"
)

type Status struct {
	QueueRejectOlderThan        time.Time
	State                       *State   `json:"state"`
	StateQueue                  []*State `json:"state_queue"`
	Pending                     *Pending `json:"pending"`
	LastReadiness               bool
	LastReadinessTimestamp      time.Time
	LastReadinessStarted        time.Time
	LastDependsSolved           bool
	LastDependsSolvedTimestamp  time.Time
	LastDependsStartedTimestamp time.Time
	StateMachine                gograph.Graph[*State] `json:"-"`
	LastUpdate                  time.Time
	mu                          sync.RWMutex `json:"-"` // Mutex for thread safety
}

type State struct {
	State         string    `json:"state"`
	PreviousState string    `json:"previous_state"`
	QueuedAt      time.Time `json:"queued_at"`
	category      int8      `json:"category"`
}

type Pending struct {
	Pending string
}

const PENDING_CREATE = "created"
const PENDING_DELETE = "delete"
const PENDING_RESTART = "restart"

const CATEGORY_PRERUN = 0
const CATEGORY_CLEAN = 0
const CATEGORY_WHILERUN = 1
const CATEGORY_POSTRUN = 2
const CATEGORY_END = 3

const INITIAL string = "initial"
const TRANSFERING string = "transfering"
const CHANGE string = "dependency_updated"
const CREATED string = "created"
const CLEAN string = "clean"
const PREPARE string = "prepare"
const INIT string = "init"
const INIT_FAILED string = "init_failed"
const DEPENDS_CHECKING string = "depends_checking"
const DEPENDS_SOLVED string = "depends_solved"
const READINESS_CHECKING string = "readiness_check"
const READY string = "readiness_ready"
const RUNNING string = "running"
const DEAD string = "dead"
const BACKOFF string = "backoff"
const DAEMON_FAILURE string = "daemon_failure"
const RESTART string = "restarting"

const DEPENDS_FAILED string = "depends_failed"
const READINESS_FAILED string = "readiness_failed"
const PENDING string = "pending"
const RUNTIME_PENDING string = "runtime_pending"

const START string = "start"
const KILL string = "kill"
const DELETE string = "delete"
