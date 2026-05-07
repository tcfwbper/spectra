# RuntimeSocketManager

## Overview

RuntimeSocketManager manages the lifecycle of a Unix domain socket for a single session. It creates, listens on, and deletes the runtime socket file (`runtime.sock`) located in the session directory. The socket enables request-response communication between spectra-agent (client) and the Workflow Runtime (server). RuntimeSocketManager provides methods to create the socket, listen for incoming connections, receive and validate JSON-formatted messages (protocol-level only), dispatch them to a MessageHandler, send JSON-formatted responses back to clients, and clean up the socket when the session terminates. It does not manage the session directory, coordinate with other storage components, or validate message semantics.

## Boundaries

- Owns: Unix domain socket file lifecycle (create, bind, listen, accept, close, delete).
- Owns: per-connection goroutine management and connection isolation.
- Owns: protocol-level message validation (well-formed JSON, required top-level fields, recognized type, payload is JSON object, message size limit).
- Owns: response serialization and transmission to clients.
- Owns: socket file permission enforcement (0600).
- Owns: residual socket file detection (existence check before create).
- Delegates: path composition to StorageLayout.
- Delegates: message semantic validation and business logic processing to the MessageHandler.
- Delegates: session directory creation to SessionDirectoryManager (called before RuntimeSocketManager usage).
- Delegates: session state management, persistence, and lifecycle orchestration to the runtime caller.
- Delegates: panic recovery within message processing to the MessageHandler implementation.
- Must not: create or delete the session directory.
- Must not: validate message semantics (event type defined in workflow, session state, agent role derivation).
- Must not: access or modify Session entity state directly.
- Must not: retry socket creation or connection failures.
- Must not: buffer messages in memory.
- Must not: implement Windows named pipe support (Unix domain socket only).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `StorageLayout` | Path composition | `GetRuntimeSocketPath` | Must not bypass for path construction |
| `MessageHandler` | Message processing interface | Invoke `Handle(sessionUUID, msg)` per validated message | Must not access MessageHandler internals or assume concrete type |
| `logger.Logger` | Structured logging | `Warn(msg string, args ...any)` | Must not use Logger for structured return values; must not use stdlib log directly |
| `RuntimeMessage` | Incoming message entity | Construct via `NewRuntimeMessage` after protocol validation | Must not construct via struct literal |
| `RuntimeResponse` | Outgoing response entity (business-logic responses only) | Read fields via getters for serialization of MessageHandler responses | Must not construct RuntimeResponse for protocol-level errors; must not use RuntimeResponse factory functions directly |

Construction constraint: Must be constructed via `NewRuntimeSocketManager(projectRoot, sessionUUID string, logger logger.Logger)`. Direct struct literal is forbidden. Constructor composes the socket path via StorageLayout and stores it internally. No I/O is performed at construction time.

### MessageHandler Interface

Defined within the storage package (interfaces belong in the package that uses them):

```
type MessageHandler interface {
    Handle(sessionUUID string, msg RuntimeMessage) RuntimeResponse
}
```

The MessageHandler implementation is injected at `Listen()` call time, not at construction. RuntimeSocketManager does not own or manage the MessageHandler lifecycle.

### Logger

The Logger interface is defined in the `logger` package (see [Logger](../logger/logger.md)). Injected at construction time. RuntimeSocketManager only uses the `Warn` method, but accepts the full `logger.Logger` interface.

## Behavior

### Construction

1. `NewRuntimeSocketManager(projectRoot, sessionUUID string, logger Logger)` composes the socket path via `StorageLayout.GetRuntimeSocketPath(projectRoot, sessionUUID)` and stores it internally.
2. Stores the `sessionUUID` for passing to the MessageHandler on each connection.
3. Stores the `logger` reference for warning output.
4. No I/O is performed at construction time.

### CreateSocket

5. Checks if the socket file already exists at the target path via `os.Stat`.
6. If the socket file exists, returns an error: `"runtime socket file already exists: <path>. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm <path>"`.
7. Creates a Unix domain socket file at the session-specific path with permissions `0600` (owner read/write only).
8. If socket creation fails (e.g., permission denied, disk full, path too long), returns an error: `"failed to create runtime socket: <underlying error>"`.
9. Does not retry on failure. Errors are propagated to the caller.

