# EventProcessor

## Overview

EventProcessor handles RuntimeMessage with `type="event"`. It validates session status, constructs an Event entity, writes it to EventStore, evaluates the transition using TransitionEvaluator, invokes Runtime's TransitionToNode to perform the state transition, and returns a RuntimeResponse. EventProcessor records all events to EventStore regardless of whether a matching transition is found, ensuring a complete audit trail. EventProcessor does not manage session lifecycle or socket cleanup; it focuses solely on event processing logic.

## Behavior

1. EventProcessor is invoked by MessageRouter when a RuntimeMessage with `type="event"` is received.
2. EventProcessor receives the RuntimeMessage containing `claudeSessionID` field and payload `{eventType, message, payload}`, along with the session UUID (determined from the socket path).
3. EventProcessor validates that `Session.Status == "running"` by reading the session's status (thread-safe via Session's internal locking). If validation fails, EventProcessor returns a RuntimeResponse with `status="error"` and `message="session not ready: status is '<actual-status>'"` without processing the event further.
4. EventProcessor loads the WorkflowDefinition for the session's workflow.
5. EventProcessor retrieves the current node definition from the workflow using `Session.CurrentState` as the node name.
6. EventProcessor validates the `claudeSessionID` from the RuntimeMessage based on the current node type:
   - If the current node's `type == "agent"`:
     - EventProcessor calls `Session.GetSessionDataSafe("<CurrentState>.ClaudeSessionID")` to retrieve the stored Claude session ID.
     - If the key does not exist (returns `(nil, false)`), EventProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID not found for node '<CurrentState>'"`. No event is recorded.
     - If the stored value does not match `RuntimeMessage.claudeSessionID`, EventProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID mismatch: expected <stored-uuid> but got <provided-uuid>"`. No event is recorded.
   - If the current node's `type == "human"`:
     - If `RuntimeMessage.claudeSessionID` is not an empty string `""`, EventProcessor returns a RuntimeResponse with `status="error"` and `message="invalid claude session ID for human node: must be empty"`. No event is recorded.
7. EventProcessor constructs an Event entity with: a generated UUID, the eventType from the payload, the message and payload from the RuntimeMessage, `EmittedBy` set to `Session.CurrentState`, `EmittedAt` set to the current POSIX timestamp, and `SessionID` set to the session UUID.
8. EventProcessor calls `Session.UpdateEventHistorySafe(event)` to append the event to the in-memory EventHistory and persist it to EventStore. This write operation occurs after Claude session ID validation to ensure only validated events are recorded.
9. If `Session.UpdateEventHistorySafe(event)` returns an error (EventStore write failed), EventProcessor constructs a RuntimeError with `Issuer="EventProcessor"`, `Message="failed to record event"`, and details, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to failed, and returns a RuntimeResponse with `status="error"` and `message="failed to record event: <error-details>"`.
10. EventProcessor invokes the package-level function `TransitionEvaluator.EvaluateTransition(WorkflowDefinition, Session.CurrentState, Event.Type)` (TransitionEvaluator is a stateless pure function, not an injected dependency).
11. If TransitionEvaluator returns `(nil, false, nil)` (no matching transition), EventProcessor returns a RuntimeResponse with `status="error"` and `message="no transition found for event '<EventType>' from node '<CurrentState>'"`. The event has already been recorded in step 8. **The session remains in `running` status**; the caller may retry with a different event.
12. If TransitionEvaluator returns a valid transition `(transition, isExitTransition, nil)`, EventProcessor invokes `TransitionToNode.Transition(Message=Event.Message, TargetNodeName=transition.ToNode, IsExitTransition=isExitTransition)`.
13. `TransitionToNode` executes the target node dispatch logic: For human nodes, it prints the message to stdout. For agent nodes, it loads the agent definition and invokes AgentInvoker to start the Claude CLI process. After the node-specific action completes, it updates `Session.CurrentState` to the target node. If `IsExitTransition == true`, it calls `Session.Done(terminationNotifier)` to mark the session as "completed". TransitionToNode uses fail-fast semantics with internal error handling: any operation failure (node lookup, agent definition loading, agent invocation, Session.Done) immediately causes TransitionToNode to construct a RuntimeError, call `Session.Fail(runtimeError, terminationNotifier)` internally, and return an error to EventProcessor. No rollback is performed for partial completions.
14. If `TransitionToNode` returns `nil` (success), EventProcessor returns a RuntimeResponse with `status="success"` and `message="event '<EventType>' processed successfully | session=<SessionID> | currentState=<NewCurrentState> | sessionStatus=<SessionStatus>"`.
15. If `TransitionToNode` returns an error, EventProcessor does NOT call Session.Fail again (TransitionToNode already did this internally). EventProcessor returns a RuntimeResponse with `status="error"` and `message="transition failed: <error-details>"`.
16. EventProcessor does not catch panics. Panic recovery is handled by MessageRouter.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Session | *Session | Reference to the Session entity shared across all runtime components | Yes |
| WorkflowDefinitionLoader | WorkflowDefinitionLoader | Loader for workflow definitions | Yes |
| TransitionToNode | TransitionToNode (struct/interface) | Initialized TransitionToNode instance with its own dependencies (Session, AgentDefinitionLoader, AgentInvoker, TerminationNotifier). EventProcessor invokes its `Transition` method. | Yes |
| TerminationNotifier | chan<- struct{} | Channel for notifying the main loop of session termination (passed to Session methods) | Yes |

