package runtime

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities/session"
)

// All tests in this file are scaffolded awaiting the production surface:
// - runtime.PersistentSession struct
// - runtime.NewPersistentSession constructor
// - Session interface (defined in runtime package, consumed by PersistentSession)
// - SessionMetadataStore interface (defined in runtime package, consumed by PersistentSession)
// - EventStore interface (defined in runtime package, consumed by PersistentSession)
//
// The mocks in helpers_test.go are designed to satisfy those interfaces once defined.

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewPersistentSession_ValidDeps(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and runtime.PersistentSession type")

	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// require.NotNil(t, ps)
}

// =============================================================================
// Validation Failures
// =============================================================================

func TestNewPersistentSession_NilSession(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor")

	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = metaStore
	_ = evStore
	_ = log

	// assert.PanicsWithValue(t, "NewPersistentSession: session must not be nil", func() {
	// 	NewPersistentSession(nil, metaStore, evStore, log)
	// })
}

func TestNewPersistentSession_NilMetadataStore(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor")

	sess := newDefaultMockSession()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = evStore
	_ = log

	// assert.PanicsWithValue(t, "NewPersistentSession: metadataStore must not be nil", func() {
	// 	NewPersistentSession(sess, nil, evStore, log)
	// })
}

func TestNewPersistentSession_NilEventStore(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor")

	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = log

	// assert.PanicsWithValue(t, "NewPersistentSession: eventStore must not be nil", func() {
	// 	NewPersistentSession(sess, metaStore, nil, log)
	// })
}

func TestNewPersistentSession_NilLogger(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor")

	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()

	_ = sess
	_ = metaStore
	_ = evStore

	// assert.PanicsWithValue(t, "NewPersistentSession: logger must not be nil", func() {
	// 	NewPersistentSession(sess, metaStore, evStore, nil)
	// })
}

// =============================================================================
// Happy Path — Run
// =============================================================================

func TestPersistentSession_Run_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Run method")

	sess := newDefaultMockSession()
	sess.runErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Run()
	//
	// require.NoError(t, err)
	// assert.Equal(t, 1, sess.runCalled)
	// assert.Equal(t, 1, metaStore.writeCalled)
	// assert.Equal(t, sess.getMetadataSnapshotResult, metaStore.writeInput)
}

// =============================================================================
// Happy Path — Done
// =============================================================================

func TestPersistentSession_Done_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Done method")

	sess := newDefaultMockSession()
	sess.doneErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	notifier := newTerminationChannel()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = notifier

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Done(notifier)
	//
	// require.NoError(t, err)
	// assert.Equal(t, 1, sess.doneCalled)
	// assert.Equal(t, 1, metaStore.writeCalled)
	// assert.Equal(t, sess.getMetadataSnapshotResult, metaStore.writeInput)
}

// =============================================================================
// Happy Path — Fail
// =============================================================================

func TestPersistentSession_Fail_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Fail method")

	sess := newDefaultMockSession()
	sess.failErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	notifier := newTerminationChannel()
	someErr := errors.New("some error")

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = notifier
	_ = someErr

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Fail(someErr, notifier)
	//
	// require.NoError(t, err)
	// assert.Equal(t, 1, sess.failCalled)
	// assert.Equal(t, someErr, sess.failInputErr)
	// assert.Equal(t, 1, metaStore.writeCalled)
	// assert.Equal(t, sess.getMetadataSnapshotResult, metaStore.writeInput)
}

// =============================================================================
// Happy Path — UpdateCurrentStateSafe
// =============================================================================

func TestPersistentSession_UpdateCurrentStateSafe_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateCurrentStateSafe method")

	sess := newDefaultMockSession()
	sess.updateCurrentStateErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateCurrentStateSafe("node_2")
	//
	// require.NoError(t, err)
	// assert.Equal(t, 1, sess.updateCurrentStateCalled)
	// assert.Equal(t, "node_2", sess.updateCurrentStateInput)
	// assert.Equal(t, 1, metaStore.writeCalled)
	// assert.Equal(t, sess.getMetadataSnapshotResult, metaStore.writeInput)
}

// =============================================================================
// Happy Path — UpdateSessionDataSafe
// =============================================================================

