# EventProcessor

## Overview

EventProcessor handles RuntimeMessage with `type="event"`. It validates session status, validates the Claude session ID, constructs an Event entity, records it to the session's event history (via PersistentSession, which automatically persists), evaluates the transition, invokes TransitionToNode for dispatch, and manages exit transition completion. EventProcessor is responsible for constructing RuntimeError and calling PersistentSession.Fail when TransitionToNode returns an error or when event recording fails. EventProcessor does not manage socket lifecycle, panic recovery, or direct persistence.

## Boundaries

- Owns: session status validation for event processing (must be "running").
- Owns: Event entity construction (UUID generation, field assembly, EmittedBy auto-assignment).
- Owns: event recording via PersistentSession.UpdateEventHistorySafe (which auto-persists to EventStore and SessionMetadataStore).
- Owns: transition evaluation invocation (calling TransitionEvaluator).
- Owns: TransitionToNode invocation and error handling (constructing RuntimeError, calling PersistentSession.Fail on TransitionToNode failure).
- Owns: exit transition handling (calling PersistentSession.Done after successful TransitionToNode dispatch).
- Owns: RuntimeError construction for event recording failure and transition failure.
- Delegates: Claude session ID validation to ValidateClaudeSessionID helper.
- Delegates: node dispatch logic (stdout print, agent invocation, state update) to TransitionToNode.
- Delegates: transition lookup to TransitionEvaluator (stateless function).
- Delegates: persistence to PersistentSession (automatic, non-fatal).
- Delegates: panic recovery to MessageRouter.
- Must not: manage socket lifecycle or connection handling.
- Must not: implement panic recovery.
- Must not: directly modify Session's internal fields (use thread-safe methods only).
- Must not: call PersistentSession.Run.
- Must not: call SessionMetadataStore.Write() or EventStore.Append() directly.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State container with auto-persist | `GetStatusSafe()`, `GetCurrentStateSafe()`, `UpdateEventHistorySafe(event)`, `Fail(err, notifier)`, `Done(notifier)`, read `ID` | Must not call `Run()`, must not modify fields directly, must not call stores directly |
| `WorkflowDefinition` | Configuration source (initialized once) | Read `Nodes()` to find node by name | Must not modify |
| `TransitionEvaluator` | Stateless function | `EvaluateTransition(workflowDef, currentState, eventType)` | — |
| `TransitionToNode` | Node dispatch | `Transition(targetNodeName, message)` | Must not access internal state |
| `ValidateClaudeSessionID` | Shared validation helper | Call with PersistentSession, Node, claudeSessionID | — |
| `TerminationNotifier` | Main loop notification | Pass to `PersistentSession.Fail()` and `PersistentSession.Done()` | Must not send directly |
| `RuntimeError` | Error entity | Construct via `NewRuntimeError` for failure cases | — |
| `RuntimeResponse` | Response entity | Construct via `SuccessResponse` or `ErrorResponse` | — |
| `Event` | Event entity | Construct via `NewEvent` | — |

Construction constraint: EventProcessor is initialized with PersistentSession, WorkflowDefinition (loaded once at initialization), TransitionToNode, and TerminationNotifier. TransitionEvaluator and ValidateClaudeSessionID are stateless package-level functions called directly.

## Behavior

