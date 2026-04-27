# MessageRouter

## Overview

MessageRouter is the concrete implementation of the MessageHandler callback interface used by RuntimeSocketManager.Listen() for processing received RuntimeMessages. MessageHandler is a function signature that defines the contract for message processing: `func(sessionUUID string, message RuntimeMessage) RuntimeResponse`. MessageRouter implements this interface by parsing the message type, dispatching to EventProcessor or ErrorProcessor based on the type, handling unknown message types, and returning a RuntimeResponse. MessageRouter implements panic recovery to ensure that failures in message processing do not crash the Runtime process. When a panic is recovered, MessageRouter triggers a RuntimeError to transition the session to "failed" status. MessageRouter does not validate message structure (handled by RuntimeSocketManager) or manage session state directly (handled by processors).

## Behavior

1. MessageRouter implements the MessageHandler interface, which is a function signature: `func(sessionUUID string, message RuntimeMessage) RuntimeResponse`. This interface defines the contract between RuntimeSocketManager and message processing logic.
2. MessageRouter is registered as the MessageHandler callback when RuntimeSocketManager.Listen() is invoked during session initialization.
3. RuntimeSocketManager invokes MessageRouter for each incoming RuntimeMessage after validating the JSON structure and required fields.
4. MessageRouter receives a RuntimeMessage struct containing `type` and `payload` fields, along with the sessionUUID extracted from the socket path.
5. MessageRouter wraps the entire processing logic in a panic recovery block using `defer` and `recover()`.
6. If a panic occurs during message processing, MessageRouter:
   - Logs the error with full stack trace (including panic message and goroutine information).
   - Constructs a RuntimeError with `Issuer="MessageRouter"`, `Message="panic during message processing"`, `Detail` containing the panic message and stack trace, `SessionID` set to the session's ID, `FailingState` set to the session's current `CurrentState`, and `OccurredAt` set to the current POSIX timestamp.
   - Calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to `"failed"` status in memory first, then attempt to persist, and notify the main loop.
   - Returns a RuntimeResponse with `status="error"` and `message="internal server error"`.
7. MessageRouter examines the `RuntimeMessage.Type` field.
8. If `type == "event"`, MessageRouter invokes `EventProcessor.ProcessEvent()` with the session UUID and RuntimeMessage, and returns the resulting RuntimeResponse.
9. If `type == "error"`, MessageRouter invokes `ErrorProcessor.ProcessError()` with the session UUID and RuntimeMessage, and returns the resulting RuntimeResponse.
10. If `type` is neither "event" nor "error", MessageRouter returns a RuntimeResponse with `status="error"` and `message="unknown message type '<type>'"`.
11. MessageRouter does not log messages or errors (except for panic recovery). Logging is the responsibility of individual processors or RuntimeSocketManager.
12. MessageRouter does not modify the RuntimeMessage or RuntimeResponse structures (except when handling panics). It acts as a pure dispatcher for normal message processing.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Session | *Session | Reference to the Session entity shared across all runtime components | Yes |
| EventProcessor | *EventProcessor | Initialized EventProcessor instance for the session | Yes |
| ErrorProcessor | *ErrorProcessor | Initialized ErrorProcessor instance for the session | Yes |
| TerminationNotifier | chan<- struct{} | Channel for notifying the main loop of session termination (used during panic recovery) | Yes |

### For RouteMessage Operation (Callback Invocation)

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string (UUID) | Valid UUID extracted from the socket path by RuntimeSocketManager | Yes |
| RuntimeMessage | RuntimeMessage | Valid RuntimeMessage struct with `type` and `payload` fields validated by RuntimeSocketManager | Yes |

## Outputs

### Success Cases

**Case 1: Event Message Routed Successfully**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | RuntimeResponse returned by EventProcessor.ProcessEvent(), with `status="success"` or `status="error"` depending on processing outcome |

**Case 2: Error Message Routed Successfully**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | RuntimeResponse returned by ErrorProcessor.ProcessError(), with `status="success"` or `status="error"` depending on processing outcome |

### Error Cases

