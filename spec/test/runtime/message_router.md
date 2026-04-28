# Test Specification: `message_router.go`

## Source File Under Test
`runtime/message_router.go`

## Test File
`runtime/message_router_test.go`

---

## `MessageRouter`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_New` | `unit` | Constructs MessageRouter with valid inputs. | Test fixture; mock Session, EventProcessor, ErrorProcessor, TerminationNotifier channel | `Session=<mock>`, `EventProcessor=<mock>`, `ErrorProcessor=<mock>`, `TerminationNotifier=<channel>` | Returns MessageRouter instance; no error |

### Happy Path — RouteMessage (Event Type)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_EventType_Success` | `unit` | Routes event message to EventProcessor successfully. | Mock EventProcessor returns RuntimeResponse with `status="success"`, `message="event processed"` | RuntimeMessage with `type="event"`, valid payload | EventProcessor.ProcessEvent called; returns RuntimeResponse with `status="success"`, `message="event processed"` |
| `TestRouteMessage_EventType_Error` | `unit` | Returns error response from EventProcessor. | Mock EventProcessor returns RuntimeResponse with `status="error"`, `message="session not ready"` | RuntimeMessage with `type="event"`, valid payload | EventProcessor.ProcessEvent called; returns RuntimeResponse with `status="error"`, `message="session not ready"` |
| `TestRouteMessage_EventType_SessionUUIDPassed` | `unit` | Passes session UUID to EventProcessor. | Mock EventProcessor tracks arguments; session UUID is `abc-123` | `sessionUUID="abc-123"`, RuntimeMessage with `type="event"` | EventProcessor.ProcessEvent called with `sessionUUID="abc-123"` |

### Happy Path — RouteMessage (Error Type)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_ErrorType_Success` | `unit` | Routes error message to ErrorProcessor successfully. | Mock ErrorProcessor returns RuntimeResponse with `status="success"`, `message="error recorded"` | RuntimeMessage with `type="error"`, valid payload | ErrorProcessor.ProcessError called; returns RuntimeResponse with `status="success"`, `message="error recorded"` |
| `TestRouteMessage_ErrorType_Error` | `unit` | Returns error response from ErrorProcessor. | Mock ErrorProcessor returns RuntimeResponse with `status="error"`, `message="session terminated"` | RuntimeMessage with `type="error"`, valid payload | ErrorProcessor.ProcessError called; returns RuntimeResponse with `status="error"`, `message="session terminated"` |
| `TestRouteMessage_ErrorType_SessionUUIDPassed` | `unit` | Passes session UUID to ErrorProcessor. | Mock ErrorProcessor tracks arguments; session UUID is `def-456` | `sessionUUID="def-456"`, RuntimeMessage with `type="error"` | ErrorProcessor.ProcessError called with `sessionUUID="def-456"` |

### Happy Path — MessageHandler Interface Implementation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_ImplementsMessageHandlerInterface` | `unit` | Verifies MessageRouter implements MessageHandler interface. | MessageRouter instance; MessageHandler interface type check | | MessageRouter satisfies MessageHandler interface: `func(sessionUUID string, message RuntimeMessage) RuntimeResponse` |

### Validation Failures — Unknown Message Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_UnknownType` | `unit` | Returns error for unknown message type. | | RuntimeMessage with `type="unknown"` | Returns RuntimeResponse with `status="error"`, `message="unknown message type 'unknown'"`; no processor called |
| `TestRouteMessage_EmptyType` | `unit` | Returns error for empty message type. | | RuntimeMessage with `type=""` | Returns RuntimeResponse with `status="error"`, `message="unknown message type ''"`; no processor called |
| `TestRouteMessage_InvalidType` | `unit` | Returns error for invalid message type. | | RuntimeMessage with `type="invalid"` | Returns RuntimeResponse with `status="error"`, `message="unknown message type 'invalid'"`; no processor called |

