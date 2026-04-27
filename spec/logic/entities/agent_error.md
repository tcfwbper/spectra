# AgentError

## Overview

An AgentError is a failure signal raised by an agent when it cannot complete its task due to an unrecoverable error (e.g., model failure, missing context, tool failure). When an AgentError is raised, the runtime immediately halts the state machine, transitions the session's `Status` to `"failed"`, records the error details, and notifies the human. The `CurrentState` remains at the node where the error occurred. The session is marked as failed and cannot be resumed in the current design. Failed sessions are retained on disk for inspection and debugging.

## Behavior

1. An agent raises an error by invoking `spectra-agent error <message> --session-id <UUID> [--claude-session-id <UUID>] [--detail <json>]`.
2. The `--session-id` flag is mandatory and used to find the runtime socket. If omitted, the spectra-agent CLI can't find the runtime socket.
3. The runtime immediately halts the state machine and captures the current state machine node name as `FailingState`.
4. The runtime records the agent role that raised the error and the current POSIX timestamp.
5. The runtime transitions the session's `Status` to `"failed"`. The `CurrentState` remains at the node where the error occurred (it does not change).
6. The runtime stores the error details in the session's `Error` field.
7. The runtime writes the error to the session's error log file and notifies the human (e.g., via console output).
8. The session is persisted to disk for inspection. No automatic retry mechanism is provided.
9. The session remains with `Status == "failed"` permanently and must be manually resolved by creating a new session.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| AgentRole | string | Agent role derived from the current node's definition by ErrorProcessor. Non-empty for agent nodes; empty string `""` for human nodes. | Yes (populated server-side) |
| Message | string | Non-empty, human-readable error description | Yes |
| Detail | JSON object | Valid JSON object, may be `null` or `{}` | No |
| SessionID | string (UUID) | Must reference an existing session | Yes |

## Outputs

### AgentError Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| AgentRole | string | Valid agent role (agent nodes) or empty string (human nodes) | Agent role that raised the error, derived from the current node's definition |
| Message | string | Non-empty, human-readable text | Error message |
| Detail | JSON object | Valid JSON object, may be `null` or `{}` | Additional error context (e.g., stack trace, API error response) |
| OccurredAt | int64 (POSIX timestamp) | Positive integer | Timestamp when the error occurred |
| SessionID | string (UUID) | Valid UUID, references an existing session | Associated session identifier |
| FailingState | string | Non-empty, valid state machine node name | State machine node where the error occurred |

## Invariants

1. **Session Status Transition**: When an `AgentError` is raised, the session's `Status` must immediately transition to `"failed"`. The `CurrentState` remains unchanged.

2. **Session Error Reference**: When an `AgentError` is raised, the session's `Error` field must be set to reference the error instance.

3. **Message Non-Empty**: `Message` must contain at least one non-whitespace character.

4. **AgentRole Derivation**: `AgentRole` is populated server-side by ErrorProcessor from the current node's `agent_role` field. For agent nodes, it must be a valid agent role defined in the workflow. For human nodes, it is an empty string `""`. AgentRole is **not** provided from the wire payload.

5. **Session Reference Integrity**: `SessionID` must reference an existing session at the time of error creation.

6. **FailingState Consistency**: `FailingState` must match the session's `CurrentState` at the moment the error is raised. `CurrentState` does not change when the error occurs.

7. **No Automatic Retry**: The runtime must not automatically retry a failed session. Manual intervention is required.

8. **Error Immutability**: Once an `AgentError` is created and recorded, none of its fields may be modified.

9. **Terminal Status**: Once a session reaches `Status == "failed"` due to an `AgentError`, the `Status` must not transition to any other value.

## Edge Cases

- **Condition**: Agent raises an error with an empty or whitespace-only message.
  **Expected**: The runtime must reject the error creation with a validation error.

- **Condition**: Agent raises an error with invalid JSON in `Detail`.
  **Expected**: The runtime must reject the error creation with a JSON parse error.

- **Condition**: Error originates from a human node.
  **Expected**: `AgentRole` is set to an empty string `""` by ErrorProcessor. The error is recorded normally.

- **Condition**: `AgentRole` is not defined in the workflow.
  **Expected**: This case cannot occur because `AgentRole` is derived from the current node's definition, which is validated at workflow load time.

- **Condition**: `SessionID` references a non-existent session.
  **Expected**: The runtime must reject the error creation with a "session not found" error.

- **Condition**: `SessionID` references a session with `Status == "failed"` or `Status == "completed"`.
  **Expected**: The runtime must reject the error creation with a "session terminated" error and log a warning.

- **Condition**: Multiple agents raise errors simultaneously.
  **Expected**: The runtime must serialize error processing. The first error to be recorded transitions the session's `Status` to `"failed"`. Subsequent errors are logged but do not overwrite the session's `Error` field.

- **Condition**: An error is raised while the session's `Status == "initializing"`.
  **Expected**: The runtime must record `FailingState` as the entry node, transition `Status` to `"failed"`, and persist the error.

- **Condition**: An error is raised while the session's `Status == "running"`.
  **Expected**: The runtime must record `FailingState` as the current `CurrentState`, transition `Status` to `"failed"`, and persist the error.

- **Condition**: Session is deleted.
  **Expected**: The `AgentError` is removed from the filesystem along with all other session resources. Subsequent error queries must fail with a "session not found" error.

- **Condition**: Human requests session recovery after failure.
  **Expected**: The runtime must reject the recovery request. The human must create a new session to retry the workflow.

- **Condition**: Error `Detail` contains sensitive information (e.g., API keys, credentials).
  **Expected**: The runtime must persist the error as-is. It is the agent's responsibility to sanitize sensitive data before raising the error.

## Related

- [Session](./session/session.md) - AgentError transitions the session's `Status` to `"failed"`
- [Event](./event.md) - Events are the primary mechanism for workflow progression; errors halt this mechanism
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
