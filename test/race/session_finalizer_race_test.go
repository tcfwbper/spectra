package race_test

import (
	"os"
	"sync"
	"testing"

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
	logger := &raceMockLoggerForFinalizer{}

	sf, err := runtime.NewSessionFinalizer(logger)
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

	// All calls should complete without panic; success message printed 5 times
	// No resource cleanup is performed by SessionFinalizer
}

func TestFinalize_ConcurrentSocketDeletion(t *testing.T) {
	logger := &raceMockLoggerForFinalizer{}

	sf, err := runtime.NewSessionFinalizer(logger)
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

	// All calls should complete without panic; no resource cleanup by SessionFinalizer
}
