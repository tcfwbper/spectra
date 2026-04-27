# SessionFinalizer

## Overview

SessionFinalizer orchestrates the cleanup flow when a session reaches a terminal state (Status="completed" or Status="failed"). It is invoked by Runtime after receiving a termination notification from the main loop. SessionFinalizer performs the following sequence: stops the socket listener (if active), deletes the runtime socket file, reads the final session status, and prints the status to stdout (if completed) or stderr (if failed), including the error details if the session failed. SessionFinalizer does not delete the session directory or session files (metadata, events); these are retained on disk for inspection and debugging. SessionFinalizer is idempotent and safe to call even if the session is in a partially initialized state.

## Behavior

### Cleanup Flow

1. SessionFinalizer is invoked by Runtime with a single input: `session` (*Session).
2. SessionFinalizer validates that `Session.Status` is either `"completed"` or `"failed"`. If Status is `"initializing"` or `"running"`, SessionFinalizer logs a warning: `"SessionFinalizer called with non-terminal session status '<status>'. This may indicate a programming error."` but proceeds with cleanup anyway.
3. SessionFinalizer calls `RuntimeSocketManager.DeleteSocket()` to stop the socket listener (if active) and delete the runtime socket file.
4. `RuntimeSocketManager.DeleteSocket()` is idempotent. If the socket file does not exist (e.g., was never created, or already deleted), it returns without error.
5. If `RuntimeSocketManager.DeleteSocket()` logs a warning about socket deletion failure, SessionFinalizer does not propagate the error. The warning is logged by RuntimeSocketManager, and SessionFinalizer continues with status printing.
6. SessionFinalizer reads the final `Session.Status` and `Session.Error` fields (thread-safe read via Session's internal lock if needed, but typically Status and Error are immutable after reaching terminal state).
7. If `Session.Status == "completed"`:
   - SessionFinalizer prints to stdout: `"Session <SessionID> completed successfully. Workflow: <WorkflowName>"`
8. If `Session.Status == "failed"`:
   - SessionFinalizer prints to stderr: `"Session <SessionID> failed. Workflow: <WorkflowName>"`
   - SessionFinalizer prints to stderr: `"Error: <Session.Error.Message>"`
   - If `Session.Error` is an AgentError:
     - SessionFinalizer prints to stderr: `"Agent: <AgentError.AgentRole>"`
     - SessionFinalizer prints to stderr: `"State: <AgentError.FailingState>"`
     - If `AgentError.Detail` is not empty (not `{}`), SessionFinalizer prints to stderr: `"Detail: <AgentError.Detail as JSON>"`
   - If `Session.Error` is a RuntimeError:
     - SessionFinalizer prints to stderr: `"Issuer: <RuntimeError.Issuer>"`
     - SessionFinalizer prints to stderr: `"State: <RuntimeError.FailingState>"`
     - If `RuntimeError.Detail` is not empty (not `{}`), SessionFinalizer prints to stderr: `"Detail: <RuntimeError.Detail as JSON>"`
9. SessionFinalizer does not return an error. All cleanup operations are best-effort, and failures are logged but do not block termination.
10. After printing, SessionFinalizer returns control to Runtime, which then exits.

## Inputs

### For Finalize Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Session | *Session | Reference to the Session entity, must have Status="completed" or Status="failed" | Yes |

## Outputs

### Success Case

No return value (void). SessionFinalizer always succeeds.

### Console Output

**Case 1: Session Completed**

Printed to stdout:
```
Session <SessionID> completed successfully. Workflow: <WorkflowName>
```

**Case 2: Session Failed with AgentError**

Printed to stderr:
```
Session <SessionID> failed. Workflow: <WorkflowName>
Error: <Message>
Agent: <AgentRole>
State: <FailingState>
Detail: <Detail as JSON>
```

(Detail line is omitted if Detail is `{}`)

**Case 3: Session Failed with RuntimeError**

Printed to stderr:
```
Session <SessionID> failed. Workflow: <WorkflowName>
Error: <Message>
Issuer: <Issuer>
State: <FailingState>
Detail: <Detail as JSON>
```

(Detail line is omitted if Detail is `{}`)

## Invariants

1. **Idempotent Cleanup**: SessionFinalizer must be safe to call multiple times for the same session. Subsequent calls have no adverse effects (socket already deleted, status already printed).

2. **Best-Effort Socket Cleanup**: Socket deletion failures must not prevent status printing. RuntimeSocketManager logs warnings, and SessionFinalizer proceeds.

3. **Terminal Status Validation**: SessionFinalizer should validate that Session.Status is "completed" or "failed". If Status is non-terminal, a warning is logged but cleanup proceeds anyway.

4. **No Session Deletion**: SessionFinalizer must not delete the session directory, metadata files, or event files. These are retained for inspection.

5. **Stdout for Success, Stderr for Failure**: Completed sessions print to stdout. Failed sessions print to stderr. This allows users to distinguish success and failure in shell scripts using exit codes and redirection.

6. **Error Field Non-Nil for Failed Sessions**: If Session.Status="failed", Session.Error must be non-nil (either AgentError or RuntimeError). SessionFinalizer assumes this invariant holds (enforced by Session.Fail()).

7. **No Panic on Nil Error**: If Session.Status="failed" but Session.Error is nil (violates Session invariant), SessionFinalizer must handle this gracefully by printing `"Error: <unknown error>"` instead of panicking.

8. **Detail JSON Serialization**: If Error.Detail is not empty, SessionFinalizer serializes it as JSON using compact format (no pretty-printing).

9. **No Return Error**: SessionFinalizer does not return errors. All operations are best-effort. Runtime proceeds to exit after SessionFinalizer completes.

## Edge Cases

- **Condition**: Session.Status is "initializing" or "running" when SessionFinalizer is called.
  **Expected**: SessionFinalizer logs a warning: `"SessionFinalizer called with non-terminal session status '<status>'. This may indicate a programming error."` and proceeds with socket cleanup and status printing. The status output will show the non-terminal status, which may confuse the user.

- **Condition**: RuntimeSocketManager.DeleteSocket() logs a warning about socket deletion failure (e.g., permission denied).
  **Expected**: The warning is logged by RuntimeSocketManager. SessionFinalizer continues with status printing and does not propagate the error.

- **Condition**: RuntimeSocketManager.DeleteSocket() is called but the socket file was never created (initialization failed before socket creation).
  **Expected**: RuntimeSocketManager.DeleteSocket() is idempotent and returns without error. SessionFinalizer proceeds with status printing.

- **Condition**: Session.Status="completed".
  **Expected**: SessionFinalizer prints to stdout: `"Session <SessionID> completed successfully. Workflow: <WorkflowName>"`. No error details are printed.

- **Condition**: Session.Status="failed" and Session.Error is an AgentError with empty Detail (`{}`).
  **Expected**: SessionFinalizer prints to stderr:
  ```
  Session <SessionID> failed. Workflow: <WorkflowName>
  Error: <Message>
  Agent: <AgentRole>
  State: <FailingState>
  ```
  The "Detail:" line is omitted.

- **Condition**: Session.Status="failed" and Session.Error is a RuntimeError with non-empty Detail.
  **Expected**: SessionFinalizer prints to stderr:
  ```
  Session <SessionID> failed. Workflow: <WorkflowName>
  Error: <Message>
  Issuer: <Issuer>
  State: <FailingState>
  Detail: {"key":"value"}
  ```
  The Detail is serialized as compact JSON.

- **Condition**: Session.Status="failed" but Session.Error is nil (violates Session invariant).
  **Expected**: SessionFinalizer detects the nil Error and prints to stderr:
  ```
  Session <SessionID> failed. Workflow: <WorkflowName>
  Error: <unknown error>
  ```
  No Agent/Issuer/State/Detail lines are printed.

- **Condition**: Session.Error.Detail contains complex nested JSON structures.
  **Expected**: SessionFinalizer serializes the entire Detail structure as compact JSON and prints it on a single line (or wrapped by terminal).

- **Condition**: Session.Error.Detail contains non-serializable data (e.g., Go channels, functions).
  **Expected**: JSON serialization fails. SessionFinalizer prints: `"Detail: <failed to serialize detail>"` and logs the serialization error.

- **Condition**: SessionFinalizer is called multiple times for the same session.
  **Expected**: Each invocation prints the same status output. Socket deletion on subsequent calls is a no-op (socket already deleted).

- **Condition**: stdout or stderr is closed or redirected to /dev/null.
  **Expected**: Print operations may fail silently. SessionFinalizer does not check for print errors.

- **Condition**: Session.Error is neither AgentError nor RuntimeError (violates Session.Fail() invariant).
  **Expected**: SessionFinalizer attempts to type-assert the error to AgentError and RuntimeError. Both fail. SessionFinalizer falls back to printing:
  ```
  Session <SessionID> failed. Workflow: <WorkflowName>
  Error: <error.Error() string representation>
  ```
  No Agent/Issuer/State/Detail lines are printed.

- **Condition**: Session.WorkflowName or Session.ID contains special characters (e.g., newlines, control characters).
  **Expected**: SessionFinalizer prints the values as-is. Terminal rendering may be affected, but no escaping is performed.

- **Condition**: Session.Error.Message is very long (e.g., 10 KB).
  **Expected**: SessionFinalizer prints the entire message to stderr. Terminal output may be truncated or wrapped by the terminal emulator, but no size limit is enforced.

- **Condition**: Session directory or files are deleted manually before SessionFinalizer is called.
  **Expected**: SessionFinalizer does not access session files (metadata, events). It only reads from the in-memory Session entity and deletes the socket. Session status printing proceeds normally.

- **Condition**: Session.Error.Detail serialization produces very large JSON output (e.g., 1 MB).
  **Expected**: SessionFinalizer serializes and prints the entire JSON to stderr. Performance may degrade, but no size limit is enforced.

## Related

- [Session](../entities/session/session.md) - Session entity structure and terminal statuses
- [AgentError](../entities/agent_error.md) - Agent-reported error type
- [RuntimeError](../entities/runtime_error.md) - Runtime component error type
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Manages socket cleanup
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture and session lifecycle
