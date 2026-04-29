package race_test

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/runtime"
)

// --- Mocks for SessionFinalizer race tests ---

// raceMockSessionForFinalizer is a thread-safe mock Session for race tests.
type raceMockSessionForFinalizer struct {
	id           string
	status       string
	workflowName string
	err          error
}

func (m *raceMockSessionForFinalizer) GetID() string           { return m.id }
func (m *raceMockSessionForFinalizer) GetStatusSafe() string   { return m.status }
func (m *raceMockSessionForFinalizer) GetWorkflowName() string { return m.workflowName }
func (m *raceMockSessionForFinalizer) GetErrorSafe() error     { return m.err }

// raceMockRSMForFinalizer is a thread-safe mock RuntimeSocketManager for race tests.
type raceMockRSMForFinalizer struct {
	mock.Mock
	mu              sync.Mutex
	deleteCallCount int
}

func (m *raceMockRSMForFinalizer) DeleteSocket() error {
	m.mu.Lock()
	m.deleteCallCount++
	m.mu.Unlock()
	args := m.Called()
	return args.Error(0)
}

func (m *raceMockRSMForFinalizer) getDeleteCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deleteCallCount
}

// raceMockLoggerForFinalizer is a thread-safe logger for race tests.
type raceMockLoggerForFinalizer struct {
	mu       sync.Mutex
	warnings []string
}

func (m *raceMockLoggerForFinalizer) Warning(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnings = append(m.warnings, msg)
}

// =====================================================================
// Concurrent Behaviour
// =====================================================================

func TestFinalize_ConcurrentCalls(t *testing.T) {
	sm := &raceMockRSMForFinalizer{}
	sm.On("DeleteSocket").Return(nil)
	logger := &raceMockLoggerForFinalizer{}

	sf, err := runtime.NewSessionFinalizer(sm, logger)
	require.NoError(t, err)

	sess := &raceMockSessionForFinalizer{
		id:           "concurrent-test",
		status:       "completed",
		workflowName: "TestWorkflow",
	}

	// Redirect stdout to discard output noise during race test
	oldStdout := os.Stdout
	devNull, devErr := os.Open(os.DevNull)
	require.NoError(t, devErr)
	defer devNull.Close()
	os.Stdout = devNull
	defer func() { os.Stdout = oldStdout }()

	var wg sync.WaitGroup
	const goroutines = 5
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			sf.Finalize(sess)
		}()
	}

	wg.Wait()

	// All calls should complete; DeleteSocket should be called 5 times
	count := sm.getDeleteCallCount()
	if count != goroutines {
		t.Errorf("expected DeleteSocket called %d times, got %d", goroutines, count)
	}
}

func TestFinalize_ConcurrentSocketDeletion(t *testing.T) {
	sm := &raceMockRSMForFinalizer{}
	// DeleteSocket is idempotent: first real delete, rest are no-ops
	sm.On("DeleteSocket").Return(nil)
	logger := &raceMockLoggerForFinalizer{}

	sf, err := runtime.NewSessionFinalizer(sm, logger)
	require.NoError(t, err)

	sess := &raceMockSessionForFinalizer{
		id:           "socket-race-test",
		status:       "completed",
		workflowName: "TestWorkflow",
	}

	// Redirect stdout to discard output noise during race test
	oldStdout := os.Stdout
	devNull, devErr := os.Open(os.DevNull)
	require.NoError(t, devErr)
	defer devNull.Close()
	os.Stdout = devNull
	defer func() { os.Stdout = oldStdout }()

	var wg sync.WaitGroup
	const goroutines = 10
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			sf.Finalize(sess)
		}()
	}

	wg.Wait()

	// All calls should complete without panic
	count := sm.getDeleteCallCount()
	if count != goroutines {
		t.Errorf("expected DeleteSocket called %d times, got %d", goroutines, count)
	}
}