1. EventProcessor is invoked by MessageRouter when a RuntimeMessage with `type="event"` is received.
2. Receives the sessionUUID and RuntimeMessage containing `ClaudeSessionID()` and payload `{eventType, message, payload}`.
3. Validates that `PersistentSession.GetStatusSafe() == "running"`. If not, returns `ErrorResponse("session not ready: status is '<actual-status>'")`.
4. Retrieves the current node name via `PersistentSession.GetCurrentStateSafe()`.
5. Retrieves the current node definition from WorkflowDefinition.Nodes() by matching node name.
6. Calls `ValidateClaudeSessionID(PersistentSession, currentNode, RuntimeMessage.ClaudeSessionID())`. If validation returns an error, returns `ErrorResponse(<validation error message>)`. No event is recorded.
7. Parses the event payload from RuntimeMessage.Payload() to extract `eventType`, `message`, and `payload` fields.
8. Constructs an Event entity with: a generated UUID, the `eventType` from the payload, the `message` and `payload` from the RuntimeMessage payload, `EmittedBy` set to the current node name (PersistentSession.GetCurrentStateSafe()), `EmittedAt` set to the current POSIX timestamp, and `SessionID` set to the sessionUUID.
9. Calls `PersistentSession.UpdateEventHistorySafe(event)` to append the event to the in-memory EventHistory. PersistentSession automatically persists the event to EventStore and updates metadata (non-fatal if persistence fails).
10. If `PersistentSession.UpdateEventHistorySafe(event)` returns an error (in-memory validation failure), constructs a RuntimeError with `Issuer="EventProcessor"`, `Message="failed to record event"`, `Detail` containing the error, `SessionID`, `FailingState` set to the current node name, and `OccurredAt` set to current POSIX timestamp. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Returns `ErrorResponse("failed to record event: <error-details>")`.
11. Invokes `TransitionEvaluator.EvaluateTransition(WorkflowDefinition, currentNodeName, event.Type())`.
12. If TransitionEvaluator returns `(nil, false)` (no matching transition), returns `ErrorResponse("no transition found for event '<eventType>' from node '<currentState>'")`. The event has been recorded. The session remains in "running" status.
13. If TransitionEvaluator returns a valid transition `(transition, isExitTransition)`, invokes `TransitionToNode.Transition(TargetNodeName=transition.ToNode(), Message=event.Message())`.
14. If TransitionToNode returns an error, constructs a RuntimeError with `Issuer="EventProcessor"`, `Message="transition failed"`, `Detail` containing the TransitionToNode error, `SessionID`, `FailingState` set to the current node name, and `OccurredAt` set to current POSIX timestamp. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Returns `ErrorResponse("transition failed: <error-details>")`.
15. If TransitionToNode returns nil (success) and `isExitTransition == false`, returns `SuccessResponse("event '<eventType>' processed successfully | session=<SessionID> | currentState=<NewCurrentState> | sessionStatus=running")`.
16. If TransitionToNode returns nil (success) and `isExitTransition == true`, calls `PersistentSession.Done(terminationNotifier)` to transition the session to "completed" status and notify the main loop. PersistentSession automatically persists the final state.
17. If `PersistentSession.Done()` returns an error, constructs a RuntimeError with `Issuer="EventProcessor"`, `Message="failed to complete session"`, `Detail` containing the Done error, `SessionID`, `FailingState` set to transition.ToNode(), and `OccurredAt` set to current POSIX timestamp. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Returns `ErrorResponse("failed to complete session: <error-details>")`.
18. If `PersistentSession.Done()` returns nil, returns `SuccessResponse("event '<eventType>' processed successfully | session=<SessionID> | currentState=<NewCurrentState> | sessionStatus=completed")`.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | PersistentSession reference | Valid, constructed via NewPersistentSession | Yes |
| WorkflowDefinition | WorkflowDefinition reference | Valid, fully validated, loaded once at initialization | Yes |
| TransitionToNode | TransitionToNode reference | Initialized with PersistentSession, WorkflowDefinition, AgentDefinitionLoader, AgentInvoker | Yes |
| TerminationNotifier | chan<- struct{} | Non-nil, buffered, capacity >= 2 | Yes |

### For ProcessEvent Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string (UUID) | Valid UUID, references an existing session | Yes |
| RuntimeMessage | RuntimeMessage | Valid RuntimeMessage with `Type()="event"` | Yes |

## Outputs

### Success Cases

**Case 1: Event Processed, Regular Transition**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `SuccessResponse("event '<eventType>' processed successfully \| session=<SessionID> \| currentState=<NewState> \| sessionStatus=running")` |

**Case 2: Event Processed, Exit Transition**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `SuccessResponse("event '<eventType>' processed successfully \| session=<SessionID> \| currentState=<NewState> \| sessionStatus=completed")` |

### Error Cases

| Case | Response |
|------|----------|
| Session not running | `ErrorResponse("session not ready: status is '<status>'")` |
| Claude session ID validation failed | `ErrorResponse(<validation error message>)` |
| Event payload parsing failed | `ErrorResponse("invalid event payload: <details>")` |
| Event recording failed | `ErrorResponse("failed to record event: <details>")` |
| No matching transition | `ErrorResponse("no transition found for event '<eventType>' from node '<currentState>'")` |
| TransitionToNode failed | `ErrorResponse("transition failed: <details>")` |
| Session.Done failed | `ErrorResponse("failed to complete session: <details>")` |

## Invariants

1. **Session Status Validation**: EventProcessor must only accept events when `PersistentSession.GetStatusSafe() == "running"`. All other statuses are rejected without event recording.

2. **Claude Session ID Validation Before Recording**: Claude session ID validation (via ValidateClaudeSessionID) must occur before event recording. Validation failures result in immediate rejection without event recording.

3. **Event Recording After Validation**: Events are written to the session (via PersistentSession.UpdateEventHistorySafe) only after both status validation and Claude session ID validation succeed.

4. **EmittedBy Auto-Assignment**: EventProcessor must set `Event.EmittedBy` to `PersistentSession.GetCurrentStateSafe()` at event creation time. The caller does not provide this field.

5. **Event Immutability**: Once an Event is constructed and recorded, EventProcessor must not modify it.

6. **TransitionToNode Error Ownership**: When TransitionToNode returns an error, EventProcessor is responsible for constructing a RuntimeError and calling PersistentSession.Fail. TransitionToNode does not perform lifecycle transitions.

