package events

type Event struct {
	Type   string
	Target string
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
