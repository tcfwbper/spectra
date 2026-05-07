package runtime

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
)

// =============================================================================
// Test Helpers — MessageRouter
// =============================================================================
//
// Production surface expected in runtime/message_router.go:
//   - type MessageRouter struct { ... }
//   - func NewMessageRouter(ps *PersistentSession, eventProcessor MessageRouterEventProcessor, errorProcessor MessageRouterErrorProcessor, terminationNotifier chan<- struct{}, logger logger.Logger) *MessageRouter
//   - func (mr *MessageRouter) Handle(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse
//   - type MessageRouterEventProcessor interface { ProcessEvent(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse }
//   - type MessageRouterErrorProcessor interface { ProcessError(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse }
//
// =============================================================================

// --- Mock: EventProcessor interface for MessageRouter ---

type mockRouterEventProcessor struct {
	mu              sync.Mutex
	processEvCalled int
	processEvUUID   string
	processEvMsg    *entities.RuntimeMessage
	processEvResp   *entities.RuntimeResponse
	processEvPanic  any
}

func (m *mockRouterEventProcessor) ProcessEvent(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse {
	m.mu.Lock()
	m.processEvCalled++
	m.processEvUUID = sessionUUID
	m.processEvMsg = msg
	m.mu.Unlock()
	if m.processEvPanic != nil {
		panic(m.processEvPanic)
	}
	return m.processEvResp
}

// --- Mock: ErrorProcessor interface for MessageRouter ---

type mockRouterErrorProcessor struct {
	mu               sync.Mutex
	processErrCalled int
	processErrUUID   string
	processErrMsg    *entities.RuntimeMessage
	processErrResp   *entities.RuntimeResponse
	processErrPanic  any
}

func (m *mockRouterErrorProcessor) ProcessError(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse {
	m.mu.Lock()
	m.processErrCalled++
	m.processErrUUID = sessionUUID
	m.processErrMsg = msg
	m.mu.Unlock()
	if m.processErrPanic != nil {
		panic(m.processErrPanic)
	}
	return m.processErrResp
}

// --- Fixture Builder: MessageRouter ---

type messageRouterFixture struct {
	session             *mockSession
	ps                  *PersistentSession
	eventProcessor      *mockRouterEventProcessor
	errorProcessor      *mockRouterErrorProcessor
	terminationNotifier chan struct{}
	logger              *mockLogger
}

func newMessageRouterFixture(t *testing.T) *messageRouterFixture {
	t.Helper()
	sess := newDefaultMockSession()
	sess.getStatusResult = "running"
	sess.getCurrentStateResult = "NodeA"

	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	return &messageRouterFixture{
		session:             sess,
		ps:                  ps,
		eventProcessor:      &mockRouterEventProcessor{},
		errorProcessor:      &mockRouterErrorProcessor{},
		terminationNotifier: newTerminationChannel(),
		logger:              log,
	}
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewMessageRouter_ValidDeps(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter constructor")

	// Setup
	f := newMessageRouterFixture(t)
	_ = f

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)

	// Assert: Returns non-nil *MessageRouter; no panic
	// require.NotNil(t, mr)
}

// =============================================================================
// Happy Path — Handle
// =============================================================================

func TestMessageRouter_Handle_EventType(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvResp = entities.SuccessResponse("ok")

	msg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "success", resp.Status())
	// assert.Equal(t, "ok", resp.Message())
	// assert.Equal(t, 1, f.eventProcessor.processEvCalled)
}

func TestMessageRouter_Handle_ErrorType(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.errorProcessor.processErrResp = entities.SuccessResponse("error recorded")

	msg := mustNewErrorRuntimeMessage(t, "cs-123", mustValidErrorPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "success", resp.Status())
	// assert.Equal(t, "error recorded", resp.Message())
	// assert.Equal(t, 1, f.errorProcessor.processErrCalled)
}

func TestMessageRouter_Handle_EventProcessorReturnsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvResp = entities.ErrorResponse("session not ready: status is 'failed'")

	msg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "session not ready: status is 'failed'", resp.Message())
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestMessageRouter_Handle_UnknownType(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle; also requires RuntimeMessage to accept unknown types or a test seam")

	// Setup
	f := newMessageRouterFixture(t)
	_ = f

	// Note: entities.NewRuntimeMessage rejects unknown types, so we need either a
	// test seam or the MessageRouter accepts raw-typed messages via an interface.
	// This test is blocked on understanding how the MessageRouter receives messages
	// with unknown types (likely from RuntimeSocketManager bypassing NewRuntimeMessage validation).

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, unknownMsg)

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "unknown message type 'unknown'", resp.Message())
}

func TestMessageRouter_Handle_PanicInEventProcessor(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvPanic = "nil pointer"
	f.session.getCurrentStateResult = "NodeA"

	msg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "internal server error", resp.Message())
	// assert.GreaterOrEqual(t, len(f.logger.errorCalls), 1)
	// assert.Equal(t, 1, f.session.failCalled)
	// rtErr, ok := f.session.failInputErr.(*entities.RuntimeError)
	// require.True(t, ok)
	// assert.Equal(t, "MessageRouter", rtErr.Issuer())
	// assert.Equal(t, "panic during message processing", rtErr.Message())
	// assert.Equal(t, testSessionID, rtErr.SessionID())
	// assert.Equal(t, "NodeA", rtErr.FailingState())
}

