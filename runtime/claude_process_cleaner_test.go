package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities/session"
)

// =============================================================================
// Test Specification: claude_process_cleaner_test.go
// Source File Under Test: runtime/claude_process_cleaner.go
//
// All tests are scaffolded: the production file claude_process_cleaner.go does
// not yet exist. The scaffolding below establishes the test surface, helper
// structure, and assertion intent for each spec row. Tests will become concrete
// once the following production symbols exist:
//   - NewClaudeProcessCleaner(persistentSession *PersistentSession, logger logger.Logger) *ClaudeProcessCleaner
//   - (*ClaudeProcessCleaner).Clean()
//   - Process inspector seam (interface for checking if PID is running / reading command)
//   - Signal sender seam (interface for sending SIGTERM/SIGKILL)
//   - Process waiter seam (interface / fake clock for waiting process exit)
// =============================================================================

// --- Test Helpers: ClaudeProcessCleaner ---

// mockProcessInspector simulates process inspection for testing.
// Production code will need a seam to inject this.
type mockProcessInspector struct {
	// isRunning maps PID -> whether process is running
	isRunning map[int]bool
	// commands maps PID -> command string
	commands map[int]string
}

func newMockProcessInspector() *mockProcessInspector {
	return &mockProcessInspector{
		isRunning: make(map[int]bool),
		commands:  make(map[int]string),
	}
}

func (m *mockProcessInspector) setRunning(pid int, command string) {
	m.isRunning[pid] = true
	m.commands[pid] = command
}

func (m *mockProcessInspector) setNotRunning(pid int) {
	m.isRunning[pid] = false
}

// mockSignalSender simulates signal delivery for testing.
type mockSignalSender struct {
	// sigtermSent tracks PIDs that received SIGTERM
	sigtermSent []int
	// sigkillSent tracks PIDs that received SIGKILL
	sigkillSent []int
	// sigtermErr maps PID -> error for SIGTERM delivery
	sigtermErr map[int]error
	// sigkillErr maps PID -> error for SIGKILL delivery
	sigkillErr map[int]error
}

func newMockSignalSender() *mockSignalSender {
	return &mockSignalSender{
		sigtermErr: make(map[int]error),
		sigkillErr: make(map[int]error),
	}
}

// mockProcessWaiter simulates process exit waiting.
type mockProcessWaiter struct {
	// exitsWithinTimeout maps PID -> whether it exits before the 2-second timeout
	exitsWithinTimeout map[int]bool
}

func newMockProcessWaiter() *mockProcessWaiter {
	return &mockProcessWaiter{
		exitsWithinTimeout: make(map[int]bool),
	}
}

