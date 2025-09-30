package dependency_test

import (
	"context"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/dependency"
	"go.uber.org/zap"
	"testing"
	"time"

	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/tests"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// EXAMPLE: Using Direct Mock Setup
// ============================================================================

func TestReady_Example_DirectMockSetup(t *testing.T) {

	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	// Create mock registry
	registry := tests.NewMockRegistry()

	// Build containers using existing container builder
	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		WithGroup("app").
		WithName("app").
		Build()

	depContainer := tests.NewContainerBuilder().
		WithGeneratedName("mysql-1").
		WithGroup("database").
		WithName("mysql").
		Build()
	depContainer.GetStatus().LastReadiness = true

	// Setup mock expectations
	registry.On("Find", "smr", "app", "app").Return(myContainer)
	registry.On("Find", "smr", "database", "mysql").Return(depContainer)

	// Define dependencies
	dependencies := []v1.ContainersDependsOn{
		{
			Prefix:  "smr",
			Group:   "database",
			Name:    "mysql",
			Timeout: "5s",
		},
	}

	// Execute
	result, err := dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	// Assert
	assert.NoError(t, err)
	assert.True(t, result)

	// Verify success state
	select {
	case state := <-channel:
		assert.Equal(t, dependency.SUCCESS, state.State)
	case <-time.After(1 * time.Second):
		t.Fatal("expected success state")
	}

	registry.AssertExpectations(t)
}

// ============================================================================
// EXAMPLE: Using Registry Builder (cleaner for complex scenarios)
// ============================================================================

func TestReady_Example_RegistryBuilder(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	// Build containers
	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	dbContainer := tests.NewContainerBuilder().
		WithGeneratedName("mysql-1").
		Build()
	dbContainer.GetStatus().LastReadiness = true

	cacheContainer := tests.NewContainerBuilder().
		WithGeneratedName("redis-1").
		Build()
	cacheContainer.GetStatus().LastReadiness = true

	// Use registry builder for cleaner setup
	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithContainer("smr", "database", "mysql", dbContainer).
		WithContainer("smr", "cache", "redis", cacheContainer).
		Build()

	dependencies := []v1.ContainersDependsOn{
		{Prefix: "smr", Group: "database", Name: "mysql", Timeout: "5s"},
		{Prefix: "smr", Group: "cache", Name: "redis", Timeout: "5s"},
	}

	result, err := dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	assert.NoError(t, err)
	assert.True(t, result)

	registry.AssertExpectations(t)
}

// ============================================================================
// EXAMPLE: Testing Failure Scenarios
// ============================================================================

func TestReady_Example_FailureNotFound(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	// Use builder with NotFound helper
	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithNotFound("smr", "database", "mysql"). // Dependency not found
		Build()

	dependencies := []v1.ContainersDependsOn{
		{Prefix: "smr", Group: "database", Name: "mysql", Timeout: "2s"},
	}

	result, err := dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "container not found")

	// Verify CHECKING and FAILED states
	foundChecking := false
	foundFailed := false

	for len(channel) > 0 {
		state := <-channel
		if state.State == dependency.CHECKING {
			foundChecking = true
		}
		if state.State == dependency.FAILED {
			foundFailed = true
		}
	}

	assert.True(t, foundChecking)
	assert.True(t, foundFailed)
}

func TestReady_Example_FailureNotReady(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	// Dependency exists but is NOT ready
	depContainer := tests.NewContainerBuilder().
		WithGeneratedName("mysql-1").
		Build()
	depContainer.GetStatus().LastReadiness = false // Not ready!

	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithContainer("smr", "database", "mysql", depContainer).
		Build()

	dependencies := []v1.ContainersDependsOn{
		{Prefix: "smr", Group: "database", Name: "mysql", Timeout: "1s"},
	}

	dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	for len(channel) > 0 {
		state := <-channel
		if state.State == dependency.CHECKING {
			assert.Error(t, state.Error)
			assert.Contains(t, state.Error.Error(), "container not ready")
		}
	}
}

