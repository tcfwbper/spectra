# Test Specification: `error.go`

## Source File Under Test
`cmd/spectra_agent/error.go`

## Test File
`cmd/spectra_agent/error_test.go`

---

## `ErrorCommand`

### Happy Path — Report Error

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_ReportSuccessWithMessage` | `unit` | Successfully reports error with message and default detail. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server created inside test fixture that responds with `{"status":"success","message":"ok"}\n` | `args=["error", "test error message", "--session-id", "<test-uuid>"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0` |
| `TestErrorCommand_ReportSuccessWithDetail` | `unit` | Successfully reports error with message and detail object. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test error", "--session-id", "<test-uuid>", "--detail", "{\"stack\":\"...\",\"code\":500}"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives correct RuntimeMessage with detail object |
| `TestErrorCommand_ReportSuccessWithNullDetail` | `unit` | Successfully reports error with null detail. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test error", "--session-id", "<test-uuid>", "--detail", "null"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives RuntimeMessage with `detail: null` |
| `TestErrorCommand_ReportSuccessDefaultDetail` | `unit` | Successfully reports error with default empty object detail. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test error", "--session-id", "<test-uuid>"]` (no --detail) | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives RuntimeMessage with `detail: {}` |
| `TestErrorCommand_WithClaudeSessionID` | `unit` | Successfully reports error with Claude session ID. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test error", "--session-id", "<test-uuid>", "--claude-session-id", "550e8400-e29b-41d4-a716-446655440000"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives RuntimeMessage with `claudeSessionID: "550e8400-e29b-41d4-a716-446655440000"` |
| `TestErrorCommand_DefaultClaudeSessionID` | `unit` | Successfully reports error with default empty Claude session ID. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test error", "--session-id", "<test-uuid>"]` (no --claude-session-id) | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives RuntimeMessage with `claudeSessionID: ""` |
| `TestErrorCommand_IgnoresRuntimeMessage` | `unit` | Uses default success message regardless of Runtime response message. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"success","message":"Session marked as failed"}\n` | `args=["error", "test error", "--session-id", "<test-uuid>"]` | Prints `"Error reported successfully"` to stdout (not custom message); returns exit code `0` |

### Happy Path — Message Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_ConstructsRuntimeMessage` | `unit` | Constructs correct RuntimeMessage JSON structure. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server captures sent message | `args=["error", "test error", "--session-id", "<test-uuid>", "--claude-session-id", "session-123", "--detail", "{\"key\":\"value\"}"]` | Mock server receives valid JSON with structure: `{"type":"error","claudeSessionID":"session-123","payload":{"message":"test error","detail":{"key":"value"}}}`, terminated by newline |
| `TestErrorCommand_MessageWithWhitespace` | `unit` | Accepts message with whitespace and sends successfully. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "   test error   ", "--session-id", "<test-uuid>"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; message sent with whitespace preserved |
| `TestErrorCommand_MessageWithSpecialChars` | `unit` | Accepts message with special characters. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "error: \"critical\" <failure>", "--session-id", "<test-uuid>"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; message properly JSON-escaped |

### Validation Failures — Missing Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_MissingMessage` | `unit` | Returns exit code 1 when message argument is missing. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "--session-id", "<test-uuid>"]` (no message) | Returns exit code `1`; stderr matches `/Error: error message is required/` |
| `TestErrorCommand_EmptyMessage` | `unit` | Returns exit code 1 when message is empty string. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "", "--session-id", "<test-uuid>"]` | Returns exit code `1`; stderr matches `/Error: error message is required/` |

### Validation Failures — Invalid Detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_DetailNotJSON` | `unit` | Returns exit code 1 when detail is invalid JSON. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "{invalid}"]` | Returns exit code `1`; stderr matches `/Error: --detail must be a JSON object or null/` |
| `TestErrorCommand_DetailPrimitiveString` | `unit` | Returns exit code 1 when detail is JSON string primitive. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "\"error detail string\""]` | Returns exit code `1`; stderr matches `/Error: --detail must be a JSON object or null/` |
| `TestErrorCommand_DetailPrimitiveNumber` | `unit` | Returns exit code 1 when detail is JSON number primitive. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "42"]` | Returns exit code `1`; stderr matches `/Error: --detail must be a JSON object or null/` |
| `TestErrorCommand_DetailPrimitiveBoolean` | `unit` | Returns exit code 1 when detail is JSON boolean primitive. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "true"]` | Returns exit code `1`; stderr matches `/Error: --detail must be a JSON object or null/` |
| `TestErrorCommand_DetailArray` | `unit` | Returns exit code 1 when detail is JSON array. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "[1,2,3]"]` | Returns exit code `1`; stderr matches `/Error: --detail must be a JSON object or null/` |

### Validation Failures — Socket Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_SocketNotFound` | `unit` | Returns exit code 2 when socket file does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; no socket file created | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `2`; stderr matches `/Error: socket file not found:.*runtime\.sock/` |
| `TestErrorCommand_ConnectionRefused` | `unit` | Returns exit code 2 when Runtime is not listening. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; socket file exists but no server listening | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `2`; stderr matches `/Error: connection refused: Runtime is not running for session/` |
| `TestErrorCommand_ConnectionTimeout` | `unit` | Returns exit code 2 when connection times out. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server delays accepting connection; SocketClient configured with 100ms timeout override for testing | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `2`; stderr matches `/Error: connection timeout after/`; operation completes within milliseconds |
| `TestErrorCommand_SendIOError` | `unit` | Returns exit code 2 when sending message fails. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server closes connection immediately after accepting | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `2`; stderr matches `/Error: failed to send message:/` |
| `TestErrorCommand_ReadIOError` | `unit` | Returns exit code 2 when reading response fails. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server closes connection after receiving message but before sending response; SocketClient configured with 100ms timeout override for testing | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `2`; stderr matches `/Error: failed to read response:/`; operation completes within milliseconds |

