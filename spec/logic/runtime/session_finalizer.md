# SessionFinalizer

## Overview

SessionFinalizer logs the final status of a session when it reaches a terminal state or when Runtime terminates, and returns an appropriate exit code. It is invoked by Runtime after socket cleanup has been performed. SessionFinalizer reads the final session status from the PersistentSession (via getter methods) and logs the status via the injected Logger. SessionFinalizer does not perform any resource cleanup (socket, directory, files); all cleanup is Runtime's responsibility.

## Boundaries

- Owns: final session status logging (success, failure, non-terminal) via Logger.
- Owns: exit code determination based on session terminal status.
- Owns: error detail formatting (AgentError vs RuntimeError distinction).
- Delegates: resource cleanup (socket, directory, files) to Runtime (performed before SessionFinalizer is called).
- Delegates: log output destination (stdout/stderr routing) to the Logger implementation.
- Must not: perform any resource cleanup (socket, directory, files, session deletion).
- Must not: modify Session state (read-only access).
- Must not: write directly to stdout or stderr (all output via Logger).
- Must not: return errors — all logging operations are best-effort.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State source | Read `ID`, `WorkflowName`, `GetStatusSafe()`, `GetErrorSafe()` | Must not mutate session state, must not call stores directly |
| `AgentError` | Error type | Type-assert and read fields via getters | Must not construct |
| `RuntimeError` | Error type | Type-assert and read fields via getters | Must not construct |
| `Logger` | Output | `Info(msg, args...)`, `Warn(msg, args...)`, `Error(msg, args...)` | Must not use for unrelated logging |

Construction constraint: SessionFinalizer is constructed with `Logger` injected. It is a lightweight unit that can be constructed once and reused across sessions.

## Behavior

1. SessionFinalizer is invoked by Runtime with a single input: `session` (*PersistentSession).
2. If `session` is nil, logs via `Logger.Error("SessionFinalizer called with nil session")` and returns exit code 1.
3. Reads `Session.Status` (via `GetStatusSafe()` or direct field access if session is in terminal state and immutable).
4. If `Session.Status == "completed"`:
   - Logs via `Logger.Info("session completed", "sessionID", session.ID, "workflow", session.WorkflowName)`.
   - Returns exit code 0.
5. If `Session.Status == "failed"`:
   - Reads `Session.Error` (via `GetErrorSafe()`).
   - If `Session.Error` is nil (violates Session invariant), logs via `Logger.Error("session failed", "sessionID", session.ID, "workflow", session.WorkflowName, "error", "unknown error")` and returns exit code 1.
   - If `Session.Error` is `*AgentError`:
     - Logs via `Logger.Error("session failed", "sessionID", session.ID, "workflow", session.WorkflowName, "error", error.Message(), "agent", error.AgentRole(), "state", error.FailingState(), "detail", detailJSON)`.
     - Detail is serialized as compact JSON. If Detail is nil or empty (`{}`), the "detail" key-value pair is omitted from the log args.
   - If `Session.Error` is `*RuntimeError`:
     - Logs via `Logger.Error("session failed", "sessionID", session.ID, "workflow", session.WorkflowName, "error", error.Message(), "issuer", error.Issuer(), "state", error.FailingState(), "detail", detailJSON)`.
     - Detail is serialized as compact JSON. If Detail is nil or empty (`{}`), the "detail" key-value pair is omitted from the log args.
   - If `Session.Error` is neither `*AgentError` nor `*RuntimeError` (unexpected type):
     - Logs via `Logger.Error("session failed", "sessionID", session.ID, "workflow", session.WorkflowName, "error", error.Error())`.
   - Returns exit code 1.
6. If `Session.Status` is `"initializing"` or `"running"` (non-terminal, likely due to signal interruption):
   - Logs via `Logger.Warn("session terminated with non-terminal status", "sessionID", session.ID, "workflow", session.WorkflowName, "status", session.Status)`.
   - If `Session.Error` is non-nil, logs error details as in step 5.
   - Returns exit code 1.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Logger | logger.Logger | Non-nil Logger interface implementation | Yes |

