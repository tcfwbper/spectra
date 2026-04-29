package race_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for ErrorProcessor race tests ---

// raceMockSessionForError is a thread-safe mock Session for race tests.
type raceMockSessionForError struct {
	mu           sync.RWMutex
	status       string
	currentState string
	sessionData  map[string]any
	sessionID    string
	workflowName string
	failCalls    int
	err          error
}

func newRaceMockSessionForError(status, currentState string) *raceMockSessionForError {
	return &raceMockSessionForError{
		status:       status,
		currentState: currentState,
		sessionData:  make(map[string]any),
		sessionID:    uuid.New().String(),
		workflowName: "TestWorkflow",
	}
}

func (m *raceMockSessionForError) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *raceMockSessionForError) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *raceMockSessionForError) GetSessionDataSafe(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.sessionData[key]
	return val, ok
}

func (m *raceMockSessionForError) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status == "failed" {
		return fmt.Errorf("session already failed")
	}
	m.status = "failed"
	m.err = err
	m.failCalls++
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *raceMockSessionForError) GetID() string {
	return m.sessionID
}

func (m *raceMockSessionForError) GetWorkflowName() string {
	return m.workflowName
}

func (m *raceMockSessionForError) getFailCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failCalls
}

// raceMockWorkflowLoaderForError is a thread-safe mock WorkflowDefinitionLoader.
type raceMockWorkflowLoaderForError struct {
	wf  *storage.WorkflowDefinition
	err error
}

func (m *raceMockWorkflowLoaderForError) Load(workflowName string) (*storage.WorkflowDefinition, error) {
	return m.wf, m.err
}

// TestProcessError_ConcurrentErrors verifies that concurrent errors from multiple agents
// are handled safely. One succeeds (first to acquire lock); others return error.
func TestProcessError_ConcurrentErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForError("running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "AgentNode", Type: "agent", AgentRole: "Reviewer"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "AgentNode"},
		},
	}

	loader := &raceMockWorkflowLoaderForError{wf: wf}
	terminationNotifier := make(chan struct{}, 10)

	ep, err := runtime.NewErrorProcessor(sess, loader, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]entities.RuntimeResponse, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			payload, _ := json.Marshal(entities.ErrorPayload{
				Message: fmt.Sprintf("error from goroutine %d", idx),
				Detail:  json.RawMessage(`{}`),
			})
			msg := entities.RuntimeMessage{
				Type:            "error",
				ClaudeSessionID: claudeSessionID,
				Payload:         payload,
			}
			results[idx] = ep.ProcessError(sess.GetID(), msg)
		}(i)
	}
	wg.Wait()

	// Exactly one should succeed
	successCount := 0
	errorCount := 0
	for _, resp := range results {
		if resp.Status == "success" {
			successCount++
		} else if resp.Status == "error" {
			errorCount++
		}
	}

	assert.Equal(t, 1, successCount, "exactly one goroutine should succeed")
	assert.Equal(t, goroutines-1, errorCount, "remaining goroutines should return error")
	assert.Equal(t, "failed", sess.GetStatusSafe())
}

// TestProcessError_ConcurrentErrorAndEvent verifies ErrorProcessor's reliance on Session's
// thread-safety when concurrent with EventProcessor.
func TestProcessError_ConcurrentErrorAndEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForError("running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "AgentNode", Type: "agent", AgentRole: "Reviewer"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "AgentNode"},
			{FromNode: "AgentNode", EventType: "Done", ToNode: "Entry"},
		},
	}

	errorLoader := &raceMockWorkflowLoaderForError{wf: wf}
	terminationNotifier := make(chan struct{}, 10)

	ep, err := runtime.NewErrorProcessor(sess, errorLoader, terminationNotifier)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)

	var errorResp entities.RuntimeResponse

	// Goroutine 1: Error processing
	go func() {
		defer wg.Done()
		payload, _ := json.Marshal(entities.ErrorPayload{
			Message: "test error",
			Detail:  json.RawMessage(`{}`),
		})
		msg := entities.RuntimeMessage{
			Type:            "error",
			ClaudeSessionID: claudeSessionID,
			Payload:         payload,
		}
		errorResp = ep.ProcessError(sess.GetID(), msg)
	}()

	// Goroutine 2: Simulate event processing by reading session status
	go func() {
		defer wg.Done()
		// Read session status (simulates EventProcessor's status check)
		_ = sess.GetStatusSafe()
		_ = sess.GetCurrentStateSafe()
	}()

	wg.Wait()

	// No data races should be detected (this test is valuable when run with -race)
	_ = errorResp
}
