package tests

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/image"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/node"
)

// ============================================================================
// MOCK CONTAINER - Implements platforms.IContainer
// ============================================================================

type MockContainer struct {
	// Pre-configured fields that bypass mock calls
	readiness     []*readiness.Readiness
	status        *status.Status
	state         *state.State
	stateError    error
	generatedName string
	group         string
	name          string
	runtime       *types.Runtime
	node          *node.Node
}

// Core methods used by solver
func (m *MockContainer) GetReadiness() []*readiness.Readiness {
	return m.readiness
}

func (m *MockContainer) GetStatus() *status.Status {
	return m.status
}

func (m *MockContainer) GetState() (*state.State, error) {
	return m.state, m.stateError
}

func (m *MockContainer) GetGeneratedName() string {
	return m.generatedName
}

func (m *MockContainer) GetGroup() string {
	return m.group
}

func (m *MockContainer) GetName() string {
	return m.name
}

func (m *MockContainer) GetRuntime() *types.Runtime {
	return m.runtime
}

func (m *MockContainer) GetNode() *node.Node {
	return m.node
}

// Stub implementations for remaining IContainer interface methods
func (m *MockContainer) Run() error { return nil }
func (m *MockContainer) PreRun(*configuration.Configuration, *clients.Http, *authentication.User) error {
	return nil
}
func (m *MockContainer) PostRun(*configuration.Configuration, *dns.Records) error { return nil }
func (m *MockContainer) InitContainer(*v1.ContainersInternal, *configuration.Configuration, *clients.Http, *authentication.User) error {
	return nil
}
func (m *MockContainer) MountResources() error                         { return nil }
func (m *MockContainer) UpdateDns(*dns.Records) error                  { return nil }
func (m *MockContainer) RemoveDns(*dns.Records, string) error          { return nil }
func (m *MockContainer) SyncNetwork() error                            { return nil }
func (m *MockContainer) HasDependencyOn(string, string, string) bool   { return false }
func (m *MockContainer) HasOwner() bool                                { return false }
func (m *MockContainer) GetEngineState() string                        { return "" }
func (m *MockContainer) GetNodeName() string                           { return "" }
func (m *MockContainer) GetId() string                                 { return "" }
func (m *MockContainer) GetGlobalDefinition() *v1.ContainersDefinition { return nil }
func (m *MockContainer) GetDefinition() idefinitions.IDefinition       { return nil }
func (m *MockContainer) GetLabels() map[string]string                  { return nil }
func (m *MockContainer) GetGroupIdentifier() string                    { return "" }
func (m *MockContainer) GetIndex() (uint64, error)                     { return 0, nil }
func (m *MockContainer) GetImageState() *image.ImageState              { return nil }
func (m *MockContainer) GetImageWithTag() string                       { return "" }
func (m *MockContainer) GetNetwork() map[string]net.IP                 { return nil }
func (m *MockContainer) GetDomain(string) string                       { return "" }
func (m *MockContainer) GetHeadlessDomain(string) string               { return "" }
func (m *MockContainer) GetInit() platforms.IPlatform                  { return nil }
func (m *MockContainer) GetInitDefinition() *v1.ContainersInternal     { return nil }
func (m *MockContainer) IsGhost() bool                                 { return false }
func (m *MockContainer) SetGhost(bool)                                 {}
func (m *MockContainer) CreateVolume(*v1.VolumeDefinition) error       { return nil }
func (m *MockContainer) DeleteVolume(string, bool) error               { return nil }
func (m *MockContainer) Start() error                                  { return nil }
func (m *MockContainer) Stop(string) error                             { return nil }
func (m *MockContainer) Kill(string) error                             { return nil }
func (m *MockContainer) Restart() error                                { return nil }
func (m *MockContainer) Delete() error                                 { return nil }
func (m *MockContainer) Wait(string) error                             { return nil }
func (m *MockContainer) Rename(string) error                           { return nil }
func (m *MockContainer) Exec(context.Context, []string, bool, string, string) (string, *bufio.Reader, net.Conn, error) {
	return "", nil, nil, nil
}
func (m *MockContainer) ExecInspect(string) (bool, int, error) { return false, 0, nil }
func (m *MockContainer) ExecResize(string, int, int) error     { return nil }
func (m *MockContainer) Logs(context.Context, bool) (io.ReadCloser, error) {
	return nil, nil
}
func (m *MockContainer) Clean() error            { return nil }
func (m *MockContainer) ToJSON() ([]byte, error) { return nil, nil }