func TestReady_Example_EmptyGroup(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	// Use builder with EmptyGroup helper
	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithEmptyGroup("smr", "workers"). // No workers available
		Build()

	dependencies := []v1.ContainersDependsOn{
		{Prefix: "smr", Group: "workers", Name: "*", Timeout: "1s"},
	}

	dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	for len(channel) > 0 {
		state := <-channel
		if state.State == dependency.CHECKING {
			assert.Error(t, state.Error)
			assert.Contains(t, state.Error.Error(), "waiting for atleast one container")
		}
	}
}

// ============================================================================
// EXAMPLE: Complex Multi-Dependency Scenario
// ============================================================================

func TestReady_Example_ComplexScenario(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 20)

	// Application container
	appContainer := tests.NewContainerBuilder().
		WithGeneratedName("api-server-1").
		WithGroup("api").
		WithName("api-server").
		Build()

	// Database dependency
	dbContainer := tests.NewContainerBuilder().
		WithGeneratedName("postgres-1").
		WithGroup("database").
		WithName("postgres").
		Build()
	dbContainer.GetStatus().LastReadiness = true

	// Cache dependency
	cacheContainer := tests.NewContainerBuilder().
		WithGeneratedName("redis-1").
		WithGroup("cache").
		WithName("redis").
		Build()
	cacheContainer.GetStatus().LastReadiness = true

	// Message queue dependency
	mqContainer := tests.NewContainerBuilder().
		WithGeneratedName("rabbitmq-1").
		WithGroup("queue").
		WithName("rabbitmq").
		Build()
	mqContainer.GetStatus().LastReadiness = true

	// Multiple worker instances
	worker1 := tests.NewContainerBuilder().
		WithGeneratedName("worker-1").
		Build()
	worker1.GetStatus().LastReadiness = true

	worker2 := tests.NewContainerBuilder().
		WithGeneratedName("worker-2").
		Build()
	worker2.GetStatus().LastReadiness = true

	// Build registry with all dependencies
	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "api", "api-server", appContainer).
		WithContainer("smr", "database", "postgres", dbContainer).
		WithContainer("smr", "cache", "redis", cacheContainer).
		WithContainer("smr", "queue", "rabbitmq", mqContainer).
		WithGroup("smr", "workers", []platforms.IContainer{worker1, worker2}).
		Build()

	// Define all dependencies
	dependencies := []v1.ContainersDependsOn{
		{Prefix: "smr", Group: "database", Name: "postgres", Timeout: "30s"},
		{Prefix: "smr", Group: "cache", Name: "redis", Timeout: "30s"},
		{Prefix: "smr", Group: "queue", Name: "rabbitmq", Timeout: "30s"},
		{Prefix: "smr", Group: "workers", Name: "*", Timeout: "30s"},
	}

	result, err := dependency.Ready(ctx, registry, "api", "api-server", dependencies, channel, zap.NewNop())

	assert.NoError(t, err)
	assert.True(t, result)

	// Verify final success state
	var finalState *dependency.State
	for len(channel) > 0 {
		finalState = <-channel
	}
	assert.NotNil(t, finalState)
	assert.Equal(t, dependency.SUCCESS, finalState.State)

	registry.AssertExpectations(t)
}

// ============================================================================
// EXAMPLE: Testing Timeout Behavior
// ============================================================================

func TestReady_Example_Timeout(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	// Dependency that will never be ready
	depContainer := tests.NewContainerBuilder().
		WithGeneratedName("mysql-1").
		Build()
	depContainer.GetStatus().LastReadiness = false

	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithContainer("smr", "database", "mysql", depContainer).
		Build()

	dependencies := []v1.ContainersDependsOn{
		{
			Prefix:  "smr",
			Group:   "database",
			Name:    "mysql",
			Timeout: "100ms", // Very short timeout
		},
	}

	start := time.Now()
	result, err := dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.False(t, result)

	// Should timeout relatively quickly (with backoff, will be slightly longer than 100ms)
	assert.Less(t, elapsed, 5*time.Second, "should timeout within reasonable time")

	// Check for CANCELED state (due to timeout)
	foundCanceled := false
	for len(channel) > 0 {
		state := <-channel
		if state.State == dependency.CANCELED {
			foundCanceled = true
			assert.Equal(t, dependency.ERROR_CONTEXT_CANCELED, state.Error)
			break
		}
	}
	assert.True(t, foundCanceled)
}

