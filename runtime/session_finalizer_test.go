package runtime

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
)

// =============================================================================
// Test Helpers — SessionFinalizer
// =============================================================================
//
// Production surface expected in runtime/session_finalizer.go:
//   - type SessionFinalizer struct { ... }
//   - func NewSessionFinalizer(logger logger.Logger) *SessionFinalizer
//   - func (sf *SessionFinalizer) Finalize(session *PersistentSession) int
//
// The SessionFinalizer reads session status and error from PersistentSession
// and logs via Logger. It returns an exit code (0 for completed, 1 otherwise).
// =============================================================================

// --- Fixture Builders: SessionFinalizer ---

// sessionFinalizerFixture provides a ready-to-use SessionFinalizer with mocks.
type sessionFinalizerFixture struct {
	logger *mockLogger
	// sf will hold the SessionFinalizer once production surface exists.
}

func newSessionFinalizerFixture(t *testing.T) *sessionFinalizerFixture {
	t.Helper()
	return &sessionFinalizerFixture{
		logger: newDefaultMockLogger(),
	}
}

// --- Helper: build mock PersistentSession for finalizer tests ---

// finalizerPSBuilder constructs a PersistentSession backed by a mockSession
// configured for SessionFinalizer test scenarios.
type finalizerPSBuilder struct {
	sess      *mockSession
	metaStore *mockSessionMetadataStore
	evStore   *mockEventStore
	log       *mockLogger
}

func newFinalizerPSBuilder() *finalizerPSBuilder {
	return &finalizerPSBuilder{
		sess:      newDefaultMockSession(),
		metaStore: newDefaultMockMetadataStore(),
		evStore:   newDefaultMockEventStore(),
		log:       newDefaultMockLogger(),
	}
}

func (b *finalizerPSBuilder) withStatus(status string) *finalizerPSBuilder {
	b.sess.getStatusResult = status
	return b
}

func (b *finalizerPSBuilder) withError(err error) *finalizerPSBuilder {
	b.sess.getErrorResult = err
	return b
}

func (b *finalizerPSBuilder) withID(id string) *finalizerPSBuilder {
	b.sess.id = id
	return b
}

func (b *finalizerPSBuilder) withWorkflowName(name string) *finalizerPSBuilder {
	b.sess.workflowName = name
	return b
}

func (b *finalizerPSBuilder) build() *PersistentSession {
	return NewPersistentSession(b.sess, b.metaStore, b.evStore, b.log)
}

// --- Helper: build AgentError for tests ---

func mustNewAgentError(t *testing.T, agentRole, message, failingState string, detail json.RawMessage) *entities.AgentError {
	t.Helper()
	ae, err := entities.NewAgentError(agentRole, message, detail, 1700000000, testSessionID, failingState)
	require.NoError(t, err, "mustNewAgentError")
	return ae
}

// --- Helper: build RuntimeError for tests ---

func mustNewRuntimeError(t *testing.T, issuer, message, failingState string, detail json.RawMessage) *entities.RuntimeError {
	t.Helper()
	re, err := entities.NewRuntimeError(issuer, message, detail, 1700000000, testSessionID, failingState)
	require.NoError(t, err, "mustNewRuntimeError")
	return re
}

// --- Assertion Helpers: SessionFinalizer ---

// assertLogDoesNotContainKey checks that no log call with the given message
// contains the specified key.
func assertLogDoesNotContainKey(t *testing.T, calls []logCall, expectedMsg string, key string) {
	t.Helper()
	for _, call := range calls {
		if call.msg != expectedMsg {
			continue
		}
		for i := 0; i+1 < len(call.args); i += 2 {
			if call.args[i] == key {
				t.Errorf("expected log call with msg=%q to NOT contain key %q, but found value=%v", expectedMsg, key, call.args[i+1])
				return
			}
		}
	}
}

