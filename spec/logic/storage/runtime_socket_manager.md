# RuntimeSocketManager

## Overview

RuntimeSocketManager manages the lifecycle of Unix domain socket (or named pipe on Windows) for a single session. It creates, listens to, and deletes the runtime socket file (`runtime.sock`) located in the session directory. The socket enables request-response communication between spectra-agent (client) and the Workflow Runtime (server). RuntimeSocketManager provides methods for the Runtime to create the socket, listen for incoming connections, receive JSON-formatted messages (event emissions and error reports), send JSON-formatted responses back to clients, and clean up the socket when the session terminates. It does not manage the session directory or coordinate with other storage components.

## Behavior

1. RuntimeSocketManager is initialized with a session UUID and uses StorageLayout to determine the path to `runtime.sock`.
2. RuntimeSocketManager provides a `CreateSocket()` method that creates a Unix domain socket file at the session-specific path.
3. Before creating the socket, `CreateSocket()` checks if the socket file already exists at the target path.
4. If the socket file exists, `CreateSocket()` returns an error: "runtime socket file already exists: <path>. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm <path>". The caller (e.g., Session initialization) is responsible for triggering a RuntimeError with `Issuer="Session"` to transition the session to "failed" status.
5. `CreateSocket()` creates the socket file with permissions `0600` (owner read/write only) to restrict access to the session owner. If socket creation fails (e.g., permission denied, disk full), `CreateSocket()` returns an error. The caller is responsible for triggering a RuntimeError.
6. On Windows, `CreateSocket()` creates a named pipe instead of a Unix domain socket, using the same path format. The implementation adapts to the platform-specific transport mechanism.
7. RuntimeSocketManager provides a `Listen()` method that binds to the created socket and starts listening for incoming connections. If the bind or listen operation fails, `Listen()` returns an error. The caller is responsible for triggering a RuntimeError with `Issuer="Session"` to transition the session to "failed" status.
8. `Listen()` accepts connections from spectra-agent clients. Each connection is handled in a separate goroutine.
9. RuntimeSocketManager provides a `Receive()` method that reads incoming messages from connected clients.
10. Each message is expected to be a single JSON object terminated by a newline (`\n`). The JSON format is defined by the message protocol (event emission or error report).
11. `Receive()` parses the JSON message and returns a structured message object containing the message type ("event" or "error") and the payload.
12. If a message is malformed (invalid JSON, missing required fields, or wrong format), RuntimeSocketManager logs a warning with connection details, sends an error response to the client (if the connection is still writable), and closes the connection.
13. After successfully parsing a message, RuntimeSocketManager extracts the session UUID from the socket path (`.spectra/sessions/<SessionUUID>/runtime.sock`) and invokes the `MessageHandler` callback with the session UUID and the parsed message. The MessageHandler processes the message and returns a `RuntimeResponse` struct.
14. RuntimeSocketManager serializes the `RuntimeResponse` to JSON and sends it back to the client over the same connection, terminated by a newline (`\n`).
15. The response format is: `{"status": "success" | "error", "message": "<human-readable-message>"}`.
16. If the `MessageHandler` returns a success response, RuntimeSocketManager sends `{"status": "success", "message": "<message>"}` to the client.
17. If the `MessageHandler` returns an error response, RuntimeSocketManager sends `{"status": "error", "message": "<error-description>"}` to the client.
18. After sending the response, RuntimeSocketManager closes the connection gracefully.
19. RuntimeSocketManager provides a `DeleteSocket()` method that stops listening, closes all active connections, and deletes the socket file.
20. `DeleteSocket()` is idempotent: if the socket file does not exist, it returns without error (no-op).
21. If `DeleteSocket()` fails to remove the socket file (e.g., due to permission error or filesystem issue), it logs a warning but does not return an error. This allows the Runtime to proceed with session termination even if socket cleanup fails.
22. RuntimeSocketManager does not retry socket creation if it fails. Socket creation errors are propagated to the Runtime, which should transition the session to "failed" status.
23. RuntimeSocketManager does not implement connection retry logic. If a spectra-agent client fails to connect, the connection fails immediately with an error. Retry logic, if needed, should be implemented by the caller of spectra-agent (e.g., external scripts, workflow orchestration).
24. RuntimeSocketManager performs **structural and protocol-level validation only**. This means: (a) the message is well-formed JSON, (b) the top-level fields `type` and `payload` are present, (c) `type` is one of the recognized values (`"event"` or `"error"`), (d) `payload` is a JSON object, (e) `claudeSessionID` (if present) is a string, (f) for `type="event"`, payload contains required field `eventType` (string) and optional `message` (string)/`payload` (object); for `type="error"`, payload contains required field `message` (non-empty string) and optional `detail` (object or null). RuntimeSocketManager does **not** validate semantic correctness (e.g., whether the event type is workflow-defined, whether the claudeSessionID matches stored values, whether the session is in a state that accepts messages, derivation of `agentRole` from the current node). Semantic validation and the population of derived fields such as `agentRole` are the MessageHandler's responsibility (EventProcessor / ErrorProcessor).
25. RuntimeSocketManager enforces a maximum message size limit of 10 MB per message as a safety guardrail. Messages exceeding this limit are rejected, an error response is sent to the client (if possible), the connection is closed, and a warning is logged. This prevents out-of-memory (OOM) vulnerabilities from maliciously large payloads.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| SessionUUID | string (UUID) | Valid UUID v4 format | Yes |

