package runtime

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities/session"
)

// =============================================================================
// Test Specification: claude_process_cleaner_test.go
// Source File Under Test: runtime/claude_process_cleaner.go
// =============================================================================

// --- Test Helpers: ClaudeProcessCleaner ---

// mockProcessInspector simulates process inspection for testing.
// Implements the ProcessInspector interface.
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

func (m *mockProcessInspector) IsRunning(pid int) bool {
	return m.isRunning[pid]
}

func (m *mockProcessInspector) Command(pid int) string {
	return m.commands[pid]
}

// mockSignalSender simulates signal delivery for testing.
// Implements the SignalSender interface.
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

func (m *mockSignalSender) SendSIGTERM(pid int) error {
	if err, ok := m.sigtermErr[pid]; ok {
		return err
	}
	m.sigtermSent = append(m.sigtermSent, pid)
	return nil
}

func (m *mockSignalSender) SendSIGKILL(pid int) error {
	if err, ok := m.sigkillErr[pid]; ok {
		return err
	}
	m.sigkillSent = append(m.sigkillSent, pid)
	return nil
}

// mockProcessWaiter simulates process exit waiting.
// Implements the ProcessWaiter interface.
type mockProcessWaiter struct {
	// exitsWithinTimeout maps PID -> whether it exits before the 2-second timeout
	exitsWithinTimeout map[int]bool
}

func newMockProcessWaiter() *mockProcessWaiter {
	return &mockProcessWaiter{
		exitsWithinTimeout: make(map[int]bool),
	}
}

func (m *mockProcessWaiter) WaitForExit(pid int) bool {
	return m.exitsWithinTimeout[pid]
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
	// Setup
	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log)

	// Assert: Returns non-nil *ClaudeProcessCleaner; no panic
	require.NotNil(t, cleaner)
}

// =============================================================================
// Happy Path — Clean
// =============================================================================

func TestClean_TerminatesRunningClaudeProcess(t *testing.T) {
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
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: SIGTERM sent to PID 1234
	assert.Contains(t, sender.sigtermSent, 1234)
	assertLoggerHasInfoMsgContaining(t, log, "sent SIGTERM to 1 claude process(es)")
	assertLoggerHasInfoMsgContaining(t, log, "claude process cleanup complete")
}

func TestClean_MultipleClaudeProcesses(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: SIGTERM sent to both PIDs
	assert.Contains(t, sender.sigtermSent, 1111)
	assert.Contains(t, sender.sigtermSent, 2222)
	assertLoggerHasInfoMsgContaining(t, log, "sent SIGTERM to 2 claude process(es)")
	assertLoggerHasInfoMsgContaining(t, log, "claude process cleanup complete")
}

func TestClean_NoPIDKeys(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log)
	cleaner.Clean()

	// Assert: No signals sent; no error; returns immediately
	assert.Empty(t, log.infoCalls)
	assert.Empty(t, log.warnCalls)
}

func TestClean_AllPIDsAlreadyExited(t *testing.T) {
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
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector))
	cleaner.Clean()

	// Assert: No signals sent; Logger.Info called with "no active claude processes to terminate"
	assertLoggerHasInfoMsgContaining(t, log, "no active claude processes to terminate")
}

func TestClean_DeduplicatesSamePID(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: SIGTERM sent to PID 5555 exactly once
	count := 0
	for _, pid := range sender.sigtermSent {
		if pid == 5555 {
			count++
		}
	}
	assert.Equal(t, 1, count, "SIGTERM should be sent to PID 5555 exactly once")
	assertLoggerHasInfoMsgContaining(t, log, "sent SIGTERM to 1 claude process(es)")
}

// =============================================================================
// State Transitions — Clean
// =============================================================================

func TestClean_EscalatesToSIGKILL(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: SIGTERM sent first; after 2-second simulated wait, SIGKILL sent to PID 1234
	assert.Contains(t, sender.sigtermSent, 1234)
	assert.Contains(t, sender.sigkillSent, 1234)
	assertLoggerHasWarnMsgContaining(t, log, "escalating to SIGKILL for PID 1234")
}