// ============================================================================
// CONTAINER BUILDER
// ============================================================================

type ContainerBuilder struct {
	container *MockContainer
}

func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		container: &MockContainer{
			status:        status.New(),
			state:         &state.State{State: "running"},
			generatedName: "test-container-1",
			group:         "test-group",
			name:          "test-container",
			readiness:     []*readiness.Readiness{},
			runtime:       &types.Runtime{},
			node:          &node.Node{},
		},
	}
}

func (b *ContainerBuilder) WithGeneratedName(name string) *ContainerBuilder {
	b.container.generatedName = name
	return b
}

func (b *ContainerBuilder) WithGroup(group string) *ContainerBuilder {
	b.container.group = group
	return b
}

func (b *ContainerBuilder) WithName(name string) *ContainerBuilder {
	b.container.name = name
	return b
}

func (b *ContainerBuilder) WithState(containerState string) *ContainerBuilder {
	b.container.state = &state.State{State: containerState}
	return b
}

func (b *ContainerBuilder) WithStateError(err error) *ContainerBuilder {
	b.container.stateError = err
	return b
}

func (b *ContainerBuilder) WithStatus(statusState string) *ContainerBuilder {
	b.container.status.SetState(statusState)
	return b
}

func (b *ContainerBuilder) WithReadiness(probes ...*readiness.Readiness) *ContainerBuilder {
	b.container.readiness = probes
	return b
}

func (b *ContainerBuilder) AddReadinessProbe(probe *readiness.Readiness) *ContainerBuilder {
	b.container.readiness = append(b.container.readiness, probe)
	return b
}

func (b *ContainerBuilder) WithRuntime(runtime *types.Runtime) *ContainerBuilder {
	b.container.runtime = runtime
	return b
}

func (b *ContainerBuilder) WithNode(n *node.Node) *ContainerBuilder {
	b.container.node = n
	return b
}

func (b *ContainerBuilder) Build() *MockContainer {
	return b.container
}

// ============================================================================
// READINESS BUILDER
// ============================================================================

