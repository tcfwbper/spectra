# ErrorProcessor

## Overview

ErrorProcessor handles RuntimeMessage with `type="error"`. It validates session status, validates the Claude session ID, derives the agent role from the current node definition, constructs an AgentError entity, and calls PersistentSession.Fail to transition the session to "failed" status. PersistentSession automatically persists the failed state. ErrorProcessor is the component responsible for translating agent-reported errors into session failure. ErrorProcessor does not perform cleanup (socket deletion, status printing); the main loop receives the termination notification and handles cleanup.

## Boundaries

- Owns: session status validation for error processing (must be "initializing" or "running").
- Owns: agent role derivation from current node definition.
- Owns: AgentError entity construction.
- Owns: PersistentSession.Fail invocation with the constructed AgentError.
- Delegates: Claude session ID validation to ValidateClaudeSessionID helper.
- Delegates: session cleanup (socket, stdout) to the main loop via termination notification.
- Delegates: persistence to PersistentSession (automatic, non-fatal).
- Delegates: panic recovery to MessageRouter.
- Must not: manage socket lifecycle or connection handling.
- Must not: implement panic recovery.
- Must not: call PersistentSession.Run or PersistentSession.Done.
- Must not: directly modify Session's internal fields (use thread-safe methods only).
- Must not: record events to EventHistory.
- Must not: call SessionMetadataStore.Write() or EventStore.Append() directly.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State container with auto-persist | `GetStatusSafe()`, `GetCurrentStateSafe()`, `Fail(err, notifier)`, read `ID` | Must not call `Run()`, `Done()`, must not modify fields directly, must not call stores directly |
| `WorkflowDefinition` | Configuration source (initialized once) | Read `Nodes()` to find node by name | Must not modify |
| `ValidateClaudeSessionID` | Shared validation helper | Call with PersistentSession, Node, claudeSessionID | — |
| `TerminationNotifier` | Main loop notification | Pass to `PersistentSession.Fail()` | Must not send directly |
| `AgentError` | Error entity | Construct via `NewAgentError` | — |
| `RuntimeResponse` | Response entity | Construct via `SuccessResponse` or `ErrorResponse` | — |

Construction constraint: ErrorProcessor is initialized with PersistentSession, WorkflowDefinition (loaded once at initialization), and TerminationNotifier. ValidateClaudeSessionID is a stateless package-level function called directly.

## Behavior

1. ErrorProcessor is invoked by MessageRouter when a RuntimeMessage with `type="error"` is received.
2. Receives the sessionUUID and RuntimeMessage containing `ClaudeSessionID()` and payload `{message, detail}`.
3. Validates that `PersistentSession.GetStatusSafe()` is either "initializing" or "running". If the status is "completed" or "failed", returns `ErrorResponse("session terminated: status is '<actual-status>'")`.
4. Retrieves the current node name via `PersistentSession.GetCurrentStateSafe()`.
5. Retrieves the current node definition from WorkflowDefinition.Nodes() by matching node name.
6. Calls `ValidateClaudeSessionID(PersistentSession, currentNode, RuntimeMessage.ClaudeSessionID())`. If validation returns an error, returns `ErrorResponse(<validation error message>)`. The error is not recorded.
7. Derives `agentRole` from the current node definition: if node `Type() == "agent"`, sets `agentRole` to `node.AgentRole()`; if node `Type() == "human"`, sets `agentRole` to empty string `""`.
8. Parses the error payload from RuntimeMessage.Payload() to extract `message` and `detail` fields.
9. Constructs an AgentError entity with: `AgentRole` set to the derived agentRole, `Message` from the payload, `Detail` from the payload (nil if absent or null), `OccurredAt` set to the current POSIX timestamp, `SessionID` set to the sessionUUID, and `FailingState` set to the current node name.
10. Calls `PersistentSession.Fail(agentError, terminationNotifier)` to transition the session to "failed" status in memory, populate Session.Error, and notify the main loop. PersistentSession automatically persists the failed state (non-fatal if persistence fails).
11. If `PersistentSession.Fail()` returns an error (e.g., session already failed by concurrent operation), returns `ErrorResponse("failed to record error: <error-details>")`.
12. Returns `SuccessResponse("error recorded | session=<SessionID> | failingState=<FailingState> | agentRole=<AgentRole> | error=<ErrorMessage>")`.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | PersistentSession reference | Valid, constructed via NewPersistentSession | Yes |
| WorkflowDefinition | WorkflowDefinition reference | Valid, fully validated, loaded once at initialization | Yes |
| TerminationNotifier | chan<- struct{} | Non-nil, buffered, capacity >= 2 | Yes |

### For ProcessError Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string (UUID) | Valid UUID, references an existing session | Yes |
| RuntimeMessage | RuntimeMessage | Valid RuntimeMessage with `Type()="error"` | Yes |

## Outputs

### Success Cases

**Case 1: Error Recorded Successfully**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `SuccessResponse("error recorded \| session=<SessionID> \| failingState=<FailingState> \| agentRole=<AgentRole> \| error=<ErrorMessage>")` |

### Error Cases

