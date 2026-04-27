# RuntimeError

## Overview

A RuntimeError is a failure signal raised by runtime components when they encounter an unrecoverable error during session execution (e.g., socket creation failure, state transition failure, panic in message processing). When a RuntimeError is raised, the runtime immediately halts operations, transitions the session's `Status` to `"failed"` (in memory first, then attempts persistence), records the error details, and notifies the human. The `CurrentState` remains at the node where the error occurred. The session is marked as failed and cannot be resumed in the current design. Failed sessions are retained on disk for inspection and debugging.

## Behavior

1. A runtime component raises a RuntimeError when it encounters an unrecoverable error that prevents the session from continuing.
2. The runtime component that detects the error is responsible for constructing the RuntimeError entity with the appropriate issuer, message, and details.
3. The runtime immediately halts the state machine and captures the current state machine node name as `FailingState`.
4. The runtime records the component name that raised the error (stored in `Issuer`) and the current POSIX timestamp.
5. The runtime transitions the session's `Status` to `"failed"` **in memory first** to ensure the runtime's behavior is consistent, then attempts to persist the updated session metadata to SessionMetadataStore.
6. If the SessionMetadataStore write succeeds, the error details are persisted. If the write fails (e.g., disk full, permission denied), the session remains in `"failed"` status in memory, and the runtime logs a warning about the persistence failure. The runtime's behavior remains correct because the in-memory status is authoritative.
7. The runtime stores the error details in the session's `Error` field (union type: AgentError or RuntimeError).
8. The runtime writes the error to the session's error log file (if persistence is available) and notifies the human (e.g., via console output).
9. The session is persisted to disk for inspection (if filesystem is available). No automatic retry mechanism is provided.
10. The session remains with `Status == "failed"` permanently and must be manually resolved by creating a new session.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Issuer | string | Non-empty string representing the runtime component name (e.g., "MessageRouter", "RuntimeSocketManager", "EventProcessor", "TransitionToNode", "Session") | Yes |
| Message | string | Non-empty, human-readable error description | Yes |
| Detail | JSON object | Valid JSON object, may be `null` or `{}` | No |
| SessionID | string (UUID) | Must reference an existing session | Yes |

## Outputs

### RuntimeError Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Issuer | string | Non-empty string | Runtime component that raised the error |
| Message | string | Non-empty, human-readable text | Error message |
| Detail | JSON object | Valid JSON object, may be `null` or `{}` | Additional error context (e.g., stack trace, system error details) |
| OccurredAt | int64 (POSIX timestamp) | Positive integer | Timestamp when the error occurred |
| SessionID | string (UUID) | Valid UUID, references an existing session | Associated session identifier |
| FailingState | string | Non-empty, valid state machine node name | State machine node where the error occurred |

## Invariants

1. **Session Status Transition**: When a `RuntimeError` is raised, the session's `Status` must immediately transition to `"failed"` **in memory first**, then the runtime attempts to persist to SessionMetadataStore. The `CurrentState` remains unchanged.

2. **Session Error Reference**: When a `RuntimeError` is raised, the session's `Error` field must be set to reference the error instance (union type: AgentError or RuntimeError).

3. **Message Non-Empty**: `Message` must contain at least one non-whitespace character.

4. **Issuer Non-Empty**: `Issuer` must be a non-empty string. The runtime does not validate that the issuer name matches a known component; any non-empty string is accepted.

5. **Session Reference Integrity**: `SessionID` must reference an existing session at the time of error creation.

6. **FailingState Consistency**: `FailingState` must match the session's `CurrentState` at the moment the error is raised. `CurrentState` does not change when the error occurs.

7. **No Automatic Retry**: The runtime must not automatically retry a failed session. Manual intervention is required.

8. **Error Immutability**: Once a `RuntimeError` is created and recorded, none of its fields may be modified.

9. **Terminal Status**: Once a session reaches `Status == "failed"` due to a `RuntimeError`, the `Status` must not transition to any other value.

10. **In-Memory Status Priority**: The in-memory session status is authoritative. If persistence fails, the session is still considered failed, and the runtime behavior must reflect this status.

## Edge Cases

- **Condition**: Runtime component raises an error with an empty or whitespace-only message.
  **Expected**: The runtime must reject the error creation with a validation error.

- **Condition**: Runtime component raises an error with invalid JSON in `Detail`.
  **Expected**: The runtime must reject the error creation with a JSON parse error.

- **Condition**: `Issuer` is an empty string or consists only of whitespace.
  **Expected**: The runtime must reject the error creation with a validation error.

- **Condition**: `SessionID` references a non-existent session.
  **Expected**: The runtime must reject the error creation with a "session not found" error.

- **Condition**: `SessionID` references a session with `Status == "failed"` or `Status == "completed"`.
  **Expected**: The runtime must reject the error creation with a "session terminated" error and log a warning.

- **Condition**: Multiple runtime components raise errors simultaneously.
  **Expected**: The runtime must serialize error processing. The first error to update the in-memory session status transitions `Status` to `"failed"`. Subsequent errors are logged but do not overwrite the session's `Error` field.

- **Condition**: An error is raised while the session's `Status == "initializing"`.
  **Expected**: The runtime must record `FailingState` as the entry node, transition `Status` to `"failed"` in memory, and attempt to persist the error.

- **Condition**: An error is raised while the session's `Status == "running"`.
  **Expected**: The runtime must record `FailingState` as the current `CurrentState`, transition `Status` to `"failed"` in memory, and attempt to persist the error.

- **Condition**: SessionMetadataStore write fails when persisting the RuntimeError.
  **Expected**: The runtime logs a warning with details (e.g., "failed to persist RuntimeError to disk: <error-details>"). The session remains in `"failed"` status in memory. The runtime continues with cleanup operations (e.g., socket deletion, human notification). The error details may be lost on disk, but the session status is still correct in memory.

- **Condition**: Session is deleted.
  **Expected**: The `RuntimeError` is removed from the filesystem along with all other session resources. Subsequent error queries must fail with a "session not found" error.

- **Condition**: Human requests session recovery after RuntimeError.
  **Expected**: The runtime must reject the recovery request. The human must create a new session to retry the workflow.

- **Condition**: Error `Detail` contains sensitive information (e.g., file paths, system configuration).
  **Expected**: The runtime must persist the error as-is. It is the issuer's responsibility to sanitize sensitive data before raising the error.

- **Condition**: RuntimeError is raised due to a panic in MessageRouter.
  **Expected**: The `Issuer` field is set to "MessageRouter". The `Detail` field contains the panic message and stack trace. The session transitions to `"failed"` status in memory, and persistence is attempted.

- **Condition**: RuntimeError is raised due to socket creation failure during session initialization.
  **Expected**: The `Issuer` field is set to the component that detected the failure (e.g., "Session"). The `Detail` field contains the underlying error details (e.g., "permission denied", "socket file already exists"). The session transitions to `"failed"` status, and initialization is aborted.

## Related

- [Session](./session/session.md) - RuntimeError transitions the session's `Status` to `"failed"`
- [AgentError](./agent_error.md) - AgentError is the other type of error that can halt a session; both share the Session.Error field via union type
- [Event](./event.md) - Events are the primary mechanism for workflow progression; runtime errors halt this mechanism
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
