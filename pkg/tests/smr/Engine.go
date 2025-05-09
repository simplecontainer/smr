package smr

import (
	"bytes"
	"fmt"
	"github.com/mattn/go-shellwords"
	"io"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

type Engine struct {
	binary string
	node   string
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func NewEngine(binary string) *Engine {
	return &Engine{
		binary: binary,
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		cmd:    nil,
	}
}

func (e *Engine) Run(t *testing.T, args string) {
	parsed, err := shellwords.Parse(args)
	if err != nil {
		t.Error(err)
		return
	}

	e.stdout.Reset()
	e.stderr.Reset()

	cmd := exec.Command(e.binary, parsed...)

	cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: "STDOUT"})
	cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: "STDERR"})

	t.Logf("[ENGINE] running command: %s %s", e.binary, args)
	err = cmd.Run()
	if err != nil {
		t.Errorf("[ENGINE] command failed: %v\nStdout: %s\nStderr: %s", err, e.stdout.String(), e.stderr.String())
	}
}

func (e *Engine) RunAndCapture(t *testing.T, args string) string {
	parsed, err := shellwords.Parse(args)
	if err != nil {
		t.Error(err)
		return ""
	}

	e.stdout.Reset()
	e.stderr.Reset()

	cmd := exec.Command(e.binary, parsed...)

	writers := []io.Writer{e.stdout, testLogWriter{t: t, prefix: "STDOUT"}}
	cmd.Stdout = io.MultiWriter(writers...)

	writers = []io.Writer{e.stderr, testLogWriter{t: t, prefix: "STDERR"}}
	cmd.Stderr = io.MultiWriter(writers...)

	t.Logf("[ENGINE] running command: %s %s", e.binary, args)
	err = cmd.Run()
	if err != nil {
		t.Errorf("[ENGINE] command failed: %v\nStdout: %s\nStderr: %s", err, e.stdout.String(), e.stderr.String())
		return ""
	}

	return e.stdout.String()
}
func (e *Engine) RunBackground(t *testing.T, args string) {
	if e.cmd == nil {
		parsed, err := shellwords.Parse(args)

		if err != nil {
			t.Error(err)
			return
		}

		e.stdout.Reset()
		e.stderr.Reset()

		cmd := exec.Command(e.binary, parsed...)

		cmd.Stdout = io.MultiWriter(e.stdout, testLogWriter{t: t, prefix: "STDOUT"})
		cmd.Stderr = io.MultiWriter(e.stderr, testLogWriter{t: t, prefix: "STDERR"})

		t.Logf("[ENGINE] starting background command: %s %s", e.binary, args)

		err = cmd.Start()
		if err != nil {
			t.Errorf("[ENGINE] failed to start command: %v", err)
		}

		e.cmd = cmd
	} else {
		t.Errorf("[ENGINE] command already started once")
	}
}

func (e *Engine) Stop(t *testing.T) {
	if e.cmd != nil && e.cmd.Process != nil {
		t.Logf("[ENGINE] stopping command process")
		err := e.cmd.Process.Signal(os.Signal(syscall.SIGTERM))
		if err != nil {
			fmt.Println("Error stopping process:", err)
		}
	}
}

func (e *Engine) GetStdout() string {
	return e.stdout.String()
}

func (e *Engine) GetStderr() string {
	return e.stderr.String()
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