func TestPersistentSession_UpdateSessionDataSafe_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateSessionDataSafe method")

	sess := newDefaultMockSession()
	sess.updateSessionDataErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateSessionDataSafe("key1", "val1")
	//
	// require.NoError(t, err)
	// assert.Equal(t, 1, sess.updateSessionDataCalled)
	// assert.Equal(t, "key1", sess.updateSessionDataInputKey)
	// assert.Equal(t, "val1", sess.updateSessionDataInputVal)
	// assert.Equal(t, 1, metaStore.writeCalled)
	// assert.Equal(t, sess.getMetadataSnapshotResult, metaStore.writeInput)
}

// =============================================================================
// Happy Path — UpdateEventHistorySafe
// =============================================================================

func TestPersistentSession_UpdateEventHistorySafe_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateEventHistorySafe method")

	sess := newDefaultMockSession()
	sess.updateEventHistoryErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	event := newTestEvent(t, testEventID)

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = event

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateEventHistorySafe(*event)
	//
	// require.NoError(t, err)
	// assert.Equal(t, 1, sess.updateEventHistoryCalled)
	// assert.Equal(t, 1, evStore.appendCalled)
	// assert.Equal(t, 1, metaStore.writeCalled)
	// assert.Equal(t, sess.getMetadataSnapshotResult, metaStore.writeInput)
}

// =============================================================================
// Happy Path — GetStatusSafe
// =============================================================================

func TestPersistentSession_GetStatusSafe(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.GetStatusSafe method")

	sess := newDefaultMockSession()
	sess.getStatusResult = "running"
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// result := ps.GetStatusSafe()
	//
	// assert.Equal(t, "running", result)
	// assert.Equal(t, 0, metaStore.writeCalled)
	// assert.Equal(t, 0, evStore.appendCalled)
}

// =============================================================================
// Happy Path — GetCurrentStateSafe
// =============================================================================

func TestPersistentSession_GetCurrentStateSafe(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.GetCurrentStateSafe method")

	sess := newDefaultMockSession()
	sess.getCurrentStateResult = "node_1"
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// result := ps.GetCurrentStateSafe()
	//
	// assert.Equal(t, "node_1", result)
	// assert.Equal(t, 0, metaStore.writeCalled)
	// assert.Equal(t, 0, evStore.appendCalled)
}

// =============================================================================
// Happy Path — GetErrorSafe
// =============================================================================

func TestPersistentSession_GetErrorSafe(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.GetErrorSafe method")

	someErr := errors.New("some error")
	sess := newDefaultMockSession()
	sess.getErrorResult = someErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// result := ps.GetErrorSafe()
	//
	// assert.Equal(t, someErr, result)
	// assert.Equal(t, 0, metaStore.writeCalled)
	// assert.Equal(t, 0, evStore.appendCalled)
}

// =============================================================================
// Happy Path — GetMetadataSnapshotSafe
// =============================================================================

func TestPersistentSession_GetMetadataSnapshotSafe(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.GetMetadataSnapshotSafe method")

	expectedMeta := session.SessionMetadata{
		ID:           testSessionID,
		WorkflowName: testWorkflowName,
		Status:       "running",
		CreatedAt:    testCreatedAt,
		UpdatedAt:    testCreatedAt + 1,
		CurrentState: testEntryNode,
		SessionData:  map[string]any{},
	}
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = expectedMeta
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// result := ps.GetMetadataSnapshotSafe()
	//
	// assert.Equal(t, expectedMeta, result)
	// assert.Equal(t, 0, metaStore.writeCalled)
	// assert.Equal(t, 0, evStore.appendCalled)
}

// =============================================================================
// Happy Path — GetSessionDataSafe
// =============================================================================

func TestPersistentSession_GetSessionDataSafe(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.GetSessionDataSafe method")

	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = "val1"
	sess.getSessionDataResultOK = true
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// val, ok := ps.GetSessionDataSafe("key1")
	//
	// assert.Equal(t, "val1", val)
	// assert.True(t, ok)
	// assert.Equal(t, 0, metaStore.writeCalled)
	// assert.Equal(t, 0, evStore.appendCalled)
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestPersistentSession_Run_SessionError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Run method")

	sessionErr := errors.New("precondition failure")
	sess := newDefaultMockSession()
	sess.runErr = sessionErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Run()
	//
	// require.Error(t, err)
	// assert.Equal(t, sessionErr, err)
	// assert.Equal(t, 0, metaStore.writeCalled)
}

