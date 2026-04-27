# SocketClient

## Overview

SocketClient is a shared component that encapsulates all socket communication logic for spectra-agent subcommands. It handles the complete lifecycle of a request-response interaction with the Runtime: connecting to the session socket, sending a `RuntimeMessage`, receiving a `RuntimeResponse`, and closing the connection. SocketClient enforces the 30-second timeout for the entire operation and translates socket-level errors into the appropriate exit code categories (transport error vs. Runtime execution error).

## Behavior

### Core Responsibilities

1. SocketClient provides a single public method: `Send(sessionID, projectRoot, runtimeMessage) → (response, exitCode, error)`.
2. `Send` performs the following lifecycle: `connect → send → receive → close → return`.
3. SocketClient uses `StorageLayout.GetRuntimeSocketPath(projectRoot, sessionID)` to determine the socket file path.
4. SocketClient connects to the Unix domain socket (or named pipe on Windows) at the computed path.
5. The connection timeout is fixed at 30 seconds. The timeout applies to the entire operation (connect + send + receive).
6. The timeout timer starts when the connection attempt begins.
7. If the socket file does not exist, `Send` returns exit code 2 and error message: `"Error: socket file not found: <path>"`
8. If the connection is refused (Runtime not listening), `Send` returns exit code 2 and error message: `"Error: connection refused: Runtime is not running for session <uuid>"`
9. If the connection times out after 30 seconds, `Send` returns exit code 2 and error message: `"Error: connection timeout after 30s"`
10. After successfully connecting, SocketClient serializes the `runtimeMessage` to JSON and appends a newline (`\n`).
11. SocketClient writes the JSON message to the socket.
12. If sending the message fails (I/O error), `Send` returns exit code 2 and error message: `"Error: failed to send message: <io-error>"`
13. After successfully sending the message, SocketClient blocks and waits to read the Runtime's response.
14. SocketClient reads the response as a single JSON object terminated by a newline (`\n`).
15. If reading the response fails (I/O error or connection closed prematurely), `Send` returns exit code 2 and error message: `"Error: failed to read response: <io-error>"`
16. SocketClient parses the response JSON and validates its structure.
17. If the response is malformed JSON, `Send` returns exit code 3 and error message: `"Error: malformed response from Runtime: <json-parse-error>"`
18. If the response is valid JSON but missing the `status` field, `Send` returns exit code 3 and error message: `"Error: response missing 'status' field"`
19. If the response contains an invalid `status` value (not "success" or "error"), `Send` returns exit code 3 and error message: `"Error: invalid response status '<value>'"`
20. If the response status is `"success"`, `Send` returns exit code 0 and the response object (containing the optional `message` field).
21. If the response status is `"error"`, `Send` returns exit code 3 and the response object (containing the error `message` field).
22. After receiving the response (or encountering an error), SocketClient closes the connection.
23. If closing the socket fails, SocketClient prints a warning to stderr: `"Warning: failed to close socket: <error>"`. The warning does not change the exit code.
24. If any error occurs after a connection attempt (successful or failed), SocketClient must attempt to close the socket before returning.
25. SocketClient does not implement retry logic. If the connection fails, the caller is responsible for retrying if needed.

### Response Handling

The Runtime response format is:
```json
{
  "status": "success" | "error",
  "message": "<optional-human-readable-message>"
}
```

SocketClient returns a structured response object to the caller:

| Field | Type | Description |
|-------|------|-------------|
| Status | string | Response status: "success" or "error" |
| Message | string | Human-readable message (may be empty) |

## Inputs

### For Send Method

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| SessionID | string (UUID) | Valid UUID v4 format (not validated by SocketClient) | Yes |
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| RuntimeMessage | struct | Contains `Type` (string: "event" or "error") and `Payload` (JSON object) | Yes |

### RuntimeMessage Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Type | string | Must be "event" or "error" | Message type |
| ClaudeSessionID | string | Any string, may be empty `""` | Claude session identifier from `--claude-session-id` flag (validated server-side based on current node type) |
| Payload | JSON object | Depends on message type | Message-specific data |