// ============================================================================
// EXAMPLE: Testing SolveDepends Directly
// ============================================================================

func TestSolveDepends_Example_Direct(t *testing.T) {
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	depContainer := tests.NewContainerBuilder().
		WithGeneratedName("mysql-1").
		Build()
	depContainer.GetStatus().LastReadiness = true

	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithContainer("smr", "database", "mysql", depContainer).
		Build()

	dep := &dependency.Dependency{
		Prefix:  "smr",
		Group:   "database",
		Name:    "mysql",
		Timeout: "30s",
		Cancel:  func() {},
	}

	err := dependency.SolveDepends(registry, "smr", "app", "app", dep, channel, zap.NewNop())

	assert.NoError(t, err)
	assert.Empty(t, channel, zap.NewNop()) // No state messages for success case

	registry.AssertExpectations(t)
}

// ============================================================================
// EXAMPLE: Progressive Readiness Test (simulating containers becoming ready)
// ============================================================================

func TestReady_Example_ProgressiveReadiness(t *testing.T) {
	t.Skip("This is a conceptual example - would need more sophisticated mocking")

	// This example shows how you might test a scenario where containers
	// become ready over time. In practice, you'd need to use channels or
	// callbacks to simulate state changes during the retry loop.

	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	depContainer := tests.NewContainerBuilder().
		WithGeneratedName("mysql-1").
		Build()

	// Initially not ready
	depContainer.GetStatus().LastReadiness = false

	registry := tests.NewMockRegistry()

	// Set up mock to change behavior after first call
	callCount := 0
	registry.On("Find", "smr", "app", "app").Return(myContainer)
	registry.On("Find", "smr", "database", "mysql").Return(func(string, string, string) platforms.IContainer {
		callCount++
		if callCount > 2 {
			// Become ready after a few attempts
			depContainer.GetStatus().LastReadiness = true
		}
		return depContainer
	})

	dependencies := []v1.ContainersDependsOn{
		{Prefix: "smr", Group: "database", Name: "mysql", Timeout: "10s"},
	}

	result, err := dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	assert.NoError(t, err)
	assert.True(t, result)
	assert.Greater(t, callCount, 2, "should have retried multiple times")
}

// ============================================================================
// EXAMPLE: Testing Wildcard Dependencies with Builder
// ============================================================================

func TestReady_Example_WildcardWithBuilder(t *testing.T) {
	ctx := context.Background()
	channel := make(chan *dependency.State, 10)

	myContainer := tests.NewContainerBuilder().
		WithGeneratedName("app-1").
		Build()

	// Create multiple ready workers
	worker1 := tests.NewContainerBuilder().
		WithGeneratedName("worker-1").
		Build()
	worker1.GetStatus().LastReadiness = true

	worker2 := tests.NewContainerBuilder().
		WithGeneratedName("worker-2").
		Build()
	worker2.GetStatus().LastReadiness = true

	worker3 := tests.NewContainerBuilder().
		WithGeneratedName("worker-3").
		Build()
	worker3.GetStatus().LastReadiness = true

	// Use builder to setup group
	registry := tests.NewRegistryBuilder().
		WithContainer("smr", "app", "app", myContainer).
		WithGroup("smr", "workers", []platforms.IContainer{worker1, worker2, worker3}).
		Build()

	dependencies := []v1.ContainersDependsOn{
		{
			Prefix:  "smr",
			Group:   "workers",
			Name:    "*", // Wait for all workers
			Timeout: "5s",
		},
	}

	result, err := dependency.Ready(ctx, registry, "app", "app", dependencies, channel, zap.NewNop())

	assert.NoError(t, err)
	assert.True(t, result)

	// Verify success state was sent
	var finalState *dependency.State
	for len(channel) > 0 {
		finalState = <-channel
	}
	assert.NotNil(t, finalState)
	assert.Equal(t, dependency.SUCCESS, finalState.State)
	assert.Nil(t, finalState.Error)

	registry.AssertExpectations(t)
}
