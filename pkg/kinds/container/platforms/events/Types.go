package events

type Events struct {
	Type  string
	Kind  string
	Group string
	Name  string
	Data  []byte
}

const EVENT_CHANGE = "change"
const EVENT_RESTART = "restart"
const EVENT_DELETE = "delete"
const EVENT_STOP = "stop"
const EVENT_RECREATE = "recreate"