### For CreateSocket

No additional inputs beyond initialization parameters.

### For Listen

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| MessageHandler | function interface | Signature: `func(sessionUUID string, message RuntimeMessage) RuntimeResponse`. This is an interface defining the contract for message processing. The concrete implementation (e.g., MessageRouter) accepts the session UUID (extracted from the socket path) and a `RuntimeMessage` struct, processes them, and returns a `RuntimeResponse` struct. The handler runs in a separate goroutine for each connection. The handler must return a response indicating success or error, which RuntimeSocketManager sends back to the client. The MessageHandler implementation (MessageRouter) is initialized with a Session reference and other runtime components (EventProcessor, ErrorProcessor, TerminationNotifier) before being passed to RuntimeSocketManager. RuntimeSocketManager does not directly access the Session; it only invokes the handler callback. | Yes |

**Return values from `Listen`**:

`Listen` returns two channels and an immediate error:

| Return | Type | Description |
|--------|------|-------------|
| ListenerErr | `<-chan error` | Buffered (capacity 1). Receives at most one error if the listener goroutine encounters a fatal failure (e.g., bind failure that surfaces after a successful first accept, accept loop unrecoverable error). **Never closed** by RuntimeSocketManager — consumers must not assume a close means "no more errors" and must instead observe `ListenerDone` for the listener-shutdown signal. The channel is garbage-collected when RuntimeSocketManager and Runtime become unreachable. |
| ListenerDone | `<-chan struct{}` | Closed exactly once by RuntimeSocketManager when the listener goroutine has fully exited (after `DeleteSocket()` is called or after a fatal listener error). This is the only channel returned by `Listen()` that gets closed. Allows Runtime to wait for goroutine cleanup before invoking SessionFinalizer. |
| Err | error | Returned synchronously if the initial bind/listen fails before the goroutine starts. Non-nil means the listener is **not** running and the channels above must not be relied on. |

If the synchronous `Err` is non-nil, the caller must not start its main monitoring loop and should treat the failure as a session-fatal error (RuntimeError → Session.Fail).

### For Receive

Called internally by `Listen()`. No direct inputs from the caller.

### For DeleteSocket

No additional inputs beyond initialization parameters.

## Outputs

### For CreateSocket

