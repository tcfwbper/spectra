# Test Specification: `client.go`

## Source File Under Test
`cmd/spectra_agent/client.go`

## Test File
`cmd/spectra_agent/client_test.go`

---

## `SocketClient`

### Happy Path — Send

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_SendSuccess` | `unit` | Successfully sends message and receives success response. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock Unix socket server created inside test fixture that responds with `{"status":"success","message":"ok"}\n` | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{Type:"event",ClaudeSessionID:"test-session",Payload:{...}}` | Returns response with `Status="success"`, `Message="ok"`, exit code `0`, error `nil` |
| `TestSocketClient_SendSuccessEmptyMessage` | `unit` | Receives success response with empty message field. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"success","message":""}\n` | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{Type:"event",ClaudeSessionID:"",Payload:{...}}` | Returns response with `Status="success"`, `Message=""`, exit code `0`, error `nil` |
| `TestSocketClient_SendEventMessage` | `unit` | Sends event-type message as valid JSON with newline terminator. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server captures sent message and responds with success | `RuntimeMessage{Type:"event",ClaudeSessionID:"session-123",Payload:{"eventType":"MyEvent","message":"test","payload":{}}}` | Mock server receives valid JSON terminated by `\n`; message can be parsed successfully; returns exit code `0` |
| `TestSocketClient_SendErrorMessage` | `unit` | Sends error-type message as valid JSON with newline terminator. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server captures sent message and responds with success | `RuntimeMessage{Type:"error",ClaudeSessionID:"session-123",Payload:{"message":"test error","detail":{"key":"value"}}}` | Mock server receives valid JSON terminated by `\n`; message can be parsed successfully; returns exit code `0` |

### Happy Path — Runtime Error Response

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_RuntimeErrorResponse` | `unit` | Returns exit code 3 when Runtime responds with error status. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"error","message":"session not found"}\n` | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns response with `Status="error"`, `Message="session not found"`, exit code `3`, error contains `"session not found"` |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_ClosesSocketOnSuccess` | `unit` | Closes socket connection and releases file descriptor after successful operation. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server tracks connection state; test monitors file descriptor count | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Mock server confirms connection closed after response sent; file descriptor count returns to baseline; exit code `0` |
| `TestSocketClient_ClosesSocketOnError` | `unit` | Attempts to close socket and release resources even after error occurs. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server closes connection prematurely; test monitors file descriptor count | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Exit code `2`; error message indicates read failure; socket close attempted; file descriptor released |
| `TestSocketClient_NoGoroutineLeak` | `unit` | Does not leak goroutines after Send completes. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success; test uses goleak to detect goroutine leaks | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `0`; no goroutines leaked after Send returns |

### Validation Failures — Socket Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_SocketFileNotFound` | `unit` | Returns exit code 2 when socket file does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; no socket file created | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: socket file not found:.*runtime\.sock/` |
| `TestSocketClient_SessionDirNotExist` | `unit` | Returns exit code 2 when session directory does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; no sessions directory | `SessionID=<nonexistent-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: socket file not found:.*<nonexistent-uuid>/` |

### Validation Failures — Connection Refused

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_ConnectionRefused` | `unit` | Returns exit code 2 when Runtime is not listening. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; socket file exists but no server listening | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: connection refused: Runtime is not running for session <uuid>/` |

### Validation Failures — Timeout

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_ConnectionTimeout` | `unit` | Returns exit code 2 when connection times out. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server delays accepting connection; SocketClient configured with 100ms timeout for testing | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: connection timeout after/`, operation completes within milliseconds |
| `TestSocketClient_ReadTimeout` | `unit` | Returns exit code 2 when reading response times out. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server accepts connection but never sends response; SocketClient configured with 100ms timeout for testing | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: connection timeout after/`, operation completes within milliseconds |

### Validation Failures — Send Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_SendMessageIOError` | `unit` | Returns exit code 2 when sending message fails with I/O error. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server closes connection immediately after accepting | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: failed to send message:/` |

### Validation Failures — Read Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|---|
| `TestSocketClient_ReadResponseIOError` | `unit` | Returns exit code 2 when reading response fails with I/O error. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server closes connection after receiving message but before sending response | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: failed to read response:/` |
| `TestSocketClient_ConnectionClosedByRuntime` | `unit` | Returns exit code 2 when Runtime closes connection without response. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server closes connection immediately after receiving message | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`, error message matches `/Error: failed to read response:.*connection closed/` |

