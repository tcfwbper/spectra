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
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for EventProcessor race tests ---

// raceMockSessionForEvent is a thread-safe mock Session for event race tests.
type raceMockSessionForEvent struct {
	mu           sync.RWMutex
	status       string
	currentState string
	sessionData  map[string]any
	sessionID    string
	workflowName string
	eventHistory []session.Event
	err          error
	failCalls    int
}

func newRaceMockSessionForEvent(status, currentState string) *raceMockSessionForEvent {
	return &raceMockSessionForEvent{
		status:       status,
		currentState: currentState,
		sessionData:  make(map[string]any),
		sessionID:    uuid.New().String(),
		workflowName: "TestWorkflow",
		eventHistory: []session.Event{},
	}
}

func (m *raceMockSessionForEvent) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *raceMockSessionForEvent) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *raceMockSessionForEvent) GetSessionDataSafe(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.sessionData[key]
	return val, ok
}

func (m *raceMockSessionForEvent) UpdateEventHistorySafe(event session.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHistory = append(m.eventHistory, event)
	return nil
}

func (m *raceMockSessionForEvent) UpdateCurrentStateSafe(newState string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentState = newState
	return nil
}

func (m *raceMockSessionForEvent) Fail(err error, terminationNotifier chan<- struct{}) error {
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

func (m *raceMockSessionForEvent) GetID() string {
	return m.sessionID
}

func (m *raceMockSessionForEvent) GetWorkflowName() string {
	return m.workflowName
}

func (m *raceMockSessionForEvent) getEventHistory() []session.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]session.Event(nil), m.eventHistory...)
}

// raceMockWorkflowLoaderForEvent is a thread-safe mock WorkflowDefinitionLoader.
type raceMockWorkflowLoaderForEvent struct {
	wf  *storage.WorkflowDefinition
	err error
}

func (m *raceMockWorkflowLoaderForEvent) Load(workflowName string) (*storage.WorkflowDefinition, error) {
	return m.wf, m.err
}

// raceMockTransitionToNode is a thread-safe mock TransitionToNode.
type raceMockTransitionToNode struct {
	mu   sync.Mutex
	sess *raceMockSessionForEvent
}

func (m *raceMockTransitionToNode) Transition(message string, targetNodeName string, isExitTransition bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sess != nil {
		m.sess.mu.Lock()
		m.sess.currentState = targetNodeName
		m.sess.mu.Unlock()
	}
	return nil
}

// TestProcessEvent_ConcurrentEvents verifies concurrent events from multiple agents
// are handled safely. All events should be recorded (serialized by lock).
func TestProcessEvent_ConcurrentEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForEvent("running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "AgentNode", Type: "agent", AgentRole: "Reviewer"},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "AgentNode"},
			{FromNode: "AgentNode", EventType: "Approved", ToNode: "NodeB"},
		},
	}

	loader := &raceMockWorkflowLoaderForEvent{wf: wf}
	transitioner := &raceMockTransitionToNode{sess: sess}
	terminationNotifier := make(chan struct{}, 10)

	ep, err := runtime.NewEventProcessor(sess, loader, transitioner, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]entities.RuntimeResponse, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			payload, _ := json.Marshal(entities.EventPayload{
				EventType: "Approved",
				Message:   fmt.Sprintf("event from goroutine %d", idx),
				Payload:   json.RawMessage(`{}`),
			})
			msg := entities.RuntimeMessage{
				Type:            "event",
				ClaudeSessionID: claudeSessionID,
				Payload:         payload,
			}
			results[idx] = ep.ProcessEvent(sess.GetID(), msg)
		}(i)
	}
	wg.Wait()

	// All 5 events should be recorded to EventHistory
	events := sess.getEventHistory()
	assert.Len(t, events, goroutines, "all events should be recorded")

	// No data races should be detected (run with -race)
}

// TestProcessEvent_EventRecordingSerialized verifies event recording is serialized
// via session-level write lock.
func TestProcessEvent_EventRecordingSerialized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForEvent("running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "AgentNode", Type: "agent", AgentRole: "Reviewer"},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "AgentNode"},
			{FromNode: "AgentNode", EventType: "Approved", ToNode: "NodeB"},
		},
	}

	loader := &raceMockWorkflowLoaderForEvent{wf: wf}
	transitioner := &raceMockTransitionToNode{sess: sess}
	terminationNotifier := make(chan struct{}, 10)

	ep, err := runtime.NewEventProcessor(sess, loader, transitioner, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 3
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			payload, _ := json.Marshal(entities.EventPayload{
				EventType: "Approved",
				Message:   fmt.Sprintf("event %d", idx),
				Payload:   json.RawMessage(`{}`),
			})
			msg := entities.RuntimeMessage{
				Type:            "event",
				ClaudeSessionID: claudeSessionID,
				Payload:         payload,
			}
			ep.ProcessEvent(sess.GetID(), msg)
		}(i)
	}
	wg.Wait()

	events := sess.getEventHistory()
	assert.Len(t, events, goroutines, "all events should be serialized and recorded")
}

// TestProcessEvent_ConcurrentEventAndError verifies EventProcessor's reliance on
// Session's thread-safety when concurrent with ErrorProcessor.
func TestProcessEvent_ConcurrentEventAndError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForEvent("running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "AgentNode", Type: "agent", AgentRole: "Reviewer"},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "AgentNode"},
			{FromNode: "AgentNode", EventType: "Approved", ToNode: "NodeB"},
		},
	}

	eventLoader := &raceMockWorkflowLoaderForEvent{wf: wf}
	transitioner := &raceMockTransitionToNode{sess: sess}
	terminationNotifier := make(chan struct{}, 10)

	ep, err := runtime.NewEventProcessor(sess, eventLoader, transitioner, terminationNotifier)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Event processing
	go func() {
		defer wg.Done()
		payload, _ := json.Marshal(entities.EventPayload{
			EventType: "Approved",
			Message:   "test event",
			Payload:   json.RawMessage(`{}`),
		})
		msg := entities.RuntimeMessage{
			Type:            "event",
			ClaudeSessionID: claudeSessionID,
			Payload:         payload,
		}
		ep.ProcessEvent(sess.GetID(), msg)
	}()

	// Goroutine 2: Simulate error processing by calling Session.Fail
	go func() {
		defer wg.Done()
		sess.Fail(
			&session.RuntimeError{Issuer: "test", Message: "concurrent error"},
			terminationNotifier,
		)
	}()

	wg.Wait()

	// No data races should be detected (run with -race)
	// Either the event wins or the error wins, but no data corruption
	finalStatus := sess.GetStatusSafe()
	assert.Contains(t, []string{"running", "failed"}, finalStatus)
}
