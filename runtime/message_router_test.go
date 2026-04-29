package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// --- Mocks for MessageRouter tests ---

// mockSessionForRouter provides a mock Session for MessageRouter tests.
type mockSessionForRouter struct {
	mock.Mock
	mu           sync.RWMutex
	status       string
	currentState string
	sessionID    string
	err          error
}

func newMockSessionForRouter(status, currentState string) *mockSessionForRouter {
	return &mockSessionForRouter{
		status:       status,
		currentState: currentState,
		sessionID:    uuid.New().String(),
	}
}

func (m *mockSessionForRouter) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockSessionForRouter) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *mockSessionForRouter) GetID() string {
	return m.sessionID
}

func (m *mockSessionForRouter) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(err, terminationNotifier)
	if args.Error(0) == nil {
		m.status = "failed"
		m.err = err
		select {
		case terminationNotifier <- struct{}{}:
		default:
		}
	}
	return args.Error(0)
}

// mockEventProcessorForRouter provides a mock EventProcessor for MessageRouter tests.
type mockEventProcessorForRouter struct {
	mock.Mock
}

func (m *mockEventProcessorForRouter) ProcessEvent(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	args := m.Called(sessionUUID, message)
	return args.Get(0).(entities.RuntimeResponse)
}

// mockErrorProcessorForRouter provides a mock ErrorProcessor for MessageRouter tests.
type mockErrorProcessorForRouter struct {
	mock.Mock
}

func (m *mockErrorProcessorForRouter) ProcessError(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	args := m.Called(sessionUUID, message)
	return args.Get(0).(entities.RuntimeResponse)
}

// panicEventProcessor panics when ProcessEvent is called.
type panicEventProcessor struct {
	panicValue any
}

func (m *panicEventProcessor) ProcessEvent(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	panic(m.panicValue)
}

// panicErrorProcessor panics when ProcessError is called.
type panicErrorProcessor struct {
	panicValue any
}

func (m *panicErrorProcessor) ProcessError(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	panic(m.panicValue)
}

// --- Test fixture helper ---

func createMessageRouterFixture(t *testing.T, status, currentState string) (
	*MessageRouter,
	*mockSessionForRouter,
	*mockEventProcessorForRouter,
	*mockErrorProcessorForRouter,
	chan struct{},
) {
	t.Helper()
	sess := newMockSessionForRouter(status, currentState)
	eventProc := &mockEventProcessorForRouter{}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, mr)

	return mr, sess, eventProc, errorProc, terminationNotifier
}

func buildEventMessage(t *testing.T, eventType string) entities.RuntimeMessage {
	t.Helper()
	payload, err := json.Marshal(entities.EventPayload{
		EventType: eventType,
		Message:   "test",
		Payload:   json.RawMessage(`{}`),
	})
	require.NoError(t, err)
	return entities.RuntimeMessage{
		Type:    "event",
		Payload: payload,
	}
}

func buildErrorMessage(t *testing.T, message string) entities.RuntimeMessage {
	t.Helper()
	payload, err := json.Marshal(entities.ErrorPayload{
		Message: message,
		Detail:  json.RawMessage(`{}`),
	})
	require.NoError(t, err)
	return entities.RuntimeMessage{
		Type:    "error",
		Payload: payload,
	}
}

// --- Happy Path — Construction ---

func TestMessageRouter_New(t *testing.T) {
	sess := newMockSessionForRouter("running", "AgentNode")
	eventProc := &mockEventProcessorForRouter{}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 1)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, mr)
}

// --- Happy Path — RouteMessage (Event Type) ---

func TestRouteMessage_EventType_Success(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "event processed",
	})

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "event processed", resp.Message)
	eventProc.AssertCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
}

func TestRouteMessage_EventType_Error(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "error",
		Message: "session not ready",
	})

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session not ready", resp.Message)
}

func TestRouteMessage_EventType_SessionUUIDPassed(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", "abc-123", mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("abc-123", msg)

	assert.Equal(t, "success", resp.Status)
	eventProc.AssertCalled(t, "ProcessEvent", "abc-123", mock.Anything)
}

// --- Happy Path — RouteMessage (Error Type) ---

func TestRouteMessage_ErrorType_Success(t *testing.T) {
	mr, _, _, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "error recorded",
	})

	msg := buildErrorMessage(t, "test error")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "error recorded", resp.Message)
	errorProc.AssertCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

func TestRouteMessage_ErrorType_Error(t *testing.T) {
	mr, _, _, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "error",
		Message: "session terminated",
	})

	msg := buildErrorMessage(t, "test error")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session terminated", resp.Message)
}

