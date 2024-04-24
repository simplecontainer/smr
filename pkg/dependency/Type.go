package dependency

import (
	"context"
)

type Dependency struct {
	Name     string         `yaml:"name"`
	Operator string         `yaml:"operator"`
	Timeout  string         `yaml:"timeout"`
	Body     map[string]any `mapstructure:"body"`
	Solved   bool
	Ctx      context.Context
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
	Data string `json:"data"`
}