### Validation Failures — Malformed Response

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|---|
| `TestSocketClient_MalformedJSON` | `unit` | Returns exit code 3 when response is invalid JSON. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{invalid\n` | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `3`, error message matches `/Error: malformed response from Runtime:/` |
| `TestSocketClient_MissingStatusField` | `unit` | Returns exit code 3 when response JSON is valid but missing status field. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"message":"ok"}\n` | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `3`, error message matches `/Error: response missing 'status' field/` |
| `TestSocketClient_InvalidStatusValue` | `unit` | Returns exit code 3 when response status is not "success" or "error". | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"unknown","message":"test"}\n` | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `3`, error message matches `/Error: invalid response status 'unknown'/` |

### Validation Failures — Socket Close Warning

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_CloseSocketFailsAfterSuccess` | `unit` | Prints warning to stderr when closing socket fails after successful operation. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket setup to fail on close | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `0`; stderr contains `/Warning: failed to close socket:/`; exit code unchanged |
| `TestSocketClient_CloseSocketFailsAfterError` | `unit` | Prints warning but preserves original exit code when closing socket fails after error. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with error status; close operation fails | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `3`; stderr contains both original error and `/Warning: failed to close socket:/`; exit code is `3` (original error preserved) |

### Validation Failures — Input Values

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_InvalidUUIDFormat` | `unit` | Proceeds with invalid UUID format without validation. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `SessionID="not-a-uuid"`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Computes socket path using invalid UUID; returns exit code `2` with socket file not found error |
| `TestSocketClient_EmptyProjectRoot` | `unit` | Returns error when project root path is empty. | Temporary test directory created programmatically within test fixture | `SessionID=<test-uuid>`, `ProjectRoot=""`, `RuntimeMessage{...}` | Returns exit code `2`; error indicates socket path computation failure or file not found |
| `TestSocketClient_GetSocketPathError` | `unit` | Returns exit code 2 when StorageLayout.GetRuntimeSocketPath fails. | Temporary test directory created programmatically within test fixture; mock StorageLayout injected to return error | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `2`; error message indicates socket path computation failed |

### Boundary Values — Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_EmptyClaudeSessionID` | `unit` | Accepts empty ClaudeSessionID in message. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{Type:"event",ClaudeSessionID:"",Payload:{...}}` | Returns exit code `0`; message sent with empty `claudeSessionID` field |
| `TestSocketClient_MultipleResponsesIgnored` | `unit` | Reads first response and closes connection, ignoring subsequent responses. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server sends two JSON responses | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Returns exit code `0` with first response; connection closed; second response never read |
| `TestSocketClient_LargePayload` | `unit` | Handles reasonably large JSON payload in message. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `RuntimeMessage` with `Payload` containing 1MB of JSON data | Returns exit code `0`; entire payload transmitted successfully |
| `TestSocketClient_VeryLargePayload` | `unit` | Returns error when payload exceeds reasonable limits. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server tracks message size | `RuntimeMessage` with `Payload` containing 100MB of JSON data | Returns exit code `2`; error indicates send failure due to size limits or timeout |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_RepeatedSend` | `unit` | Multiple Send invocations produce independent results. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds to multiple connections | Call `Send` three times with same inputs | All three calls succeed independently; each opens new connection; all return exit code `0` |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_ConcurrentSend` | `race` | Multiple goroutines send messages concurrently with different sessions. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid1>/` and `.spectra/sessions/<uuid2>/` directories created inside test fixture; mock socket servers handle multiple concurrent connections | 20 goroutines call `Send` with different session IDs and messages | All calls succeed independently; no data races detected; all return exit code `0` |
| `TestSocketClient_ConcurrentSendSameSession` | `race` | Multiple goroutines send messages to same session concurrently. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server handles multiple concurrent connections to same socket | 10 goroutines each call `Send` with same session ID but different messages | All calls succeed independently; no data races detected; all return exit code `0`; each gets independent socket connection |
| `TestSocketClient_ConcurrentWithSocketDeletion` | `race` | Send operations handle concurrent socket file deletion gracefully. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; socket file deleted by another goroutine during Send operations | Multiple goroutines call `Send`; one goroutine deletes socket file after some succeed | Early calls may succeed; later calls return exit code `2` with socket not found error; no data races or panics |
| `TestSocketClient_ConcurrentConnectionAndCleanup` | `race` | Socket creation, connection, and cleanup operations are race-free. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server with instrumented lifecycle hooks | 20 goroutines call `Send` concurrently | No races detected during socket creation, connection, send, receive, or close operations; all file descriptors properly released |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSocketClient_UsesStorageLayout` | `unit` | Calls StorageLayout.GetRuntimeSocketPath with correct parameters. | Temporary test directory created programmatically within test fixture; mock StorageLayout injected | `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage{...}` | Mock StorageLayout receives call with `projectRoot=<test-fixture>`, `sessionID=<test-uuid>`; returned socket path used for connection |