**For "event" type**:
```json
{
  "type": "event",
  "claudeSessionID": "<string>",
  "payload": {
    "eventType": "<string>",
    "message": "<string>",
    "payload": <json-object>
  }
}
```

**For "error" type**:
```json
{
  "type": "error",
  "claudeSessionID": "<string>",
  "payload": {
    "message": "<string>",
    "detail": <json-object-or-null>
  }
}
```

## Outputs

### Return Values from Send

| Output | Type | Description |
|--------|------|-------------|
| Response | struct (nullable) | Contains `Status` and `Message` fields. Null if transport error occurred. |
| ExitCode | int | Exit code: 0 (success), 2 (transport error), 3 (Runtime execution error) |
| Error | error (nullable) | Error message to be printed to stderr. Null if exit code is 0. |

### Exit Code Mapping

| Exit Code | Category | Trigger Conditions |
|-----------|----------|-------------------|
| 0 | Success | Runtime responded with `"status": "success"` |
| 2 | Transport Error | Socket file not found, connection refused, connection timeout, I/O error during send/receive |
| 3 | Runtime Execution Error | Runtime responded with `"status": "error"`, malformed response JSON, missing/invalid `status` field |

### stderr Output

**Exit Code 2 Examples**:
```
Error: socket file not found: /path/to/.spectra/sessions/<uuid>/runtime.sock
Error: connection refused: Runtime is not running for session <uuid>
Error: connection timeout after 30s
Error: failed to send message: <io-error>
Error: failed to read response: <io-error>
```

**Exit Code 3 Examples**:
```
Error: malformed response from Runtime: <json-parse-error>
Error: response missing 'status' field
Error: invalid response status 'unknown'
```

**Warning (does not affect exit code)**:
```
Warning: failed to close socket: <error>
```

## Invariants

1. **Synchronous Blocking**: `Send` must block until the Runtime responds or the timeout occurs. It must not return prematurely.

2. **Timeout Scope**: The 30-second timeout applies to the entire operation (connect + send + receive). The timer starts when the connection attempt begins.

3. **Socket Lifecycle**: If a connection attempt is made (successful or failed), the socket must be closed before `Send` returns, regardless of success or failure.

4. **Exit Code Determinism**: The exit code must be deterministic based on the error category: transport errors (2) vs. Runtime execution errors (3).

5. **JSON Message Protocol**: All messages sent to the Runtime must be valid JSON objects conforming to the `RuntimeMessage` format, terminated by `\n`.

6. **Response Validation**: All responses from the Runtime must be parsed as JSON and validated for the required `status` field with valid values ("success" or "error").

7. **Error Prefix**: All error messages must be prefixed with `"Error: "`. Warnings must be prefixed with `"Warning: "`.

8. **No Retry Logic**: SocketClient must not retry failed operations. The caller is responsible for implementing retry logic if needed.

9. **Platform Abstraction**: SocketClient must abstract the differences between Unix domain sockets and Windows named pipes. The API and behavior remain consistent across platforms.

10. **Stateless Execution**: SocketClient must not maintain any state between `Send` invocations. Each invocation is independent.

11. **Single Response per Connection**: SocketClient must read exactly one response from the Runtime and then close the connection. Multiple responses are not supported.

## Edge Cases

- **Condition**: Socket file does not exist at the computed path.
  **Expected**: Connection fails immediately. Return exit code 2, error message: `"Error: socket file not found: <path>"`. Attempt to close socket (no-op if connection never opened).

- **Condition**: Socket file exists but connection is refused (Runtime crashed or not listening).
  **Expected**: Return exit code 2, error message: `"Error: connection refused: Runtime is not running for session <uuid>"`. Attempt to close socket.

- **Condition**: Connection times out after 30 seconds (Runtime not responding).
  **Expected**: Return exit code 2, error message: `"Error: connection timeout after 30s"`. Attempt to close socket.

- **Condition**: Connection succeeds but sending the message fails (I/O error).
  **Expected**: Return exit code 2, error message: `"Error: failed to send message: <io-error>"`. Attempt to close socket.

