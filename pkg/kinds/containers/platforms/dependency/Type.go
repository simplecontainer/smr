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

const CHECKING = 0
const SUCCESS = 1
const FAILED = 2
const CANCELED = 3

type Result struct {
	Data string
}