**Success Case**: No return value (void / nil error in Go).

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"runtime socket file already exists: <path>. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux \| grep spectra), then remove the socket file manually with: rm <path>"` | Socket file exists at the target path |
| `"failed to create runtime socket: <error>"` | Socket creation failed (e.g., permission denied, disk full, path too long) |

### For Listen

**Success Case**: Synchronous bind/listen succeeds; returns `(listenerErrCh, listenerDoneCh, nil)`. The listener runs in the background, accepting connections and invoking the `MessageHandler` callback for each received message.

**Error Cases (synchronous)**:

| Error Message Format | Description |
|---------------------|-------------|
| `"runtime socket not created: call CreateSocket() first"` | `Listen()` was called before `CreateSocket()` |
| `"failed to listen on runtime socket: <error>"` | Initial socket bind or listen operation failed |

**Error Cases (asynchronous, delivered via `listenerErrCh`)**:

| Error Message Format | Description |
|---------------------|-------------|
| `"listener accept loop failed: <error>"` | Accept loop encountered an unrecoverable error after the listener had started |

### For Receive (via MessageHandler callback)

**Input to MessageHandler**:

| Field | Type | Description |
|-------|------|-------------|
| SessionUUID | string (UUID) | Session UUID extracted from the socket path (`.spectra/sessions/<SessionUUID>/runtime.sock`) |
| RuntimeMessage | RuntimeMessage struct | `RuntimeMessage` struct containing `Type` ("event" or "error") and `Payload` (JSON object). For "event" type: `{eventType: string, message: string, payload: object}`. For "error" type: `{message: string, detail: object}` (note: `agentRole` is **not** part of the wire payload — it is derived server-side by ErrorProcessor from the current node's definition). |

**Output from MessageHandler**: `RuntimeResponse` struct containing:

| Field | Type | Description |
|-------|------|-------------|
| Status | string | Response status: "success" or "error" |
| Message | string | Human-readable message describing the result or error |

**Error Cases** (connection dropped, error response sent if possible, no callback invocation):

- Malformed JSON
- Missing required fields (`type`, `payload`)
- Invalid message type (not "event" or "error")
- Message size exceeds 10 MB limit
- Connection closed by client
- Read timeout or I/O error

### For DeleteSocket

**Success Case**: No return value (void / nil error in Go). Socket file is removed if it exists.

**Warnings** (logged, not returned as errors):

| Warning Message Format | Description |
|----------------------|-------------|
| `"failed to delete runtime socket: <error>. The socket file may need to be manually removed."` | Socket file deletion failed but did not block the operation |

## Invariants

1. **Socket Path Uniqueness**: Each session must have a unique runtime socket path. The path is determined by the session UUID.

2. **Single Socket per Session**: RuntimeSocketManager must ensure only one socket file exists per session at any given time. Residual socket files from previous sessions must be detected and reported as errors.

3. **Owner-Only Permissions**: Newly created socket files must have permissions `0600` (owner read/write only) to prevent unauthorized access.

4. **Socket Lifecycle Tied to Session**: The socket must be created when the session transitions to "running" status and deleted when the session reaches "completed" or "failed" status.

5. **Request-Response Communication**: The socket supports request-response communication. Each client request (spectra-agent → Runtime) must receive exactly one response (Runtime → spectra-agent) before the connection is closed.

6. **JSON Message Protocol**: All messages must be JSON objects terminated by a newline (`\n`). Non-JSON messages are rejected.

7. **Message Type Validation**: Each message must have a `type` field with value "event" or "error". Other values are rejected.

8. **Idempotent Deletion**: `DeleteSocket()` must be idempotent. Calling it multiple times or when the socket does not exist must not cause errors.

9. **Non-Blocking Cleanup**: Socket deletion failures must not block session termination. Warnings are logged, but errors are not propagated.

10. **Platform-Specific Transport**: On Unix-like systems, use Unix domain sockets. On Windows, use named pipes. The path composition and API remain the same; the underlying implementation differs.

11. **Connection Isolation**: Each spectra-agent connection is handled in a separate goroutine. Malformed messages or errors on one connection must not affect other connections.

12. **Listener Lifecycle Channels**: `Listen()` must return a buffered (capacity 1) error channel and a done channel. The error channel receives at most one error if the listener fails after starting and is **never closed** (closing it would risk a `send on closed channel` panic if a fatal error and `DeleteSocket()` race; consumers observe shutdown via the done channel instead). The done channel is closed exactly once by RuntimeSocketManager when the listener goroutine has fully exited. The caller must not close either channel.

13. **No Message Buffering**: RuntimeSocketManager does not buffer messages in memory. Messages are processed synchronously by the `MessageHandler` callback.

14. **Message Size Limit**: RuntimeSocketManager must enforce a maximum message size of 10 MB per message. Messages exceeding this limit must be rejected with a warning logged and the connection closed. This prevents out-of-memory (OOM) vulnerabilities.

15. **MessageHandler Response Responsibility**: The `MessageHandler` callback must return a `RuntimeResponse` struct indicating success or error. The handler is responsible for all business logic validation (e.g., invalid event type, session not ready, state update failures). If validation or processing fails, the handler returns an error response with a descriptive message. RuntimeSocketManager serializes and sends the response to the client. RuntimeSocketManager only handles connection-layer and protocol-layer errors (malformed JSON, message size violations, I/O errors).

16. **MessageHandler Panic Recovery**: RuntimeSocketManager assumes the MessageHandler implementation (e.g., MessageRouter) handles panic recovery internally. RuntimeSocketManager does not implement panic recovery for MessageHandler invocations. If the MessageHandler panics and does not recover, the panic will propagate up and may crash the connection handler goroutine or the Runtime process, depending on the Go runtime behavior. It is the MessageHandler implementation's responsibility to ensure robustness through panic recovery.

17. **Response Format Validation**: All responses sent to clients must conform to the JSON format: `{"status": "success" | "error", "message": "<string>"}`, terminated by a newline (`\n`).

18. **Single Response per Connection**: Each client connection must receive exactly one response. After sending the response, RuntimeSocketManager must close the connection to signal completion to the client.

## Edge Cases

- **Condition**: Socket file already exists at the target path when `CreateSocket()` is called.
  **Expected**: RuntimeSocketManager returns an error with a descriptive message. The caller (e.g., Session initialization) should trigger a RuntimeError with `Issuer="Session"` to transition the session to "failed" status and notify the user to manually remove the residual socket file.

- **Condition**: Session directory does not exist when `CreateSocket()` is called.
  **Expected**: Socket creation fails with a filesystem error (e.g., "no such file or directory"). RuntimeSocketManager propagates the error. The caller should trigger a RuntimeError with `Issuer="Session"` to transition the session to "failed".

- **Condition**: Socket path exceeds the platform's maximum path length (e.g., ~108 characters on some Unix systems).
  **Expected**: Socket creation fails with a filesystem error (e.g., "file name too long"). RuntimeSocketManager propagates the error. Consider using shorter project root paths or session UUID abbreviations.

- **Condition**: `CreateSocket()` fails due to permission denied (e.g., session directory is read-only).
  **Expected**: RuntimeSocketManager returns an error: `"failed to create runtime socket: permission denied"`. The caller should trigger a RuntimeError with `Issuer="Session"` to transition the session to "failed".

- **Condition**: `Listen()` is called before `CreateSocket()`.
  **Expected**: RuntimeSocketManager returns an error: `"runtime socket not created: call CreateSocket() first"`. The caller should trigger a RuntimeError if appropriate.

- **Condition**: Multiple goroutines call `CreateSocket()` concurrently.
  **Expected**: The filesystem-level atomic check for socket file existence prevents race conditions. One goroutine succeeds; the others receive "file already exists" errors.

- **Condition**: A spectra-agent client sends a malformed JSON message (e.g., missing closing brace).
  **Expected**: RuntimeSocketManager logs a warning: `"dropping connection <client-id>: malformed JSON: unexpected end of JSON input"`. The connection is closed. Other connections continue to function.

- **Condition**: A spectra-agent client sends a valid JSON object but with a missing `type` field.
  **Expected**: RuntimeSocketManager logs a warning: `"dropping connection <client-id>: missing required field 'type'"`. The connection is closed.

- **Condition**: A spectra-agent client sends a message with `type: "unknown"`.
  **Expected**: RuntimeSocketManager logs a warning: `"dropping connection <client-id>: invalid message type 'unknown'"`. The connection is closed.

- **Condition**: A spectra-agent client sends a very large JSON payload exceeding 10 MB, a reasonable size limit.
  **Expected**: RuntimeSocketManager detects the message size exceeds the 10 MB limit before fully reading the payload. It logs a warning: `"dropping connection <client-id>: message size exceeds 10 MB limit"`, closes the connection, and does not invoke the MessageHandler.

- **Condition**: A spectra-agent client closes the connection without sending a message.
  **Expected**: RuntimeSocketManager detects the closed connection and exits the connection handler goroutine. No error is logged (normal connection close).

- **Condition**: `DeleteSocket()` is called while `Listen()` is active with open connections.
  **Expected**: `DeleteSocket()` stops listening, closes all active connections gracefully, and deletes the socket file. In-flight message processing may be interrupted.

- **Condition**: `DeleteSocket()` fails to remove the socket file due to a filesystem error.
  **Expected**: RuntimeSocketManager logs a warning: `"failed to delete runtime socket: <error>. The socket file may need to be manually removed."` The method returns without error. The Runtime proceeds with session termination.

- **Condition**: `DeleteSocket()` is called when the socket file does not exist (e.g., already deleted or never created).
  **Expected**: `DeleteSocket()` returns immediately without error (no-op). No warning is logged.

- **Condition**: On Windows, the underlying named pipe implementation behaves differently from Unix domain sockets.
  **Expected**: RuntimeSocketManager abstracts the platform differences. The API and behavior remain consistent. Path composition methods return the same format; the I/O layer adapts to named pipes.

- **Condition**: Runtime process crashes or is forcibly terminated (e.g., kill -9) before calling `DeleteSocket()`.
  **Expected**: The socket file remains on the filesystem (residual socket). The next session creation will detect the residual socket file and return an error, prompting the user to clean up manually.

- **Condition**: spectra-agent attempts to connect before the socket is ready (race condition during session initialization).
  **Expected**: The connection attempt fails with exit code 2 (socket file not found or connection refused). The caller of spectra-agent (e.g., external script, workflow orchestration) may implement retry logic with exponential backoff or poll until the session reaches "running" status.

- **Condition**: Multiple spectra-agent clients connect simultaneously to the same socket.
  **Expected**: RuntimeSocketManager accepts all connections and handles each in a separate goroutine. Messages from different clients are processed concurrently by the `MessageHandler`.

- **Condition**: Message processing in the `MessageHandler` is slow (e.g., blocks for several seconds).
  **Expected**: The connection remains open until the handler completes. Other connections are not affected (each runs in a separate goroutine).

- **Condition**: The `MessageHandler` callback panics while processing a message.
  **Expected**: The MessageHandler implementation (e.g., MessageRouter) is expected to handle panic recovery internally. If the MessageHandler does not recover, the panic may propagate and crash the connection handler goroutine. RuntimeSocketManager does not provide panic recovery as a safety net. It is the MessageHandler's responsibility to implement robust error handling.

- **Condition**: The `MessageHandler` returns a response with an empty `Message` field.
  **Expected**: RuntimeSocketManager sends the response as-is: `{"status": "success", "message": ""}` or `{"status": "error", "message": ""}`.

- **Condition**: Sending the response to the client fails due to an I/O error (e.g., client disconnected prematurely).
  **Expected**: RuntimeSocketManager logs a warning: `"failed to send response to client: <error>"` and closes the connection. No retry is attempted.

- **Condition**: The client closes the connection after sending a request but before reading the response.
  **Expected**: RuntimeSocketManager processes the request, attempts to send the response, detects the closed connection, logs a warning, and exits the handler goroutine.

- **Condition**: A spectra-agent client sends multiple JSON messages over the same connection (newline-delimited stream).
  **Expected**: RuntimeSocketManager reads the first message, processes it, sends a response, and immediately closes the connection. Only one request-response cycle is allowed per connection. Subsequent messages are not processed.

## Related

- [Session](../entities/session/session.md) - Defines the session lifecycle and socket creation/deletion timing
- [Event](../entities/event.md) - Events are emitted by spectra-agent via the runtime socket
- [AgentError](../entities/agent_error.md) - Errors are reported by spectra-agent via the runtime socket
- [StorageLayout](./storage_layout.md) - Provides the path to `runtime.sock`
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Agent communication and workflow runtime architecture