// marshalDetailJSON is a helper to produce compact JSON detail for assertions.
func marshalDetailJSON(t *testing.T, detail map[string]string) string {
	t.Helper()
	b, err := json.Marshal(detail)
	require.NoError(t, err)
	return string(b)
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewSessionFinalizer_ValidLogger(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer does not exist yet")

	f := newSessionFinalizerFixture(t)

	// Act
	sf := NewSessionFinalizer(f.logger)

	// Assert
	require.NotNil(t, sf)
}

// =============================================================================
// Validation Failures
// =============================================================================

func TestNewSessionFinalizer_NilLogger(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer does not exist yet")

	// Act & Assert: panics with message indicating nil logger
	assert.PanicsWithValue(t, "NewSessionFinalizer: logger must not be nil", func() {
		NewSessionFinalizer(nil)
	})
}

// =============================================================================
// Happy Path — Finalize
// =============================================================================

func TestSessionFinalizer_Finalize_Completed(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	ps := newFinalizerPSBuilder().
		withStatus("completed").
		withID("sess-1").
		withWorkflowName("wf-1").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 0, exitCode)
	assertLogHasMessage(t, f.logger.infoCalls, "session completed")
	assertLogContainsArg(t, f.logger.infoCalls, "session completed", "sessionID", "sess-1")
	assertLogContainsArg(t, f.logger.infoCalls, "session completed", "workflow", "wf-1")
}

func TestSessionFinalizer_Finalize_FailedWithAgentError(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	detail := json.RawMessage(`{"key":"val"}`)
	agentErr := mustNewAgentError(t, "parser", "agent broke", "node_3", detail)

	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(agentErr).
		withID("sess-1").
		withWorkflowName("wf-1").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "agent", "parser")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "state", "node_3")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "detail", `{"key":"val"}`)
}

func TestSessionFinalizer_Finalize_FailedWithRuntimeError(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	detail := json.RawMessage(`{"elapsed":"30s"}`)
	rtErr := mustNewRuntimeError(t, "SessionInitializer", "timeout", "entry", detail)

	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(rtErr).
		withID("sess-1").
		withWorkflowName("wf-1").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "issuer", "SessionInitializer")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "state", "entry")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "detail", `{"elapsed":"30s"}`)
}

func TestSessionFinalizer_Finalize_FailedWithAgentError_EmptyDetail(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	// AgentError with empty detail (empty JSON object)
	agentErr := mustNewAgentError(t, "runner", "err", "s1", json.RawMessage(`{}`))

	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(agentErr).
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogDoesNotContainKey(t, f.logger.errorCalls, "session failed", "detail")
}

func TestSessionFinalizer_Finalize_FailedWithAgentError_NilDetail(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	// AgentError with nil detail
	agentErr := mustNewAgentError(t, "runner", "err", "s1", nil)

	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(agentErr).
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogDoesNotContainKey(t, f.logger.errorCalls, "session failed", "detail")
}

// =============================================================================
// Null / Empty Input
// =============================================================================

func TestSessionFinalizer_Finalize_NilSession(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(nil)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "SessionFinalizer called with nil session")
}

func TestSessionFinalizer_Finalize_FailedWithNilError(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(nil).
		withID("sess-1").
		withWorkflowName("wf-1").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "error", "unknown error")
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestSessionFinalizer_Finalize_FailedWithUnexpectedErrorType(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(errors.New("something went wrong")).
		withID("sess-1").
		withWorkflowName("wf-1").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogContainsArg(t, f.logger.errorCalls, "session failed", "error", "something went wrong")
}

func TestSessionFinalizer_Finalize_FailedWithDetailSerializationError(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet; also needs AgentError with non-serializable Detail which is not representable via current entity API (Detail is json.RawMessage)")

	// This test verifies fallback when detail JSON serialization fails.
	// NOTE: Since Detail() returns json.RawMessage (pre-serialized bytes), the
	// production code must handle the case where Detail contains invalid JSON that
	// cannot be compacted. We use an invalid JSON byte slice as the detail.
	//
	// The test spec mentions math.Inf(1) or channel, but those apply to a
	// map[string]any representation. With json.RawMessage, we simulate failure
	// via malformed JSON bytes.
	_ = math.Inf(1) // referenced by spec; used as conceptual marker only

	f := newSessionFinalizerFixture(t)

	// We need a way to produce an AgentError whose Detail() returns bytes
	// that json.Compact or similar would reject. Since NewAgentError validates
	// detail, we may need a special mock or the production surface may handle
	// this differently. This test remains scaffolded.

	_ = f
}

