package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/mattn/go-shellwords"
)

type CmdSource interface {
	ToCmdArgs() ([]string, error)
	String() string
}

type StringCmd string

func NewStringCmd(command string, args ...any) StringCmd {
	return StringCmd(fmt.Sprintf(command, args))
}

func (s StringCmd) ToCmdArgs() ([]string, error) {
	if s == "" {
		return []string{}, nil
	}
	return shellwords.Parse(string(s))
}

func (s StringCmd) String() string {
	return string(s)
}

type SliceCmd []string

func (s SliceCmd) ToCmdArgs() ([]string, error) {
	return []string(s), nil
}

func (s SliceCmd) String() string {
	return strings.Join([]string(s), " ")
}

type EngineOptions struct {
	FailOnError bool
}

func DefaultEngineOptions() EngineOptions {
	return EngineOptions{
		FailOnError: true,
	}
}

type Engine struct {
	command []string
	cmd     *exec.Cmd
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	options EngineOptions
}

func NewEngine(binary string) *Engine {
	return NewEngineWithOptions(binary, DefaultEngineOptions())
}

func NewEngineWithOptions(binary string, options EngineOptions) *Engine {
	cmdParts, err := shellwords.Parse(binary)
	if err != nil {

		cmdParts = strings.Fields(binary)
	}

	return &Engine{
		command: cmdParts,
		stdout:  &bytes.Buffer{},
		stderr:  &bytes.Buffer{},
		cmd:     nil,
		options: options,
	}
}

func NewEngineFromSlice(cmdParts []string) *Engine {
	return NewEngineFromSliceWithOptions(cmdParts, DefaultEngineOptions())
}

func NewEngineFromSliceWithOptions(cmdParts []string, options EngineOptions) *Engine {
	return &Engine{
		command: cmdParts,
		stdout:  &bytes.Buffer{},
		stderr:  &bytes.Buffer{},
		cmd:     nil,
		options: options,
	}
}

func (e *Engine) SetFailOnError(failOnError bool) {
	e.options.FailOnError = failOnError
}

func (e *Engine) Run(t *testing.T, command ...CmdSource) error {
	var args []string
	var err error

	if len(command) > 0 {
		args, err = command[0].ToCmdArgs()
		if err != nil {
			if e.options.FailOnError {
				t.Fatalf("[ENGINE] Failed to parse command arguments: %v", err)
			} else {
				t.Errorf("[ENGINE] Failed to parse command arguments: %v", err)
			}
			return err
		}
	}

	e.stdout.Reset()
	e.stderr.Reset()

	fullCmd := append(append([]string{}, e.command...), args...)

	executable := fullCmd[0]
	var execArgs []string
	if len(fullCmd) > 1 {
		execArgs = fullCmd[1:]
	}

	cmd := exec.Command(executable, execArgs...)

	cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: "STDOUT"})
	cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: "STDERR"})

	cmdStr := ""
	if len(command) > 0 {
		cmdStr = command[0].String()
	}
	t.Logf("[ENGINE] Running command: %s %s", strings.Join(e.command, " "), cmdStr)

	err = cmd.Run()
	if err != nil {
		errorMsg := fmt.Sprintf("[ENGINE] Command failed: %v\nStdout: %s\nStderr: %s",
			err, e.stdout.String(), e.stderr.String())

		if e.options.FailOnError {
			t.Fatalf("%s", errorMsg)
		} else {
			t.Errorf("%s", errorMsg)
		}
		return err
	}

	return nil
}

func (e *Engine) RunAndCapture(t *testing.T, command ...CmdSource) (string, error) {
	var args []string
	var err error

	if len(command) > 0 {
		args, err = command[0].ToCmdArgs()
		if err != nil {
			if e.options.FailOnError {
				t.Fatalf("[ENGINE] Failed to parse command arguments: %v", err)
			} else {
				t.Errorf("[ENGINE] Failed to parse command arguments: %v", err)
			}
			return "", err
		}
	}

	e.stdout.Reset()
	e.stderr.Reset()

	fullCmd := append(append([]string{}, e.command...), args...)

	executable := fullCmd[0]
	var execArgs []string
	if len(fullCmd) > 1 {
		execArgs = fullCmd[1:]
	}

	cmd := exec.Command(executable, execArgs...)

	cmdStr := ""
	if len(command) > 0 {
		cmdStr = command[0].String()
	}

	writers := []io.Writer{e.stdout, testLogWriter{t: t, prefix: fmt.Sprintf("STDOUT %s", cmdStr)}}
	cmd.Stdout = io.MultiWriter(writers...)

	writers = []io.Writer{e.stderr, testLogWriter{t: t, prefix: fmt.Sprintf("STDERR %s", cmdStr)}}
	cmd.Stderr = io.MultiWriter(writers...)

	t.Logf("[ENGINE] Running command with capture: %s %s", strings.Join(e.command, " "), cmdStr)
	err = cmd.Run()
	if err != nil {
		errorMsg := fmt.Sprintf("[ENGINE] Command failed: %v\nStdout: %s\nStderr: %s",
			err, e.stdout.String(), e.stderr.String())

		if e.options.FailOnError {
			t.Fatalf("%s", errorMsg)
		} else {
			t.Errorf("%s", errorMsg)
		}
		return "", err
	}

	return e.stdout.String(), nil
}