7. **Exit Transition Completion**: For exit transitions, EventProcessor calls PersistentSession.Done after TransitionToNode succeeds. TransitionToNode does not know about exit transitions.

8. **No Panic Handling**: EventProcessor does not implement panic recovery. That is MessageRouter's responsibility.

9. **No PersistentSession.Run**: EventProcessor must never call PersistentSession.Run. That belongs to the runtime initialization layer.

10. **Session Access via Thread-Safe Methods**: All reads and writes must go through PersistentSession's exported methods.

11. **Event Recording Failure Is Fatal**: If PersistentSession.UpdateEventHistorySafe returns an error (in-memory validation failure), EventProcessor constructs a RuntimeError and calls PersistentSession.Fail. This is a critical failure. Note: persistence failures within PersistentSession are non-fatal and handled internally.

12. **No Transition Does Not Fail Session**: When no matching transition is found, the session remains "running". The event is recorded but no state change occurs.

13. **WorkflowDefinition Loaded Once**: WorkflowDefinition is loaded at initialization, not per-invocation. It is immutable during the session lifecycle.

14. **No Direct Store Access**: EventProcessor must not call SessionMetadataStore.Write() or EventStore.Append() directly. All persistence is delegated to PersistentSession.

## Edge Cases

- Condition: `PersistentSession.GetStatusSafe()` returns "initializing".
  Expected: Returns `ErrorResponse("session not ready: status is 'initializing'")`. No event recorded.

- Condition: `PersistentSession.GetStatusSafe()` returns "completed" or "failed".
  Expected: Returns `ErrorResponse("session not ready: status is '<status>'")`. No event recorded.

- Condition: ValidateClaudeSessionID returns an error (ID not found, mismatch, or non-empty for human node).
  Expected: Returns `ErrorResponse(<error message>)`. No event recorded.

- Condition: RuntimeMessage payload is missing `eventType` field.
  Expected: Returns `ErrorResponse("invalid event payload: missing eventType")`.

- Condition: `PersistentSession.UpdateEventHistorySafe()` returns error (in-memory validation failure — e.g., missing required event field).
  Expected: Constructs RuntimeError with Issuer="EventProcessor". Calls PersistentSession.Fail. Returns error response. Transition not attempted.

- Condition: TransitionEvaluator returns `(nil, false)`.
  Expected: Event already recorded. Returns error response. Session remains "running".

- Condition: TransitionToNode returns error (node not found, agent invocation failed, state update failed).
  Expected: Constructs RuntimeError with Issuer="EventProcessor". Calls PersistentSession.Fail. Returns error response.

- Condition: TransitionToNode succeeds but PersistentSession.Done fails (exit transition, status not "running" due to concurrent Fail).
  Expected: Constructs RuntimeError with Issuer="EventProcessor". Calls PersistentSession.Fail (which may also fail if already failed). Returns error response.

- Condition: PersistentSession.Fail returns error when EventProcessor attempts to fail the session (session already failed by concurrent operation).
  Expected: EventProcessor still returns the error RuntimeResponse. The session is already in a terminal state.

- Condition: Multiple events processed concurrently from different agents.
  Expected: Each ProcessEvent invocation runs in a separate goroutine. PersistentSession methods serialize access via internal Session lock. TransitionToNode's UpdateCurrentStateSafe serializes state writes. Last successful transition wins for CurrentState.

- Condition: Event message or payload contains very large data (e.g., 5 MB).
  Expected: Processed normally in memory. Persistence may fail if EventStore's MaxPayloadSize is exceeded, but PersistentSession logs the failure and continues.

- Condition: Current node name from PersistentSession.GetCurrentStateSafe() does not match any node in WorkflowDefinition.
  Expected: Returns `ErrorResponse("current node '<name>' not found in workflow definition")`.

## Related

- [PersistentSession](./persistent_session.md) — State container with automatic persistence
- [MessageRouter](./message_router.md) — dispatches event messages to EventProcessor
- [ValidateClaudeSessionID](./validate_claude_session_id.md) — shared Claude session ID validation helper
- [TransitionEvaluator](./transition_evaluator.md) — stateless transition lookup function
- [TransitionToNode](./transition_to_node.md) — executes node dispatch and state update
- [Event](../entities/event.md) — event entity constructed by EventProcessor
- [RuntimeError](../entities/runtime_error.md) — error entity constructed on failure
- [RuntimeResponse](../entities/runtime_response.md) — response entity returned to caller
- [Session lifecycle](../entities/session/lifecycle.md) — Fail and Done methods (wrapped by PersistentSession)
- [Session event_history](../entities/session/event_history.md) — UpdateEventHistorySafe method (wrapped by PersistentSession)
- [WorkflowDefinition](../components/workflow_definition.md) — workflow structure (loaded once)
