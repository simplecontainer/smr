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

type Logger interface {
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type TestLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t: t}
}

func (tl *TestLogger) Logf(format string, args ...interface{}) {
	tl.t.Logf(format, args...)
}

func (tl *TestLogger) Errorf(format string, args ...interface{}) {
	tl.t.Errorf(format, args...)
}

func (tl *TestLogger) Fatalf(format string, args ...interface{}) {
	tl.t.Fatalf(format, args...)
}

type Config struct {
	FailOnError bool
	Suffix      string
}

func DefaultConfig() Config {
	return Config{
		FailOnError: true,
	}
}

type CommandBuilder struct {
	baseCommand []string
	config      Config
}

func NewCommandBuilder(binary string, config Config) (*CommandBuilder, error) {
	cmdParts, err := shellwords.Parse(binary)
	if err != nil {
		cmdParts = strings.Fields(binary)
	}

	return &CommandBuilder{
		baseCommand: cmdParts,
		config:      config,
	}, nil
}

func NewCommandBuilderFromSlice(cmdParts []string, config Config) *CommandBuilder {
	return &CommandBuilder{
		baseCommand: cmdParts,
		config:      config,
	}
}

func (cb *CommandBuilder) BuildCommand(sources ...CmdSource) ([]string, string, error) {
	var args []string
	var cmdStr string

	if len(sources) > 0 {
		parsedArgs, err := sources[0].ToCmdArgs()
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse command arguments: %w", err)
		}
		args = parsedArgs
		cmdStr = sources[0].String()
	}

	fullCmd := append(append([]string{}, cb.baseCommand...), args...)

	if cb.config.Suffix != "" {
		suffixParsed, err := shellwords.Parse(cb.config.Suffix)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse suffix: %w", err)
		}
		fullCmd = append(fullCmd, suffixParsed...)
	}

	return fullCmd, cmdStr, nil
}

type OutputCapture struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func NewOutputCapture() *OutputCapture {
	return &OutputCapture{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
}

func (oc *OutputCapture) Reset() {
	oc.stdout.Reset()
	oc.stderr.Reset()
}

func (oc *OutputCapture) GetStdout() string {
	return oc.stdout.String()
}

func (oc *OutputCapture) GetStderr() string {
	return oc.stderr.String()
}

func (oc *OutputCapture) CreateWriters(logger Logger, cmdStr string, suffix string) (io.Writer, io.Writer) {
	stdoutWriter := io.MultiWriter(oc.stdout, &logWriter{
		logger: logger,
		prefix: fmt.Sprintf("STDOUT %s %s", cmdStr, suffix),
	})

	stderrWriter := io.MultiWriter(oc.stderr, &logWriter{
		logger: logger,
		prefix: fmt.Sprintf("STDERR %s %s", cmdStr, suffix),
	})

	return stdoutWriter, stderrWriter
}

type logWriter struct {
	logger Logger
	prefix string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\n"))
	for _, line := range lines {
		if len(line) > 0 {
			w.logger.Logf("[%s] %s", w.prefix, string(line))
		}
	}
	return len(p), nil
}

type ErrorHandler struct {
	config Config
	logger Logger
}

func NewErrorHandler(config Config, logger Logger) *ErrorHandler {
	return &ErrorHandler{
		config: config,
		logger: logger,
	}
}

func (eh *ErrorHandler) HandleError(fullCmd []string, output *OutputCapture, err error) error {
	if err == nil {
		return nil
	}

	errorMsg := fmt.Sprintf("[%s] Command failed: %v\nStdout: %s\nStderr: %s",
		strings.Join(fullCmd, " "), err, output.GetStdout(), output.GetStderr())

	if eh.config.FailOnError {
		eh.logger.Fatalf("%s", errorMsg)
	} else {
		eh.logger.Logf("%s", errorMsg)
	}

	return err
}

type ProcessManager struct {
	process *exec.Cmd
	logger  Logger
	config  Config
}

func NewProcessManager(config Config, logger Logger) *ProcessManager {
	return &ProcessManager{
		logger: logger,
		config: config,
	}
}

func (pm *ProcessManager) IsRunning() bool {
	return pm.process != nil
}

func (pm *ProcessManager) Start(cmd *exec.Cmd) error {
	if pm.process != nil {
		err := fmt.Errorf("process already running")
		if pm.config.FailOnError {
			pm.logger.Fatalf("[ENGINE] %s", err.Error())
		} else {
			pm.logger.Errorf("[ENGINE] %s", err.Error())
		}
		return err
	}

	if err := cmd.Start(); err != nil {
		errorMsg := fmt.Errorf("failed to start command: %v", err)
		if pm.config.FailOnError {
			pm.logger.Fatalf("[ENGINE] %s", errorMsg)
		} else {
			pm.logger.Errorf("[ENGINE] %s", errorMsg)
		}
		return errorMsg
	}

	pm.process = cmd
	return nil
}

func (pm *ProcessManager) Stop() error {
	if pm.process == nil || pm.process.Process == nil {
		return nil
	}

	pm.logger.Logf("[ENGINE] Stopping command process")
	err := pm.process.Process.Signal(os.Signal(syscall.SIGTERM))
	if err != nil {
		errorMsg := fmt.Errorf("error stopping process: %v", err)
		if pm.config.FailOnError {
			pm.logger.Fatalf("%s", errorMsg)
		} else {
			pm.logger.Errorf("%s", errorMsg)
		}
		return errorMsg
	}

	pm.process = nil
	return nil
}

func (pm *ProcessManager) StopWithoutLogging() error {
	if pm.process == nil || pm.process.Process == nil {
		return nil
	}
	return pm.process.Process.Signal(os.Signal(syscall.SIGTERM))
}

type Engine struct {
	builder        *CommandBuilder
	output         *OutputCapture
	errorHandler   *ErrorHandler
	processManager *ProcessManager
	logger         Logger
	config         Config
	prepared       bool
}

func NewEngine(binary string) (*Engine, error) {
	return NewEngineWithConfig(binary, DefaultConfig())
}

func NewEngineWithConfig(binary string, config Config) (*Engine, error) {
	builder, err := NewCommandBuilder(binary, config)
	if err != nil {
		return nil, err
	}

	logger := &noOpLogger{}

	return &Engine{
		builder:        builder,
		output:         NewOutputCapture(),
		errorHandler:   NewErrorHandler(config, logger),
		processManager: NewProcessManager(config, logger),
		logger:         logger,
		config:         config,
		prepared:       false,
	}, nil
}

func NewEngineFromSlice(cmdParts []string) *Engine {
	return NewEngineFromSliceWithConfig(cmdParts, DefaultConfig())
}

func NewEngineFromSliceWithConfig(cmdParts []string, config Config) *Engine {
	builder := NewCommandBuilderFromSlice(cmdParts, config)
	logger := &noOpLogger{}

	return &Engine{
		builder:        builder,
		output:         NewOutputCapture(),
		errorHandler:   NewErrorHandler(config, logger),
		processManager: NewProcessManager(config, logger),
		logger:         logger,
		config:         config,
		prepared:       false,
	}
}

func (e *Engine) updateLogger(logger Logger) {
	e.logger = logger
	e.errorHandler = NewErrorHandler(e.config, logger)
	e.processManager = NewProcessManager(e.config, logger)
}

func (e *Engine) buildCommandWithSuffix(sources ...CmdSource) ([]string, string, error) {
	fullCmd, cmdStr, err := e.builder.BuildCommand(sources...)
	if err != nil {
		return nil, "", err
	}

	if !e.prepared && e.config.Suffix != "" {
		suffixParsed, err := shellwords.Parse(e.config.Suffix)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse suffix: %w", err)
		}
		fullCmd = append(fullCmd, suffixParsed...)
		e.prepared = true
	}

	return fullCmd, cmdStr, nil
}