### Error Propagation — Panic Recovery

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_EventProcessorPanics` | `unit` | Recovers from panic in EventProcessor. | Mock EventProcessor panics with "test panic"; mock Session with `CurrentState="node_a"`; TerminationNotifier channel monitored | RuntimeMessage with `type="event"` | Panic recovered; error logged with stack trace; RuntimeError constructed with `Issuer="MessageRouter"`, `Message="panic during message processing"`; Session.Fail called; returns RuntimeResponse with `status="error"`, `message="internal server error"` |
| `TestRouteMessage_ErrorProcessorPanics` | `unit` | Recovers from panic in ErrorProcessor. | Mock ErrorProcessor panics with "test panic"; mock Session with `CurrentState="node_b"`; TerminationNotifier channel monitored | RuntimeMessage with `type="error"` | Panic recovered; error logged with stack trace; RuntimeError constructed with `Issuer="MessageRouter"`; Session.Fail called; returns RuntimeResponse with `status="error"`, `message="internal server error"` |
| `TestRouteMessage_PanicLogIncludesStackTrace` | `unit` | Logs full stack trace during panic recovery. | Mock EventProcessor panics; capture log output | RuntimeMessage with `type="event"` | Log output contains panic message, stack trace, and goroutine information; RuntimeResponse with `status="error"` returned |
| `TestRouteMessage_PanicSessionFailCalled` | `unit` | Calls Session.Fail with RuntimeError during panic recovery. | Mock EventProcessor panics; mock Session tracks Session.Fail calls | RuntimeMessage with `type="event"` | Session.Fail called once with RuntimeError containing: `Issuer="MessageRouter"`, `Message="panic during message processing"`, `Detail` with panic message and stack trace, `SessionID`, `FailingState`, `OccurredAt` |
| `TestRouteMessage_PanicTerminationNotifierSignaled` | `unit` | Signals termination notifier after panic recovery. | Mock EventProcessor panics; mock Session with valid state; TerminationNotifier channel monitored | RuntimeMessage with `type="event"` | Session.Fail called; TerminationNotifier receives signal; returns RuntimeResponse with `status="error"`, `message="internal server error"` |

### Error Propagation — Panic Recovery Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_PanicWithNilValue` | `unit` | Handles panic with nil value. | Mock EventProcessor panics with nil | RuntimeMessage with `type="event"` | Panic recovered; RuntimeError constructed; Session.Fail called; returns RuntimeResponse with `status="error"`, `message="internal server error"` |
| `TestRouteMessage_PanicWithNonStringValue` | `unit` | Handles panic with non-string value. | Mock EventProcessor panics with integer value `123` | RuntimeMessage with `type="event"` | Panic recovered; RuntimeError detail includes panic value representation; Session.Fail called; returns RuntimeResponse with `status="error"` |
| `TestRouteMessage_PanicDuringSessionFail` | `unit` | Handles panic during Session.Fail in panic recovery. | Mock EventProcessor panics; Session.Fail also panics | RuntimeMessage with `type="event"` | Inner panic recovered (best-effort); returns RuntimeResponse with `status="error"`, `message="internal server error"`; process does NOT crash |

### Error Propagation — Session.Fail Failure During Panic

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_PanicSessionFailReturnsError` | `unit` | Handles Session.Fail error during panic recovery. | Mock EventProcessor panics; Session.Fail returns error "session already failed" | RuntimeMessage with `type="event"` | Panic recovered; Session.Fail called (returns error); error logged; returns RuntimeResponse with `status="error"`, `message="internal server error"` |
| `TestRouteMessage_PanicPersistenceFailureBestEffort` | `unit` | Continues when persistence fails during panic recovery. | Mock EventProcessor panics; SessionMetadataStore.Write fails during Session.Fail; Session.Fail logs warning but returns nil | RuntimeMessage with `type="event"` | Panic recovered; Session.Fail called (persistence fails but returns nil); session remains "failed" in memory; returns RuntimeResponse with `status="error"` |

### Boundary Values — Response Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_ResponseNotModified` | `unit` | Returns processor response without modification. | Mock EventProcessor returns RuntimeResponse with `status="success"`, `message="custom message with special chars: 🎉"` | RuntimeMessage with `type="event"` | Returns exact RuntimeResponse from EventProcessor without modification |
| `TestRouteMessage_EmptyMessageField` | `unit` | Handles response with empty message field. | Mock ErrorProcessor returns RuntimeResponse with `status="success"`, `message=""` | RuntimeMessage with `type="error"` | Returns RuntimeResponse with empty message field as-is |
| `TestRouteMessage_MalformedProcessorResponse` | `unit` | Returns malformed processor response as-is. | Mock EventProcessor returns RuntimeResponse with `status="invalid"` (neither "success" nor "error") | RuntimeMessage with `type="event"` | Returns malformed RuntimeResponse as-is; validation deferred to RuntimeSocketManager |

