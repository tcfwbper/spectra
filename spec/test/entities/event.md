# Test Specification: `event.go`

## Source File Under Test
`entities/event.go`

## Test File
`entities/event_test.go`

---

## `Event`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_ValidConstruction` | `unit` | Creates Event with all valid fields. | Session exists with `Status="running"`, `CurrentState="processing"` | `Type="TaskCompleted"`, `Message="success"`, `Payload={"result": "done"}`, `SessionID=<valid-uuid>` | Returns valid Event; `ID` is UUID; `EmittedBy="processing"`; `EmittedAt` is current timestamp; all fields match input |
| `TestEvent_EmptyMessage` | `unit` | Creates Event with empty message (defaults to empty string). | Session exists with `Status="running"`, `CurrentState="review"` | `Type="Approved"`, `Message` omitted, `Payload={}`, `SessionID=<valid-uuid>` | Returns valid Event; `Message=""`; other fields valid |
| `TestEvent_EmptyPayload` | `unit` | Creates Event with empty payload (defaults to empty object). | Session exists with `Status="running"`, `CurrentState="init"` | `Type="Started"`, `Message="beginning"`, `Payload` omitted, `SessionID=<valid-uuid>` | Returns valid Event; `Payload={}`; other fields valid |
| `TestEvent_BothMessageAndPayloadOmitted` | `unit` | Creates Event with both Message and Payload omitted. | Session exists with `Status="running"`, `CurrentState="waiting"` | `Type="Continue"`, `Message` omitted, `Payload` omitted, `SessionID=<valid-uuid>` | Returns valid Event; `Message=""`; `Payload={}`; other fields valid |

### Happy Path — EmittedBy Automatic Assignment

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_EmittedBySetToCurrentState` | `unit` | EmittedBy is automatically set to CurrentState at emission time. | Session with `Status="running"`, `CurrentState="processing"` | `Type="Progress"`, `SessionID=<session-uuid>` | Event created with `EmittedBy="processing"` |
| `TestEvent_EmittedByNotProvidedByCaller` | `unit` | EmittedBy cannot be provided by caller; runtime sets it. | Session with `Status="running"`, `CurrentState="review"` | Attempt to provide `EmittedBy="wrong_node"` in request | `EmittedBy` field ignored if provided; runtime sets `EmittedBy="review"` from session's `CurrentState` |

### Validation Failures — Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_EmptyType` | `unit` | Rejects Event with empty Type. | Session with `Status="running"` | `Type=""`, `Message="msg"`, `SessionID=<valid-uuid>` | Returns error; error message matches `/type.*non-empty/i` |
| `TestEvent_UndefinedType` | `unit` | Rejects Event with Type not defined in workflow. | Session with `Status="running"` | `Type="UndefinedEvent"`, `SessionID=<valid-uuid>` | Returns error; error message matches `/event type.*not defined|undefined.*type/i` |
| `TestEvent_InvalidTypeFormat` | `unit` | Rejects Event with Type not in PascalCase. | Session with `Status="running"` | `Type="invalid_format"`, `SessionID=<valid-uuid>` | Returns error; error message matches `/type.*PascalCase|invalid.*format/i` |

### Validation Failures — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_NullMessage` | `unit` | Rejects Event with null Message. | Session with `Status="running"` | `Type="Event"`, `Message=null`, `SessionID=<valid-uuid>` | Returns error; error message matches `/message.*string/i` |
| `TestEvent_NonStringMessage` | `unit` | Rejects Event with non-string Message value. | Session with `Status="running"` | `Type="Event"`, `Message=123`, `SessionID=<valid-uuid>` | Returns error; error message matches `/message.*string/i` |

