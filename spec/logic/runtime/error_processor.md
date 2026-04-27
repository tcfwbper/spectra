# ErrorProcessor

## Overview

ErrorProcessor handles RuntimeMessage with `type="error"`. It validates session status, constructs an AgentError entity, and calls `Session.Fail()` to transition the session to `"failed"` status, record the error, attempt persistence, and notify the main loop via terminationNotifier. ErrorProcessor is the only component responsible for transitioning sessions to `"failed"` status when agents report unrecoverable errors. ErrorProcessor does not perform cleanup (socket deletion, status printing); the main loop receives the termination notification and invokes SessionFinalizer.

## Behavior

1. ErrorProcessor is invoked by MessageRouter when a RuntimeMessage with `type="error"` is received.
2. ErrorProcessor receives the RuntimeMessage containing `claudeSessionID` field and payload `{message, detail}`, along with the session UUID (determined from the socket path). The `agentRole` is **not** part of the wire payload; it is derived server-side in step 6.
3. ErrorProcessor validates that `Session.Status` is either "initializing" or "running" by reading the session's status (thread-safe via Session's internal locking). If the status is "completed" or "failed", ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="session terminated: status is '<actual-status>'"` without processing the error further.
4. ErrorProcessor loads the WorkflowDefinition for the session's workflow.
5. ErrorProcessor retrieves the current node definition from the workflow using `Session.CurrentState` as the node name.
6. ErrorProcessor derives `agentRole` from the current node definition: if the node's `type == "agent"`, `agentRole` is set to the node's `agent_role` field; if the node's `type == "human"`, `agentRole` is set to an empty string `""`.
7. ErrorProcessor validates the `claudeSessionID` from the RuntimeMessage based on the current node type:
   - If the current node's `type == "agent"`:
     - ErrorProcessor calls `Session.GetSessionDataSafe("<CurrentState>.ClaudeSessionID")` to retrieve the stored Claude session ID.
     - If the key does not exist (returns `(nil, false)`), ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID not found for node '<CurrentState>'"`. The error is not recorded.
     - If the stored value does not match `RuntimeMessage.claudeSessionID`, ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID mismatch: expected <stored-uuid> but got <provided-uuid>"`. The error is not recorded.
   - If the current node's `type == "human"`:
     - If `RuntimeMessage.claudeSessionID` is not an empty string `""`, ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="invalid claude session ID for human node: must be empty"`. The error is not recorded.
8. ErrorProcessor constructs an AgentError entity with: the `agentRole` derived from the current node definition (step 6), the message and detail from the RuntimeMessage, `OccurredAt` set to the current POSIX timestamp, `SessionID` set to the session UUID, and `FailingState` set to `Session.CurrentState`.
9. ErrorProcessor calls `Session.Fail(agentError, terminationNotifier)` to transition the session to `"failed"` status in memory first, populate `Session.Error` with the AgentError, attempt to persist the session to SessionMetadataStore (best-effort), and notify the main loop via terminationNotifier.
10. If `Session.Fail()` returns an error (e.g., session already failed), ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="failed to record error: <error-details>"`.
11. ErrorProcessor returns a RuntimeResponse with `status="success"` and `message="error recorded | session=<SessionID> | failingState=<FailingState> | agentRole=<AgentRole> | error=<ErrorMessage>"`.
12. ErrorProcessor does not catch panics. Panic recovery is handled by MessageRouter.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Session | *Session | Reference to the Session entity shared across all runtime components | Yes |
| WorkflowDefinitionLoader | WorkflowDefinitionLoader | Loader for workflow definitions | Yes |
| TerminationNotifier | chan<- struct{} | Channel for notifying the main loop of session termination (passed to Session.Fail()) | Yes |

### For ProcessError Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string (UUID) | Valid UUID, references an existing session | Yes |
| RuntimeMessage | RuntimeMessage | Valid RuntimeMessage with `type="error"` and payload `{message, detail}` | Yes |

## Outputs

### Success Cases

**Case 1: Error Recorded Successfully**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="success"`, `message="error recorded \| session=<SessionID> \| failingState=<FailingState> \| agentRole=<AgentRole> \| error=<ErrorMessage>"` |

### Error Cases

**Case 2: Session Terminated**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="session terminated: status is '<actual-status>'"` |