**Note**: TransitionEvaluator is a stateless package-level function, not an injected dependency. EventProcessor calls it directly without holding a reference.

### For ProcessEvent Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string (UUID) | Valid UUID, references an existing session | Yes |
| RuntimeMessage | RuntimeMessage | Valid RuntimeMessage with `type="event"` and payload `{eventType, message, payload}` | Yes |

## Outputs

### Success Cases

**Case 1: Event Processed and Transition Successful**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="success"`, `message="event '<EventType>' processed successfully \| session=<SessionID> \| currentState=<NewCurrentState> \| sessionStatus=<SessionStatus>"` |

### Error Cases

**Case 2: Session Not Ready**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="session not ready: status is '<actual-status>'"` |

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

**Case 6: Failed to Record Event**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="failed to record event: <error-details>"` |

**Case 7: No Matching Transition**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="no transition found for event '<EventType>' from node '<CurrentState>'"` |

**Case 8: Transition Execution Failed**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="transition failed: <error-details>"` |

### Function Signature (Go-like pseudocode)

```
func (ep *EventProcessor) ProcessEvent(
    sessionUUID string,
    runtimeMessage RuntimeMessage,
) RuntimeResponse
```

## Invariants

1. **Claude Session ID Validation Priority**: EventProcessor must validate the `claudeSessionID` from RuntimeMessage against the current node's requirements before recording the event. Validation failures result in immediate rejection without event recording.

2. **Event Recording After Validation**: Events must be written to EventStore (via `Session.UpdateEventHistorySafe()`) only after Claude session ID validation succeeds. This ensures only validated events are recorded in the audit trail.

3. **Session Status Validation**: EventProcessor must only accept events when `Session.Status == "running"`. Other statuses are rejected without event recording.

4. **Event Immutability**: Once an Event is created and written to EventStore via `Session.UpdateEventHistorySafe()`, EventProcessor must not modify it.

5. **EmittedBy Auto-Assignment**: EventProcessor must automatically set `Event.EmittedBy` to `Session.CurrentState` at the time of event creation. The caller must not provide this field.

6. **TransitionToNode Internal Error Handling**: The `TransitionToNode.Transition` method uses fail-fast semantics with internal error handling. If any step fails during transition (node lookup, agent definition loading, agent invocation, Session.Done), TransitionToNode immediately constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)` internally, and returns an error to EventProcessor. EventProcessor must NOT call Session.Fail again when TransitionToNode returns an error.

7. **Session Modification via Methods**: EventProcessor must not directly modify the Session structure's internal fields. All state updates (event history, session data) must be performed via Session's thread-safe methods (`UpdateEventHistorySafe()`, `UpdateSessionDataSafe()`, `GetSessionDataSafe()`). Session status transitions are delegated to TransitionToNode which calls Session methods (`UpdateCurrentStateSafe()`, `Done()`, `Fail()`).

8. **No Panic Handling**: EventProcessor does not implement panic recovery. Panic handling is the responsibility of MessageRouter.

9. **Single Responsibility**: EventProcessor focuses solely on event processing logic. It does not manage session lifecycle, socket cleanup, or human notifications.

10. **Node Type Determination**: EventProcessor must determine the current node's type by loading the WorkflowDefinition and looking up the node definition using `Session.CurrentState`. It must not infer node type from RuntimeMessage fields.

11. **Event Recording Failure Triggers Session Failure**: If `Session.UpdateEventHistorySafe()` returns an error, EventProcessor must construct a RuntimeError and call `Session.Fail()` to transition the session to failed. Event recording failure is considered a critical error that halts the session.

## Edge Cases

- **Condition**: Session reference is nil or invalid during EventProcessor initialization.
  **Expected**: This is a programming error. EventProcessor initialization should fail or panic during startup, not during message processing.

- **Condition**: `Session.Status == "initializing"`.
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="session not ready: status is 'initializing'"`. No event is recorded.

- **Condition**: `Session.Status == "completed"`.
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="session not ready: status is 'completed'"`. No event is recorded.

- **Condition**: `Session.Status == "failed"`.
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="session not ready: status is 'failed'"`. No event is recorded.

- **Condition**: Current node is an agent node, and the stored `<NodeName>.ClaudeSessionID` does not exist in SessionData.
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID not found for node '<NodeName>'"`. No event is recorded.