### Boundary Values — Message Processing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_VeryLargeMessage` | `unit` | Handles very large RuntimeMessage. | Mock EventProcessor succeeds; RuntimeMessage with 5 MB payload | RuntimeMessage with `type="event"`, very large payload | EventProcessor called with complete message; returns success response |
| `TestRouteMessage_UnicodeInMessage` | `unit` | Handles Unicode characters in RuntimeMessage. | Mock EventProcessor succeeds | RuntimeMessage with Unicode in payload: `{type:"event", payload:{eventType:"测试🎉"}}` | EventProcessor called with Unicode preserved; returns success response |

### Boundary Values — Message Type Case Sensitivity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_TypeCaseSensitive` | `unit` | Message type matching is case-sensitive. | | RuntimeMessage with `type="Event"` (capital E) | Returns RuntimeResponse with `status="error"`, `message="unknown message type 'Event'"`; no processor called |
| `TestRouteMessage_TypeEventLowercase` | `unit` | Accepts lowercase event type. | Mock EventProcessor succeeds | RuntimeMessage with `type="event"` (lowercase) | EventProcessor called; returns success |
| `TestRouteMessage_TypeErrorLowercase` | `unit` | Accepts lowercase error type. | Mock ErrorProcessor succeeds | RuntimeMessage with `type="error"` (lowercase) | ErrorProcessor called; returns success |

### Mock / Dependency Interaction — Processors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_EventProcessorCalledOnce` | `unit` | Calls EventProcessor exactly once for event message. | Mock EventProcessor tracks invocation count | RuntimeMessage with `type="event"` | EventProcessor.ProcessEvent called exactly once |
| `TestRouteMessage_ErrorProcessorCalledOnce` | `unit` | Calls ErrorProcessor exactly once for error message. | Mock ErrorProcessor tracks invocation count | RuntimeMessage with `type="error"` | ErrorProcessor.ProcessError called exactly once |
| `TestRouteMessage_OnlyEventProcessorCalled` | `unit` | Only EventProcessor called for event message. | Mock EventProcessor and ErrorProcessor track calls | RuntimeMessage with `type="event"` | EventProcessor called once; ErrorProcessor NOT called |
| `TestRouteMessage_OnlyErrorProcessorCalled` | `unit` | Only ErrorProcessor called for error message. | Mock EventProcessor and ErrorProcessor track calls | RuntimeMessage with `type="error"` | ErrorProcessor called once; EventProcessor NOT called |

### Mock / Dependency Interaction — RuntimeMessage

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_RuntimeMessageNotModified` | `unit` | Does not modify RuntimeMessage before passing to processor. | Mock EventProcessor tracks received RuntimeMessage | RuntimeMessage with `type="event"`, specific payload structure | EventProcessor receives exact RuntimeMessage without modifications |
| `TestRouteMessage_RuntimeMessagePassedByValue` | `unit` | RuntimeMessage passed to processor (verify no shared mutation). | Mock EventProcessor modifies received RuntimeMessage; call MessageRouter again | Same RuntimeMessage instance used twice | Second call receives original RuntimeMessage (no mutations from first call); each processor receives independent copy or immutable reference |

### Concurrent Behaviour — Multiple Simultaneous Calls

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_ConcurrentCalls` | `race` | Handles concurrent RouteMessage calls safely. | MessageRouter instance shared across goroutines; mock processors are thread-safe; 10 goroutines call RouteMessage simultaneously | 10 concurrent RuntimeMessages (mix of event and error types) | All 10 calls complete successfully; correct processor called for each type; no data races detected |
| `TestRouteMessage_ConcurrentPanics` | `race` | Handles concurrent panics from processors safely. | Mock processors panic on some calls; 5 goroutines call RouteMessage simultaneously | 5 concurrent RuntimeMessages, some trigger panics | All panics recovered; RuntimeErrors constructed; Session.Fail called for each panic; all goroutines return error responses; no process crash |