// =============================================================================
// State Transitions
// =============================================================================

func TestSessionFinalizer_Finalize_NonTerminalInitializing(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	ps := newFinalizerPSBuilder().
		withStatus("initializing").
		withError(nil).
		withID("sess-2").
		withWorkflowName("wf-2").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.warnCalls, "session terminated with non-terminal status")
	assertLogContainsArg(t, f.logger.warnCalls, "session terminated with non-terminal status", "status", "initializing")
}

func TestSessionFinalizer_Finalize_NonTerminalRunning(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	ps := newFinalizerPSBuilder().
		withStatus("running").
		withError(nil).
		withID("sess-3").
		withWorkflowName("wf-3").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.warnCalls, "session terminated with non-terminal status")
	assertLogContainsArg(t, f.logger.warnCalls, "session terminated with non-terminal status", "status", "running")
}

func TestSessionFinalizer_Finalize_NonTerminalWithError(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	agentErr := mustNewAgentError(t, "worker", "interrupted", "node_5", nil)

	ps := newFinalizerPSBuilder().
		withStatus("running").
		withError(agentErr).
		withID("sess-3").
		withWorkflowName("wf-3").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.warnCalls, "session terminated with non-terminal status")
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
}

// =============================================================================
// Idempotency
// =============================================================================

func TestSessionFinalizer_Finalize_CalledTwice(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	ps := newFinalizerPSBuilder().
		withStatus("completed").
		withID("sess-4").
		withWorkflowName("wf-4").
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode1 := sf.Finalize(ps)
	exitCode2 := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 0, exitCode1)
	assert.Equal(t, 0, exitCode2)
	assert.Len(t, f.logger.infoCalls, 2)
	// Both calls should log identical messages
	assert.Equal(t, f.logger.infoCalls[0].msg, f.logger.infoCalls[1].msg)
	assert.Equal(t, f.logger.infoCalls[0].args, f.logger.infoCalls[1].args)
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestSessionFinalizer_Finalize_NopLogger(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	// Use a no-op logger that silently drops all messages.
	nopLogger := &nopLoggerMock{}
	rtErr := mustNewRuntimeError(t, "SomeIssuer", "timeout", "node_1", nil)

	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(rtErr).
		build()

	// Act
	sf := NewSessionFinalizer(nopLogger)
	exitCode := sf.Finalize(ps)

	// Assert: returns correct exit code without panic
	assert.Equal(t, 1, exitCode)
}

func TestSessionFinalizer_Finalize_RuntimeError_EmptyDetail(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionFinalizer/Finalize does not exist yet")

	f := newSessionFinalizerFixture(t)
	rtErr := mustNewRuntimeError(t, "SomeIssuer", "timeout", "node_1", json.RawMessage(`{}`))

	ps := newFinalizerPSBuilder().
		withStatus("failed").
		withError(rtErr).
		build()

	// Act
	sf := NewSessionFinalizer(f.logger)
	exitCode := sf.Finalize(ps)

	// Assert
	assert.Equal(t, 1, exitCode)
	assertLogHasMessage(t, f.logger.errorCalls, "session failed")
	assertLogDoesNotContainKey(t, f.logger.errorCalls, "session failed", "detail")
}

// --- nopLoggerMock: silently drops all log messages ---

type nopLoggerMock struct{}

func (n *nopLoggerMock) Debug(msg string, args ...any) {}
func (n *nopLoggerMock) Info(msg string, args ...any)  {}
func (n *nopLoggerMock) Warn(msg string, args ...any)  {}
func (n *nopLoggerMock) Error(msg string, args ...any) {}
