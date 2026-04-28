# Test Specification: `runtime_socket_manager.go`

## Source File Under Test
`storage/runtime_socket_manager.go`

## Test File
`storage/runtime_socket_manager_test.go`

---

## `RuntimeSocketManager`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeSocketManager_New` | `unit` | Constructs RuntimeSocketManager with valid inputs. | Test fixture with `.spectra/sessions/` subdirectory | `ProjectRoot=<test-fixture>`, `SessionUUID=<valid-uuid-v4>` | Returns RuntimeSocketManager instance; no error |
| `TestRuntimeSocketManager_PathResolution` | `unit` | Resolves correct socket path from session UUID. | Test fixture with `.spectra/sessions/<uuid>/` directory | `ProjectRoot=<test-fixture>`, `SessionUUID=<uuid>` | Socket path resolves to `<test-fixture>/.spectra/sessions/<uuid>/runtime.sock` |

### Happy Path — CreateSocket

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_NewSocket` | `unit` | Creates socket file at correct path. | Session directory exists; no socket file | | Returns nil; socket file exists at expected path |
| `TestCreateSocket_Permissions` | `unit` | Creates socket with correct permissions 0600. | Session directory exists; no socket file | | Returns nil; socket file has permissions `0600` (owner read/write only) |

### Happy Path — Listen

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_StartsListener` | `unit` | Starts listener on created socket. | Socket created via `CreateSocket()` | `MessageHandler` callback function | Returns `(listenerErrCh, listenerDoneCh, nil)`; channels are not nil |
| `TestListen_AcceptsConnection` | `unit` | Accepts client connection successfully. | Socket created; listener started | Client connects to socket | Connection accepted; MessageHandler callback invoked |
| `TestListen_MultipleConnections` | `unit` | Accepts multiple client connections sequentially. | Socket created; listener started | 3 clients connect sequentially | All 3 connections accepted; MessageHandler invoked 3 times |

### Happy Path — Receive and MessageHandler Invocation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_ValidEventMessage` | `unit` | Parses valid event message and invokes handler. | Socket created; listener started; mock MessageHandler | JSON: `{"type":"event","payload":{"eventType":"test","message":"msg"}}` followed by newline | MessageHandler invoked with `Type="event"`, `Payload` containing `eventType` and `message`; response sent to client |
| `TestReceive_ValidErrorMessage` | `unit` | Parses valid error message and invokes handler. | Socket created; listener started; mock MessageHandler | JSON: `{"type":"error","payload":{"message":"error msg"}}` followed by newline | MessageHandler invoked with `Type="error"`, `Payload` containing `message`; response sent to client |
| `TestReceive_SessionUUIDExtracted` | `unit` | Extracts session UUID from socket path and passes to handler. | Test fixture with session UUID `<uuid>`; socket created; listener started; mock MessageHandler | Valid message | MessageHandler invoked with `sessionUUID=<uuid>` |
| `TestReceive_ComplexPayload` | `unit` | Handles complex nested payload structure. | Socket created; listener started; mock MessageHandler | JSON with nested payload (min 3 levels: arrays, objects, numbers, booleans, nulls) | MessageHandler invoked with complete payload structure preserved |
| `TestReceive_OptionalFields` | `unit` | Handles message with optional fields. | Socket created; listener started; mock MessageHandler | Event message with optional `claudeSessionID` field | MessageHandler invoked with all fields including optional ones |

### Happy Path — Response Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestResponse_Success` | `unit` | Sends success response to client. | Socket created; listener started; MessageHandler returns `RuntimeResponse{Status:"success", Message:"ok"}` | Valid message from client | Client receives `{"status":"success","message":"ok"}\n` |
| `TestResponse_Error` | `unit` | Sends error response to client. | Socket created; listener started; MessageHandler returns `RuntimeResponse{Status:"error", Message:"failed"}` | Valid message from client | Client receives `{"status":"error","message":"failed"}\n` |
| `TestResponse_EmptyMessage` | `unit` | Sends response with empty message field. | Socket created; listener started; MessageHandler returns `RuntimeResponse{Status:"success", Message:""}` | Valid message from client | Client receives `{"status":"success","message":""}\n` |
| `TestResponse_ConnectionClosedAfterSend` | `unit` | Closes connection after sending response. | Socket created; listener started | Valid message from client | Response sent; connection closed gracefully; subsequent read from client returns EOF |

