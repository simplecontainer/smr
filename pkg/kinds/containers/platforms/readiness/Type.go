package readiness

import "context"

type Readiness struct {
	Name            string
	Method          string
	URL             string
	CommandUnpacked []string
	Command         []string
	Type            string
	Timeout         string
	Solved          bool
	BodyUnpack      map[string]string  `json:"-"`
	Body            map[string]string  `json:"-"`
	Function        func() error       `json:"-"`
	Ctx             context.Context    `json:"-"`
	Cancel          context.CancelFunc `json:"-"`
}

type ReadinessState struct {
	State int8
}

const CHECKING = 0
const SUCCESS = 1
const FAILED = 2

const TYPE_URL = "url"
const TYPE_COMMAND = "command"

type ReadinessResult struct {
	Data string
}