### Validation Failures — Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_InvalidPayloadJSON` | `unit` | Rejects Event with malformed JSON in Payload. | Session with `Status="running"` | `Type="Event"`, `Payload=[]byte("{invalid json}")`, `SessionID=<valid-uuid>` | Returns error; error message matches `/JSON.*parse|unmarshal/i` |
| `TestEvent_PayloadPrimitive` | `unit` | Rejects Event with JSON primitive Payload. | Session with `Status="running"` | `Type="Event"`, `Payload=[]byte("\"string\"")`, `SessionID=<valid-uuid>` | Returns error; error message matches `/payload.*object/i` |
| `TestEvent_PayloadArray` | `unit` | Rejects Event with JSON array Payload. | Session with `Status="running"` | `Type="Event"`, `Payload=[]byte("[1,2,3]")`, `SessionID=<valid-uuid>` | Returns error; error message matches `/payload.*object/i` |
| `TestEvent_PayloadNull` | `unit` | Rejects Event with null Payload. | Session with `Status="running"` | `Type="Event"`, `Payload=null`, `SessionID=<valid-uuid>` | Returns error; error message matches `/payload.*not.*null|payload.*required/i` |

### Validation Failures — SessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_NonExistentSession` | `unit` | Rejects Event with non-existent SessionID. | Session with given UUID does not exist | `Type="Event"`, `SessionID=<non-existent-uuid>` | Returns error; error message matches `/session.*not found/i` |
| `TestEvent_InitializingSession` | `unit` | Rejects Event for session with Status=initializing. | Session exists with `Status="initializing"` | `Type="Event"`, `SessionID=<session-uuid>` | Returns error; error message matches `/session.*not ready|initializing/i` |
| `TestEvent_CompletedSession` | `unit` | Rejects Event for session with Status=completed. | Session exists with `Status="completed"` | `Type="Event"`, `SessionID=<session-uuid>` | Returns error; error message matches `/session.*terminated|completed/i` |
| `TestEvent_FailedSession` | `unit` | Rejects Event for session with Status=failed. | Session exists with `Status="failed"` | `Type="Event"`, `SessionID=<session-uuid>` | Returns error; error message matches `/session.*terminated|failed/i` |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_TriggersStateTransition` | `unit` | Event triggers workflow state transition. | Session with `Status="running"`, `CurrentState="review"`; workflow defines transition from "review" on "Approved" to "deploy" | Event with `Type="Approved"` | Session `CurrentState` transitions to "deploy"; event appended to `EventHistory` |
| `TestEvent_NoMatchingTransition` | `unit` | Event with no matching transition is rejected. | Session with `CurrentState="review"`; no transition defined for "Rejected" from "review" | Event with `Type="Rejected"` | Returns error; error message matches `/no.*transition|invalid.*transition/i`; event not recorded; session state unchanged |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_FieldsImmutable` | `unit` | Event fields cannot be modified after creation. | Event instance created | Attempt to modify `Type`, `Message`, `Payload`, or other fields | Field modification attempt fails or has no effect; original values remain |

### Ordering — Chronological

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_OrderingChronological` | `unit` | Events in EventHistory are ordered by EmittedAt ascending. | Session with multiple events emitted at different times | Query `EventHistory` | Events returned in ascending `EmittedAt` order |
| `TestEvent_OrderingTiebreaker` | `unit` | Events with same EmittedAt are ordered by ID lexicographically. | Session with two events emitted simultaneously (same `EmittedAt`) | Query `EventHistory` | Events with identical `EmittedAt` ordered by `ID` lexicographically |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_SimultaneousEmission` | `race` | Multiple events emitted simultaneously are serialized. | Session with `Status="running"` | Two Event instances emitted at same time for same session | Both events recorded in `EventHistory`; first event to acquire session lock is processed first; serialized processing ensures deterministic order |