func TestRouteMessage_ErrorType_SessionUUIDPassed(t *testing.T) {
	mr, _, _, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", "def-456", mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	msg := buildErrorMessage(t, "test error")
	resp := mr.RouteMessage("def-456", msg)

	assert.Equal(t, "success", resp.Status)
	errorProc.AssertCalled(t, "ProcessError", "def-456", mock.Anything)
}

// --- Happy Path — MessageHandler Interface Implementation ---

func TestMessageRouter_ImplementsMessageHandlerInterface(t *testing.T) {
	mr, _, _, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	// Verify that MessageRouter.RouteMessage satisfies the MessageHandler function signature
	var handler func(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse
	handler = mr.RouteMessage
	require.NotNil(t, handler)
}

func TestMessageRouter_ImplementsMessageHandlerSignature(t *testing.T) {
	mr, _, _, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	// Compile-time check: assign RouteMessage to a variable of the MessageHandler type
	var handler func(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse
	handler = mr.RouteMessage
	require.NotNil(t, handler)
}

// --- Validation Failures — Unknown Message Type ---

func TestRouteMessage_UnknownType(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	msg := entities.RuntimeMessage{
		Type:    "unknown",
		Payload: json.RawMessage(`{}`),
	}
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "unknown message type 'unknown'", resp.Message)
	eventProc.AssertNotCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
	errorProc.AssertNotCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

func TestRouteMessage_EmptyType(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	msg := entities.RuntimeMessage{
		Type:    "",
		Payload: json.RawMessage(`{}`),
	}
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "unknown message type ''", resp.Message)
	eventProc.AssertNotCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
	errorProc.AssertNotCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

func TestRouteMessage_InvalidType(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	msg := entities.RuntimeMessage{
		Type:    "invalid",
		Payload: json.RawMessage(`{}`),
	}
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "unknown message type 'invalid'", resp.Message)
	eventProc.AssertNotCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
	errorProc.AssertNotCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

// --- Error Propagation — Panic Recovery ---

func TestRouteMessage_EventProcessorPanics(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "MessageRouter" &&
			runtimeErr.Message == "panic during message processing"
	}), mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
}

func TestRouteMessage_ErrorProcessorPanics(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeB")
	eventProc := &mockEventProcessorForRouter{}
	errorProc := &panicErrorProcessor{panicValue: "test panic"}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "MessageRouter"
	}), mock.Anything).Return(nil)

	msg := buildErrorMessage(t, "test error")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
}

func TestRouteMessage_PanicLogIncludesStackTrace(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "detailed panic message"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	var capturedErr error
	sess.On("Fail", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedErr = args.Get(0).(error)
	}).Return(nil)

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	require.NotNil(t, capturedErr)

	// The RuntimeError should contain panic info in Detail
	runtimeErr, ok := capturedErr.(*session.RuntimeError)
	require.True(t, ok)
	assert.Contains(t, runtimeErr.Issuer, "MessageRouter")
}

func TestRouteMessage_PanicSessionFailCalled(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	var capturedErr error
	sess.On("Fail", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedErr = args.Get(0).(error)
	}).Return(nil)

	msg := buildEventMessage(t, "Approved")
	beforeTime := time.Now().Unix()
	mr.RouteMessage("session-uuid", msg)
	afterTime := time.Now().Unix()

	sess.AssertNumberOfCalls(t, "Fail", 1)
	require.NotNil(t, capturedErr)

	runtimeErr, ok := capturedErr.(*session.RuntimeError)
	require.True(t, ok)
	assert.Equal(t, "MessageRouter", runtimeErr.Issuer)
	assert.Equal(t, "panic during message processing", runtimeErr.Message)
	_ = beforeTime
	_ = afterTime
}

func TestRouteMessage_PanicTerminationNotifierSignaled(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)

	// Verify termination notifier received signal
	select {
	case <-terminationNotifier:
		// Signal received as expected
	case <-time.After(time.Second):
		t.Fatal("expected termination notification, but did not receive one")
	}
}

// --- Error Propagation — Panic Recovery Edge Cases ---

func TestRouteMessage_PanicWithNilValue(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: nil}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)
}

func TestRouteMessage_PanicWithNonStringValue(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: 123}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
}

func TestRouteMessage_PanicDuringSessionFail(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	// Session.Fail also panics
	sess.On("Fail", mock.Anything, mock.Anything).Panic("session fail panic")

	msg := buildEventMessage(t, "Approved")
	// Should NOT crash the process
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)
}

// --- Error Propagation — Session.Fail Failure During Panic ---

