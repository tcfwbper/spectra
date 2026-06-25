# ClaudeProcessCleaner

## Overview

ClaudeProcessCleaner is responsible for finding and terminating orphaned Claude CLI processes that were spawned by the current session. It reads PID values from the session's `SessionData` (keys matching `<NodeName>.PID`), verifies each PID is still running and belongs to a `claude` process, sends SIGTERM, waits up to 2 seconds, then escalates to SIGKILL for any survivors. ClaudeProcessCleaner does not modify session state, does not perform persistence, and does not construct RuntimeError — it is a best-effort cleanup utility invoked by Runtime during shutdown.

## Boundaries

- Owns: discovery of stored PIDs from session data snapshot.
- Owns: verification that a PID is still running and its command matches "claude".
- Owns: SIGTERM delivery to all verified claude processes.
- Owns: 2-second wait for process exit after SIGTERM.
- Owns: SIGKILL escalation for processes that survive SIGTERM.
- Owns: logging of cleanup actions and failures.
- Delegates: session data access to PersistentSession (read-only via `GetMetadataSnapshotSafe`).
- Delegates: process signal delivery to the OS.
- Must not: modify session state (no calls to any mutation method on PersistentSession).
- Must not: construct RuntimeError or call `PersistentSession.Fail()`.
- Must not: perform persistence (no store writes).
- Must not: block longer than the 2-second SIGTERM wait plus SIGKILL delivery time.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State source (read-only) | `GetMetadataSnapshotSafe()` | Must not call any mutation method, must not call stores |
| `Logger` | Structured logging | `Info(msg, args...)`, `Warn(msg, args...)` | Must not use for session status output |
| OS/process | Process inspection and signaling | Find process by PID, read process command name, send SIGTERM, send SIGKILL, wait for process exit | Must not start new processes |

Construction constraint: ClaudeProcessCleaner is constructed with `PersistentSession` (read-only reference) and `Logger`. It is a lightweight struct constructed by Runtime after session initialization. Must be constructed via a constructor function `NewClaudeProcessCleaner(persistentSession, logger)`.

## Behavior

### Construction

`NewClaudeProcessCleaner(persistentSession *PersistentSession, logger logger.Logger) *ClaudeProcessCleaner`

1. Stores references to `persistentSession` and `logger`.
2. Returns the constructed instance.

### Clean

`Clean()`

1. Calls `PersistentSession.GetMetadataSnapshotSafe()` to obtain a snapshot of session metadata.
2. Iterates over all keys in `SessionData` from the snapshot.
3. For each key matching the suffix `.PID` (case-sensitive), reads the value.
4. If the value is not an integer type, logs a warning: `"skipping non-integer PID value"` with key and actual type, and continues to the next key.
5. Collects all valid PID integers into a list of candidate PIDs. Deduplicates by PID value (if the same PID appears under multiple keys, it is included only once).
6. If the candidate list is empty, returns immediately (nothing to clean).
7. For each candidate PID, verifies the process:
   a. Checks if the process with that PID is still running.
   b. If not running (process already exited), skips it.
   c. If running, reads the process command name / command line.
   d. Checks whether the command contains "claude" (case-insensitive substring match on the executable name or full command line).
   e. If the command does not match "claude", logs a warning: `"PID does not belong to a claude process, skipping"` with PID and observed command, and skips it.
8. Collects all verified PIDs (running and confirmed as claude) into a kill list.
9. If the kill list is empty, logs: `"no active claude processes to terminate"` at Info level and returns.
10. Sends SIGTERM to all processes in the kill list.
11. For each process where SIGTERM delivery fails (e.g., permission denied, process exited between check and signal), logs a warning and removes it from the wait list.
12. Logs: `"sent SIGTERM to N claude process(es)"` at Info level, with the list of PIDs.
13. Waits up to 2 seconds for all signaled processes to exit. Checks periodically (e.g., every 100ms) whether each process has exited.
14. After 2 seconds, for any process still running, sends SIGKILL.
15. For each SIGKILL delivery, logs a warning: `"escalating to SIGKILL for PID <pid>"`.
16. If SIGKILL fails (permission denied), logs a warning: `"failed to kill claude process"` with PID and error.
17. Logs summary: `"claude process cleanup complete"` with count of terminated processes.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| persistentSession | *PersistentSession | Valid, non-nil | Yes |
| logger | logger.Logger | Non-nil Logger interface implementation | Yes |