func TestMessageRouter_Handle_PanicInErrorProcessor(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.errorProcessor.processErrPanic = "index out of range"
	f.session.getCurrentStateResult = "NodeB"

	msg := mustNewErrorRuntimeMessage(t, "cs-123", mustValidErrorPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "internal server error", resp.Message())
	// assert.GreaterOrEqual(t, len(f.logger.errorCalls), 1)
	// assert.Equal(t, 1, f.session.failCalled)
	// rtErr, ok := f.session.failInputErr.(*entities.RuntimeError)
	// require.True(t, ok)
	// assert.Equal(t, "MessageRouter", rtErr.Issuer())
}

func TestMessageRouter_Handle_PanicRecovery_FailReturnsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvPanic = "boom"
	f.session.failErr = errAlreadyFailed
	f.session.getCurrentStateResult = "NodeA"

	msg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert: Still returns "internal server error" even when Fail fails
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "internal server error", resp.Message())
	// Logger.Error called (logs Fail error)
	// assert.GreaterOrEqual(t, len(f.logger.errorCalls), 1)
}

func TestMessageRouter_Handle_PanicRecovery_RuntimeErrorConstructionFails(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup: conditions that cause NewRuntimeError to fail (e.g., empty sessionID and failingState)
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvPanic = "boom"
	f.session.id = ""
	f.session.getCurrentStateResult = ""

	// Re-create PersistentSession with blank ID
	// Note: PersistentSession.ID is set from session metadata, so we need a session
	// that returns empty metadata
	f.session.getMetadataSnapshotResult.ID = ""

	msg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert: Still returns "internal server error"
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "internal server error", resp.Message())
	// Logger.Error called
	// assert.GreaterOrEqual(t, len(f.logger.errorCalls), 1)
	// Fail() not called (no valid RuntimeError to pass)
	// assert.Equal(t, 0, f.session.failCalled)
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestMessageRouter_Handle_DoesNotModifyMessage(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvResp = entities.SuccessResponse("ok")

	payload := mustValidEventPayload()
	msg := mustNewEventRuntimeMessage(t, "cs-specific", payload)

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// _ = mr.Handle(testSessionID, msg)
	_ = msg

	// Assert: EventProcessor receives the exact same RuntimeMessage reference
	// assert.Equal(t, msg, f.eventProcessor.processEvMsg)
	// assert.Equal(t, "cs-specific", f.eventProcessor.processEvMsg.ClaudeSessionID())
}

func TestMessageRouter_Handle_DoesNotModifyResponse(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	expectedResp := entities.SuccessResponse("custom-message-42")
	f.errorProcessor.processErrResp = expectedResp

	msg := mustNewErrorRuntimeMessage(t, "cs-123", mustValidErrorPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// resp := mr.Handle(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, expectedResp, resp)
}

func TestMessageRouter_Handle_NoLogOnNormalDispatch(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvResp = entities.SuccessResponse("ok")

	msg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())

	// Act
	// mr := NewMessageRouter(f.ps, f.eventProcessor, f.errorProcessor, f.terminationNotifier, f.logger)
	// _ = mr.Handle(testSessionID, msg)
	_ = msg

	// Assert: Logger.Error not called
	// assert.Len(t, f.logger.errorCalls, 0)
}

// =============================================================================
// Concurrent Behaviour
// =============================================================================

func TestMessageRouter_Handle_ConcurrentMessages(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/message_router.go — NewMessageRouter, MessageRouter.Handle")

	// Setup
	f := newMessageRouterFixture(t)
	f.eventProcessor.processEvResp = entities.SuccessResponse("ev-ok")
	f.errorProcessor.processErrResp = entities.SuccessResponse("err-ok")

	evMsg := mustNewEventRuntimeMessage(t, "cs-123", mustValidEventPayload())
	errMsg := mustNewErrorRuntimeMessage(t, "cs-123", mustValidErrorPayload())
	_ = evMsg
	_ = errMsg

	// Act: Call mr.Handle concurrently from two goroutines
	// var wg sync.WaitGroup
	// var resp1, resp2 *entities.RuntimeResponse
	// wg.Add(2)
	// go func() { defer wg.Done(); resp1 = mr.Handle(testSessionID, evMsg) }()
	// go func() { defer wg.Done(); resp2 = mr.Handle(testSessionID, errMsg) }()
	// wg.Wait()

	// Assert: Both complete without data race
	// assert.NotNil(t, resp1)
	// assert.NotNil(t, resp2)
}

// =============================================================================
// Payload Helpers
// =============================================================================

// mustValidEventPayload returns a minimal valid event payload for test messages.
func mustValidEventPayload() json.RawMessage {
	return json.RawMessage(`{"eventType":"MsgSent","message":"hi","payload":{}}`)
}

// mustValidErrorPayload returns a minimal valid error payload for test messages.
func mustValidErrorPayload() json.RawMessage {
	return json.RawMessage(`{"message":"something went wrong"}`)
}

// errAlreadyFailed is a sentinel for tests where session.Fail returns an error.
var errAlreadyFailed = errSentinel("session already failed")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }

// =============================================================================
// Compile guards — suppress unused import warnings
// =============================================================================

var (
	_ = json.RawMessage{}
	_ = (*entities.RuntimeMessage)(nil)
	_ = (*entities.RuntimeResponse)(nil)
	_ = assert.Equal
	_ = require.NoError
	_ = (*sync.WaitGroup)(nil)
)