### Listen

10. Returns an error synchronously if `CreateSocket()` has not been called: `"runtime socket not created: call CreateSocket() first"`.
11. Accepts a `MessageHandler` argument.
12. Binds to the created socket and starts listening for incoming connections.
13. If the initial bind/listen fails, returns a synchronous error: `"failed to listen on runtime socket: <underlying error>"`.
14. On successful bind/listen, spawns a background goroutine for the accept loop and returns `(listenerErrCh, listenerDoneCh, nil)`.
15. The accept loop runs until `DeleteSocket()` is called.
16. Each accepted connection is handled in a separate goroutine.
17. Per-connection handling: reads one message, validates at protocol level, dispatches to MessageHandler, serializes response, sends response, closes connection.

### Per-Connection Protocol

18. Reads a single JSON message from the connection, terminated by a newline (`\n`).
19. Enforces a maximum message size of 10 MB. If exceeded, logs a warning via Logger: `"dropping connection: message size exceeds 10 MB limit"`, sends a protocol-level error response (raw JSON, see step 26a) to the client if possible, and closes the connection. MessageHandler is not invoked.
20. Parses the JSON message. If malformed (invalid JSON), logs a warning: `"dropping connection: malformed JSON: <parse error>"`, sends a protocol-level error response if possible, and closes the connection.
21. Validates that the top-level field `type` is present and is one of `"event"` or `"error"`. If missing or invalid, logs a warning: `"dropping connection: <reason>"`, sends a protocol-level error response if possible, and closes the connection.
22. Validates that the top-level field `payload` is present and is a JSON object. If missing or not an object, logs a warning, sends a protocol-level error response if possible, and closes.
23. Extracts the optional `claudeSessionID` field (string, defaults to empty string if absent).
24. Constructs a `RuntimeMessage` via `NewRuntimeMessage(type, payload, claudeSessionID)`. If construction fails (should not happen given prior validation), logs a warning, sends a protocol-level error response if possible, and closes.
25. Invokes `MessageHandler.Handle(sessionUUID, runtimeMessage)` which returns a `RuntimeResponse`.
26. For business-logic responses (from MessageHandler): reads RuntimeResponse fields via getters and serializes to JSON format `{"status": "<status>", "message": "<message>"}` terminated by a newline.
26a. For protocol-level error responses (steps 19–24, before MessageHandler): serializes directly as raw JSON `{"status": "error", "message": "<description>"}` terminated by a newline. Does not use the RuntimeResponse entity.
27. Sends the serialized response to the client. If sending fails (client disconnected), logs a warning: `"failed to send response to client: <error>"`.
28. Closes the connection after sending the response (or after any error).
29. Only one request-response cycle is allowed per connection. Subsequent messages on the same connection are not read.

### DeleteSocket

30. Stops the listener (closes the underlying socket listener).
31. Closes all active connections immediately (hard interrupt, no drain).
32. Deletes the socket file from the filesystem.
33. Is idempotent: if the socket file does not exist, returns without error.
34. If deletion of the socket file fails (e.g., permission error), logs a warning: `"failed to delete runtime socket: <error>. The socket file may need to be manually removed."` and returns without error.
35. After `DeleteSocket()` completes, the background listener goroutine exits and `listenerDoneCh` is closed.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| sessionUUID | string | Valid UUID v4 format | Yes |
| logger | logger.Logger | Non-nil Logger interface implementation from the logger package | Yes |

### For CreateSocket

No additional inputs.

### For Listen

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| handler | MessageHandler | Non-nil MessageHandler interface implementation | Yes |

### For DeleteSocket

No additional inputs.

## Outputs

### For Construction

| Field | Type | Description |
|-------|------|-------------|
| manager | *RuntimeSocketManager | Configured instance holding the socket path |

No error — constructor does not perform I/O.

### For CreateSocket

