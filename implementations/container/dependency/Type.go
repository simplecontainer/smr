package dependency

import (
	"context"
)

type Dependency struct {
	Name    string
	Timeout string
	Ctx     context.Context
}

type State struct {
	Name       string
	Success    bool
	Missing    bool
	Timeout    bool
	Error      error
	TryToSolve bool
	Depend     *Dependency
}

type Result struct {
	Data string
}
