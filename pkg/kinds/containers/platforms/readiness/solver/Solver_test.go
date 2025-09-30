package solver_test

import (
	"context"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness/solver"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/tests"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ============================================================================
// SIMPLE USAGE EXAMPLES
// ============================================================================

func TestReady_Success_SimpleBuilder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := tests.NewHttpClientBuilder().
		WithClient("test", server.Client()).
		Build()

	user := tests.NewUserBuilder().
		WithUsername("test").
		Build()

	probe := tests.NewHttpProbe(server.URL, "GET")
	container := tests.NewContainerWithSingleProbe(probe)

	result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.NoError(t, err)
	assert.True(t, result)
	assert.True(t, probe.Solved)
}

// ============================================================================
// USING PRESET BUILDERS
// ============================================================================

func TestReady_Success_WithPreset(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	container := tests.NewContainerWithURLProbe(server.URL)

	httpClient := tests.NewHttpClientBuilder().
		WithDefaultClient(server.Client()).
		Build()

	user := tests.NewUserBuilder().Build()

	result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.NoError(t, err)
	assert.True(t, result)
}

// ============================================================================
// MULTIPLE READINESS PROBES
// ============================================================================

func TestReady_MultipleProbes_UsingBuilders(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 20)

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	probe1 := tests.NewHttpProbe(server1.URL, "GET")
	probe2 := tests.NewHttpProbeWithBody(server2.URL, "POST", map[string]string{"key": "value"})

	container := tests.NewContainerWithMultipleProbes(probe1, probe2)

	httpClient := tests.NewHttpClientBuilder().
		WithClient("default", server1.Client()).
		Build()

	user := tests.NewUserBuilder().Build()

	result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.NoError(t, err)
	assert.True(t, result)
	assert.True(t, probe1.Solved)
	assert.True(t, probe2.Solved)
}

// ============================================================================
// TESTING DIFFERENT STATES
// ============================================================================

func TestReady_ContainerNotRunning_UsingBuilder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	container := tests.NewContainerWithCommandProbe("echo", "test")
	container = tests.NewContainerBuilder().
		WithState("exited").
		WithStatus(status.DEAD).
		WithReadiness(container.GetReadiness()...).
		Build()

	httpClient := tests.NewHttpClientBuilder().Build()
	user := tests.NewUserBuilder().Build()

	_, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container is not in valid state for readiness checking")
}

// ============================================================================
// TESTING INVALID STATES
// ============================================================================

func TestReady_InvalidState_UsingBuilder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	container := tests.NewContainerWithURLProbe("http://example.com")
	container = tests.NewContainerBuilder().
		WithState("running").
		WithStatus("invalid-state").
		WithReadiness(container.GetReadiness()...).
		Build()

	httpClient := tests.NewHttpClientBuilder().Build()
	user := tests.NewUserBuilder().Build()

	_, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container is not in valid state")
}

// ============================================================================
// TESTING TIMEOUT SCENARIOS
// ============================================================================

func TestReady_ShortTimeout_UsingBuilder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	probe := tests.NewReadinessBuilder().
		WithURL(server.URL).
		WithTimeout("100ms").
		Build()

	container := tests.NewContainerWithSingleProbe(probe)

	httpClient := tests.NewHttpClientBuilder().
		WithClient("test", server.Client()).
		Build()

	user := tests.NewUserBuilder().WithUsername("test").Build()

	result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.Error(t, err)
	assert.False(t, result)
}

// ============================================================================
// CUSTOM SCENARIO
// ============================================================================

func TestReady_CustomScenario_UsingBuilder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	container := tests.NewContainerBuilder().
		WithGeneratedName("custom-app-1").
		WithGroup("production").
		WithName("custom-app").
		WithState("running").
		WithStatus(status.READINESS_CHECKING).
		AddReadinessProbe(tests.NewHttpProbe(server.URL+"/health", "GET")).
		AddReadinessProbe(tests.NewHttpProbe(server.URL+"/ready", "GET")).
		Build()

	httpClient := tests.NewHttpClientBuilder().
		WithClient("production-user", server.Client()).
		Build()

	user := tests.NewUserBuilder().
		WithUsername("production-user").
		Build()

	result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, "custom-app-1", container.GetGeneratedName())
	assert.Equal(t, "production", container.GetGroup())
}

// ============================================================================
// TABLE-DRIVEN TESTS
// ============================================================================

func TestReady_VariousHTTPStatuses(t *testing.T) {
	testsTable := []struct {
		name           string
		statusCode     int
		expectedError  bool
		expectedSolved bool
	}{
		{"200 OK", http.StatusOK, false, true},
		{"201 Created", http.StatusCreated, true, false},
		{"400 Bad Request", http.StatusBadRequest, true, false},
		{"404 Not Found", http.StatusNotFound, true, false},
		{"500 Internal Error", http.StatusInternalServerError, true, false},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true, false},
	}

	for _, tt := range testsTable {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := zap.NewNop()
			channel := make(chan *readiness.ReadinessState, 10)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			probe := tests.NewHttpProbe(server.URL, "GET")
			probe.Timeout = "1s"
			container := tests.NewContainerWithSingleProbe(probe)

			httpClient := tests.NewHttpClientBuilder().
				WithClient("test", server.Client()).
				Build()

			user := tests.NewUserBuilder().WithUsername("test").Build()

			result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

			if tt.expectedError {
				assert.Error(t, err)
				assert.False(t, result)
			} else {
				assert.NoError(t, err)
				assert.True(t, result)
			}
			assert.Equal(t, tt.expectedSolved, probe.Solved)
		})
	}
}

// ============================================================================
// CONTEXT CANCELLATION
// ============================================================================

func TestReady_ContextCancellation_UsingBuilder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := zap.NewNop()
	channel := make(chan *readiness.ReadinessState, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	probe := tests.NewHttpProbe(server.URL, "GET")
	container := tests.NewContainerWithSingleProbe(probe)

	httpClient := tests.NewHttpClientBuilder().
		WithClient("test", server.Client()).
		Build()

	user := tests.NewUserBuilder().WithUsername("test").Build()

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	result, err := solver.Ready(ctx, httpClient, container, user, channel, logger)

	assert.Error(t, err)
	assert.False(t, result)
	assert.Equal(t, context.Canceled, err)
}