### Happy Path — DeleteSocket

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_StopsListener` | `unit` | Stops listening when socket deleted. | Socket created; listener started | | Listener stops; `listenerDoneCh` closed; socket file removed |
| `TestDeleteSocket_ClosesActiveConnections` | `unit` | Closes active connections during deletion. | Socket created; listener started; 2 client connections active | | Both connections closed; socket file removed |
| `TestDeleteSocket_RemovesSocketFile` | `unit` | Removes socket file from filesystem. | Socket created | | Returns nil; socket file does not exist |
| `TestDeleteSocket_ListenerNeverStarted` | `unit` | Deletes socket when listener was never started. | Socket created; `Listen()` never called | | Returns nil; socket file removed; no panic from nil channels |

### Idempotency — DeleteSocket

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_SocketDoesNotExist` | `unit` | No error when socket file does not exist. | Session directory exists; no socket file | | Returns nil; no error; no warning logged |
| `TestDeleteSocket_CalledTwice` | `unit` | Second call to DeleteSocket is no-op. | Socket created then deleted | | First call removes socket; second call returns nil without error |

### Validation Failures — CreateSocket (Socket Already Exists)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_SocketAlreadyExists` | `unit` | Returns error when socket file already exists. | Session directory exists; `runtime.sock` file created programmatically within test fixture | | Returns error matching `/runtime socket file already exists:.*runtime\.sock.*This may indicate a previous runtime process did not clean up properly/i` |
| `TestCreateSocket_ResidualSocket` | `unit` | Detects residual socket from previous session. | Old socket file created programmatically within test fixture at target path | | Returns error with cleanup instructions matching `/rm.*runtime\.sock/i` |

### Validation Failures — CreateSocket (Directory Missing)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_SessionDirDoesNotExist` | `unit` | Returns error when session directory missing. | Test fixture; `.spectra/sessions/<uuid>/` does not exist | | Returns error matching `/failed to create runtime socket:.*no such file or directory/i` |

### Validation Failures — CreateSocket (Permission Denied)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_PermissionDenied` | `unit` | Returns error when session directory is read-only. | Session directory created with permissions `0444` inside test fixture | | Returns error matching `/failed to create runtime socket:.*permission denied/i` |

### Validation Failures — CreateSocket (Path Too Long)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_PathTooLong` | `unit` | Returns error when socket path exceeds platform limit. | Deeply nested directory structure (within test fixture) exceeding ~108 characters | | Returns error matching `/failed to create runtime socket:.*file name too long/i` |

### Validation Failures — Listen (Socket Not Created)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_SocketNotCreated` | `unit` | Returns error when called before CreateSocket. | Test fixture; `CreateSocket()` not called | `MessageHandler` callback | Returns `(nil, nil, error)` matching `/runtime socket not created: call CreateSocket\(\) first/i` |

### Validation Failures — Listen (Bind Failure)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_BindFails` | `unit` | Returns error when socket bind fails. | Socket file created but bind operation fails (e.g., mock failure) | `MessageHandler` callback | Returns `(nil, nil, error)` matching `/failed to listen on runtime socket:/i`; returned channels are nil |
| `TestListen_InitialBindFailure` | `unit` | Returns nil channels when initial bind/listen fails. | Socket file exists but another process holds it (simulated in test fixture) | `MessageHandler` callback | Returns `(nil, nil, error)`; both channels are nil; synchronous error returned |

### Validation Failures — Receive (Malformed JSON)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MalformedJSON` | `unit` | Rejects message with invalid JSON. | Socket created; listener started | `{"type":"event","payload":` (missing closing braces) | Warning logged matching `/dropping connection.*malformed JSON.*unexpected end of JSON input/i`; connection closed; no MessageHandler invocation |
| `TestReceive_NotJSONObject` | `unit` | Rejects non-object JSON. | Socket created; listener started | `["array","not","object"]\n` | Warning logged; connection closed; no MessageHandler invocation |

