package readiness

import "context"

type Readiness struct {
	Name       string
	Operator   string
	Timeout    string
	Body       map[string]string
	Solved     bool
	BodyUnpack map[string]string  `json:"-"`
	Function   func() error       `json:"-"`
	Ctx        context.Context    `json:"-"`
	Cancel     context.CancelFunc `json:"-"`
}

type ReadinessState struct {
	State int8
}

const CHECKING = 0
const SUCCESS = 1
const FAILED = 2

type ReadinessResult struct {
	Data string
}
