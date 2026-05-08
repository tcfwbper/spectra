# Test Specification: `send_and_handle_test.go`

## Source File Under Test
`internal/cmdutil/send_and_handle.go`

## Test File
`internal/cmdutil/send_and_handle_test.go`

---

## `SendAndHandle`

### Happy Path — SendAndHandle

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSendAndHandle_Success` | `unit` | Returns exit code 0 with successText on stdout when SocketClient returns success. | Mock `SocketClient.Send` to return `(&Response{Status:"success", Message:"done"}, 0, nil)`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=validStruct{}`, `successText="Event emitted"` | Returns `exitCode == 0`, `stdout == "Event emitted"`, `stderr == ""` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSendAndHandle_SerializationFailure` | `unit` | Returns exit code 1 when message cannot be serialized to JSON. | None (use a message struct containing an unencodable field such as a channel). | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=unserializableStruct{}`, `successText="ok"` | Returns `exitCode == 1`, `stdout == ""`, `stderr` contains `"failed to serialize message"` |
| `TestSendAndHandle_TransportError` | `unit` | Returns exit code 2 when SocketClient returns transport error. | Mock `SocketClient.Send` to return `(nil, 2, errors.New("socket file not found: /path"))`. Mock `ErrorFormatter.FormatError` to return `"Error: socket file not found: /path"`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=validStruct{}`, `successText="ok"` | Returns `exitCode == 2`, `stdout == ""`, `stderr == "Error: socket file not found: /path"` |
| `TestSendAndHandle_RuntimeErrorWithResponse` | `unit` | Returns exit code 3 with formatted error message when SocketClient returns error response. | Mock `SocketClient.Send` to return `(&Response{Status:"error", Message:"session not found: abc"}, 3, nil)`. Mock `ErrorFormatter.FormatError` to return `"Error: session not found: abc"`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=validStruct{}`, `successText="ok"` | Returns `exitCode == 3`, `stdout == ""`, `stderr == "Error: session not found: abc"` |
| `TestSendAndHandle_MalformedResponseNilResponse` | `unit` | Returns exit code 3 with SocketClient error when response is nil (malformed). | Mock `SocketClient.Send` to return `(nil, 3, errors.New("malformed response from Runtime: ..."))`. Mock `ErrorFormatter.FormatError` to return `"Error: malformed response from Runtime: ..."`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=validStruct{}`, `successText="ok"` | Returns `exitCode == 3`, `stdout == ""`, `stderr == "Error: malformed response from Runtime: ..."` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSendAndHandle_SerializesMessageToJSON` | `unit` | Passes JSON-serialized message bytes to SocketClient.Send. | Mock `SocketClient.Send` to capture the `message` argument and return success. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=TestMsg{Type:"event", ClaudeSessionID:"c-1"}`, `successText="ok"` | `SocketClient.Send` called with `message` equal to `[]byte("{\"type\":\"event\",\"claudeSessionID\":\"c-1\"}")` (or equivalent JSON encoding) |
| `TestSendAndHandle_PassesSessionIDAndProjectRoot` | `unit` | Passes sessionID and projectRoot unchanged to SocketClient.Send. | Mock `SocketClient.Send` to capture arguments and return success. | `sessionID="my-session"`, `projectRoot="/home/user/project"`, `message=validStruct{}`, `successText="ok"` | `SocketClient.Send` called with `sessionID="my-session"`, `projectRoot="/home/user/project"` |
| `TestSendAndHandle_CallsFormatErrorOnRuntimeError` | `unit` | Calls ErrorFormatter.FormatError with response message on exit code 3 with non-nil response. | Mock `SocketClient.Send` to return `(&Response{Status:"error", Message:"bad request"}, 3, nil)`. Mock `ErrorFormatter.FormatError` to record call and return formatted string. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=validStruct{}`, `successText="ok"` | `FormatError` called once with `msg="bad request"` |
| `TestSendAndHandle_DoesNotCallFormatErrorOnSuccess` | `unit` | Does not call ErrorFormatter.FormatError when operation succeeds. | Mock `SocketClient.Send` to return success. Mock `ErrorFormatter.FormatError` to record calls. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=validStruct{}`, `successText="ok"` | `FormatError` not called |

---

## `PublicSendAndHandle`

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPublicSendAndHandle_DelegatesToSendAndHandle` | `unit` | Delegates to SendAndHandle with production SocketClient and FormatError. | Stub `storage.GetRuntimeSocketPath` to return a known path. Set up a real Unix socket listener at that path that returns a success response. | `sessionID="sess-1"`, `projectRoot=tmpDir`, `message=validStruct{Type:"event", ClaudeSessionID:"c-1"}`, `successText="Done"` | Returns `exitCode == 0`, `stdout == "Done"`, `stderr == ""` |
| `TestPublicSendAndHandle_PropagatesTransportError` | `unit` | Propagates transport error when socket is unreachable. | Create a temp directory as projectRoot. Do not create a socket file (socket path does not exist). | `sessionID="sess-1"`, `projectRoot=tmpDir`, `message=validStruct{}`, `successText="ok"` | Returns `exitCode == 2`, `stdout == ""`, `stderr` contains error about socket |
| `TestPublicSendAndHandle_UsesFormatErrorForErrorFormatting` | `unit` | Uses FormatError to format error messages in stderr output. | Set up a real Unix socket listener that returns an error response `{Status:"error", Message:"bad session"}`. Stub `storage.GetRuntimeSocketPath` to return socket path. | `sessionID="sess-1"`, `projectRoot=tmpDir`, `message=validStruct{}`, `successText="ok"` | Returns `exitCode == 3`, `stderr` contains `"Error: bad session"` |
