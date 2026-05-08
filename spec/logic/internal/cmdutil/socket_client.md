# SocketClient

## Overview

SocketClient encapsulates the complete lifecycle of a single request-response interaction with the Runtime over a Unix domain socket. It connects to the session socket, sends a JSON message, receives a JSON response, and closes the connection. SocketClient enforces a 30-second timeout for the entire operation and classifies errors into transport errors (exit code 2) or runtime execution errors (exit code 3). It does not manage retries or maintain state between invocations.

## Boundaries

- Owns: socket connection lifecycle (connect, send, receive, close) for a single request-response.
- Owns: 30-second timeout enforcement across the entire operation.
- Owns: error classification into transport errors vs. runtime errors.
- Owns: response JSON parsing and structural validation (status field presence and value).
- Delegates: socket path composition to `storage.StorageLayout.GetRuntimeSocketPath`.
- Must not: implement retry logic.
- Must not: maintain state between `Send` invocations.
- Must not: validate RuntimeMessage content (that is the caller's responsibility).
- Must not: perform business-logic interpretation of the response message content.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `storage.StorageLayout` | Path composition | `GetRuntimeSocketPath(projectRoot, sessionID)` | Must not use any other StorageLayout function |

Construction constraint: SocketClient is a stateless module. `Send` is a package-level function (or method on a zero-value struct for testability). No constructor is needed.

## Behavior

1. `Send(sessionID, projectRoot string, message []byte) (response *Response, exitCode int, err error)`.
2. Computes the socket path via `storage.StorageLayout.GetRuntimeSocketPath(projectRoot, sessionID)`.
3. Checks if the socket file exists. If not, returns exit code 2 and error: `"socket file not found: <path>"`.
4. Connects to the Unix domain socket at the computed path with a 30-second deadline applied to the entire operation (connect + send + receive).
5. If connection is refused, returns exit code 2 and error: `"connection refused: Runtime is not running for session <sessionID>"`.
6. If connection times out, returns exit code 2 and error: `"connection timeout after 30s"`.
7. Writes the JSON message bytes followed by a newline (`\n`) to the socket.
8. If write fails, returns exit code 2 and error: `"failed to send message: <io-error>"`.
9. Reads the response as a newline-terminated JSON line.
10. If read fails (I/O error or connection closed prematurely), returns exit code 2 and error: `"failed to read response: <io-error>"`.
11. Parses the response JSON into a Response struct.
12. If JSON parsing fails, returns exit code 3 and error: `"malformed response from Runtime: <parse-error>"`.
13. Validates that the `status` field is present. If missing, returns exit code 3 and error: `"response missing 'status' field"`.
14. Validates that `status` is `"success"` or `"error"`. If invalid, returns exit code 3 and error: `"invalid response status '<value>'"`.
15. Returns the parsed Response, exit code 0 (if status is "success") or exit code 3 (if status is "error"), and nil error for success.
16. Closes the connection after receiving the response (or encountering an error). If close fails, prints a warning to stderr but does not change the exit code.

## Inputs

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| sessionID | string | Non-empty (not validated for UUID format) | Yes |
| projectRoot | string | Absolute path to directory containing `.spectra` | Yes |
| message | []byte | Valid JSON bytes (caller is responsible for well-formedness) | Yes |

## Outputs

### Response Struct

| Field | Type | Description |
|-------|------|-------------|
| Status | string | `"success"` or `"error"` |
| Message | string | Human-readable message (may be empty) |

### Return Values

| Output | Type | Description |
|--------|------|-------------|
| response | *Response | Parsed response (nil if transport error) |
| exitCode | int | 0, 2, or 3 |
| err | error | Error message for stderr (nil if exitCode is 0) |

## Invariants

1. **Timeout Scope**: The 30-second timeout applies to the entire operation (connect + send + receive). Timer starts at connection attempt.
2. **Socket Lifecycle**: Connection is always closed before `Send` returns, regardless of success or failure.
3. **Exit Code Determinism**: Transport errors always map to exit code 2. Response validation errors and runtime-reported errors always map to exit code 3.
4. **JSON Newline Protocol**: Messages are sent as JSON terminated by `\n`. Responses are read as a single JSON line terminated by `\n`.
5. **No Retry**: SocketClient never retries failed operations.
6. **Stateless**: Each `Send` invocation is independent. No shared state between calls.
7. **Single Response**: Reads exactly one response per connection. Connection is closed after reading.
8. **Close Warning**: Close failures produce a stderr warning (`"Warning: failed to close socket: <error>"`) but do not alter the exit code.

## Edge Cases

- Condition: Socket file does not exist.
  Expected: Returns exit code 2, error `"socket file not found: <path>"`.

- Condition: Socket file exists but connection refused.
  Expected: Returns exit code 2, error `"connection refused: Runtime is not running for session <sessionID>"`.

- Condition: Connection times out (Runtime not responding within 30s).
  Expected: Returns exit code 2, error `"connection timeout after 30s"`.

- Condition: Write fails after connection.
  Expected: Returns exit code 2, error `"failed to send message: <io-error>"`. Closes socket.

- Condition: Read fails (I/O error or premature close).
  Expected: Returns exit code 2, error `"failed to read response: <io-error>"`. Closes socket.

- Condition: Response is malformed JSON.
  Expected: Returns exit code 3, error `"malformed response from Runtime: <parse-error>"`.

- Condition: Response missing `status` field.
  Expected: Returns exit code 3, error `"response missing 'status' field"`.

- Condition: Response `status` is neither "success" nor "error".
  Expected: Returns exit code 3, error `"invalid response status '<value>'"`.

- Condition: Response `status` is "success".
  Expected: Returns exit code 0, parsed Response, nil error.

- Condition: Response `status` is "error".
  Expected: Returns exit code 3, parsed Response (with error message), nil error. Caller reads Response.Message for stderr output.

- Condition: Close fails after successful operation.
  Expected: Prints warning to stderr. Returns original exit code (0) unchanged.

- Condition: sessionID is malformed (not a UUID).
  Expected: Proceeds with malformed path. Connection likely fails with exit code 2.

## Related

- [StorageLayout](../../storage/storage_layout.md) — provides `GetRuntimeSocketPath`
- [ExitCodes](./exit_codes.md) — exit code constants used by SocketClient
- [ErrorFormatter](./error_formatter.md) — formats error messages for stderr output
- [RuntimeSocketManager](../../storage/runtime_socket_manager.md) — server-side counterpart
- [SendAndHandle](./send_and_handle.md) — higher-level function that wraps SocketClient