### For Clean

No parameters. All state is read from the injected PersistentSession.

## Outputs

### Clean

| Field | Type | Description |
|-------|------|-------------|
| (none) | — | Clean does not return a value or error. All failures are logged and absorbed. |

## Invariants

1. **Read-Only Session Access**: ClaudeProcessCleaner must only read session data via `GetMetadataSnapshotSafe()`. It must never mutate session state.

2. **Best-Effort**: All process inspection and signaling failures are logged and absorbed. `Clean()` never returns an error or panics on signal delivery failure.

3. **PID Verification Before Kill**: A process is only signaled if (a) it is still running AND (b) its command matches "claude". Both conditions must be verified before sending any signal.

4. **SIGTERM Before SIGKILL**: SIGKILL is only sent after SIGTERM has been sent and the 2-second wait has elapsed. SIGKILL is never sent without a prior SIGTERM attempt.

5. **Bounded Duration**: The total execution time of `Clean()` is bounded: PID iteration + SIGTERM delivery + 2-second wait + SIGKILL delivery. No unbounded blocking.

6. **No Partial State**: `Clean()` operates on a snapshot of session data taken at invocation time. Changes to session data during cleanup do not affect the operation.

7. **Platform-Specific Process Inspection**: The method for reading a process's command name is platform-dependent (e.g., `/proc/<pid>/cmdline` on Linux, `ps` on macOS). The implementation must support the host OS.

8. **Case-Insensitive Command Match**: The "claude" substring check on the process command is case-insensitive to handle variations in executable naming.

9. **PID Deduplication**: If the same PID value appears under multiple `.PID` keys, it is collected and signaled only once.

## Edge Cases

- Condition: SessionData contains no `.PID` keys.
  Expected: Returns immediately after snapshot iteration. No processes signaled. No log output beyond the implicit early return.

- Condition: All stored PIDs have already exited (stale PIDs).
  Expected: Each PID fails the "is running" check in step 7a. Kill list is empty. Logs "no active claude processes to terminate" and returns.

- Condition: A stored PID is still running but belongs to a non-claude process (PID reuse).
  Expected: Fails the command match check in step 7d. Logs warning with PID and observed command. Process is not signaled.

- Condition: A PID value in SessionData is not an integer (data corruption).
  Expected: Logs warning with key and type. Skips to next key.

- Condition: SIGTERM is sent but process exits before SIGKILL check.
  Expected: Process is no longer in the wait list at the SIGKILL step. No SIGKILL sent.

- Condition: SIGTERM delivery fails (permission denied).
  Expected: Logs warning. Removes PID from wait list. Does not attempt SIGKILL for that PID.

- Condition: Process survives SIGTERM for 2 seconds.
  Expected: SIGKILL sent after timeout. Log warning about escalation.

- Condition: SIGKILL fails (permission denied — should not happen for own child processes but handled defensively).
  Expected: Logs warning with PID and error. Continues with remaining processes.

- Condition: Multiple nodes have running claude processes.
  Expected: All are discovered from SessionData, all verified, all receive SIGTERM concurrently, all subject to same 2-second wait and SIGKILL escalation.

- Condition: Same PID appears under multiple keys (should not happen in practice).
  Expected: Deduplicated during collection. SIGTERM sent once.

- Condition: `GetMetadataSnapshotSafe()` returns SessionData with many keys (hundreds).
  Expected: Only keys matching `.PID` suffix are extracted. Iteration is fast and bounded.

## Related

- [Runtime](./runtime.md) — Invokes ClaudeProcessCleaner during cleanup sequence
- [AgentInvoker](./agent_invoker.md) — Writes `<NodeName>.PID` to session data after process startup
- [PersistentSession](./persistent_session.md) — Provides read-only access to session data
- [Session Data](../entities/session/data.md) — Defines `.PID` key namespace and type validation