func TestPersistentSession_Done_SessionError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Done method")

	sessionErr := errors.New("cannot complete session")
	sess := newDefaultMockSession()
	sess.doneErr = sessionErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	notifier := newTerminationChannel()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = notifier

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Done(notifier)
	//
	// require.Error(t, err)
	// assert.Equal(t, sessionErr, err)
	// assert.Equal(t, 0, metaStore.writeCalled)
}

func TestPersistentSession_Fail_SessionError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Fail method")

	sessionErr := errors.New("session already failed")
	sess := newDefaultMockSession()
	sess.failErr = sessionErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	notifier := newTerminationChannel()
	someErr := errors.New("some error")

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = notifier
	_ = someErr

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Fail(someErr, notifier)
	//
	// require.Error(t, err)
	// assert.Equal(t, sessionErr, err)
	// assert.Equal(t, 0, metaStore.writeCalled)
}

func TestPersistentSession_UpdateCurrentStateSafe_SessionError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateCurrentStateSafe method")

	sessionErr := errors.New("current state cannot be empty")
	sess := newDefaultMockSession()
	sess.updateCurrentStateErr = sessionErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateCurrentStateSafe("x")
	//
	// require.Error(t, err)
	// assert.Equal(t, sessionErr, err)
	// assert.Equal(t, 0, metaStore.writeCalled)
}

func TestPersistentSession_UpdateSessionDataSafe_SessionError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateSessionDataSafe method")

	sessionErr := errors.New("session data key cannot be empty")
	sess := newDefaultMockSession()
	sess.updateSessionDataErr = sessionErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateSessionDataSafe("k", "v")
	//
	// require.Error(t, err)
	// assert.Equal(t, sessionErr, err)
	// assert.Equal(t, 0, metaStore.writeCalled)
}

func TestPersistentSession_UpdateEventHistorySafe_SessionError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateEventHistorySafe method")

	sessionErr := errors.New("invalid event: ID is required")
	sess := newDefaultMockSession()
	sess.updateEventHistoryErr = sessionErr
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	event := newTestEvent(t, testEventID)

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = event

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateEventHistorySafe(*event)
	//
	// require.Error(t, err)
	// assert.Equal(t, sessionErr, err)
	// assert.Equal(t, 0, evStore.appendCalled)
	// assert.Equal(t, 0, metaStore.writeCalled)
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestPersistentSession_Run_MetadataWriteFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Run method")

	writeErr := errors.New("disk full")
	sess := newDefaultMockSession()
	sess.id = "sess-123"
	sess.runErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Run()
	//
	// require.NoError(t, err)
	// require.Len(t, log.errorCalls, 1)
	// assert.Equal(t, "failed to persist session metadata after Run", log.errorCalls[0].msg)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after Run", "error", writeErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after Run", "sessionID", "sess-123")
}

func TestPersistentSession_Done_MetadataWriteFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Done method")

	writeErr := errors.New("disk full")
	sess := newDefaultMockSession()
	sess.id = "sess-456"
	sess.doneErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	notifier := newTerminationChannel()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = notifier

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Done(notifier)
	//
	// require.NoError(t, err)
	// require.Len(t, log.errorCalls, 1)
	// assert.Equal(t, "failed to persist session metadata after Done", log.errorCalls[0].msg)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after Done", "error", writeErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after Done", "sessionID", "sess-456")
}

func TestPersistentSession_Fail_MetadataWriteFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.Fail method")

	writeErr := errors.New("disk full")
	sess := newDefaultMockSession()
	sess.id = "sess-789"
	sess.failErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	notifier := newTerminationChannel()
	someErr := errors.New("some error")

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = notifier
	_ = someErr

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.Fail(someErr, notifier)
	//
	// require.NoError(t, err)
	// require.Len(t, log.errorCalls, 1)
	// assert.Equal(t, "failed to persist session metadata after Fail", log.errorCalls[0].msg)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after Fail", "error", writeErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after Fail", "sessionID", "sess-789")
}

func TestPersistentSession_UpdateCurrentStateSafe_MetadataWriteFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateCurrentStateSafe method")

	writeErr := errors.New("disk full")
	sess := newDefaultMockSession()
	sess.id = "sess-100"
	sess.updateCurrentStateErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateCurrentStateSafe("node_x")
	//
	// require.NoError(t, err)
	// require.Len(t, log.errorCalls, 1)
	// assert.Equal(t, "failed to persist session metadata after UpdateCurrentStateSafe", log.errorCalls[0].msg)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateCurrentStateSafe", "error", writeErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateCurrentStateSafe", "sessionID", "sess-100")
}