### Happy Path — Persistence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_AppendedToHistory` | `unit` | Event is appended to session's EventHistory. | Temporary test directory created; session with existing `EventHistory` of 2 events stored in test directory; all file operations occur within test fixtures | New valid Event emitted | Event appended as 3rd entry in `EventHistory`; chronological order maintained |
| `TestEvent_SessionDeletion` | `unit` | Events removed when session is deleted. | Temporary test directory created; session exists with EventHistory containing events in test directory; all file operations occur within test fixtures | Delete session | Events removed from filesystem; subsequent queries match error `/session.*not found/i` |

### Happy Path — Message Delivery

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_MessageDeliveredToRecipient` | `unit` | Message field delivered to recipient determined by workflow routing. | Workflow routes "TaskCompleted" events to "orchestrator" node | Event with `Type="TaskCompleted"`, `Message="task done"` | Message "task done" delivered to orchestrator node |
| `TestEvent_MessageQueuedForFutureNode` | `unit` | Message queued when target node not yet active. | Event triggers transition to "deploy" node; `Message` intended for "deploy" | Event emitted; session transitions | Message queued for delivery when "deploy" becomes active (`CurrentState="deploy"`) |
| `TestEvent_UndeliveredMessageLogged` | `unit` | Undelivered message logged when session terminates before node activates. | Session transitions to completed before target node activates | Event with `Message` for future node | Message marked undelivered in session log |

### Boundary Values — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_LargeMessage` | `unit` | Accepts Event with very large message string. | Session with `Status="running"` | `Type="Event"`, `Message=<1MB string>`, `SessionID=<valid-uuid>` | Returns valid Event; message stored correctly |
| `TestEvent_UnicodeMessage` | `unit` | Accepts Event with Unicode characters in message. | Session with `Status="running"` | `Type="Event"`, `Message="通知: Process complete 🎉"`, `SessionID=<valid-uuid>` | Returns valid Event; Unicode preserved correctly |

### Boundary Values — Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_LargePayload` | `unit` | Accepts Event with very large Payload JSON object. | Session with `Status="running"` | `Type="Event"`, `Payload=<10MB JSON object>`, `SessionID=<valid-uuid>` | Returns valid Event; Payload stored correctly |
| `TestEvent_DeepNestedPayload` | `unit` | Accepts Event with deeply nested JSON in Payload. | Session with `Status="running"` | `Type="Event"`, `Payload=<JSON nested 100 levels deep>`, `SessionID=<valid-uuid>` | Returns valid Event; nested structure preserved |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_CLIInvocation` | `e2e` | Agent emits event via spectra-agent CLI. | Temporary test directory created; runtime running with session files in test directory; session active with socket in test directory and `Status="running"`; all file operations occur within test fixtures | Execute `spectra-agent event emit TaskCompleted --session-id <uuid> --message "done" --payload '{"code": 0}'` | Command succeeds; Event recorded in test directory; workflow transitions; message delivered |
| `TestEvent_CLIMissingSessionID` | `e2e` | CLI rejects event without session-id flag. | Temporary test directory created; runtime running; all file operations occur within test fixtures | Execute `spectra-agent event emit TaskCompleted --message "done"` | Command fails; error message matches `/session-id.*required/i` |
| `TestEvent_CLIDefaultsMessageToEmpty` | `e2e` | CLI defaults Message to empty string when omitted. | Temporary test directory created; runtime running with session files in test directory; session active with `Status="running"`; all file operations occur within test fixtures | Execute `spectra-agent event emit Started --session-id <uuid>` | Event created with `Message=""`; other fields valid |
| `TestEvent_CLIDefaultsPayloadToEmpty` | `e2e` | CLI defaults Payload to empty object when omitted. | Temporary test directory created; runtime running with session files in test directory; session active with `Status="running"`; all file operations occur within test fixtures | Execute `spectra-agent event emit Started --session-id <uuid> --message "go"` | Event created with `Payload={}`; other fields valid |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_RepeatedQueryIdempotent` | `unit` | Repeated queries for EventHistory return same results. | Session with EventHistory of 3 events | Query EventHistory multiple times | All queries return identical results; no mutations |