### Concurrent Behaviour — Panic Recovery Isolation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_PanicIsolation` | `race` | Panic in one goroutine does not affect others. | Mock EventProcessor that panics based on message payload marker (e.g., `eventType="trigger-panic"`); 3 goroutines launched with sync.WaitGroup: one with panic-triggering message, two with normal messages | Three concurrent RouteMessage calls: one triggers panic, two are normal (order non-deterministic) | All three goroutines complete; the one with panic-trigger recovers, calls Session.Fail, returns error response; the two normal calls succeed; no process crash; verified via sync.WaitGroup and collecting all responses |

### Resource Cleanup — No Logging (Except Panics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_NoLoggingForNormalMessages` | `unit` | Does not log for normal message routing. | Redirect stderr to test buffer or use test logger with mock verification; mock processors return success | RuntimeMessage with `type="event"` | No log output captured in test buffer; mock verification confirms MessageRouter did not call logger |
| `TestRouteMessage_NoLoggingForProcessorErrors` | `unit` | Does not log when processors return error responses. | Redirect stderr to test buffer or use test logger with mock verification; mock processor returns error response | RuntimeMessage with `type="event"` | No log output captured in test buffer; processor error communicated via RuntimeResponse only |
| `TestRouteMessage_LogsOnlyForPanic` | `unit` | Logs only during panic recovery. | Redirect stderr to test buffer; mock processor panics | RuntimeMessage with `type="event"` | Test buffer contains panic information and stack trace captured from stderr; no other logging |

### State Transitions — Session Failure via Panic

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_PanicTransitionsSessionToFailed` | `unit` | Panic recovery transitions session to failed status. | Mock Session with `Status="running"`; mock EventProcessor panics; Session.Fail updates Status to "failed" | RuntimeMessage with `type="event"` | Session.Status transitions from "running" to "failed"; Session.Fail called during panic recovery |
| `TestRouteMessage_PanicCapturesFailingState` | `unit` | Captures correct FailingState during panic recovery. | Mock Session with `CurrentState="node_x"`; mock EventProcessor panics | RuntimeMessage with `type="event"` | RuntimeError has `FailingState="node_x"` captured from Session.CurrentState at time of panic |

### Happy Path — MessageHandler Interface Conformance

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMessageRouter_ImplementsMessageHandlerSignature` | `unit` | Verifies MessageRouter.RouteMessage conforms to MessageHandler interface signature. | MessageRouter instance; interface type assertion or compile-time check | Assign `MessageRouter.RouteMessage` to variable of type `func(sessionUUID string, message RuntimeMessage) RuntimeResponse` | Assignment succeeds; type matches exactly; verifies signature compatibility with RuntimeSocketManager's callback contract |

### Happy Path — Independent Message Processing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRouteMessage_IndependentMessageProcessing` | `unit` | Each message processed independently. | MessageRouter instance; mock processors; process 3 messages sequentially | 3 RuntimeMessages: event, error, event | Each message routed to appropriate processor; no state carried between calls; each returns independent RuntimeResponse |
| `TestRouteMessage_NoSharedState` | `unit` | No shared state between message processing calls. | MessageRouter instance; process two messages with delay | First message processed; second message processed after delay | Second message processing unaffected by first; no shared mutable state between calls |