**Case 3: Unknown Message Type**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="unknown message type '<type>'"` |

**Case 4: Panic Recovered**

| Field | Type | Description |
|-------|------|-------------|
| RuntimeResponse | RuntimeResponse | `status="error"`, `message="internal server error"` (logged with full stack trace) |

### Function Signature (Go-like pseudocode)

```
func (mr *MessageRouter) RouteMessage(
    sessionUUID string,
    runtimeMessage RuntimeMessage,
) RuntimeResponse
```

## Invariants

1. **Panic Recovery with RuntimeError**: MessageRouter must recover from all panics that occur during message processing. When a panic is recovered, MessageRouter must: (1) log the error with full stack trace, (2) construct a RuntimeError with `Issuer="MessageRouter"`, (3) call `Session.Fail(runtimeError, terminationNotifier)` to transition the session to `"failed"` status in memory first, then attempt persistence, and notify the main loop, and (4) return a RuntimeResponse with `status="error"` and `message="internal server error"`. Panics must not crash the Runtime process.

2. **Type-Based Dispatch**: MessageRouter must dispatch messages to processors based solely on the `RuntimeMessage.Type` field. It must not inspect the payload structure.

3. **No Message Modification**: MessageRouter must not modify the RuntimeMessage struct before passing it to processors. It acts as a transparent dispatcher.

4. **No Response Modification**: MessageRouter must return the RuntimeResponse from processors without modification (except in panic recovery or unknown type cases).

5. **Message Processing Independence**: Each message is processed independently. MessageRouter holds references to Session, EventProcessor, ErrorProcessor, and TerminationNotifier, which are initialized once and reused across invocations. The session UUID is provided by RuntimeSocketManager on each invocation for validation purposes.

6. **No Logging Except Panics**: MessageRouter does not log routine messages or errors. Only panic recovery logs are emitted. Individual processors are responsible for their own logging if needed.

7. **Single Callback per Connection**: MessageRouter is invoked once per RuntimeSocketManager connection. Each connection is handled in a separate goroutine by RuntimeSocketManager.

8. **MessageHandler Interface Implementation**: MessageRouter is a concrete implementation of the MessageHandler interface. MessageHandler is defined as a function signature: `func(sessionUUID string, message RuntimeMessage) RuntimeResponse`. RuntimeSocketManager depends on this interface, and MessageRouter fulfills the contract.

## Edge Cases

- **Condition**: RuntimeMessage has `type="event"`.
  **Expected**: MessageRouter invokes `EventProcessor.ProcessEvent()` and returns the resulting RuntimeResponse. The response may have `status="success"` or `status="error"` depending on EventProcessor's outcome.

- **Condition**: RuntimeMessage has `type="error"`.
  **Expected**: MessageRouter invokes `ErrorProcessor.ProcessError()` and returns the resulting RuntimeResponse. The response may have `status="success"` or `status="error"` depending on ErrorProcessor's outcome.

- **Condition**: RuntimeMessage has `type="unknown"` (not "event" or "error").
  **Expected**: MessageRouter returns a RuntimeResponse with `status="error"` and `message="unknown message type 'unknown'"`.

- **Condition**: RuntimeMessage has `type=""` (empty string).
  **Expected**: MessageRouter returns a RuntimeResponse with `status="error"` and `message="unknown message type ''"`.

- **Condition**: EventProcessor panics during `ProcessEvent()` execution.
  **Expected**: MessageRouter's panic recovery catches the panic, logs the error with full stack trace (including goroutine ID), constructs a RuntimeError with `Issuer="MessageRouter"` and details about the panic, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to `"failed"` status in memory first, attempt to persist, and notify the main loop, and returns a RuntimeResponse with `status="error"` and `message="internal server error"`.

- **Condition**: ErrorProcessor panics during `ProcessError()` execution.
  **Expected**: MessageRouter's panic recovery catches the panic, logs the error with full stack trace, constructs a RuntimeError with `Issuer="MessageRouter"` and details about the panic, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to `"failed"` status in memory first, attempt to persist, and notify the main loop, and returns a RuntimeResponse with `status="error"` and `message="internal server error"`.

- **Condition**: EventProcessor returns a RuntimeResponse with `status="error"` (e.g., session not ready).
  **Expected**: MessageRouter returns the RuntimeResponse as-is without modification. The error is communicated to the client via the response.

- **Condition**: ErrorProcessor returns a RuntimeResponse with `status="error"` (e.g., session terminated).
  **Expected**: MessageRouter returns the RuntimeResponse as-is without modification. The error is communicated to the client via the response.

- **Condition**: RuntimeMessage is malformed (missing `type` or `payload` fields).
  **Expected**: This case is prevented by RuntimeSocketManager validation. RuntimeSocketManager rejects such messages before invoking MessageRouter. If this case occurs due to a bug, MessageRouter may panic, which is caught by panic recovery.

- **Condition**: SessionUUID provided by RuntimeSocketManager is invalid or does not reference an existing session.
  **Expected**: EventProcessor and ErrorProcessor will fail when attempting to load session metadata. They return error RuntimeResponses, which MessageRouter returns to the client.

- **Condition**: Multiple RuntimeMessages are received on the same socket connection.
  **Expected**: This case is prevented by RuntimeSocketManager. Each socket connection processes only one message, then closes. MessageRouter is invoked once per connection.

- **Condition**: RuntimeMessage processing takes a very long time (e.g., 10 minutes due to slow disk I/O).
  **Expected**: MessageRouter waits for the processor to complete. RuntimeSocketManager keeps the connection open. No timeout is enforced by MessageRouter. Connection timeout, if needed, is handled by RuntimeSocketManager or the underlying transport layer.

- **Condition**: EventProcessor or ErrorProcessor return a malformed RuntimeResponse (e.g., `status` is neither "success" nor "error").
  **Expected**: MessageRouter returns the malformed response as-is. RuntimeSocketManager validates the response before sending to the client and may replace it with an error response if validation fails (as per RuntimeResponse specification).

- **Condition**: Panic recovery log write fails (e.g., stderr is unavailable).
  **Expected**: The panic is recovered, but the log message is lost. MessageRouter still triggers the RuntimeError, transitions the session to `"failed"` status in memory, attempts persistence, and returns a RuntimeResponse with `status="error"` and `message="internal server error"`. The Runtime process does not crash.

- **Condition**: SessionMetadataStore write fails when persisting the RuntimeError during panic recovery via `Session.Fail()`.
  **Expected**: `Session.Fail()` logs a warning about the persistence failure but returns `nil` (best-effort persistence). The session remains in `"failed"` status in memory. The main loop is notified via terminationNotifier. MessageRouter returns a RuntimeResponse with `status="error"` and `message="internal server error"`.

## Related

- [RuntimeMessage](../entities/runtime_message.md) - Input message format
- [RuntimeResponse](../entities/runtime_response.md) - Output response format
- [EventProcessor](./event_processor.md) - Handles event messages
- [ErrorProcessor](./error_processor.md) - Handles error messages
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Invokes MessageRouter as the MessageHandler callback implementation
- [RuntimeError](../entities/runtime_error.md) - RuntimeError triggered by MessageRouter during panic recovery
- [Session](../entities/session/session.md) - Session status transitioned to "failed" when RuntimeError is triggered
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
