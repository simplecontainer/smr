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
	return StringCmd(fmt.Sprintf(command, args...))
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
	Suffix      string
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

func (e *Engine) prepareCommand(t *testing.T, command ...CmdSource) ([]string, string, error) {
	var args []string
	var err error
	cmdStr := ""

	if len(command) > 0 {
		args, err = command[0].ToCmdArgs()
		if err != nil {
			errorMsg := fmt.Sprintf("[ENGINE] Failed to parse command arguments: %v", err)
			if e.options.FailOnError {
				t.Fatalf("%s", errorMsg)
			} else {
				t.Errorf("%s", errorMsg)
			}
			return nil, "", err
		}
		cmdStr = command[0].String()
	}

	e.stdout.Reset()
	e.stderr.Reset()

	fullCmd := append(append([]string{}, e.command...), args...)

	if e.options.Suffix != "" {
		suffixParsed, err := shellwords.Parse(e.options.Suffix)
		if err != nil {
			return nil, "", err
		}
		fullCmd = append(fullCmd, suffixParsed...)
	}

	return fullCmd, cmdStr, nil
}

func (e *Engine) createCommand(fullCmd []string) *exec.Cmd {
	executable := fullCmd[0]
	var execArgs []string
	if len(fullCmd) > 1 {
		execArgs = fullCmd[1:]
	}

	return exec.Command(executable, execArgs...)
}

func (e *Engine) handleCommandError(t *testing.T, fullCmd []string, err error) error {
	if err != nil {
		errorMsg := fmt.Sprintf("[%s] Command failed: %v\nStdout: %s\nStderr: %s",
			strings.Join(fullCmd, " "), err, e.stdout.String(), e.stderr.String())

		if e.options.FailOnError {
			t.Fatalf("%s", errorMsg)
		} else {
			t.Logf("%s", errorMsg)
		}
	}
	return err
}

func (e *Engine) Run(t *testing.T, command ...CmdSource) error {
	fullCmd, cmdStr, err := e.prepareCommand(t, command...)
	if err != nil {
		return err
	}

	cmd := e.createCommand(fullCmd)

	cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: "STDOUT"})
	cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: "STDERR"})

	t.Logf("[ENGINE] Running command: %s %s %s", strings.Join(e.command, " "), cmdStr, e.options.Suffix)

	err = cmd.Run()
	return e.handleCommandError(t, fullCmd, err)
}

func (e *Engine) RunAndCapture(t *testing.T, command ...CmdSource) (string, error) {
	fullCmd, cmdStr, err := e.prepareCommand(t, command...)
	if err != nil {
		return "", err
	}

	cmd := e.createCommand(fullCmd)

	stdoutPrefix := fmt.Sprintf("STDOUT %s %s\n", cmdStr, e.options.Suffix)
	stderrPrefix := fmt.Sprintf("STDERR %s %s\n", cmdStr, e.options.Suffix)

	cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: stdoutPrefix})
	cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: stderrPrefix})

	t.Logf("[ENGINE] Running command with capture: %s %s", strings.Join(e.command, " "), cmdStr)

	err = cmd.Run()
	if e.handleCommandError(t, fullCmd, err) != nil {
		return "", err
	}

	return e.stdout.String(), nil
}

func (e *Engine) RunBackground(t *testing.T, command ...CmdSource) error {
	if e.cmd != nil {
		errorMsg := "[ENGINE] Command already started once"
		if e.options.FailOnError {
			t.Fatalf("%s", errorMsg)
		} else {
			t.Errorf("%s", errorMsg)
		}
		return fmt.Errorf(errorMsg)
	}

	fullCmd, cmdStr, err := e.prepareCommand(t, command...)
	if err != nil {
		return err
	}

	cmd := e.createCommand(fullCmd)

	stdoutPrefix := fmt.Sprintf("STDOUT %s %s\n", cmd.String(), e.options.Suffix)
	stderrPrefix := fmt.Sprintf("STDERR %s %s\n", cmd.String(), e.options.Suffix)

	cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: stdoutPrefix})
	cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: stderrPrefix})

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