func TestPersistentSession_UpdateSessionDataSafe_MetadataWriteFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateSessionDataSafe method")

	writeErr := errors.New("disk full")
	sess := newDefaultMockSession()
	sess.id = "sess-200"
	sess.updateSessionDataErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateSessionDataSafe("myKey", "myVal")
	//
	// require.NoError(t, err)
	// require.Len(t, log.errorCalls, 1)
	// assert.Equal(t, "failed to persist session metadata after UpdateSessionDataSafe", log.errorCalls[0].msg)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateSessionDataSafe", "error", writeErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateSessionDataSafe", "sessionID", "sess-200")
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateSessionDataSafe", "key", "myKey")
}

func TestPersistentSession_UpdateEventHistorySafe_AppendFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateEventHistorySafe method")

	appendErr := errors.New("append failed")
	sess := newDefaultMockSession()
	sess.id = "sess-300"
	sess.updateEventHistoryErr = nil
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	evStore.appendErr = appendErr
	log := newDefaultMockLogger()
	event := newTestEvent(t, testEventID)

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = event

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateEventHistorySafe(*event)
	//
	// require.NoError(t, err)
	// assertLogHasMessage(t, log.errorCalls, "failed to persist event")
	// assertLogContainsArg(t, log.errorCalls, "failed to persist event", "error", appendErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist event", "sessionID", "sess-300")
	// assertLogContainsArg(t, log.errorCalls, "failed to persist event", "eventID", testEventID)
	// assert.Equal(t, 1, metaStore.writeCalled) // metadata persist still attempted
}

func TestPersistentSession_UpdateEventHistorySafe_MetadataWriteFails_LogsError(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateEventHistorySafe method")

	writeErr := errors.New("disk full")
	sess := newDefaultMockSession()
	sess.id = "sess-400"
	sess.updateEventHistoryErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	evStore.appendErr = nil
	log := newDefaultMockLogger()
	event := newTestEvent(t, testEventID)

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = event

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateEventHistorySafe(*event)
	//
	// require.NoError(t, err)
	// assertLogHasMessage(t, log.errorCalls, "failed to persist session metadata after UpdateEventHistorySafe")
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateEventHistorySafe", "error", writeErr)
	// assertLogContainsArg(t, log.errorCalls, "failed to persist session metadata after UpdateEventHistorySafe", "sessionID", "sess-400")
}

func TestPersistentSession_UpdateEventHistorySafe_BothFail_LogsBothErrors(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.UpdateEventHistorySafe method")

	appendErr := errors.New("append failed")
	writeErr := errors.New("write failed")
	sess := newDefaultMockSession()
	sess.id = "sess-500"
	sess.updateEventHistoryErr = nil
	metaStore := newDefaultMockMetadataStore()
	metaStore.writeErr = writeErr
	evStore := newDefaultMockEventStore()
	evStore.appendErr = appendErr
	log := newDefaultMockLogger()
	event := newTestEvent(t, testEventID2)

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log
	_ = event

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// err := ps.UpdateEventHistorySafe(*event)
	//
	// require.NoError(t, err)
	// require.Len(t, log.errorCalls, 2)
	// assertLogHasMessage(t, log.errorCalls, "failed to persist event")
	// assertLogHasMessage(t, log.errorCalls, "failed to persist session metadata after UpdateEventHistorySafe")
}

// =============================================================================
// Happy Path — ID
// =============================================================================

func TestPersistentSession_ID(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.ID field access")

	sess := newDefaultMockSession()
	sess.id = "sess-abc"
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// assert.Equal(t, "sess-abc", ps.ID)
}

// =============================================================================
// Happy Path — WorkflowName
// =============================================================================

func TestPersistentSession_WorkflowName(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime.NewPersistentSession constructor and PersistentSession.WorkflowName field access")

	sess := newDefaultMockSession()
	sess.workflowName = "my-workflow"
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()

	_ = sess
	_ = metaStore
	_ = evStore
	_ = log

	// ps := NewPersistentSession(sess, metaStore, evStore, log)
	// assert.Equal(t, "my-workflow", ps.WorkflowName)
}

// Ensure unused imports are satisfied at compile time.
var (
	_ = assert.Equal
	_ = require.NoError
	_ = session.SessionMetadata{}
)