| Case | Response |
|------|----------|
| Session terminated | `ErrorResponse("session terminated: status is '<status>'")` |
| Claude session ID validation failed | `ErrorResponse(<validation error message>)` |
| Error payload parsing failed | `ErrorResponse("invalid error payload: <details>")` |
| Session.Fail returned error | `ErrorResponse("failed to record error: <details>")` |
| Current node not found in workflow | `ErrorResponse("current node '<name>' not found in workflow definition")` |

## Invariants

1. **Terminal Status Rejection**: ErrorProcessor must reject errors when `PersistentSession.GetStatusSafe()` is "completed" or "failed". Only "initializing" and "running" are accepted.

2. **Claude Session ID Validation Before Recording**: Claude session ID validation (via ValidateClaudeSessionID) must occur before AgentError construction. Validation failures result in immediate rejection.

3. **AgentRole Derivation Server-Side**: ErrorProcessor must derive `agentRole` from the current node's definition (node type → agent_role field or empty string). It must not read agentRole from the RuntimeMessage payload.

4. **FailingState Capture**: ErrorProcessor must capture `PersistentSession.GetCurrentStateSafe()` at processing time and store it in `AgentError.FailingState`. The current state must not change as a result of error processing (only status changes).

5. **First Error Wins**: PersistentSession.Fail enforces the first-error-wins policy. If a concurrent operation already failed the session, ErrorProcessor receives an error from PersistentSession.Fail and returns it to the caller.

6. **Cleanup Delegation**: ErrorProcessor does not handle cleanup operations. PersistentSession.Fail sends notification to the main loop via terminationNotifier. The main loop handles cleanup.

7. **No Panic Handling**: ErrorProcessor does not implement panic recovery. That is MessageRouter's responsibility.

8. **No PersistentSession.Done or PersistentSession.Run**: ErrorProcessor must never call PersistentSession.Done or PersistentSession.Run.

9. **Session Access via Thread-Safe Methods**: All reads and writes must go through PersistentSession's exported methods.

10. **WorkflowDefinition Loaded Once**: WorkflowDefinition is loaded at initialization, not per-invocation.

11. **No Event Recording**: ErrorProcessor does not write to EventHistory. Error processing is separate from event recording.

12. **No Direct Store Access**: ErrorProcessor must not call SessionMetadataStore.Write() or EventStore.Append() directly.

## Edge Cases

- Condition: `PersistentSession.GetStatusSafe()` returns "completed".
  Expected: Returns `ErrorResponse("session terminated: status is 'completed'")`. Error not recorded.

- Condition: `PersistentSession.GetStatusSafe()` returns "failed".
  Expected: Returns `ErrorResponse("session terminated: status is 'failed'")`. Error not recorded.

- Condition: `PersistentSession.GetStatusSafe()` returns "initializing".
  Expected: Accepted. Proceeds with validation, AgentError construction, and PersistentSession.Fail.

- Condition: `PersistentSession.GetStatusSafe()` returns "running".
  Expected: Accepted. Normal processing flow.

- Condition: ValidateClaudeSessionID returns an error.
  Expected: Returns error response. Error not recorded.

- Condition: Current node is a human node.
  Expected: `agentRole` derived as empty string "". AgentError constructed with empty AgentRole. Session transitions to "failed" normally.

- Condition: RuntimeMessage payload is missing `message` field.
  Expected: Returns `ErrorResponse("invalid error payload: missing required field 'message'")`.

- Condition: RuntimeMessage payload has `detail` set to null or missing.
  Expected: Treats detail as nil. AgentError constructed with nil Detail. Proceeds normally.

- Condition: PersistentSession.Fail returns error "session already failed".
  Expected: Returns `ErrorResponse("failed to record error: session already failed")`.

- Condition: Multiple agents raise errors simultaneously via concurrent connections.
  Expected: Each ProcessError runs in a separate goroutine. PersistentSession.Fail serializes via Session's internal lock. First error to call Fail succeeds. Subsequent calls return "session already failed".

- Condition: Current node name from PersistentSession.GetCurrentStateSafe() does not match any node in WorkflowDefinition.
  Expected: Returns `ErrorResponse("current node '<name>' not found in workflow definition")`.

- Condition: Error message or detail contains very large data (e.g., 5 MB stack trace).
  Expected: Processed normally. Performance may degrade but no error unless persistence limits hit.

- Condition: Concurrent event and error for the same session.
  Expected: Session's internal lock serializes access. If error is processed first (Fail succeeds), subsequent event processing sees status "failed" and rejects. If event is processed first, error still finds session in "running" and Fail succeeds.

- Condition: AgentError construction fails (e.g., empty message after parse).
  Expected: Returns `ErrorResponse("invalid error payload: <details>")`. Session.Fail is not called.

## Related

- [PersistentSession](./persistent_session.md) — State container with automatic persistence
- [MessageRouter](./message_router.md) — dispatches error messages to ErrorProcessor
- [ValidateClaudeSessionID](./validate_claude_session_id.md) — shared Claude session ID validation helper
- [AgentError](../entities/agent_error.md) — error entity constructed by ErrorProcessor
- [RuntimeResponse](../entities/runtime_response.md) — response entity returned to caller
- [Session lifecycle](../entities/session/lifecycle.md) — Session.Fail method (wrapped by PersistentSession)
- [WorkflowDefinition](../components/workflow_definition.md) — workflow structure (loaded once)
- [Node](../components/node.md) — provides node type and agent role
