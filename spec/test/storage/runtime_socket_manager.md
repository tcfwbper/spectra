# Test Specification: `runtime_socket_manager_test.go`

## Source File Under Test
`storage/runtime_socket_manager.go`

## Test File
`storage/runtime_socket_manager_test.go`

---

## `RuntimeSocketManager`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeSocketManager_StoresSocketPath` | `unit` | Constructor composes socket path via StorageLayout and stores it. | Stub `StorageLayout.GetRuntimeSocketPath` to return a known path. | `projectRoot="/tmp/proj"`, `sessionUUID="aaaa-bbbb"`, `logger=mockLogger` | Returns non-nil manager; internal socket path equals the stubbed path |
| `TestNewRuntimeSocketManager_NoIO` | `unit` | Constructor does not perform filesystem I/O. | Stub `StorageLayout.GetRuntimeSocketPath`. Do not create any directory on disk. | `projectRoot="/nonexistent/path"`, `sessionUUID="aaaa-bbbb"`, `logger=mockLogger` | Returns non-nil manager without error or panic |

### Happy Path — CreateSocket

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_CreatesSocketFile` | `unit` | Creates a Unix domain socket file at the expected path. | Create a temp directory to serve as the session directory. Construct manager with socket path pointing inside the temp directory. | Call `CreateSocket()` | Returns nil; socket file exists at path |
| `TestCreateSocket_FilePermissions` | `unit` | Socket file is created with 0600 permissions. | Create a temp directory. Construct manager with socket path inside it. | Call `CreateSocket()` | Returns nil; socket file permissions are `0600` |

### Happy Path — Listen

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_ReturnsChannels` | `unit` | Returns error channel, done channel, and nil error on success. | Create temp directory; call `CreateSocket()` successfully. Provide a no-op `MessageHandler` and a context. | Call `Listen(ctx, handler)` | Returns non-nil `listenerErrCh`, non-nil `listenerDoneCh`, nil error |
| `TestListen_AcceptsConnection` | `unit` | Accepts a client connection on the socket. | Create temp directory; call `CreateSocket()` and `Listen()`. Connect a client via Unix domain socket dial. | Client dials the socket path | Client connection succeeds without error |
| `TestListen_DispatchesToHandler` | `unit` | Dispatches a valid message to MessageHandler with correct sessionUUID. | Create temp directory; call `CreateSocket()` and `Listen()` with a mock `MessageHandler`. Connect client and send valid JSON message. | Client sends `{"type":"event","payload":{"key":"val"}}` | Mock `MessageHandler.Handle` called with the stored `sessionUUID` and a `RuntimeMessage` with type `"event"` |
| `TestListen_SendsResponseToClient` | `unit` | Serializes and sends RuntimeResponse back to the client. | Create temp directory; call `CreateSocket()` and `Listen()` with a mock `MessageHandler` that returns a success response. Connect client and send valid message. | Client sends valid message and reads response | Client receives `{"status":"success","message":"ok"}\n` |
| `TestListen_ClosesConnectionAfterResponse` | `unit` | Connection is closed after sending the response. | Create temp directory; call `CreateSocket()` and `Listen()`. Connect client and send valid message. | Client sends valid message, reads response, then attempts second read | Second read returns EOF or connection closed |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_SocketAlreadyExists` | `unit` | Returns descriptive error when socket file already exists. | Create temp directory; create a file at the socket path before calling `CreateSocket()`. | Call `CreateSocket()` | Returns error containing `"runtime socket file already exists:"` and the socket path |
| `TestCreateSocket_DirectoryMissing` | `unit` | Returns error when parent directory does not exist. | Construct manager with socket path in a nonexistent directory. | Call `CreateSocket()` | Returns error containing `"failed to create runtime socket:"` |
| `TestCreateSocket_PermissionDenied` | `unit` | Returns error when directory is not writable. | Create temp directory with permissions `0555` (read/execute only). Construct manager with socket path inside it. | Call `CreateSocket()` | Returns error containing `"failed to create runtime socket:"` |
| `TestListen_BeforeCreateSocket` | `unit` | Returns error when Listen is called before CreateSocket. | Construct manager without calling `CreateSocket()`. | Call `Listen(ctx, handler)` | Returns nil channels and error `"runtime socket not created: call CreateSocket() first"` |
| `TestListen_BindFailure` | `unit` | Returns synchronous error when bind/listen fails. | Create temp directory; call `CreateSocket()` then delete the socket file externally before `Listen()`. | Call `Listen(ctx, handler)` | Returns error containing `"failed to listen on runtime socket:"` |

### Happy Path — DeleteSocket

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_RemovesFile` | `unit` | Deletes the socket file from the filesystem. | Create temp directory; call `CreateSocket()` and `Listen()`. | Call `DeleteSocket(ctx)` | Socket file no longer exists at path |
| `TestDeleteSocket_ClosesListener` | `unit` | Stops the listener and closes the done channel. | Create temp directory; call `CreateSocket()` and `Listen()`. | Call `DeleteSocket(ctx)` then read from `listenerDoneCh` | `listenerDoneCh` is closed (read returns immediately) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_Idempotent` | `unit` | Calling DeleteSocket multiple times does not error. | Create temp directory; call `CreateSocket()` and `Listen()`. Call `DeleteSocket(ctx)` once. | Call `DeleteSocket(ctx)` again | Returns without error or panic |
| `TestDeleteSocket_FileAlreadyGone` | `unit` | Returns without error when socket file does not exist. | Create temp directory; call `CreateSocket()` and `Listen()`. Manually remove socket file. | Call `DeleteSocket(ctx)` | Returns without error; no warning logged |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPerConnection_MalformedJSON` | `unit` | Rejects malformed JSON and sends error response. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends `not valid json\n` | Client receives `{"status":"error","message":"..."}` with newline; logger.Warn called with message containing `"malformed JSON"` |
| `TestPerConnection_MissingTypeField` | `unit` | Rejects message with missing type field. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends `{"payload":{"k":"v"}}\n` | Client receives error response; logger.Warn called with message about missing/invalid type |
| `TestPerConnection_InvalidTypeValue` | `unit` | Rejects message with unrecognized type value. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends `{"type":"unknown","payload":{"k":"v"}}\n` | Client receives error response; logger.Warn called |
| `TestPerConnection_MissingPayload` | `unit` | Rejects message with missing payload field. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends `{"type":"event"}\n` | Client receives error response; logger.Warn called |
| `TestPerConnection_PayloadNotObject` | `unit` | Rejects message with payload that is not a JSON object. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends `{"type":"event","payload":"string"}\n` | Client receives error response; logger.Warn called |
| `TestPerConnection_PayloadArray` | `unit` | Rejects message with payload that is a JSON array. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends `{"type":"event","payload":[1,2]}\n` | Client receives error response; logger.Warn called |
| `TestPerConnection_ClientClosesWithoutSending` | `unit` | No warning logged when client disconnects before sending a message. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client connects then immediately closes the connection. | No logger.Warn call; connection handler exits cleanly |