type ReadinessBuilder struct {
	readiness *readiness.Readiness
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewReadinessBuilder() *ReadinessBuilder {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	return &ReadinessBuilder{
		readiness: &readiness.Readiness{
			Name:       "default-probe",
			Method:     "GET",
			Timeout:    "30s",
			Ctx:        ctx,
			Cancel:     cancel,
			BodyUnpack: make(map[string]string),
			Body:       make(map[string]string),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (b *ReadinessBuilder) WithName(name string) *ReadinessBuilder {
	b.readiness.Name = name
	return b
}

func (b *ReadinessBuilder) WithURL(url string) *ReadinessBuilder {
	b.readiness.URL = url
	b.readiness.Type = readiness.TYPE_URL
	return b
}

func (b *ReadinessBuilder) WithMethod(method string) *ReadinessBuilder {
	b.readiness.Method = method
	return b
}

func (b *ReadinessBuilder) WithCommand(command ...string) *ReadinessBuilder {
	b.readiness.Command = command
	b.readiness.Type = readiness.TYPE_COMMAND
	return b
}

func (b *ReadinessBuilder) WithTimeout(timeout string) *ReadinessBuilder {
	b.readiness.Timeout = timeout

	if b.cancel != nil {
		b.cancel()
	}

	duration, err := time.ParseDuration(timeout)
	if err == nil {
		b.ctx, b.cancel = context.WithTimeout(context.Background(), duration)
		b.readiness.Ctx = b.ctx
		b.readiness.Cancel = b.cancel
	}

	return b
}

func (b *ReadinessBuilder) WithContext(ctx context.Context, cancel context.CancelFunc) *ReadinessBuilder {
	b.readiness.Ctx = ctx
	b.readiness.Cancel = cancel
	b.ctx = ctx
	b.cancel = cancel
	return b
}

func (b *ReadinessBuilder) WithBody(body map[string]string) *ReadinessBuilder {
	b.readiness.BodyUnpack = body
	return b
}

func (b *ReadinessBuilder) Build() *readiness.Readiness {
	return b.readiness
}

// ============================================================================
// HTTP CLIENT BUILDER
// ============================================================================

type HttpClientBuilder struct {
	client *clients.Http
}

func NewHttpClientBuilder() *HttpClientBuilder {
	return &HttpClientBuilder{
		client: &clients.Http{
			Clients: make(map[string]*clients.Client),
		},
	}
}

func (b *HttpClientBuilder) WithClient(username string, httpClient *http.Client) *HttpClientBuilder {
	b.client.Clients[username] = &clients.Client{
		Http: httpClient,
	}
	return b
}

func (b *HttpClientBuilder) WithDefaultClient(httpClient *http.Client) *HttpClientBuilder {
	return b.WithClient("default", httpClient)
}

func (b *HttpClientBuilder) Build() *clients.Http {
	return b.client
}

// ============================================================================
// USER BUILDER
// ============================================================================

type UserBuilder struct {
	user *authentication.User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: &authentication.User{
			Username: "default",
		},
	}
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.Username = username
	return b
}

func (b *UserBuilder) Build() *authentication.User {
	return b.user
}

// ============================================================================
// PRESET BUILDERS (Common scenarios)
// ============================================================================

func NewRunningContainer() *MockContainer {
	return NewContainerBuilder().
		WithState("running").
		WithStatus(status.READINESS_CHECKING).
		WithGeneratedName("running-container-1").
		Build()
}

func NewStoppedContainer() *MockContainer {
	return NewContainerBuilder().
		WithState("exited").
		WithStatus(status.DEAD).
		WithGeneratedName("stopped-container-1").
		Build()
}

func NewContainerWithURLProbe(url string) *MockContainer {
	probe := NewReadinessBuilder().
		WithURL(url).
		WithMethod("GET").
		Build()

	return NewContainerBuilder().
		WithState("running").
		WithStatus(status.READINESS_CHECKING).
		WithReadiness(probe).
		Build()
}

func NewContainerWithCommandProbe(command ...string) *MockContainer {
	probe := NewReadinessBuilder().
		WithCommand(command...).
		Build()

	return NewContainerBuilder().
		WithState("running").
		WithStatus(status.READINESS_CHECKING).
		WithReadiness(probe).
		Build()
}

func NewContainerWithMultipleProbes(probes ...*readiness.Readiness) *MockContainer {
	return NewContainerBuilder().
		WithState("running").
		WithStatus(status.READINESS_CHECKING).
		WithReadiness(probes...).
		Build()
}

func NewHttpProbe(url, method string) *readiness.Readiness {
	return NewReadinessBuilder().
		WithURL(url).
		WithMethod(method).
		Build()
}

func NewHttpProbeWithBody(url, method string, body map[string]string) *readiness.Readiness {
	return NewReadinessBuilder().
		WithURL(url).
		WithMethod(method).
		WithBody(body).
		Build()
}

// Quick container builder for a single probe
func NewContainerWithSingleProbe(probe *readiness.Readiness) *MockContainer {
	return NewContainerBuilder().
		WithState("running").
		WithStatus(status.READINESS_CHECKING).
		WithReadiness(probe).
		Build()
}