- **Condition**: Message sent successfully but reading the response fails (I/O error).
  **Expected**: Return exit code 2, error message: `"Error: failed to read response: <io-error>"`. Attempt to close socket.

- **Condition**: Connection succeeds, message sent, but response times out after 30 seconds.
  **Expected**: Return exit code 2, error message: `"Error: connection timeout after 30s"`. Attempt to close socket.

- **Condition**: Runtime responds with `{"status": "success", "message": ""}` (empty message).
  **Expected**: Return exit code 0, response object with `Status: "success"`, `Message: ""`.

- **Condition**: Runtime responds with `{"status": "success", "message": "Custom success message"}`.
  **Expected**: Return exit code 0, response object with `Status: "success"`, `Message: "Custom success message"`.

- **Condition**: Runtime responds with `{"status": "error", "message": "session not found"}`.
  **Expected**: Return exit code 3, response object with `Status: "error"`, `Message: "session not found"`.

- **Condition**: Runtime responds with malformed JSON (e.g., `{invalid`).
  **Expected**: Return exit code 3, error message: `"Error: malformed response from Runtime: unexpected end of JSON input"`.

- **Condition**: Runtime responds with valid JSON but missing `status` field (e.g., `{"message": "ok"}`).
  **Expected**: Return exit code 3, error message: `"Error: response missing 'status' field"`.

- **Condition**: Runtime responds with `{"status": "unknown", "message": "..."}` (invalid status value).
  **Expected**: Return exit code 3, error message: `"Error: invalid response status 'unknown'"`.

- **Condition**: Closing the socket fails after a successful operation (exit code 0).
  **Expected**: Print warning to stderr: `"Warning: failed to close socket: <error>"`. Return exit code 0 (unchanged).

- **Condition**: Closing the socket fails after a failed operation (exit code 2 or 3).
  **Expected**: Print warning to stderr: `"Warning: failed to close socket: <error>"`. Return the original exit code (unchanged).

- **Condition**: Socket file is deleted (session terminates) while SocketClient is attempting to connect.
  **Expected**: Connection fails. Return exit code 2, error message: `"Error: socket file not found: <path>"` or `"Error: connection refused: Runtime is not running for session <uuid>"`.

- **Condition**: Session directory exists but `runtime.sock` file does not (session in "initializing" or "completed" status).
  **Expected**: Connection fails. Return exit code 2, error message: `"Error: socket file not found: <path>"`.

- **Condition**: SessionID is an invalid UUID format (e.g., `"abc-123"`).
  **Expected**: SocketClient proceeds with the invalid UUID (does not validate format). Socket path is computed using the invalid value. Connection likely fails with exit code 2. UUID validation is the Runtime's responsibility.

- **Condition**: ProjectRoot does not contain a `.spectra/` directory.
  **Expected**: SocketClient computes the socket path as if `.spectra/` exists. Connection fails with exit code 2 (socket file not found). Directory existence validation is the root command's responsibility.

- **Condition**: RuntimeMessage contains invalid JSON in the `Payload` field (e.g., circular references).
  **Expected**: JSON serialization fails. This is a programming error. SocketClient should log an internal error and return exit code 1 (or panic, depending on implementation philosophy).

- **Condition**: Runtime sends multiple JSON responses (violating the single-response protocol).
  **Expected**: SocketClient reads and parses the first response, closes the connection, and returns. Subsequent responses are ignored.

- **Condition**: Runtime closes the connection without sending a response.
  **Expected**: SocketClient detects the closed connection when attempting to read. Return exit code 2, error message: `"Error: failed to read response: connection closed by Runtime"`.

## Related

- [Root Command](./root.md) - Initializes project root and session ID
- [event emit Subcommand](./event_emit.md) - Constructs event messages and invokes SocketClient
- [error Subcommand](./error.md) - Constructs error messages and invokes SocketClient
- [StorageLayout](../../storage/storage_layout.md) - Provides socket path composition
- [RuntimeSocketManager](../../storage/runtime_socket_manager.md) - Server-side socket management and response generation
