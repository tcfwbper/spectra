# Test Specification: `message_router_test.go`

## Source File Under Test

`runtime/message_router.go`

## Test File

`runtime/message_router_test.go`

---

## `MessageRouter`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewMessageRouter_ValidDeps` | `unit` | Constructs MessageRouter with all valid dependencies. | Create mock PersistentSession, mock EventProcessor, mock ErrorProcessor, a buffered TerminationNotifier channel (cap >= 2), and mock Logger. | `NewMessageRouter(persistentSession, eventProcessor, errorProcessor, terminationNotifier, logger)` | Returns non-nil `*MessageRouter`; no panic |

### Happy Path — Handle

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_Handle_EventType` | `unit` | Dispatches event-type message to EventProcessor and returns its response. | Mock EventProcessor: `ProcessEvent("sess-uuid", msg)` returns `SuccessResponse("ok")`. Create RuntimeMessage with `Type()="event"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns the exact RuntimeResponse from EventProcessor (`SuccessResponse("ok")`) |
| `TestMessageRouter_Handle_ErrorType` | `unit` | Dispatches error-type message to ErrorProcessor and returns its response. | Mock ErrorProcessor: `ProcessError("sess-uuid", msg)` returns `SuccessResponse("error recorded")`. Create RuntimeMessage with `Type()="error"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns the exact RuntimeResponse from ErrorProcessor (`SuccessResponse("error recorded")`) |
| `TestMessageRouter_Handle_EventProcessorReturnsError` | `unit` | Returns ErrorResponse from EventProcessor unchanged. | Mock EventProcessor: `ProcessEvent(...)` returns `ErrorResponse("session not ready: status is 'failed'")`. Create RuntimeMessage with `Type()="event"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("session not ready: status is 'failed'")` unchanged |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_Handle_UnknownType` | `unit` | Returns error response for unknown message type. | Create RuntimeMessage with `Type()="unknown"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("unknown message type 'unknown'")` |
| `TestMessageRouter_Handle_PanicInEventProcessor` | `unit` | Recovers panic from EventProcessor, fails session, and returns internal server error. | Mock EventProcessor: `ProcessEvent(...)` panics with `"nil pointer"`. Mock PersistentSession: `ID` returns `"sess-uuid"`, `GetCurrentStateSafe()` returns `"nodeA"`, `Fail()` returns nil. Mock Logger: `Error(...)` captures call. Create RuntimeMessage with `Type()="event"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("internal server error")`; `Logger.Error` called once; `Fail()` called with RuntimeError having `Issuer="MessageRouter"`, `Message="panic during message processing"`, `SessionID="sess-uuid"`, `FailingState="nodeA"` |
| `TestMessageRouter_Handle_PanicInErrorProcessor` | `unit` | Recovers panic from ErrorProcessor, fails session, and returns internal server error. | Mock ErrorProcessor: `ProcessError(...)` panics with `"index out of range"`. Mock PersistentSession: `ID` returns `"sess-uuid"`, `GetCurrentStateSafe()` returns `"nodeB"`, `Fail()` returns nil. Mock Logger: `Error(...)` captures call. Create RuntimeMessage with `Type()="error"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("internal server error")`; `Logger.Error` called once; `Fail()` called with RuntimeError having `Issuer="MessageRouter"` |
| `TestMessageRouter_Handle_PanicRecovery_FailReturnsError` | `unit` | Still returns internal server error when PersistentSession.Fail fails during panic recovery. | Mock EventProcessor: `ProcessEvent(...)` panics. Mock PersistentSession: `Fail()` returns `errors.New("session already failed")`, `ID` returns `"sess-uuid"`, `GetCurrentStateSafe()` returns `"nodeA"`. Mock Logger. Create RuntimeMessage with `Type()="event"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("internal server error")`; Logger.Error called (logs Fail error) |
| `TestMessageRouter_Handle_PanicRecovery_RuntimeErrorConstructionFails` | `unit` | Returns internal server error even if RuntimeError construction fails during panic recovery. | Mock EventProcessor: `ProcessEvent(...)` panics. Setup conditions that cause `NewRuntimeError` to fail (e.g., invalid parameters mock). Mock PersistentSession: `ID` returns `""`, `GetCurrentStateSafe()` returns `""`. Mock Logger. Create RuntimeMessage with `Type()="event"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("internal server error")`; Logger.Error called; `Fail()` not called |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_Handle_DoesNotModifyMessage` | `unit` | Passes RuntimeMessage to processor without modification. | Mock EventProcessor: `ProcessEvent(...)` captures the RuntimeMessage arg and returns success. Create RuntimeMessage with `Type()="event"`, specific `ClaudeSessionID()` and `Payload()`. | `mr.Handle("sess-uuid", runtimeMessage)` | EventProcessor receives the exact same RuntimeMessage reference (same ClaudeSessionID, same Payload) |
| `TestMessageRouter_Handle_DoesNotModifyResponse` | `unit` | Returns processor response without modification. | Mock ErrorProcessor: `ProcessError(...)` returns a specific RuntimeResponse with custom message. Create RuntimeMessage with `Type()="error"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Returned RuntimeResponse is identical to what ErrorProcessor returned |
| `TestMessageRouter_Handle_NoLogOnNormalDispatch` | `unit` | Does not call Logger during normal (non-panic) dispatch. | Mock EventProcessor returns success. Mock Logger to capture calls. Create RuntimeMessage with `Type()="event"`. | `mr.Handle("sess-uuid", runtimeMessage)` | Logger.Error not called |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_Handle_ConcurrentMessages` | `race` | Multiple concurrent Handle calls are independent and do not race. | Mock EventProcessor: returns success. Mock ErrorProcessor: returns success. Create one RuntimeMessage with `Type()="event"` and one with `Type()="error"`. | Call `mr.Handle(...)` concurrently from two goroutines with different messages. | Both calls complete without data race; each returns appropriate response |
