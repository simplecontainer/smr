package events

type EventGroup struct {
	Events []Event
}

type Event struct {
	Type   string
	Target string
	Prefix string
	Kind   string
	Group  string
	Name   string
	Data   []byte
}

const EVENT_INSPECT = "inspect"
const EVENT_CHANGED = "changed"
const EVENT_CHANGE = "change"
const EVENT_RESTART = "restart"
const EVENT_DELETED = "deleted"
const EVENT_STOP = "stop"
const EVENT_RECREATE = "recreate"
const EVENT_SYNC = "sync"
const EVENT_REFRESH = "refresh"
const EVENT_DEPENDENCY = "refresh"

const EVENT_CONTROL_START = "upgrade_start"
const EVENT_CONTROL_FAILED = "upgrade_failed"
const EVENT_CONTROL_SUCCESS = "upgrade_success"