### Validation Failures — Receive (Missing Required Fields)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MissingTypeField` | `unit` | Rejects message missing type field. | Socket created; listener started | `{"payload":{"eventType":"test"}}\n` | Warning logged matching `/dropping connection.*missing required field 'type'/i`; connection closed |
| `TestReceive_MissingPayloadField` | `unit` | Rejects message missing payload field. | Socket created; listener started | `{"type":"event"}\n` | Warning logged matching `/dropping connection.*missing required field 'payload'/i`; connection closed |
| `TestReceive_InvalidMessageType` | `unit` | Rejects message with invalid type value. | Socket created; listener started | `{"type":"unknown","payload":{}}\n` | Warning logged matching `/dropping connection.*invalid message type 'unknown'/i`; connection closed |

### Validation Failures — Receive (Payload Structure)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_PayloadNotObject` | `unit` | Rejects message where payload is not an object. | Socket created; listener started | `{"type":"event","payload":"string"}\n` | Warning logged; connection closed; no MessageHandler invocation |
| `TestReceive_EventMissingEventType` | `unit` | Rejects event message missing eventType in payload. | Socket created; listener started | `{"type":"event","payload":{"message":"test"}}\n` | Warning logged matching `/missing required field.*eventType/i`; connection closed |
| `TestReceive_ErrorEmptyMessage` | `unit` | Rejects error message with empty message field (protocol-level validation). | Socket created; listener started | `{"type":"error","payload":{"message":""}}\n` | Warning logged matching `/missing required field.*message/i`; connection closed |

### Validation Failures — Receive (Field Types)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_ClaudeSessionIDNotString` | `unit` | Rejects message with non-string claudeSessionID. | Socket created; listener started | `{"type":"event","payload":{"eventType":"test","claudeSessionID":123}}\n` | Warning logged; connection closed; no MessageHandler invocation |
| `TestReceive_EventTypeNotString` | `unit` | Rejects event with non-string eventType. | Socket created; listener started | `{"type":"event","payload":{"eventType":123}}\n` | Warning logged; connection closed; no MessageHandler invocation |

### Boundary Values — Message Size Limit

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MessageExceeds10MBLimit` | `unit` | Rejects message exceeding 10 MB limit. | Socket created; listener started | JSON message with payload totaling 11 MB | Warning logged matching `/dropping connection.*message size exceeds 10 MB limit/i`; connection closed; no MessageHandler invocation |
| `TestReceive_MessageExactly10MB` | `unit` | Accepts message at exactly 10 MB limit. | Socket created; listener started; mock MessageHandler | JSON message totaling exactly 10 MB | MessageHandler invoked successfully; response sent |
| `TestReceive_MessageJustUnder10MB` | `unit` | Accepts message just under 10 MB limit. | Socket created; listener started; mock MessageHandler | JSON message totaling 10 MB - 1 byte | MessageHandler invoked successfully; response sent |

### Boundary Values — Special Characters

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MessageWithNewlines` | `unit` | Handles escaped newlines in message field. | Socket created; listener started; mock MessageHandler | Event message with `message` field containing `\n` escape sequences | MessageHandler invoked with newlines preserved in string; response sent |
| `TestReceive_MessageWithUnicode` | `unit` | Handles Unicode characters in payload. | Socket created; listener started; mock MessageHandler | Event message with Unicode in payload: `emoji: 🎉, CJK: 中文` | MessageHandler invoked with Unicode preserved; response sent |
| `TestReceive_MessageWithQuotes` | `unit` | Handles escaped quotes in payload. | Socket created; listener started; mock MessageHandler | Message with `\"` escaped quotes | MessageHandler invoked with quotes preserved; response sent |