- **Condition**: Current node is an agent node, and `RuntimeMessage.claudeSessionID` does not match the stored value in SessionData.
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="claude session ID mismatch: expected <stored-uuid> but got <provided-uuid>"`. No event is recorded.

- **Condition**: Current node is a human node, and `RuntimeMessage.claudeSessionID` is not an empty string.
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="invalid claude session ID for human node: must be empty"`. No event is recorded.

- **Condition**: Current node is a human node, and `RuntimeMessage.claudeSessionID` is an empty string.
  **Expected**: Validation succeeds. EventProcessor proceeds to record the event and evaluate transitions.

- **Condition**: Current node is an agent node, and `RuntimeMessage.claudeSessionID` matches the stored value in SessionData.
  **Expected**: Validation succeeds. EventProcessor proceeds to record the event and evaluate transitions.

- **Condition**: RuntimeMessage payload is missing `eventType` field.
  **Expected**: This case is prevented by RuntimeSocketManager validation. If it occurs, EventProcessor returns a RuntimeResponse with `status="error"` and `message="invalid event payload: missing eventType"`.

- **Condition**: EventStore write fails due to disk full or permission error (via `Session.UpdateEventHistorySafe()`).
  **Expected**: `Session.UpdateEventHistorySafe()` returns an error. EventProcessor constructs a RuntimeError with `Issuer="EventProcessor"`, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to failed, and returns a RuntimeResponse with `status="error"` and `message="failed to record event: <error-details>"`. The transition is not attempted.

- **Condition**: WorkflowDefinition cannot be loaded (file missing or parse error).
  **Expected**: EventProcessor returns a RuntimeResponse with `status="error"` and `message="failed to load workflow definition: <error-details>"`.

- **Condition**: TransitionEvaluator returns `(nil, false, nil)` (no matching transition).
  **Expected**: The event has already been recorded. EventProcessor returns a RuntimeResponse with `status="error"` and `message="no transition found for event '<EventType>' from node '<CurrentState>'"`. **The session remains in `running` status**; the caller may retry with a different event. EventProcessor does not call Session.Fail.

- **Condition**: `TransitionToNode.Transition` fails due to target node not found, agent definition load error, or agent invocation error.
  **Expected**: `TransitionToNode` internally constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status, and returns an error. EventProcessor does NOT call Session.Fail again. EventProcessor returns a RuntimeResponse with `status="error"` and `message="transition failed: <error-details>"`.

- **Condition**: `TransitionToNode.Transition` completes node-specific action but Session.Done fails (for exit transitions).
  **Expected**: `TransitionToNode` internally constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status, and returns an error. EventProcessor does NOT call Session.Fail again. EventProcessor returns a RuntimeResponse with `status="error"` and `message="transition failed: <error-details>"`. Some intermediate steps (e.g., agent process started, CurrentState updated) have partially completed and are not rolled back.

- **Condition**: TransitionEvaluator returns an exit transition (`isExitTransition == true`).
  **Expected**: EventProcessor passes this flag to `TransitionToNode.Transition`, which **skips the target node's dispatch action** (no stdout print for human nodes, no AgentInvoker call for agent nodes), updates `Session.CurrentState` to the target node, and calls `Session.Done(terminationNotifier)` to transition `Session.Status` to "completed" and notify the main loop. If `Session.Done()` fails (e.g., status is not "running"), `TransitionToNode` constructs a RuntimeError and calls `Session.Fail()`. The success message reflects the new status: `"sessionStatus=completed"`.

- **Condition**: Event message or payload contains very large data (e.g., 5 MB).
  **Expected**: EventProcessor processes the event normally. EventStore serializes the large data to JSONL. Performance may degrade, but no error occurs unless EventStore encounters filesystem limits.

- **Condition**: Multiple events are emitted simultaneously from different agents via concurrent socket connections.
  **Expected**: Each EventProcessor invocation runs in a separate goroutine (handled by RuntimeSocketManager). `Session.UpdateEventHistorySafe()` serializes writes to EventHistory with the session-level write lock. Session's other methods also use the session-level lock for thread safety. The last successful transition wins for `CurrentState`.

## Related

- [RuntimeMessage](../entities/runtime_message.md) - Input message format
- [RuntimeResponse](../entities/runtime_response.md) - Output response format
- [Event](../entities/event.md) - Event entity structure
- [Session](../entities/session/session.md) - Session state and lifecycle
- [RuntimeError](../entities/runtime_error.md) - RuntimeError constructed by EventProcessor (event recording failure) and by TransitionToNode (transition failures)
- [EventStore](../storage/event_store.md) - Persistent event storage
- [SessionMetadataStore](../storage/session_metadata_store.md) - Session metadata persistence
- [TransitionEvaluator](./transition_evaluator.md) - Transition matching logic
- [TransitionToNode](./transition_to_node.md) - Executes node dispatch logic and state transitions, handles errors internally
- [WorkflowDefinition](../components/workflow_definition.md) - Workflow structure
- [MessageRouter](./message_router.md) - Dispatches messages to EventProcessor
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