**Case 3: Claude Session ID Validation Failed (Agent Node, Not Found)**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="claude session ID not found for node '<NodeName>'"` |

**Case 4: Claude Session ID Validation Failed (Agent Node, Mismatch)**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="claude session ID mismatch: expected <stored-uuid> but got <provided-uuid>"` |

**Case 5: Claude Session ID Validation Failed (Human Node, Non-Empty)**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="invalid claude session ID for human node: must be empty"` |

**Case 6: Failed to Record Error (Session.Fail() returned error)**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="failed to record error: <error-details>"` |

### Function Signature (Go-like pseudocode)

```
func (ep *ErrorProcessor) ProcessError(
    sessionUUID string,
    runtimeMessage RuntimeMessage,
) RuntimeResponse
```

## Invariants

1. **Terminal Status Validation**: ErrorProcessor must only accept errors when `Session.Status` is "initializing" or "running". Sessions with status "completed" or "failed" must be rejected.

2. **AgentRole Derivation**: ErrorProcessor must derive `agentRole` from the current node's definition (`Session.CurrentState` → node lookup → `agent_role` field). For agent nodes, `agentRole` is set to the node's `agent_role` field. For human nodes, `agentRole` is set to an empty string `""`. ErrorProcessor must **not** expect or read `agentRole` from the RuntimeMessage payload.

3. **Claude Session ID Validation Priority**: ErrorProcessor must validate the `claudeSessionID` from RuntimeMessage (via `Session.GetSessionDataSafe()`) against the current node's requirements before recording the error. Validation failures result in immediate rejection without error recording.

4. **Status Transition via Session.Fail()**: ErrorProcessor constructs an AgentError and calls `Session.Fail(agentError, terminationNotifier)` to transition the session to "failed" status. `Session.Fail()` handles in-memory-first updates, best-effort persistence, and main loop notification internally.

5. **FailingState Capture**: ErrorProcessor must capture `Session.CurrentState` at the time of error processing and store it in `AgentError.FailingState`. The `CurrentState` field in Session must not change when an error occurs.

6. **First Error Wins**: `Session.Fail()` enforces the first-error-wins policy. If a session has already transitioned to "failed" status, `Session.Fail()` returns an error, which ErrorProcessor returns to the caller.

7. **Cleanup Delegation**: ErrorProcessor does not handle cleanup operations (socket deletion, human notification). `Session.Fail()` notifies the main loop via terminationNotifier. The main loop (Runtime) calls SessionFinalizer to perform cleanup.

8. **No Panic Handling**: ErrorProcessor does not implement panic recovery. Panic handling is the responsibility of MessageRouter.

9. **Single Responsibility**: ErrorProcessor focuses solely on error validation, AgentError construction, and calling `Session.Fail()`. It does not handle event processing, transition evaluation, or socket management.

10. **Node Type Determination**: ErrorProcessor must determine the current node's type by loading the WorkflowDefinition and looking up the node definition using `Session.CurrentState`. It must not infer node type from RuntimeMessage fields.

## Edge Cases

- **Condition**: Session reference is nil or invalid during ErrorProcessor initialization.
  **Expected**: This is a programming error. ErrorProcessor initialization should fail or panic during startup, not during message processing.

- **Condition**: `Session.Status == "completed"`.
  **Expected**: ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="session terminated: status is 'completed'"`. The error is not recorded.

- **Condition**: `Session.Status == "failed"`.
  **Expected**: ErrorProcessor validates status and returns a RuntimeResponse with `status="error"` and `message="session terminated: status is 'failed'"`. The error is not recorded. If ErrorProcessor mistakenly calls `Session.Fail()`, it returns an error: `"session already failed"`, which ErrorProcessor returns to the caller.

- **Condition**: `Session.Status == "initializing"`.
  **Expected**: ErrorProcessor accepts the error, constructs an AgentError with `FailingState` set to the entry node (Session.CurrentState at initialization), calls `Session.Fail(agentError, terminationNotifier)` to transition Status to "failed", persist (best-effort), and notify the main loop.

- **Condition**: `Session.Status == "running"`.
  **Expected**: ErrorProcessor accepts the error, constructs an AgentError with `FailingState` set to the current node (Session.CurrentState), calls `Session.Fail(agentError, terminationNotifier)` to transition Status to "failed", persist (best-effort), and notify the main loop.

- **Condition**: Current node is an agent node, and the stored `<NodeName>.ClaudeSessionID` does not exist in SessionData.
  **Expected**: ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID not found for node '<NodeName>'"`. The error is not recorded.