### Boundary Values — Empty and Minimal Payloads

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MinimalEventPayload` | `unit` | Accepts event with only required fields. | Socket created; listener started; mock MessageHandler | `{"type":"event","payload":{"eventType":"test"}}\n` | MessageHandler invoked successfully |
| `TestReceive_MinimalErrorPayload` | `unit` | Accepts error with only required fields. | Socket created; listener started; mock MessageHandler | `{"type":"error","payload":{"message":"error"}}\n` | MessageHandler invoked successfully |
| `TestReceive_EmptyNestedObjects` | `unit` | Handles empty nested objects in payload. | Socket created; listener started; mock MessageHandler | Event with `payload.payload={}` | MessageHandler invoked with empty nested object |

### Boundary Values — Invalid Session UUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeSocketManager_InvalidSessionUUID` | `unit` | Handles malformed session UUID. | Test fixture | `ProjectRoot=<test-fixture>`, `SessionUUID="not-a-uuid"` | Construction succeeds; subsequent operations fail with filesystem errors (malformed path) |
| `TestRuntimeSocketManager_EmptySessionUUID` | `unit` | Handles empty session UUID. | Test fixture | `ProjectRoot=<test-fixture>`, `SessionUUID=""` | Construction succeeds; subsequent operations fail with filesystem errors |

### Error Propagation — Connection Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_ClientClosesConnection` | `unit` | Handles client closing connection without sending message. | Socket created; listener started | Client connects then immediately closes | Connection closed; handler goroutine exits cleanly; no error logged |
| `TestReceive_ClientClosesAfterMessage` | `unit` | Handles client closing after valid message but before reading response. | Socket created; listener started | Client sends valid message then closes | MessageHandler invoked; response send fails; warning logged matching `/failed to send response to client:/i` |
| `TestReceive_ReadTimeout` | `unit` | Handles read timeout (if implemented). | Socket created; listener started with read timeout | Client connects but does not send data | Connection closed after timeout; no MessageHandler invocation |

### Error Propagation — Response Send Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestResponse_SendFailsIOError` | `unit` | Handles I/O error when sending response. | Socket created; listener started; client connection becomes unwritable after MessageHandler returns | Valid message from client | Warning logged matching `/failed to send response to client:/i`; connection closed |

### Error Propagation — Asynchronous Listener Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_AcceptLoopFails` | `unit` | Delivers asynchronous listener error via channel. | Socket created; listener started; listener encounters unrecoverable accept error after starting | | Error delivered to `listenerErrCh` matching `/listener accept loop failed:/i`; `listenerDoneCh` closed |
| `TestListen_ListenerErrChannelBuffered` | `unit` | Error channel has capacity 1 and is never closed. | Socket created; listener started | | `listenerErrCh` is buffered with capacity 1; channel never closed by RuntimeSocketManager |
| `TestListen_ListenerDoneSignalsShutdown` | `unit` | Done channel closed on listener exit. | Socket created; listener started; `DeleteSocket()` called | | `listenerDoneCh` closed after listener goroutine exits |

### Resource Cleanup — DeleteSocket Warnings

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_RemovalFailsWarning` | `unit` | Logs warning when socket removal fails but does not error. | Socket created; filesystem error on removal simulated (e.g., mock returns permission error) | | Warning logged matching `/failed to delete runtime socket:.*The socket file may need to be manually removed/i`; method returns nil |
| `TestDeleteSocket_ProceedsAfterFailure` | `unit` | Continues gracefully even if removal fails. | Socket created; listener started; filesystem error on removal | | Listener stops; connections closed; warning logged; no error returned |

### Resource Cleanup — Connection Isolation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MalformedMessageIsolation` | `unit` | Malformed message on one connection does not affect others. | Socket created; listener started; mock MessageHandler | Client 1 sends malformed JSON; Client 2 sends valid message | Client 1 connection closed with warning; Client 2 processed successfully |

### Concurrent Behaviour — Multiple Simultaneous Connections

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_ConcurrentConnections` | `race` | Handles multiple simultaneous connections safely. | Socket created; listener started | 10 clients connect simultaneously and send valid messages | All 10 connections accepted; MessageHandler invoked 10 times (concurrently); all responses sent; all connections closed |
| `TestListen_ConnectionGoroutineIsolation` | `race` | Each connection runs in separate goroutine. | Socket created; listener started; MessageHandler with 100ms delay | 3 clients send messages; handler delays 100ms | All 3 handlers execute concurrently; total execution time ~100ms, not 300ms |

### Concurrent Behaviour — CreateSocket Race

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_ConcurrentCalls` | `race` | Filesystem-level atomic check prevents race conditions. | Session directory exists; no socket file | 5 goroutines call `CreateSocket()` simultaneously | One succeeds; other 4 receive error matching `/runtime socket file already exists/i` |

