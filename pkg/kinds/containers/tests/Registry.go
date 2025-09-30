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

// Stub implementations for remaining Registry interface methods
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

func (m *MockRegistry) BackOff(group string, name string) error {
	args := m.Called(group, name)
	return args.Error(0)
}

func (m *MockRegistry) BackOffReset(group string, name string) {
	m.Called(group, name)
}

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

func (b *RegistryBuilder) Build() *MockRegistry {
	return b.registry
}