### Validation Failures — Runtime Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_RuntimeError` | `unit` | Returns exit code 3 when Runtime responds with error status. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"error","message":"session not found: abc-123"}\n` | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `3`; stderr matches `/Error: session not found: abc-123/` |
| `TestErrorCommand_SessionTerminated` | `unit` | Returns exit code 3 when session is already terminated. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"error","message":"session terminated: session is in 'completed' status"}\n` | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `3`; stderr matches `/Error: session terminated:/` |
| `TestErrorCommand_ClaudeSessionIDMismatch` | `unit` | Returns exit code 3 when Claude session ID does not match. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"error","message":"claude session ID mismatch: expected 550e8400-e29b-41d4-a716-446655440000 but got 660e8400-e29b-41d4-a716-446655440001"}\n` | `args=["error", "test", "--session-id", "<test-uuid>", "--claude-session-id", "660e8400-e29b-41d4-a716-446655440001"]` | Returns exit code `3`; stderr matches `/Error: claude session ID mismatch:/` |
| `TestErrorCommand_InvalidClaudeSessionIDForHumanNode` | `unit` | Returns exit code 3 when Claude session ID provided for human node. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"status":"error","message":"invalid claude session ID for human node: must be empty"}\n` | `args=["error", "test", "--session-id", "<test-uuid>", "--claude-session-id", "550e8400-e29b-41d4-a716-446655440000"]` | Returns exit code `3`; stderr matches `/Error: invalid claude session ID for human node: must be empty/` |
| `TestErrorCommand_MalformedRuntimeResponse` | `unit` | Returns exit code 3 when Runtime response is malformed JSON. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{invalid\n` | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `3`; stderr matches `/Error: malformed response from Runtime:/` |
| `TestErrorCommand_MissingStatusField` | `unit` | Returns exit code 3 when Runtime response missing status field. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with `{"message":"ok"}\n` | `args=["error", "test", "--session-id", "<test-uuid>"]` | Returns exit code `3`; stderr matches `/Error: response missing 'status' field/` |

### Boundary Values — Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_WhitespaceOnlyMessage` | `unit` | Accepts whitespace-only message and sends successfully. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "   ", "--session-id", "<test-uuid>"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; whitespace-only message sent |
| `TestErrorCommand_VeryLongMessage` | `unit` | Accepts very long error message. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "<10000-character-string>", "--session-id", "<test-uuid>"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0` |
| `TestErrorCommand_EmptyDetailObject` | `unit` | Accepts empty detail object explicitly. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "{}"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives `detail: {}` |
| `TestErrorCommand_ComplexDetailObject` | `unit` | Accepts complex nested detail object. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test", "--session-id", "<test-uuid>", "--detail", "{\"stack\":[{\"file\":\"a.go\",\"line\":10}],\"context\":{\"user\":\"test\"}}"]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; complex detail transmitted correctly |
| `TestErrorCommand_EmptyClaudeSessionID` | `unit` | Accepts empty Claude session ID explicitly. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test", "--session-id", "<test-uuid>", "--claude-session-id", ""]` | Prints `"Error reported successfully"` to stdout; returns exit code `0`; mock server receives `claudeSessionID: ""` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_UsesSocketClient` | `unit` | Delegates socket communication to SocketClient.Send. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock SocketClient injected | `args=["error", "test error", "--session-id", "<test-uuid>", "--detail", "{\"key\":\"value\"}"]` | Mock SocketClient.Send called once with correct parameters: `SessionID=<test-uuid>`, `ProjectRoot=<test-fixture>`, `RuntimeMessage` with `Type="error"`, `Payload.message="test error"`, `Payload.detail={"key":"value"}` |
| `TestErrorCommand_ReceivesRootCommandContext` | `unit` | Receives session ID and project root from root command. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test", "--session-id", "<test-uuid>"]` | Error subcommand receives session ID and project root from root command initialization; uses them for SocketClient call |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_RepeatedInvocation` | `unit` | Multiple invocations with same arguments produce consistent results. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds to multiple connections | Execute error command three times with identical arguments | All three invocations return exit code `0` and produce identical output; no state leakage between invocations |

### Error Output Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCommand_ErrorPrefixFormat` | `unit` | All error messages are prefixed with "Error: ". | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "--session-id", "<test-uuid>"]` (missing message) | stderr output starts with `/^Error: /` |
| `TestErrorCommand_ErrorOutputToStderr` | `unit` | Error messages printed to stderr, not stdout. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | `args=["error", "--session-id", "<test-uuid>"]` (missing message) | Error message appears in stderr; stdout is empty |
| `TestErrorCommand_SuccessOutputToStdout` | `unit` | Success message printed to stdout, not stderr. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory created inside test fixture; mock socket server responds with success | `args=["error", "test", "--session-id", "<test-uuid>"]` | Success message appears in stdout; stderr is empty |