### Concurrent Behaviour — DeleteSocket During Active Connections

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDeleteSocket_DuringActiveConnections` | `race` | Safely closes connections during DeleteSocket. | Socket created; listener started; 5 active connections with in-flight processing | | All connections closed (may interrupt in-flight processing); listener stops; socket removed |
| `TestDeleteSocket_NoRaceCondition` | `race` | No race between DeleteSocket and connection handlers. | Socket created; listener started; multiple connections processing | Call `DeleteSocket()` concurrently with message processing | No data races detected; all goroutines exit cleanly |

### Concurrent Behaviour — Listen Error and Done Channels

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestListen_ErrorChannelNeverClosed` | `unit` | Error channel is never closed by RuntimeSocketManager. | Socket created; listener started; listener exits normally via `DeleteSocket()` | | `listenerErrCh` remains open (not closed); `listenerDoneCh` closed; no send on closed channel panic |
| `TestListen_DoneChannelClosedOnce` | `unit` | Done channel closed exactly once. | Socket created; listener started; call `DeleteSocket()` multiple times | | `listenerDoneCh` closed exactly once; no double-close panic |

### Mock / Dependency Interaction — MessageHandler

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MessageHandlerSessionUUID` | `unit` | MessageHandler receives session UUID extracted from path. | Test fixture with session UUID `abc-123`; socket created; listener started; mock MessageHandler records arguments | Valid message | MessageHandler invoked with `sessionUUID="abc-123"` |
| `TestReceive_MessageHandlerRuntimeMessage` | `unit` | MessageHandler receives correctly parsed RuntimeMessage. | Socket created; listener started; mock MessageHandler records arguments | `{"type":"event","payload":{"eventType":"test","message":"msg"}}` | MessageHandler receives `RuntimeMessage{Type:"event", Payload:{eventType:"test", message:"msg"}}` |
| `TestReceive_MessageHandlerResponseSerialized` | `unit` | MessageHandler response correctly serialized to JSON. | Socket created; listener started; MessageHandler returns `RuntimeResponse{Status:"success", Message:"processed"}` | Valid message | Client receives `{"status":"success","message":"processed"}\n` |

### Mock / Dependency Interaction — MessageHandler Robustness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_MessageHandlerSlow` | `unit` | Slow MessageHandler does not block other connections. | Socket created; listener started; MessageHandler delays 5 seconds | Client 1 sends message (slow handler); Client 2 sends message (fast handler) | Client 2 receives response before Client 1; both complete successfully |

### Happy Path — Platform-Specific (Windows)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_WindowsNamedPipe` | `unit` | Creates named pipe on Windows instead of Unix socket. | Test runs only on Windows (skip on Unix); test fixture with session directory | | Named pipe created at expected path; Listen succeeds; client can connect |

### Happy Path — Platform-Specific (Unix)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSocket_UnixDomainSocket` | `unit` | Creates Unix domain socket on Unix-like systems. | Test runs only on Unix (skip on Windows); test fixture with session directory | | Unix domain socket created; Listen succeeds; client can connect |

### Boundary Values — Single Request-Response Per Connection

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_SingleRequestResponseCycle` | `unit` | Only one request-response cycle per connection. | Socket created; listener started | Client sends valid message, receives response, attempts to send second message | First message processed and response sent; connection closed after response; second message not processed (connection closed) |
| `TestReceive_MultipleMessagesInStream` | `unit` | Subsequent messages in stream are ignored. | Socket created; listener started | Client sends two newline-delimited JSON messages in single stream | First message processed; response sent; connection closed; second message not processed |

### Happy Path — No Message Buffering

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestReceive_NoMessageBuffering` | `unit` | Messages processed synchronously without buffering. | Socket created; listener started; MessageHandler records invocation order | 3 clients send messages with slight delays | Messages processed in order received; no intermediate buffering; MessageHandler invocations correspond to arrival order |
