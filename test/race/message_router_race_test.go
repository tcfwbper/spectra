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
)

// --- Mocks for MessageRouter race tests ---

// raceMockSessionForRouter is a thread-safe mock Session for MessageRouter race tests.
type raceMockSessionForRouter struct {
	mu           sync.RWMutex
	status       string
	currentState string
	sessionID    string
	err          error
	failCalls    int
}

func newRaceMockSessionForRouter(status, currentState string) *raceMockSessionForRouter {
	return &raceMockSessionForRouter{
		status:       status,
		currentState: currentState,
		sessionID:    uuid.New().String(),
	}
}

func (m *raceMockSessionForRouter) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *raceMockSessionForRouter) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *raceMockSessionForRouter) GetID() string {
	return m.sessionID
}

func (m *raceMockSessionForRouter) Fail(err error, terminationNotifier chan<- struct{}) error {
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

// raceSafeEventProcessor is a thread-safe mock EventProcessor for race tests.
type raceSafeEventProcessor struct {
	mu       sync.Mutex
	response entities.RuntimeResponse
}

func (m *raceSafeEventProcessor) ProcessEvent(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.response
}

// raceSafeErrorProcessor is a thread-safe mock ErrorProcessor for race tests.
type raceSafeErrorProcessor struct {
	mu       sync.Mutex
	response entities.RuntimeResponse
}

func (m *raceSafeErrorProcessor) ProcessError(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.response
}

// racePanicEventProcessor panics conditionally based on payload content.
type racePanicEventProcessor struct {
	mu          sync.Mutex
	response    entities.RuntimeResponse
	panicMarker string // If eventType matches this, panic
}

func (m *racePanicEventProcessor) ProcessEvent(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	// Parse eventType from payload
	var ep entities.EventPayload
	_ = json.Unmarshal(message.Payload, &ep)
	if ep.EventType == m.panicMarker {
		panic("triggered panic for " + ep.EventType)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.response
}

// TestRouteMessage_ConcurrentCalls verifies concurrent RouteMessage calls are handled safely.
func TestRouteMessage_ConcurrentCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForRouter("running", "AgentNode")
	eventProc := &raceSafeEventProcessor{
		response: entities.RuntimeResponse{Status: "success", Message: "event ok"},
	}
	errorProc := &raceSafeErrorProcessor{
		response: entities.RuntimeResponse{Status: "success", Message: "error ok"},
	}
	terminationNotifier := make(chan struct{}, 20)

	mr, err := runtime.NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]entities.RuntimeResponse, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			var msg entities.RuntimeMessage
			if idx%2 == 0 {
				// Event message
				payload, _ := json.Marshal(entities.EventPayload{
					EventType: "Test",
					Message:   fmt.Sprintf("event %d", idx),
					Payload:   json.RawMessage(`{}`),
				})
				msg = entities.RuntimeMessage{Type: "event", Payload: payload}
			} else {
				// Error message
				payload, _ := json.Marshal(entities.ErrorPayload{
					Message: fmt.Sprintf("error %d", idx),
					Detail:  json.RawMessage(`{}`),
				})
				msg = entities.RuntimeMessage{Type: "error", Payload: payload}
			}
			results[idx] = mr.RouteMessage("session-uuid", msg)
		}(i)
	}
	wg.Wait()

	// All calls should complete successfully
	for i, resp := range results {
		assert.Equal(t, "success", resp.Status, "goroutine %d should succeed", i)
	}
}

// TestRouteMessage_ConcurrentPanics verifies concurrent panics from processors are handled safely.
func TestRouteMessage_ConcurrentPanics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForRouter("running", "AgentNode")
	// EventProcessor that always panics
	eventProc := &racePanicEventProcessor{
		response:    entities.RuntimeResponse{Status: "success", Message: "ok"},
		panicMarker: "TriggerPanic",
	}
	errorProc := &raceSafeErrorProcessor{
		response: entities.RuntimeResponse{Status: "success", Message: "error ok"},
	}
	terminationNotifier := make(chan struct{}, 20)

	mr, err := runtime.NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]entities.RuntimeResponse, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			payload, _ := json.Marshal(entities.EventPayload{
				EventType: "TriggerPanic",
				Message:   fmt.Sprintf("panic event %d", idx),
				Payload:   json.RawMessage(`{}`),
			})
			msg := entities.RuntimeMessage{Type: "event", Payload: payload}
			results[idx] = mr.RouteMessage("session-uuid", msg)
		}(i)
	}
	wg.Wait()

	// All panics should be recovered; all should return error responses
	for i, resp := range results {
		assert.Equal(t, "error", resp.Status, "goroutine %d should return error", i)
		assert.Equal(t, "internal server error", resp.Message, "goroutine %d should return internal server error", i)
	}
	// Process should not crash
}

// TestRouteMessage_PanicIsolation verifies panic in one goroutine does not affect others.
func TestRouteMessage_PanicIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sess := newRaceMockSessionForRouter("running", "AgentNode")
	// EventProcessor that panics only for "TriggerPanic" eventType
	eventProc := &racePanicEventProcessor{
		response:    entities.RuntimeResponse{Status: "success", Message: "ok"},
		panicMarker: "TriggerPanic",
	}
	errorProc := &raceSafeErrorProcessor{
		response: entities.RuntimeResponse{Status: "success", Message: "error ok"},
	}
	terminationNotifier := make(chan struct{}, 20)

	mr, err := runtime.NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(3)

	results := make([]entities.RuntimeResponse, 3)

	// Goroutine 0: panic-triggering message
	go func() {
		defer wg.Done()
		payload, _ := json.Marshal(entities.EventPayload{
			EventType: "TriggerPanic",
			Message:   "this will panic",
			Payload:   json.RawMessage(`{}`),
		})
		msg := entities.RuntimeMessage{Type: "event", Payload: payload}
		results[0] = mr.RouteMessage("session-uuid", msg)
	}()

	// Goroutine 1: normal message
	go func() {
		defer wg.Done()
		payload, _ := json.Marshal(entities.EventPayload{
			EventType: "NormalEvent",
			Message:   "this is normal",
			Payload:   json.RawMessage(`{}`),
		})
		msg := entities.RuntimeMessage{Type: "event", Payload: payload}
		results[1] = mr.RouteMessage("session-uuid", msg)
	}()

	// Goroutine 2: another normal message
	go func() {
		defer wg.Done()
		payload, _ := json.Marshal(entities.EventPayload{
			EventType: "AnotherNormal",
			Message:   "also normal",
			Payload:   json.RawMessage(`{}`),
		})
		msg := entities.RuntimeMessage{Type: "event", Payload: payload}
		results[2] = mr.RouteMessage("session-uuid", msg)
	}()

	wg.Wait()

	// All three should complete (no process crash)
	// The panic one should return error
	assert.Equal(t, "error", results[0].Status, "panic goroutine should return error")
	assert.Equal(t, "internal server error", results[0].Message)

	// The normal ones should succeed
	assert.Equal(t, "success", results[1].Status, "normal goroutine 1 should succeed")
	assert.Equal(t, "success", results[2].Status, "normal goroutine 2 should succeed")
}
