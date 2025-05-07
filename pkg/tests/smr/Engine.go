package smr

import (
	"os/exec"
	"testing"
)

type Engine struct {
	binary string
}

func NewEngine(binary string) *Engine {
	return &Engine{
		binary: binary,
	}
}

func (e *Engine) Create(t *testing.T, args ...string) {

}

func (e *Engine) Start(t *testing.T, args ...string) {
	cmd := exec.Command(e.binary, args...)

	//cmd.Env = append(os.Environ(), ...)
	//cmd.Stdout = ...
	//cmd.Stderr = ...

	cmd.Start()
}
func (e *Engine) Stop()  {}
func (e *Engine) Join()  {}
func (e *Engine) Leave() {}