### For Finalize Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | *PersistentSession | Reference to the PersistentSession wrapper. May be nil (handled gracefully). | Yes |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| ExitCode | int | 0 if session completed successfully, 1 otherwise (failed, non-terminal, nil session) |

## Invariants

1. **No Resource Cleanup**: SessionFinalizer must not perform any resource cleanup. All cleanup is Runtime's responsibility and is performed before invoking SessionFinalizer.

2. **All Output Via Logger**: SessionFinalizer must not write directly to stdout or stderr. All output goes through the injected Logger interface. The Logger implementation determines the actual output destination and format.

3. **Logger Level Semantics**: Completed sessions use `Logger.Info`. Failed sessions use `Logger.Error`. Non-terminal status sessions use `Logger.Warn`.

4. **Exit Code Contract**: Returns 0 only when `Session.Status == "completed"`. All other cases (failed, non-terminal, nil session) return 1.

5. **Idempotent**: SessionFinalizer is safe to call multiple times for the same session. Each call logs the same output and returns the same exit code.

6. **No Return Error**: SessionFinalizer does not return errors. All logging operations are best-effort via Logger (fire-and-forget).

7. **Nil Session Safety**: SessionFinalizer handles nil session input gracefully by logging an error and returning exit code 1.

8. **Detail Omission**: If Error.Detail is nil or an empty JSON object (`{}`), the "detail" key-value pair is omitted from the Logger args entirely (not logged as empty).

9. **No Panic on Unexpected Error Type**: If Session.Error is not *AgentError or *RuntimeError, SessionFinalizer falls back to logging error.Error() string representation.

10. **Detail JSON Serialization**: Error.Detail is serialized as compact JSON (no pretty-printing). If serialization fails, the detail is logged as `"<failed to serialize detail>"`.

## Edge Cases

- **Condition**: Session is nil.
  **Expected**: Logs error via Logger.Error. Returns exit code 1.

- **Condition**: Session.Status is "completed".
  **Expected**: Logs info. Returns exit code 0.

- **Condition**: Session.Status is "failed" with AgentError and empty Detail (`{}`).
  **Expected**: Logs error with agent, state fields. "detail" key-value pair is omitted. Returns exit code 1.

- **Condition**: Session.Status is "failed" with RuntimeError and non-empty Detail.
  **Expected**: Logs error with issuer, state, detail fields. Detail is compact JSON. Returns exit code 1.

- **Condition**: Session.Status is "failed" but Session.Error is nil (violates Session invariant).
  **Expected**: Logs error with "unknown error" message. Returns exit code 1.

- **Condition**: Session.Status is "failed" and Session.Error is neither *AgentError nor *RuntimeError.
  **Expected**: Logs error using error.Error() string. Returns exit code 1.

- **Condition**: Session.Status is "initializing" or "running" (signal interruption).
  **Expected**: Logs warning with non-terminal status. If Error is non-nil, also logs error details. Returns exit code 1.

- **Condition**: Session.Error.Detail contains non-serializable data.
  **Expected**: JSON serialization fails. Detail logged as `"<failed to serialize detail>"`.

- **Condition**: SessionFinalizer called multiple times for the same session.
  **Expected**: Each call logs the same output. No side effects.

- **Condition**: Logger implementation silently drops messages (e.g., NopLogger in tests).
  **Expected**: SessionFinalizer still returns correct exit code. Logging is best-effort.

## Related

- [PersistentSession](./persistent_session.md) — State container with automatic persistence, read-only access for status reporting
- [Session](../entities/session/session.md) — Underlying Session entity
- [AgentError](../entities/agent_error.md) — Agent-reported error type
- [RuntimeError](../entities/runtime_error.md) — Runtime component error type
- [Logger](../logger/logger.md) — Structured logging interface