// buildSessionDataWithPIDs creates a SessionMetadata with specific SessionData entries.
func buildSessionDataWithPIDs(entries map[string]any) session.SessionMetadata {
	return session.SessionMetadata{
		ID:           testSessionID,
		WorkflowName: testWorkflowName,
		Status:       "running",
		CreatedAt:    testCreatedAt,
		UpdatedAt:    testCreatedAt + 1,
		CurrentState: testEntryNode,
		SessionData:  entries,
	}
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewClaudeProcessCleaner_ValidDeps(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner constructor")

	// Setup
	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Act
	_ = ps  // placeholder: cleaner := NewClaudeProcessCleaner(ps, log)
	_ = log // placeholder

	// Assert: Returns non-nil *ClaudeProcessCleaner; no panic
	// require.NotNil(t, cleaner)
}

// =============================================================================
// Happy Path — Clean
// =============================================================================

func TestClean_TerminatesRunningClaudeProcess(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam, process waiter seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[1234] = true

	// Act
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	// cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	// cleaner.Clean()

	// Assert: SIGTERM sent to PID 1234
	// assert.Contains(t, sender.sigtermSent, 1234)
	// assertLoggerHasInfoMsgContaining(t, log, "sent SIGTERM to 1 claude process(es)")
	// assertLoggerHasInfoMsgContaining(t, log, "claude process cleanup complete")
	assert.NotNil(t, ps) // placeholder assertion
}

func TestClean_MultipleClaudeProcesses(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam, process waiter seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1111,
		"NodeB.PID": 2222,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1111, "claude")
	inspector.setRunning(2222, "claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[1111] = true
	waiter.exitsWithinTimeout[2222] = true

	// Act & Assert (placeholder)
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	assert.NotNil(t, ps)
}

func TestClean_NoPIDKeys(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"logicSpec.output":      "data",
		"NodeA.ClaudeSessionID": "uuid",
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Act & Assert: No signals sent; no error; returns immediately
	_ = ps
	assert.NotNil(t, ps)
}

func TestClean_AllPIDsAlreadyExited(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setNotRunning(1234)

	// Act & Assert: No signals sent; Logger.Info called with "no active claude processes to terminate"
	_ = ps
	_ = inspector
	assert.NotNil(t, ps)
}

func TestClean_DeduplicatesSamePID(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 5555,
		"NodeB.PID": 5555,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(5555, "claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[5555] = true

	// Act & Assert: SIGTERM sent to PID 5555 exactly once
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	assert.NotNil(t, ps)
}

// =============================================================================
// State Transitions — Clean
// =============================================================================

func TestClean_EscalatesToSIGKILL(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam, fake timer/clock seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[1234] = false // does NOT exit within 2 seconds

	// Act & Assert: SIGTERM sent first; after 2-second simulated wait, SIGKILL sent to PID 1234
	// Logger.Warn called with "escalating to SIGKILL for PID 1234"
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	assert.NotNil(t, ps)
}

func TestClean_ProcessExitsBeforeSIGKILL(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam, fake timer/clock seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[1234] = true // exits before timeout

	// Act & Assert: SIGTERM sent; SIGKILL NOT sent; Logger.Info with cleanup complete
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	assert.NotNil(t, ps)
}

// =============================================================================
// Error Propagation — Clean
// =============================================================================

func TestClean_SIGTERMFailsPermissionDenied(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "claude")
	sender := newMockSignalSender()
	// SIGTERM fails with permission denied
	// sender.sigtermErr[1234] = errors.New("operation not permitted")

	// Act & Assert: Logger.Warn about SIGTERM failure; SIGKILL not attempted; Clean() does not panic
	_ = ps
	_ = inspector
	_ = sender
	assert.NotNil(t, ps)
}

func TestClean_SIGKILLFailsPermissionDenied(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam, fake timer/clock seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[1234] = false
	// SIGKILL fails
	// sender.sigkillErr[1234] = errors.New("operation not permitted")

	// Act & Assert: Logger.Warn with "failed to kill claude process"; Clean() does not panic
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	assert.NotNil(t, ps)
}

func TestClean_NonIntegerPIDValue(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": "not-an-int",
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Act & Assert: Logger.Warn with "skipping non-integer PID value"; no signals sent
	_ = ps
	assert.NotNil(t, ps)
}

// =============================================================================
// Mock / Dependency Interaction — Clean
// =============================================================================

func TestClean_SkipsNonClaudeProcess(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "python my_script.py")

	// Act & Assert: Logger.Warn "PID does not belong to a claude process, skipping";
	// no signals sent; Logger.Info "no active claude processes to terminate"
	_ = ps
	_ = inspector
	assert.NotNil(t, ps)
}

func TestClean_CommandMatchCaseInsensitive(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam, signal sender seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setRunning(1234, "/usr/local/bin/Claude")
	sender := newMockSignalSender()
	waiter := newMockProcessWaiter()
	waiter.exitsWithinTimeout[1234] = true

	// Act & Assert: SIGTERM sent to PID 1234 (case-insensitive match)
	_ = ps
	_ = inspector
	_ = sender
	_ = waiter
	assert.NotNil(t, ps)
}

func TestClean_ReadOnlySessionAccess(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": 1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setNotRunning(1234)

	// Act
	_ = ps
	_ = inspector

	// Assert: PersistentSession.GetMetadataSnapshotSafe called once;
	// no mutation methods called (UpdateSessionDataSafe, Fail, Run, Done never called)
	require.Equal(t, 0, sess.updateSessionDataCalled)
	require.Equal(t, 0, sess.failCalled)
	require.Equal(t, 0, sess.runCalled)
	require.Equal(t, 0, sess.doneCalled)
}

func TestClean_OnlyExtractsPIDSuffixKeys(t *testing.T) {
	t.Skip("scaffolded: production file runtime/claude_process_cleaner.go does not yet exist — missing NewClaudeProcessCleaner, Clean, process inspector seam")

	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.ClaudeSessionID": "uuid-val",
		"logicSpec.count":       42,
		"NodeA.PID":             1234,
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	inspector := newMockProcessInspector()
	inspector.setNotRunning(1234)

	// Act & Assert: Only PID 1234 is inspected; value 42 under "logicSpec.count" is ignored
	_ = ps
	_ = inspector
	assert.NotNil(t, ps)
}