func (e *Engine) RunBackground(t *testing.T, command ...CmdSource) error {
	if e.cmd == nil {
		var args []string
		var err error

		if len(command) > 0 {
			args, err = command[0].ToCmdArgs()
			if err != nil {
				if e.options.FailOnError {
					t.Fatalf("[ENGINE] Failed to parse command arguments: %v", err)
				} else {
					t.Errorf("[ENGINE] Failed to parse command arguments: %v", err)
				}
				return err
			}
		}

		e.stdout.Reset()
		e.stderr.Reset()

		fullCmd := append(append([]string{}, e.command...), args...)

		executable := fullCmd[0]
		var execArgs []string
		if len(fullCmd) > 1 {
			execArgs = fullCmd[1:]
		}

		cmd := exec.Command(executable, execArgs...)

		cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: "STDOUT"})
		cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: "STDERR"})

		cmdStr := ""
		if len(command) > 0 {
			cmdStr = command[0].String()
		}
		t.Logf("[ENGINE] Starting background command: %s %s", strings.Join(e.command, " "), cmdStr)

		err = cmd.Start()
		if err != nil {
			errorMsg := fmt.Sprintf("[ENGINE] Failed to start command: %v", err)
			if e.options.FailOnError {
				t.Fatalf("%s", errorMsg)
			} else {
				t.Errorf("%s", errorMsg)
			}
			return err
		}

		e.cmd = cmd
		return nil
	}

	errorMsg := "[ENGINE] Command already started once"
	if e.options.FailOnError {
		t.Fatalf("%s", errorMsg)
	} else {
		t.Errorf("%s", errorMsg)
	}
	return fmt.Errorf(errorMsg)
}

func (e *Engine) Stop(t *testing.T) error {
	if e.cmd != nil && e.cmd.Process != nil {
		t.Logf("[ENGINE] Stopping command process")
		err := e.cmd.Process.Signal(os.Signal(syscall.SIGTERM))
		if err != nil {
			errorMsg := fmt.Sprintf("Error stopping process: %v", err)
			if e.options.FailOnError {
				t.Fatalf("%s", errorMsg)
			} else {
				t.Errorf("%s", errorMsg)
			}
			return err
		}
		e.cmd = nil
		return nil
	}
	return nil
}

func (e *Engine) GetStdout() string {
	return e.stdout.String()
}

func (e *Engine) GetStderr() string {
	return e.stderr.String()
}

func (e *Engine) StopWithError() error {
	if e.cmd != nil && e.cmd.Process != nil {
		return e.cmd.Process.Signal(os.Signal(syscall.SIGTERM))
	}
	return nil
}

func (e *Engine) RunString(t *testing.T, args string) error {
	return e.Run(t, StringCmd(args))
}

func (e *Engine) RunAndCaptureString(t *testing.T, args string) (string, error) {
	return e.RunAndCapture(t, StringCmd(args))
}

func (e *Engine) RunBackgroundString(t *testing.T, args string) error {
	return e.RunBackground(t, StringCmd(args))
}

func (e *Engine) MustRunString(t *testing.T, args string) {
	err := e.RunString(t, args)
	if err != nil && !e.options.FailOnError {
		t.Fatalf("[ENGINE] Command must succeed but failed: %v", err)
	}
}

func (e *Engine) MustRunAndCaptureString(t *testing.T, args string) string {
	output, err := e.RunAndCaptureString(t, args)
	if err != nil && !e.options.FailOnError {
		t.Fatalf("[ENGINE] Command must succeed but failed: %v", err)
	}
	return output
}

type testLogWriter struct {
	t      *testing.T
	prefix string
}

func (w testLogWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\n"))
	for _, line := range lines {
		if len(line) > 0 {
			w.t.Logf("[%s] %s", w.prefix, string(line))
		}
	}
	return len(p), nil
}
