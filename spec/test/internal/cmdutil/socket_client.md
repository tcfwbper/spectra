# Test Specification: `socket_client_test.go`

## Source File Under Test
`internal/cmdutil/socket_client.go`

## Test File
`internal/cmdutil/socket_client_test.go`

---

## `Send`

### Happy Path — Send

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSend_SuccessResponse` | `unit` | Returns parsed Response with exit code 0 when Runtime responds with status "success". | Create a temporary directory with a Unix domain socket listener that accepts one connection, reads a line, and responds with `{"status":"success","message":"ok"}\n`. Stub `StorageLayout.GetRuntimeSocketPath` to return the socket path. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{\"type\":\"event\"}")` | Returns `response.Status == "success"`, `response.Message == "ok"`, `exitCode == 0`, `err == nil` |
| `TestSend_ErrorStatusResponse` | `unit` | Returns parsed Response with exit code 3 when Runtime responds with status "error". | Create a temporary Unix socket listener that responds with `{"status":"error","message":"session not found"}\n`. Stub `StorageLayout.GetRuntimeSocketPath` to return the socket path. | `sessionID="sess-2"`, `projectRoot="/tmp/project"`, `message=[]byte("{\"type\":\"event\"}")` | Returns `response.Status == "error"`, `response.Message == "session not found"`, `exitCode == 3`, `err == nil` |
| `TestSend_SendsMessageWithNewline` | `unit` | Writes the message bytes followed by a newline to the socket. | Create a temporary Unix socket listener that captures received bytes before responding with a success JSON. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-3"`, `projectRoot="/tmp/project"`, `message=[]byte("{\"type\":\"event\"}")` | Listener receives `"{\"type\":\"event\"}\n"`; function returns `exitCode == 0` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSend_SocketFileNotFound` | `unit` | Returns exit code 2 when socket file does not exist. | Stub `StorageLayout.GetRuntimeSocketPath` to return a non-existent path in a temp directory. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 2`, `err` contains `"socket file not found: <path>"` |
| `TestSend_ConnectionRefused` | `unit` | Returns exit code 2 when socket file exists but no listener is active. | Create a temp file at the socket path (not a real listener). Stub `StorageLayout.GetRuntimeSocketPath` to return that path. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 2`, `err` contains `"connection refused"` |
| `TestSend_ConnectionTimeout` | `unit` | Returns exit code 2 when connection does not complete within the timeout. | Create a Unix socket listener that accepts but never responds. Inject a short deadline (e.g., 50ms) via a test-only timeout option or fake clock to avoid real waiting. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 2`, `err` contains `"connection timeout"` |
| `TestSend_WriteFails` | `unit` | Returns exit code 2 when write to socket fails. | Create a Unix socket listener that accepts and immediately closes the connection before a write can complete. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 2`, `err` contains `"failed to send message"` |
| `TestSend_ReadFails` | `unit` | Returns exit code 2 when read from socket fails. | Create a Unix socket listener that accepts, reads the message, then closes the connection without sending a response. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 2`, `err` contains `"failed to read response"` |
| `TestSend_MalformedJSON` | `unit` | Returns exit code 3 when response is not valid JSON. | Create a Unix socket listener that responds with `"not json\n"`. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 3`, `err` contains `"malformed response from Runtime"` |
| `TestSend_MissingStatusField` | `unit` | Returns exit code 3 when response JSON lacks "status" field. | Create a Unix socket listener that responds with `{"message":"hello"}\n`. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 3`, `err` contains `"response missing 'status' field"` |
| `TestSend_InvalidStatusValue` | `unit` | Returns exit code 3 when response status is not "success" or "error". | Create a Unix socket listener that responds with `{"status":"unknown"}\n`. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 3`, `err` contains `"invalid response status 'unknown'"` |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSend_ClosesConnectionOnSuccess` | `unit` | Connection is closed after a successful response. | Create a Unix socket listener that responds with success JSON and tracks whether the client-side connection is closed (detects EOF on server after response). Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Server observes connection closure (EOF) after response is sent; function returns `exitCode == 0` |
| `TestSend_ClosesConnectionOnError` | `unit` | Connection is closed after a transport error. | Create a Unix socket listener that accepts and responds with malformed data. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Server observes connection closure after error; function returns `exitCode == 3` |
| `TestSend_CloseFailureWarning` | `unit` | Close failure prints warning to stderr but does not alter exit code. | Create a Unix socket listener that responds with success JSON. Inject a connection wrapper where `Close()` returns an error. Stub `StorageLayout.GetRuntimeSocketPath`. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `exitCode == 0`; stderr output contains `"Warning: failed to close socket"` |

### Boundary Values — sessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSend_MalformedSessionID` | `unit` | Proceeds with malformed sessionID; connection fails with exit code 2. | Stub `StorageLayout.GetRuntimeSocketPath` to return a path derived from the malformed ID (which does not exist). | `sessionID="not-a-uuid"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | Returns `response == nil`, `exitCode == 2`, `err` contains `"socket file not found"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSend_CallsGetRuntimeSocketPath` | `unit` | Calls StorageLayout.GetRuntimeSocketPath with correct arguments. | Mock `StorageLayout` to record calls and return a non-existent path. | `sessionID="sess-1"`, `projectRoot="/tmp/project"`, `message=[]byte("{}")` | `GetRuntimeSocketPath` called once with `projectRoot="/tmp/project"`, `sessionID="sess-1"` |