**Success Case**: nil error.

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"runtime socket file already exists: <path>. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux \| grep spectra), then remove the socket file manually with: rm <path>"` | Socket file already exists |
| `"failed to create runtime socket: <error>"` | Socket creation failed (permission denied, disk full, path too long, etc.) |

### For Listen

**Success Case (synchronous)**: Returns `(listenerErrCh, listenerDoneCh, nil)`.

| Return | Type | Description |
|--------|------|-------------|
| ListenerErr | `<-chan error` | Buffered (capacity 1). Receives at most one error if the listener goroutine encounters a fatal failure after start. Never closed by RuntimeSocketManager. |
| ListenerDone | `<-chan struct{}` | Closed exactly once when the listener goroutine has fully exited. The only channel that gets closed. |
| Err | error | Non-nil if the initial bind/listen fails. Channels must not be relied on if non-nil. |

**Error Cases (synchronous)**:

| Error Message Format | Description |
|---------------------|-------------|
| `"runtime socket not created: call CreateSocket() first"` | Listen called before CreateSocket |
| `"failed to listen on runtime socket: <error>"` | Initial bind/listen failed |

**Error Cases (asynchronous, via ListenerErr channel)**:

| Error Message Format | Description |
|---------------------|-------------|
| `"listener accept loop failed: <error>"` | Accept loop encountered unrecoverable error after start |

### For DeleteSocket

**Success Case**: nil error. Socket file removed (or was already absent).

No error return — deletion failures are logged as warnings, not returned.

## Invariants

1. **Socket Path Uniqueness**: Each session has a unique runtime socket path determined by its UUID via StorageLayout.

2. **Single Socket per Session**: Only one socket file may exist per session at any time. Residual socket files are detected and reported as errors on CreateSocket.

3. **Owner-Only Permissions**: Socket files are created with permissions `0600` (owner read/write only).

4. **Request-Response Protocol**: Each client connection receives exactly one response. After sending, the connection is closed.

5. **JSON Newline Protocol**: All messages and responses are single JSON objects terminated by a newline (`\n`).

6. **Message Type Validation**: Each message must have a `type` field with value `"event"` or `"error"`. Other values are rejected at protocol level.

7. **Idempotent Deletion**: `DeleteSocket()` may be called multiple times or when the socket does not exist without error.

8. **Non-Blocking Cleanup**: Socket deletion failures are logged via Logger but do not block the caller or return errors.

9. **Connection Isolation**: Each connection is handled in a separate goroutine. Malformed messages or errors on one connection do not affect others.

10. **Listener Lifecycle Channels**: `Listen()` returns a buffered (capacity 1) error channel and a done channel. The error channel is never closed (avoids send-on-closed-channel panic races). The done channel is closed exactly once when the listener goroutine fully exits.

11. **No Message Buffering**: Messages are processed synchronously by the MessageHandler per connection. No internal queue or buffer.

12. **Message Size Limit**: Messages exceeding 10 MB are rejected before full read, with a warning logged and connection closed.

13. **MessageHandler Panic Non-Recovery**: RuntimeSocketManager does not recover panics from MessageHandler invocations. The MessageHandler implementation is responsible for internal panic recovery.

14. **Response Wire Format**: All responses sent to clients conform to the wire format `{"status": "success"|"error", "message": "<string>"}` terminated by newline. Business-logic responses are serialized from RuntimeResponse entity (via getters). Protocol-level error responses (sent before MessageHandler invocation) are serialized as raw JSON directly by RuntimeSocketManager without constructing a RuntimeResponse entity.

15. **Single Response per Connection**: After sending one response, the connection is closed. Multiple messages per connection are not supported.

16. **Hard Interrupt on Delete**: `DeleteSocket()` immediately closes the listener and all active connections without waiting for in-flight processing to complete.

17. **No Constructor I/O**: The constructor performs only path composition. No filesystem access occurs until CreateSocket is called.

18. **Unix Only**: Only Unix domain sockets are supported. No Windows named pipe abstraction.

19. **RuntimeMessage Construction**: Protocol-validated messages must be constructed via `NewRuntimeMessage`. Direct struct literal construction of RuntimeMessage is forbidden.

## Edge Cases

- Condition: Socket file already exists at the target path when `CreateSocket()` is called.
  Expected: Returns an error with the descriptive message including the path and remediation instructions.

- Condition: Session directory does not exist when `CreateSocket()` is called.
  Expected: Socket creation fails with a filesystem error (e.g., "no such file or directory"). Error is propagated as `"failed to create runtime socket: <error>"`.

- Condition: Socket path exceeds Unix domain socket path limit (~108 characters on some systems).
  Expected: Socket creation fails with a filesystem error. Error is propagated.

- Condition: `CreateSocket()` fails due to permission denied.
  Expected: Returns `"failed to create runtime socket: permission denied"`.

- Condition: `Listen()` is called before `CreateSocket()`.
  Expected: Returns synchronous error `"runtime socket not created: call CreateSocket() first"`.

- Condition: A spectra-agent client sends malformed JSON.
  Expected: Logger.Warn is called with connection details. Error response sent if possible. Connection closed. Other connections unaffected.

- Condition: A spectra-agent client sends valid JSON with missing `type` field.
  Expected: Logger.Warn called. Error response sent if possible. Connection closed.

- Condition: A spectra-agent client sends `type: "unknown"`.
  Expected: Logger.Warn called. Error response sent if possible. Connection closed.

- Condition: A spectra-agent client sends a message exceeding 10 MB.
  Expected: Size limit detected before full payload read. Logger.Warn called. Error response sent if possible. Connection closed. MessageHandler not invoked.

- Condition: A spectra-agent client closes the connection without sending a message.
  Expected: Connection handler detects closed connection and exits. No warning logged (normal close).

- Condition: `DeleteSocket()` is called while connections are active.
  Expected: Listener stops, all active connections closed immediately (in-flight processing may be interrupted), socket file deleted.

- Condition: `DeleteSocket()` fails to remove the socket file.
  Expected: Logger.Warn called. Returns without error. Caller proceeds normally.

- Condition: `DeleteSocket()` is called when socket file does not exist.
  Expected: Returns without error. No warning logged.

- Condition: Runtime process crashes before `DeleteSocket()`.
  Expected: Socket file remains (residual). Next `CreateSocket()` call detects it and returns an error.

- Condition: spectra-agent connects before socket is ready (race during initialization).
  Expected: Connection fails (connection refused). Retry logic is the caller's responsibility.

- Condition: Multiple spectra-agent clients connect simultaneously.
  Expected: All connections accepted and handled in separate goroutines concurrently.

- Condition: MessageHandler is slow (blocks for seconds).
  Expected: That connection remains open until handler completes. Other connections unaffected.

- Condition: MessageHandler panics without recovery.
  Expected: Panic propagates and crashes the connection handler goroutine. Other connections unaffected (separate goroutines). RuntimeSocketManager does not recover.

- Condition: MessageHandler returns a response with empty Message field.
  Expected: Response sent as-is: `{"status": "...", "message": ""}`.

- Condition: Sending response fails (client disconnected prematurely).
  Expected: Logger.Warn called. Connection closed. No retry.

- Condition: Client sends multiple messages over one connection.
  Expected: Only the first message is read and processed. Connection closed after first response. Subsequent messages ignored.

- Condition: DeleteSocket is called while the accept loop is waiting for a new connection.
  Expected: Accept loop exits. Active connections are closed. Socket file is deleted. listenerDoneCh is closed.

- Condition: Multiple goroutines call `CreateSocket()` concurrently.
  Expected: Filesystem-level atomicity determines winner. One succeeds, others get "file already exists" errors.

## Related

- [StorageLayout](./storage_layout.md) — Provides `GetRuntimeSocketPath` for socket file path
- [RuntimeMessage](../entities/runtime_message.md) — Message entity constructed from validated wire data
- [RuntimeResponse](../entities/runtime_response.md) — Response entity returned by MessageHandler
- [SessionDirectoryManager](./session_directory_manager.md) — Creates session directory before socket usage
- [Constants](./constants.md) — Provides `MaxPayloadSize` (10 MB message size limit)
- [Logger](../logger/logger.md) — Structured logging interface used for protocol warnings and cleanup failures
