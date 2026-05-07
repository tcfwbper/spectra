# MessageRouter

## Overview

MessageRouter is the concrete implementation of the MessageHandler interface used by RuntimeSocketManager. It receives RuntimeMessages, dispatches them to EventProcessor or ErrorProcessor based on message type, and implements panic recovery to prevent failures from crashing the runtime process. When a panic is recovered, MessageRouter constructs a RuntimeError, calls Session.Fail to transition the session to "failed" status, and returns an error response. MessageRouter does not validate message structure (handled by RuntimeSocketManager) or manage session state directly (handled by processors).

## Boundaries

- Owns: message type dispatch (routing to EventProcessor or ErrorProcessor).
- Owns: panic recovery for all message processing logic.
- Owns: RuntimeError construction during panic recovery.
- Owns: Session.Fail invocation during panic recovery.
- Delegates: message structure validation to RuntimeSocketManager (upstream).
- Delegates: event processing logic to EventProcessor.
- Delegates: error processing logic to ErrorProcessor.
- Delegates: session lifecycle management (except panic-triggered Fail) to processors.
- Must not: inspect or parse RuntimeMessage payload structure.
- Must not: modify RuntimeMessage before passing to processors.
- Must not: modify RuntimeResponse returned by processors.
- Must not: log routine messages or errors (only panic recovery logs).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State container with auto-persist | Read `ID`, `GetCurrentStateSafe()`; call `Fail()` during panic recovery only | Must not call `Run()`, `Done()`, or any mutating method except `Fail` during panic |
| `EventProcessor` | Event message handler | `ProcessEvent(sessionUUID, runtimeMessage)` | Must not access internal state |
| `ErrorProcessor` | Error message handler | `ProcessError(sessionUUID, runtimeMessage)` | Must not access internal state |
| `TerminationNotifier` | Main loop notification | Pass to `PersistentSession.Fail()` during panic recovery | Must not send directly |
| `RuntimeError` | Error entity | Construct via `NewRuntimeError` during panic recovery | Must not construct outside panic recovery |
| `RuntimeResponse` | Response entity | Construct via `ErrorResponse` during panic recovery or unknown type | Must not construct for normal dispatch |
| `logger.Logger` | Structured logging | `Error(msg string, args ...any)` during panic recovery only | Must not log routine operations |

Construction constraint: Must be constructed with PersistentSession, EventProcessor, ErrorProcessor, TerminationNotifier, and Logger. All dependencies are injected at initialization and reused across invocations.

## Behavior