### Boundary Values — Message Size

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPerConnection_ExceedsMaxSize` | `unit` | Rejects message exceeding 10 MB limit. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger. Connect client. | Client sends a message larger than 10 MB terminated by newline | Client receives error response if possible; logger.Warn called with message containing `"message size exceeds 10 MB limit"`; MessageHandler not invoked |
| `TestPerConnection_AtMaxSize` | `unit` | Accepts message at exactly 10 MB. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. Connect client. | Client sends a valid JSON message that is exactly 10 MB (including newline) | MessageHandler is invoked; client receives success response |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPerConnection_EmptyClaudeSessionID` | `unit` | Accepts message without claudeSessionID field (defaults to empty string). | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. Connect client. | Client sends `{"type":"event","payload":{"k":"v"}}\n` (no claudeSessionID) | MessageHandler.Handle invoked with RuntimeMessage having empty ClaudeSessionID |
| `TestPerConnection_WithClaudeSessionID` | `unit` | Passes claudeSessionID to RuntimeMessage when present. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. Connect client. | Client sends `{"type":"event","payload":{"k":"v"},"claudeSessionID":"sess-123"}\n` | MessageHandler.Handle invoked with RuntimeMessage having ClaudeSessionID `"sess-123"` |
| `TestPerConnection_EmptyMessageInResponse` | `unit` | Sends response with empty message field as-is. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler` that returns a success response with empty Message. Connect client. | Client sends valid message and reads response | Client receives `{"status":"success","message":""}\n` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestConstruction_CallsStorageLayout` | `unit` | Constructor calls StorageLayout.GetRuntimeSocketPath with correct args. | Stub `StorageLayout.GetRuntimeSocketPath` to capture arguments. | `projectRoot="/proj"`, `sessionUUID="uuid-1"` | `GetRuntimeSocketPath` called with `"/proj"` and `"uuid-1"` |
| `TestPerConnection_InvokesNewRuntimeMessage` | `unit` | Constructs RuntimeMessage via NewRuntimeMessage after protocol validation. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. Connect client. | Client sends `{"type":"error","payload":{"detail":"x"},"claudeSessionID":"c1"}\n` | MessageHandler.Handle receives a RuntimeMessage with Type `"error"`, Payload containing `{"detail":"x"}`, ClaudeSessionID `"c1"` |
| `TestDeleteSocket_LogsOnFileDeletionFailure` | `unit` | Logs a warning when socket file deletion fails. | Create temp directory; call `CreateSocket()` and `Listen()`. Make socket file undeletable (e.g., remove write permission on parent dir). | Call `DeleteSocket(ctx)` | logger.Warn called with message containing `"failed to delete runtime socket:"` |
| `TestPerConnection_LogsOnSendFailure` | `unit` | Logs a warning when response send fails due to client disconnect. | Create temp directory; call `CreateSocket()` and `Listen()` with mock logger and a slow `MessageHandler`. Connect client, send valid message, then close client connection before handler returns. | MessageHandler returns after client disconnect | logger.Warn called with message containing `"failed to send response to client:"` |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_MultipleSimultaneousConnections` | `unit` | Handles multiple concurrent client connections in separate goroutines. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. | Connect 3 clients simultaneously, each sending a valid message | All 3 clients receive responses; MessageHandler invoked 3 times |
| `TestPerConnection_IsolationOnError` | `unit` | Malformed message on one connection does not affect others. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. | Client A sends malformed JSON; Client B sends valid message concurrently | Client A receives error response; Client B receives success response from MessageHandler |
| `TestPerConnection_HandlerPanicCrashesGoroutine` | `unit` | MessageHandler panic is not recovered by RuntimeSocketManager. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler` that panics on invocation. Connect two clients. | Client A sends valid message (triggers panic); Client B sends valid message concurrently | Client A receives no graceful error response (connection closed or broken); Client B receives normal response unaffected |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_ClosesActiveConnections` | `unit` | DeleteSocket closes all active connections immediately. | Create temp directory; call `CreateSocket()` and `Listen()` with a slow `MessageHandler` (blocks on a channel). Connect a client and send a valid message (handler blocks). | Call `DeleteSocket(ctx)` while handler is blocked | Client connection is closed (read returns error); `listenerDoneCh` is closed |
| `TestListen_ContextCancellation` | `unit` | Cancelling context stops the accept loop. | Create temp directory; call `CreateSocket()` and `Listen()` with a cancellable context. | Cancel the context | `listenerDoneCh` is closed; new connection attempts are refused |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPerConnection_SingleRequestResponse` | `unit` | Only first message per connection is processed. | Create temp directory; call `CreateSocket()` and `Listen()` with mock `MessageHandler`. Connect client. | Client sends two valid messages on the same connection | Only one invocation of MessageHandler.Handle; client receives one response then connection closed |

### Asynchronous Flow

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_ListenerErrChannel` | `unit` | Listener error channel receives fatal accept-loop error. | Create temp directory; call `CreateSocket()` and `Listen()`. Simulate accept failure by closing the underlying listener externally (not via DeleteSocket). | Read from `listenerErrCh` | Receives error containing `"listener accept loop failed:"` |
| `TestListen_DoneChannelClosedAfterDelete` | `unit` | Done channel is closed after DeleteSocket completes. | Create temp directory; call `CreateSocket()` and `Listen()`. | Call `DeleteSocket(ctx)` | `listenerDoneCh` is closed (select on it returns immediately) |
