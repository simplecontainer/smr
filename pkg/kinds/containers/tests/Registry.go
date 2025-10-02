package tests

import (
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCK REGISTRY - Implements platforms.Registry
// ============================================================================

type MockRegistry struct {
	mock.Mock
	containers map[string]platforms.IContainer
}

func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		containers: make(map[string]platforms.IContainer),
	}
}

// Core methods used by dependency solver
func (m *MockRegistry) Find(prefix string, group string, name string) platforms.IContainer {
	args := m.Called(prefix, group, name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(platforms.IContainer)
}

func (m *MockRegistry) FindGroup(prefix string, group string) []platforms.IContainer {
	args := m.Called(prefix, group)
	if args.Get(0) == nil {
		return []platforms.IContainer{}
	}
	return args.Get(0).([]platforms.IContainer)
}

func (m *MockRegistry) FindLocal(group string, name string) platforms.IContainer {
	args := m.Called(group, name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(platforms.IContainer)
}

// Helper for manual container management (optional)
func (m *MockRegistry) AddContainer(key string, container platforms.IContainer) {
	m.containers[key] = container
}

// Container Management
func (m *MockRegistry) AddOrUpdate(group string, name string, containerAddr platforms.IContainer) {
	m.Called(group, name, containerAddr)
}

func (m *MockRegistry) Sync(group string, name string) error {
	args := m.Called(group, name)
	return args.Error(0)
}

func (m *MockRegistry) Remove(prefix string, group string, name string) error {
	args := m.Called(prefix, group, name)
	return args.Error(0)
}

func (m *MockRegistry) Name(client *clients.Http, prefix string, group string, name string) (string, []uint64, error) {
	args := m.Called(client, prefix, group, name)
	return args.String(0), args.Get(1).([]uint64), args.Error(2)
}

func (m *MockRegistry) NameReplica(group string, name string, index uint64) string {
	args := m.Called(group, name, index)
	return args.String(0)
}

// Restart Tracking - Lifecycle Events
func (m *MockRegistry) MarkContainerStarted(group string, name string) {
	m.Called(group, name)
}

func (m *MockRegistry) MarkContainerStopped(group string, name string) error {
	args := m.Called(group, name)
	return args.Error(0)
}

// Backoff Management
func (m *MockRegistry) BackOff(group string, name string) error {
	args := m.Called(group, name)
	return args.Error(0)
}

func (m *MockRegistry) BackOffReset(group string, name string) {
	m.Called(group, name)
}

func (m *MockRegistry) ResetRestartTracking(group string, name string) {
	m.Called(group, name)
}

// Restart Statistics
func (m *MockRegistry) GetBackOffCount(group string, name string) uint64 {
	args := m.Called(group, name)
	return args.Get(0).(uint64)
}

func (m *MockRegistry) GetRestartCount(group string, name string) int {
	args := m.Called(group, name)
	return args.Int(0)
}

// Cleanup and Monitoring
func (m *MockRegistry) GetTrackerSize() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRegistry) IsBanned(group string, name string) bool {
	args := m.Called(group, name)
	return args.Bool(0)
}

func (m *MockRegistry) UnbanContainer(group string, name string) {
	m.Called(group, name)
}

// Legacy methods (if still needed)
func (m *MockRegistry) GetIndexes(prefix string, group string, name string) ([]uint64, error) {
	args := m.Called(prefix, group, name)
	if args.Get(0) == nil {
		return []uint64{}, args.Error(1)
	}
	return args.Get(0).([]uint64), args.Error(1)
}

func (m *MockRegistry) GetIndexesLocal(prefix string, group string, name string) ([]uint64, error) {
	args := m.Called(prefix, group, name)
	if args.Get(0) == nil {
		return []uint64{}, args.Error(1)
	}
	return args.Get(0).([]uint64), args.Error(1)
}

// ============================================================================
// REGISTRY BUILDER (Optional - for more complex scenarios)
// ============================================================================

type RegistryBuilder struct {
	registry *MockRegistry
}

func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		registry: NewMockRegistry(),
	}
}

// WithContainer adds a container that will be returned by Find
func (b *RegistryBuilder) WithContainer(prefix, group, name string, container platforms.IContainer) *RegistryBuilder {
	b.registry.On("Find", prefix, group, name).Return(container)
	return b
}

// WithGroup adds a group of containers that will be returned by FindGroup
func (b *RegistryBuilder) WithGroup(prefix, group string, containers []platforms.IContainer) *RegistryBuilder {
	b.registry.On("FindGroup", prefix, group).Return(containers)
	return b
}

// WithLocalContainer adds a container that will be returned by FindLocal
func (b *RegistryBuilder) WithLocalContainer(group, name string, container platforms.IContainer) *RegistryBuilder {
	b.registry.On("FindLocal", group, name).Return(container)
	return b
}

// WithEmptyGroup configures FindGroup to return empty slice
func (b *RegistryBuilder) WithEmptyGroup(prefix, group string) *RegistryBuilder {
	b.registry.On("FindGroup", prefix, group).Return([]platforms.IContainer{})
	return b
}

// WithNotFound configures Find to return nil
func (b *RegistryBuilder) WithNotFound(prefix, group, name string) *RegistryBuilder {
	b.registry.On("Find", prefix, group, name).Return(nil)
	return b
}

// Restart Tracking Builders
func (b *RegistryBuilder) WithRestartAllowed(group, name string) *RegistryBuilder {
	b.registry.On("MarkContainerStopped", group, name).Return(nil)
	b.registry.On("IsBanned", group, name).Return(false)
	return b
}

func (b *RegistryBuilder) WithRestartBlocked(group, name string, err error) *RegistryBuilder {
	b.registry.On("MarkContainerStopped", group, name).Return(err)
	b.registry.On("IsBanned", group, name).Return(true)
	return b
}

func (b *RegistryBuilder) WithBackoffCount(group, name string, count uint64) *RegistryBuilder {
	b.registry.On("GetBackOffCount", group, name).Return(count)
	return b
}

func (b *RegistryBuilder) WithRestartCount(group, name string, count int) *RegistryBuilder {
	b.registry.On("GetRestartCount", group, name).Return(count)
	return b
}

func (b *RegistryBuilder) WithTrackerSize(size int) *RegistryBuilder {
	b.registry.On("GetTrackerSize").Return(size)
	return b
}

func (b *RegistryBuilder) AllowAllRestarts() *RegistryBuilder {
	b.registry.On("MarkContainerStarted", mock.Anything, mock.Anything).Return()
	b.registry.On("MarkContainerStopped", mock.Anything, mock.Anything).Return(nil)
	b.registry.On("BackOff", mock.Anything, mock.Anything).Return(nil)
	b.registry.On("BackOffReset", mock.Anything, mock.Anything).Return()
	b.registry.On("ResetRestartTracking", mock.Anything, mock.Anything).Return()
	b.registry.On("IsBanned", mock.Anything, mock.Anything).Return(false)
	return b
}

func (b *RegistryBuilder) Build() *MockRegistry {
	return b.registry
}