- **Condition**: Current node is an agent node, and `RuntimeMessage.claudeSessionID` does not match the stored value in SessionData.
  **Expected**: ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID mismatch: expected <stored-uuid> but got <provided-uuid>"`. The error is not recorded.

- **Condition**: Current node is a human node, and `RuntimeMessage.claudeSessionID` is not an empty string.
  **Expected**: ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="invalid claude session ID for human node: must be empty"`. The error is not recorded.

- **Condition**: Current node is a human node, and `RuntimeMessage.claudeSessionID` is an empty string.
  **Expected**: Validation succeeds. ErrorProcessor proceeds to record the error and transition the session to "failed".

- **Condition**: Current node is an agent node, and `RuntimeMessage.claudeSessionID` matches the stored value in SessionData.
  **Expected**: Validation succeeds. ErrorProcessor proceeds to record the error and transition the session to "failed".

- **Condition**: RuntimeMessage payload is missing `message` field.
  **Expected**: This case is prevented by RuntimeSocketManager validation. If it occurs, ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="invalid error payload: missing required field"`.

- **Condition**: RuntimeMessage payload has `detail` field set to `null` or missing.
  **Expected**: ErrorProcessor treats the detail as an empty JSON object `{}` (as per AgentError and RuntimeMessage specifications) and proceeds normally.

- **Condition**: SessionMetadataStore write fails due to disk full or permission error during `Session.Fail()`.
  **Expected**: `Session.Fail()` logs a warning about the persistence failure but returns `nil` (best-effort persistence). The session remains in `"failed"` status **in memory**, ensuring the runtime's behavior is correct. The error details may not be persisted to disk, but the session is still considered failed by the runtime. The main loop is notified via terminationNotifier. ErrorProcessor returns a success RuntimeResponse.

- **Condition**: Multiple agents raise errors simultaneously via concurrent socket connections.
  **Expected**: Each ErrorProcessor invocation runs in a separate goroutine. The first error to call `Session.Fail()` succeeds (session-level write lock serializes calls). Subsequent `Session.Fail()` calls return an error: `"session already failed"`, which ErrorProcessor returns as a RuntimeResponse with `status="error"`. The first error wins.

- **Condition**: Error message or detail contains very large data (e.g., 5 MB stack trace).
  **Expected**: ErrorProcessor processes the error normally. SessionMetadataStore serializes the large data to JSON. Performance may degrade, but no error occurs unless SessionMetadataStore encounters filesystem limits.

- **Condition**: Current node is a human node, and the error is accepted.
  **Expected**: ErrorProcessor derives `agentRole` as an empty string `""` (human nodes have no `agent_role`). The AgentError is constructed with `AgentRole=""`. The session transitions to "failed" normally.

- **Condition**: WorkflowDefinition cannot be loaded (file missing or parse error).
  **Expected**: ErrorProcessor returns a RuntimeResponse with `status="error"` and `message="failed to load workflow definition: <error-details>"`. The error is not recorded.

- **Condition**: An error is raised while the session is processing an event (concurrent event and error).
  **Expected**: EventProcessor and ErrorProcessor run in separate goroutines. Session's internal lock serializes state access. If the error is processed first (via `Session.Fail()`), subsequent event validation will see `Session.Status == "failed"` and reject the event with "session not ready: status is 'failed'". If the event is processed first, the error will find the session still in "running" status and succeed.

## Related

- [RuntimeMessage](../entities/runtime_message.md) - Input message format
- [RuntimeResponse](../entities/runtime_response.md) - Output response format
- [AgentError](../entities/agent_error.md) - Error entity structure
- [Session](../entities/session/session.md) - Session state and lifecycle
- [SessionMetadataStore](../storage/session_metadata_store.md) - Session metadata persistence
- [MessageRouter](./message_router.md) - Dispatches messages to ErrorProcessor
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