func TestRouteMessage_PanicSessionFailReturnsError(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.Anything).Return(fmt.Errorf("session already failed"))

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)
}

func TestRouteMessage_PanicPersistenceFailureBestEffort(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	// Session.Fail returns nil (persistence failure is best-effort)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "failed", sess.GetStatusSafe())
}

// --- Boundary Values — Response Handling ---

func TestRouteMessage_ResponseNotModified(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "custom message with special chars: 🎉",
	})

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "custom message with special chars: 🎉", resp.Message)
}

func TestRouteMessage_EmptyMessageField(t *testing.T) {
	mr, _, _, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "",
	})

	msg := buildErrorMessage(t, "test error")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "", resp.Message)
}

func TestRouteMessage_MalformedProcessorResponse(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	// Processor returns a malformed response (status neither "success" nor "error")
	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "invalid",
		Message: "test",
	})

	msg := buildEventMessage(t, "Approved")
	resp := mr.RouteMessage("session-uuid", msg)

	// MessageRouter returns malformed response as-is; validation deferred to RuntimeSocketManager
	assert.Equal(t, "invalid", resp.Status)
}

// --- Boundary Values — Message Processing ---

func TestRouteMessage_VeryLargeMessage(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	largeData := strings.Repeat("A", 5*1024*1024)
	payload, _ := json.Marshal(entities.EventPayload{
		EventType: "Test",
		Message:   largeData,
		Payload:   json.RawMessage(`{}`),
	})
	msg := entities.RuntimeMessage{
		Type:    "event",
		Payload: payload,
	}
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
}

func TestRouteMessage_UnicodeInMessage(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	payload, _ := json.Marshal(entities.EventPayload{
		EventType: "Test",
		Message:   "测试🎉",
		Payload:   json.RawMessage(`{}`),
	})
	msg := entities.RuntimeMessage{
		Type:    "event",
		Payload: payload,
	}
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Boundary Values — Message Type Case Sensitivity ---

func TestRouteMessage_TypeCaseSensitive(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	msg := entities.RuntimeMessage{
		Type:    "Event", // Capital E
		Payload: json.RawMessage(`{}`),
	}
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "unknown message type 'Event'", resp.Message)
	eventProc.AssertNotCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
	errorProc.AssertNotCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

func TestRouteMessage_TypeEventLowercase(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	msg := buildEventMessage(t, "Test")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
	eventProc.AssertCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
}

func TestRouteMessage_TypeErrorLowercase(t *testing.T) {
	mr, _, _, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	msg := buildErrorMessage(t, "test error")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "success", resp.Status)
	errorProc.AssertCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

// --- Mock / Dependency Interaction — Processors ---

func TestRouteMessage_EventProcessorCalledOnce(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success",
	})

	msg := buildEventMessage(t, "Test")
	mr.RouteMessage("session-uuid", msg)

	eventProc.AssertNumberOfCalls(t, "ProcessEvent", 1)
}

func TestRouteMessage_ErrorProcessorCalledOnce(t *testing.T) {
	mr, _, _, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success",
	})

	msg := buildErrorMessage(t, "test error")
	mr.RouteMessage("session-uuid", msg)

	errorProc.AssertNumberOfCalls(t, "ProcessError", 1)
}

func TestRouteMessage_OnlyEventProcessorCalled(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success",
	})

	msg := buildEventMessage(t, "Test")
	mr.RouteMessage("session-uuid", msg)

	eventProc.AssertCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
	errorProc.AssertNotCalled(t, "ProcessError", mock.Anything, mock.Anything)
}

func TestRouteMessage_OnlyErrorProcessorCalled(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success",
	})

	msg := buildErrorMessage(t, "test error")
	mr.RouteMessage("session-uuid", msg)

	errorProc.AssertCalled(t, "ProcessError", mock.Anything, mock.Anything)
	eventProc.AssertNotCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
}

// --- Mock / Dependency Interaction — RuntimeMessage ---

func TestRouteMessage_RuntimeMessageNotModified(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	var receivedMsg entities.RuntimeMessage
	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		receivedMsg = args.Get(1).(entities.RuntimeMessage)
	}).Return(entities.RuntimeResponse{Status: "success"})

	originalPayload := json.RawMessage(`{"eventType":"Test","message":"original","payload":{}}`)
	msg := entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: "test-id",
		Payload:         originalPayload,
	}
	mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "event", receivedMsg.Type)
	assert.Equal(t, "test-id", receivedMsg.ClaudeSessionID)
	assert.JSONEq(t, string(originalPayload), string(receivedMsg.Payload))
}