func TestClean_ProcessExitsBeforeSIGKILL(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: SIGTERM sent; SIGKILL NOT sent; Logger.Info with cleanup complete
	assert.Contains(t, sender.sigtermSent, 1234)
	assert.Empty(t, sender.sigkillSent)
	assertLoggerHasInfoMsgContaining(t, log, "claude process cleanup complete")
}

// =============================================================================
// Error Propagation — Clean
// =============================================================================

func TestClean_SIGTERMFailsPermissionDenied(t *testing.T) {
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
	sender.sigtermErr[1234] = errors.New("operation not permitted")
	waiter := newMockProcessWaiter()

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: Logger.Warn about SIGTERM failure; SIGKILL not attempted; Clean() does not panic
	assertLoggerHasWarnMsgContaining(t, log, "failed to send SIGTERM to PID 1234")
	assert.Empty(t, sender.sigkillSent)
}

func TestClean_SIGKILLFailsPermissionDenied(t *testing.T) {
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
	sender.sigkillErr[1234] = errors.New("operation not permitted")

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: Logger.Warn with "failed to kill claude process"; Clean() does not panic
	assertLoggerHasWarnMsgContaining(t, log, "failed to kill claude process")
}

func TestClean_NonIntegerPIDValue(t *testing.T) {
	// Setup
	sess := newDefaultMockSession()
	sess.getMetadataSnapshotResult = buildSessionDataWithPIDs(map[string]any{
		"NodeA.PID": "not-an-int",
	})
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log)
	cleaner.Clean()

	// Assert: Logger.Warn with "skipping non-integer PID value"; no signals sent
	assertLoggerHasWarnMsgContaining(t, log, "skipping non-integer PID value")
}

// =============================================================================
// Mock / Dependency Interaction — Clean
// =============================================================================

func TestClean_SkipsNonClaudeProcess(t *testing.T) {
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
	sender := newMockSignalSender()

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender))
	cleaner.Clean()

	// Assert: Logger.Warn "PID does not belong to a claude process, skipping";
	// no signals sent; Logger.Info "no active claude processes to terminate"
	assertLoggerHasWarnMsgContaining(t, log, "PID does not belong to a claude process, skipping")
	assertLoggerHasInfoMsgContaining(t, log, "no active claude processes to terminate")
	assert.Empty(t, sender.sigtermSent)
}

func TestClean_CommandMatchCaseInsensitive(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector), WithSignalSender(sender), WithProcessWaiter(waiter))
	cleaner.Clean()

	// Assert: SIGTERM sent to PID 1234 (case-insensitive match)
	assert.Contains(t, sender.sigtermSent, 1234)
}

func TestClean_ReadOnlySessionAccess(t *testing.T) {
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
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector))
	cleaner.Clean()

	// Assert: PersistentSession.GetMetadataSnapshotSafe called once;
	// no mutation methods called (UpdateSessionDataSafe, Fail, Run, Done never called)
	require.Equal(t, 0, sess.updateSessionDataCalled)
	require.Equal(t, 0, sess.failCalled)
	require.Equal(t, 0, sess.runCalled)
	require.Equal(t, 0, sess.doneCalled)
}

func TestClean_OnlyExtractsPIDSuffixKeys(t *testing.T) {
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

	// Act
	cleaner := NewClaudeProcessCleaner(ps, log, WithProcessInspector(inspector))
	cleaner.Clean()

	// Assert: Only PID 1234 is inspected; value 42 under "logicSpec.count" is ignored
	// The fact that inspector.setNotRunning(1234) is the only PID setup and we get
	// "no active claude processes" confirms only .PID keys were extracted.
	assertLoggerHasInfoMsgContaining(t, log, "no active claude processes to terminate")
}
