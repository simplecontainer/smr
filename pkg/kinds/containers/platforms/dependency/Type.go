package dependency

import (
	"context"
)

type Dependency struct {
	Prefix   string
	Name     string
	Group    string
	Timeout  string
	Ctx      context.Context
	Function func() error
	Cancel   context.CancelFunc
}

type State struct {
	State int8
	Error error
}

const CHECKING = int8(0)
const SUCCESS = int8(1)
const FAILED = int8(2)
const CANCELED = int8(3)

type Result struct {
	Data string
}