func (e *Engine) SetFailOnError(failOnError bool) {
	e.config.FailOnError = failOnError
	e.builder.config.FailOnError = failOnError
	e.errorHandler = NewErrorHandler(e.config, e.logger)
	e.processManager = NewProcessManager(e.config, e.logger)
}

func (e *Engine) Run(t *testing.T, command ...CmdSource) error {
	e.updateLogger(NewTestLogger(t))

	fullCmd, cmdStr, err := e.buildCommandWithSuffix(command...)
	if err != nil {
		return e.errorHandler.HandleError(fullCmd, e.output, err)
	}

	e.output.Reset()
	cmd := exec.Command(fullCmd[0], fullCmd[1:]...)

	stdoutWriter, stderrWriter := e.output.CreateWriters(e.logger, cmdStr, e.config.Suffix)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	e.logger.Logf("[ENGINE] Running command: %s %s %s",
		strings.Join(e.builder.baseCommand, " "), cmdStr, e.config.Suffix)

	err = cmd.Run()
	return e.errorHandler.HandleError(fullCmd, e.output, err)
}

func (e *Engine) RunAndCapture(t *testing.T, command ...CmdSource) (string, error) {
	err := e.Run(t, command...)
	if err != nil {
		return "", err
	}
	return e.output.GetStdout(), nil
}

func (e *Engine) RunBackground(t *testing.T, command ...CmdSource) error {
	e.updateLogger(NewTestLogger(t))

	if e.processManager.IsRunning() {
		return fmt.Errorf("command already started")
	}

	fullCmd, cmdStr, err := e.buildCommandWithSuffix(command...)
	if err != nil {
		return e.errorHandler.HandleError(fullCmd, e.output, err)
	}

	e.output.Reset()
	cmd := exec.Command(fullCmd[0], fullCmd[1:]...)

	stdoutWriter, stderrWriter := e.output.CreateWriters(e.logger, cmdStr, e.config.Suffix)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	e.logger.Logf("[ENGINE] Starting background command: %s %s",
		strings.Join(e.builder.baseCommand, " "), cmdStr)

	return e.processManager.Start(cmd)
}

func (e *Engine) Stop(t *testing.T) error {
	e.updateLogger(NewTestLogger(t))
	return e.processManager.Stop()
}

func (e *Engine) StopWithError() error {
	return e.processManager.StopWithoutLogging()
}

func (e *Engine) GetStdout() string {
	return e.output.GetStdout()
}

func (e *Engine) GetStderr() string {
	return e.output.GetStderr()
}

type noOpLogger struct{}

func (n *noOpLogger) Logf(format string, args ...interface{})   {}
func (n *noOpLogger) Errorf(format string, args ...interface{}) {}
func (n *noOpLogger) Fatalf(format string, args ...interface{}) {}

// Legacy constructor functions for backward compatibility
func NewEngineWithOptions(binary string, options Options) *Engine {
	config := Config{
		FailOnError: options.FailOnError,
		Suffix:      options.Suffix,
	}
	engine, _ := NewEngineWithConfig(binary, config)
	return engine
}

func NewEngineFromSliceWithOptions(cmdParts []string, options Options) *Engine {
	config := Config{
		FailOnError: options.FailOnError,
		Suffix:      options.Suffix,
	}
	return NewEngineFromSliceWithConfig(cmdParts, config)
}

func DefaultEngineOptions() Options {
	return Options{
		FailOnError: true,
	}
}

type Options struct {
	FailOnError bool
	Suffix      string
}