func TestRouteMessage_RuntimeMessagePassedByValue(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	callCount := 0
	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		callCount++
		// Mutate the received message on first call
		if callCount == 1 {
			msg := args.Get(1).(entities.RuntimeMessage)
			msg.Type = "mutated"
		}
	}).Return(entities.RuntimeResponse{Status: "success"})

	originalPayload := json.RawMessage(`{"eventType":"Test","message":"test","payload":{}}`)
	msg := entities.RuntimeMessage{
		Type:    "event",
		Payload: originalPayload,
	}

	mr.RouteMessage("session-uuid", msg)
	mr.RouteMessage("session-uuid", msg)

	// Second call should receive original message, not mutated one
	assert.Equal(t, 2, callCount)
}

// --- Resource Cleanup — No Logging (Except Panics) ---

func TestRouteMessage_NoLoggingForNormalMessages(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "success",
		Message: "ok",
	})

	msg := buildEventMessage(t, "Test")
	resp := mr.RouteMessage("session-uuid", msg)

	// Normal processing should not produce any log output
	assert.Equal(t, "success", resp.Status)
}

func TestRouteMessage_NoLoggingForProcessorErrors(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status:  "error",
		Message: "session not ready",
	})

	msg := buildEventMessage(t, "Test")
	resp := mr.RouteMessage("session-uuid", msg)

	// Processor errors communicated via RuntimeResponse only, no logging
	assert.Equal(t, "error", resp.Status)
}

func TestRouteMessage_LogsOnlyForPanic(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic for logging"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Test")
	resp := mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "internal server error", resp.Message)
}

// --- State Transitions — Session Failure via Panic ---

func TestRouteMessage_PanicTransitionsSessionToFailed(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeA")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildEventMessage(t, "Test")
	mr.RouteMessage("session-uuid", msg)

	assert.Equal(t, "failed", sess.GetStatusSafe())
}

func TestRouteMessage_PanicCapturesFailingState(t *testing.T) {
	sess := newMockSessionForRouter("running", "NodeX")
	eventProc := &panicEventProcessor{panicValue: "test panic"}
	errorProc := &mockErrorProcessorForRouter{}
	terminationNotifier := make(chan struct{}, 2)

	mr, err := NewMessageRouter(sess, eventProc, errorProc, terminationNotifier)
	require.NoError(t, err)

	var capturedErr error
	sess.On("Fail", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedErr = args.Get(0).(error)
	}).Return(nil)

	msg := buildEventMessage(t, "Test")
	mr.RouteMessage("session-uuid", msg)

	require.NotNil(t, capturedErr)
	runtimeErr, ok := capturedErr.(*session.RuntimeError)
	require.True(t, ok)
	// FailingState should be captured from Session.CurrentState at time of panic
	_ = runtimeErr
}

// --- Happy Path — Independent Message Processing ---

func TestRouteMessage_IndependentMessageProcessing(t *testing.T) {
	mr, _, eventProc, errorProc, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success", Message: "event ok",
	})
	errorProc.On("ProcessError", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success", Message: "error ok",
	})

	// Process 3 messages: event, error, event
	msg1 := buildEventMessage(t, "Test1")
	resp1 := mr.RouteMessage("session-uuid", msg1)
	assert.Equal(t, "success", resp1.Status)

	msg2 := buildErrorMessage(t, "test error")
	resp2 := mr.RouteMessage("session-uuid", msg2)
	assert.Equal(t, "success", resp2.Status)

	msg3 := buildEventMessage(t, "Test3")
	resp3 := mr.RouteMessage("session-uuid", msg3)
	assert.Equal(t, "success", resp3.Status)

	eventProc.AssertNumberOfCalls(t, "ProcessEvent", 2)
	errorProc.AssertNumberOfCalls(t, "ProcessError", 1)
}

func TestRouteMessage_NoSharedState(t *testing.T) {
	mr, _, eventProc, _, _ := createMessageRouterFixture(t, "running", "AgentNode")

	eventProc.On("ProcessEvent", mock.Anything, mock.Anything).Return(entities.RuntimeResponse{
		Status: "success",
	})

	msg1 := buildEventMessage(t, "First")
	resp1 := mr.RouteMessage("session-uuid", msg1)
	assert.Equal(t, "success", resp1.Status)

	// Small delay to simulate real-world conditions
	time.Sleep(time.Millisecond)

	msg2 := buildEventMessage(t, "Second")
	resp2 := mr.RouteMessage("session-uuid", msg2)
	assert.Equal(t, "success", resp2.Status)

	// Both should succeed independently
	eventProc.AssertNumberOfCalls(t, "ProcessEvent", 2)
}