1. MessageRouter implements the `MessageHandler` interface: `Handle(sessionUUID string, msg RuntimeMessage) RuntimeResponse`.
2. MessageRouter is registered as the MessageHandler when RuntimeSocketManager.Listen() is invoked.
3. MessageRouter wraps the entire dispatch logic in a panic recovery block using `defer` and `recover()`.
4. If a panic occurs during message processing, MessageRouter:
   - Logs the error with full stack trace via Logger.Error.
   - Constructs a RuntimeError with `Issuer="MessageRouter"`, `Message="panic during message processing"`, `Detail` containing the panic value and stack trace as a JSON object, `SessionID` set to `PersistentSession.ID`, `FailingState` set to `PersistentSession.GetCurrentStateSafe()`, and `OccurredAt` set to the current POSIX timestamp.
   - Calls `PersistentSession.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status and notify the main loop. PersistentSession automatically persists the failed state.
   - Returns `ErrorResponse("internal server error")`.
5. If no panic occurs, MessageRouter examines `RuntimeMessage.Type()`.
6. If `Type() == "event"`, invokes `EventProcessor.ProcessEvent(sessionUUID, runtimeMessage)` and returns the resulting RuntimeResponse.
7. If `Type() == "error"`, invokes `ErrorProcessor.ProcessError(sessionUUID, runtimeMessage)` and returns the resulting RuntimeResponse.
8. If `Type()` is neither "event" nor "error", returns `ErrorResponse("unknown message type '<type>'")`.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | PersistentSession reference | Valid, constructed via NewPersistentSession | Yes |
| EventProcessor | EventProcessor reference | Initialized EventProcessor instance | Yes |
| ErrorProcessor | ErrorProcessor reference | Initialized ErrorProcessor instance | Yes |
| TerminationNotifier | chan<- struct{} | Non-nil, buffered, capacity >= 2 | Yes |
| Logger | logger.Logger | Non-nil Logger interface implementation | Yes |

### For Handle Operation (MessageHandler Interface)

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string | Valid UUID extracted from socket path by RuntimeSocketManager | Yes |
| RuntimeMessage | RuntimeMessage | Valid RuntimeMessage constructed by RuntimeSocketManager after protocol validation | Yes |

## Outputs

### Success Cases

**Case 1: Event Message Dispatched**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | Response returned by EventProcessor.ProcessEvent() (may be success or error) |

**Case 2: Error Message Dispatched**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | Response returned by ErrorProcessor.ProcessError() (may be success or error) |

### Error Cases

**Case 3: Unknown Message Type**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `ErrorResponse("unknown message type '<type>'")` |

**Case 4: Panic Recovered**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `ErrorResponse("internal server error")` |

## Invariants

1. **Panic Recovery Guarantee**: MessageRouter must recover from all panics that occur during message dispatch. Panics must never propagate to RuntimeSocketManager or crash the runtime process.

2. **RuntimeError on Panic**: When a panic is recovered, MessageRouter must construct a valid RuntimeError (via NewRuntimeError) and call PersistentSession.Fail to transition the session to "failed" status.

3. **Type-Based Dispatch Only**: MessageRouter dispatches messages based solely on `RuntimeMessage.Type()`. It must not inspect payload content.

4. **Transparent Dispatch**: MessageRouter must not modify the RuntimeMessage before passing to processors, and must not modify the RuntimeResponse returned by processors.

5. **No Logging Except Panics**: MessageRouter does not log routine operations. Only panic recovery uses Logger.Error.

6. **Single Responsibility**: MessageRouter acts as a pure dispatcher with panic safety. Business logic is entirely within processors.

7. **MessageHandler Interface Compliance**: MessageRouter conforms to the `MessageHandler` interface defined in the storage package: `Handle(sessionUUID string, msg RuntimeMessage) RuntimeResponse`.

8. **PersistentSession.Fail Only During Panic**: MessageRouter must not call PersistentSession.Fail during normal message dispatch. Only panic recovery triggers PersistentSession.Fail.

## Edge Cases

- Condition: `RuntimeMessage.Type()` is "event".
  Expected: Invokes EventProcessor.ProcessEvent() and returns the result unchanged.

- Condition: `RuntimeMessage.Type()` is "error".
  Expected: Invokes ErrorProcessor.ProcessError() and returns the result unchanged.

- Condition: `RuntimeMessage.Type()` is an unrecognized value.
  Expected: Returns `ErrorResponse("unknown message type '<type>'")`. Note: RuntimeSocketManager validates type at protocol level, so this case should not occur in practice, but MessageRouter handles it defensively.

- Condition: EventProcessor panics during ProcessEvent().
  Expected: Panic recovered. Logger.Error called with stack trace. RuntimeError constructed with Issuer="MessageRouter". PersistentSession.Fail called. Returns `ErrorResponse("internal server error")`.

- Condition: ErrorProcessor panics during ProcessError().
  Expected: Same as above — panic recovered, session failed, error response returned.

- Condition: PersistentSession.Fail returns error during panic recovery (e.g., session already failed).
  Expected: MessageRouter logs the Fail error but still returns `ErrorResponse("internal server error")`. The session is already in a terminal state, so the notification is best-effort.

- Condition: Logger.Error fails during panic recovery (e.g., writer unavailable).
  Expected: Log is lost. MessageRouter still constructs RuntimeError, calls Session.Fail, and returns error response. Process does not crash.

- Condition: RuntimeError construction fails during panic recovery (e.g., invalid parameters).
  Expected: MessageRouter logs the construction error, does not call PersistentSession.Fail (no valid RuntimeError to pass), and returns `ErrorResponse("internal server error")`. The session may not transition to "failed" — this is an edge case of a double failure.

- Condition: EventProcessor returns a RuntimeResponse with `status="error"`.
  Expected: MessageRouter returns it unchanged. No additional action.

- Condition: Multiple messages processed concurrently (separate goroutines via RuntimeSocketManager).
  Expected: Each invocation is independent. PersistentSession access is serialized by Session's internal lock. Panic recovery is per-invocation.

## Related

- [PersistentSession](./persistent_session.md) — State container with automatic persistence
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) — invokes MessageRouter as the MessageHandler implementation
- [EventProcessor](./event_processor.md) — handles event-type messages
- [ErrorProcessor](./error_processor.md) — handles error-type messages
- [RuntimeMessage](../entities/runtime_message.md) — input message entity
- [RuntimeResponse](../entities/runtime_response.md) — output response entity
- [RuntimeError](../entities/runtime_error.md) — constructed during panic recovery
- [Session lifecycle](../entities/session/lifecycle.md) — Session.Fail method (wrapped by PersistentSession)
- [Logger](../logger/logger.md) — structured logging for panic recovery
