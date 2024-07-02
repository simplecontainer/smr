package status

import (
	"github.com/hmdsefi/gograph"
	"time"
)

type Status struct {
	State        string
	StateMachine gograph.Graph[string]
	Reconciling  bool
	LastUpdate   time.Time
}

type Statuses struct {
	Created         bool
	Readiness       bool
	ReadinessFailed bool
	DependsSolving  bool
	DependsFailed   bool
	DependsSolved   bool
	BackOffRestart  bool
	Ready           bool
	Running         bool
	Dead            bool
	Reconciling     bool
	DefinitionDrift bool
	PendingDelete   bool
}

const STATUS_CREATED string = "created"
const STATUS_DEPENDS_SOLVED string = "depends_solved"
const STATUS_DEPENDS_SOLVING string = "depends_solving"
const STATUS_DEPENDS_FAILED string = "depends_failed"
const STATUS_RUNNING string = "running"
const STATUS_RECONCILING string = "reconciling"
const STATUS_DRIFTED string = "drifted"
const STATUS_DEAD string = "dead"
const STATUS_READINESS string = "readiness_check"
const STATUS_BACKOFF string = "backoff"
const STATUS_READY string = "readiness_ready"
const STATUS_READINESS_FAILED string = "readiness_failed"
const STATUS_PENDING_DELETE string = "pending_delete"
const STATUS_KILLED string = "killed"
