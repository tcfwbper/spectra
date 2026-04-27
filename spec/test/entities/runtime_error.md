# Test Specification: `runtime_error.go`

## Source File Under Test
`entities/runtime_error.go`

## Test File
`entities/runtime_error_test.go`

---

## `RuntimeError`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_ValidConstruction` | `unit` | Creates RuntimeError with all valid required fields. | | `Issuer="MessageRouter"`, `Message="socket creation failed"`, `Detail={"errno": 13}`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError; all fields match input |
| `TestRuntimeError_NullDetail` | `unit` | Creates RuntimeError with null Detail. | | `Issuer="EventProcessor"`, `Message="transition failed"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns valid RuntimeError; `Detail=null` |
| `TestRuntimeError_EmptyDetail` | `unit` | Creates RuntimeError with empty JSON object Detail. | | `Issuer="Session"`, `Message="initialization error"`, `Detail={}`, `SessionID=<valid-uuid>`, `FailingState="entry"`, `OccurredAt=1714147200` | Returns valid RuntimeError; `Detail={}` |

### Validation Failures — Issuer

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_EmptyIssuer` | `unit` | Rejects RuntimeError with empty Issuer. | | `Issuer=""`, `Message="error"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/issuer.*non-empty/i` |
| `TestRuntimeError_WhitespaceOnlyIssuer` | `unit` | Rejects RuntimeError with whitespace-only Issuer. | | `Issuer="   "`, `Message="error"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/issuer.*whitespace/i` |

### Validation Failures — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_EmptyMessage` | `unit` | Rejects RuntimeError with empty message. | | `Issuer="MessageRouter"`, `Message=""`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/message.*non-empty/i` |
| `TestRuntimeError_WhitespaceOnlyMessage` | `unit` | Rejects RuntimeError with whitespace-only message. | | `Issuer="MessageRouter"`, `Message="   "`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/message.*whitespace/i` |

### Validation Failures — Detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_InvalidDetailJSON` | `unit` | Rejects RuntimeError with malformed JSON in Detail. | | `Issuer="MessageRouter"`, `Message="error"`, `Detail=[]byte("{invalid json}")`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/JSON.*parse|unmarshal/i` |

### Validation Failures — SessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_NonExistentSession` | `unit` | Rejects RuntimeError with non-existent SessionID. | Session with given UUID does not exist | `Issuer="MessageRouter"`, `Message="error"`, `Detail=null`, `SessionID=<non-existent-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/session.*not found/i` |
| `TestRuntimeError_FailedSessionID` | `unit` | Rejects RuntimeError for session with Status=failed. | Session exists with `Status="failed"` | `Issuer="MessageRouter"`, `Message="error"`, `Detail=null`, `SessionID=<failed-session-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/session.*terminated/i`; warning logged |
| `TestRuntimeError_CompletedSessionID` | `unit` | Rejects RuntimeError for session with Status=completed. | Session exists with `Status="completed"` | `Issuer="MessageRouter"`, `Message="error"`, `Detail=null`, `SessionID=<completed-session-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns error; error message matches `/session.*terminated/i`; warning logged |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_TransitionsSessionToFailedInMemory` | `unit` | Session Status transitions to failed in memory first when RuntimeError is raised. | Session exists with `Status="running"`, `CurrentState="processing"` | RuntimeError with valid fields; `SessionID=<session-uuid>`, `FailingState="processing"` | Session `Status="failed"` in memory immediately; `CurrentState` unchanged; `Error` field set to RuntimeError instance |
| `TestRuntimeError_PersistenceAttemptedAfterMemoryUpdate` | `unit` | Persistence to SessionMetadataStore attempted after in-memory update. | Session exists with `Status="running"` | RuntimeError raised | In-memory status updated to "failed" first; then persistence attempted; persistence success or failure does not affect in-memory status |
| `TestRuntimeError_InitializingSessionFails` | `unit` | Session in initializing status transitions to failed. | Session exists with `Status="initializing"`, `CurrentState="entry"` | RuntimeError with `SessionID=<session-uuid>`, `FailingState="entry"` | Session `Status="failed"` in memory; `CurrentState="entry"`; `FailingState="entry"` |
| `TestRuntimeError_RunningSessionFails` | `unit` | Session in running status transitions to failed. | Session exists with `Status="running"`, `CurrentState="processing"` | RuntimeError with `SessionID=<session-uuid>`, `FailingState="processing"` | Session `Status="failed"` in memory; `CurrentState="processing"`; `FailingState="processing"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_FieldsImmutable` | `unit` | RuntimeError fields cannot be modified after creation. | RuntimeError instance created | Attempt to modify `Issuer`, `Message`, `Detail`, or other fields | Field modification attempt fails or has no effect; original values remain |

### Happy Path — Issuer Component Names

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_MessageRouterIssuer` | `unit` | RuntimeError from MessageRouter component. | | `Issuer="MessageRouter"`, `Message="panic in routing"`, `Detail={"panic": "index out of bounds"}`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError with `Issuer="MessageRouter"` |
| `TestRuntimeError_RuntimeSocketManagerIssuer` | `unit` | RuntimeError from RuntimeSocketManager component. | | `Issuer="RuntimeSocketManager"`, `Message="socket file exists"`, `Detail={"path": "/tmp/socket"}`, `SessionID=<valid-uuid>`, `FailingState="entry"`, `OccurredAt=1714147200` | Returns valid RuntimeError with `Issuer="RuntimeSocketManager"` |
| `TestRuntimeError_EventProcessorIssuer` | `unit` | RuntimeError from EventProcessor component. | | `Issuer="EventProcessor"`, `Message="invalid event type"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="waiting"`, `OccurredAt=1714147200` | Returns valid RuntimeError with `Issuer="EventProcessor"` |
| `TestRuntimeError_TransitionToNodeIssuer` | `unit` | RuntimeError from TransitionToNode component. | | `Issuer="TransitionToNode"`, `Message="target node not found"`, `Detail={"target": "unknown"}`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError with `Issuer="TransitionToNode"` |
| `TestRuntimeError_SessionIssuer` | `unit` | RuntimeError from Session component. | | `Issuer="Session"`, `Message="initialization failed"`, `Detail={"reason": "permission denied"}`, `SessionID=<valid-uuid>`, `FailingState="entry"`, `OccurredAt=1714147200` | Returns valid RuntimeError with `Issuer="Session"` |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_MultipleErrorsSerialized` | `race` | Multiple simultaneous RuntimeErrors are serialized; first error wins. | Session with `Status="running"` | Two RuntimeError instances raised simultaneously for same session | First error to acquire session lock updates in-memory status and records in session's `Error` field; second error logged but does not overwrite; session `Status="failed"` |

### Happy Path — Persistence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_PersistenceSuccess` | `unit` | RuntimeError persisted to disk when SessionMetadataStore write succeeds. | Temporary test directory created; session files placed within this directory; SessionMetadataStore operational; all file operations occur within test fixtures | Valid RuntimeError raised | In-memory status updated to "failed"; persistence succeeds; error details written to disk within test directory; session metadata updated |
| `TestRuntimeError_PersistenceFailureLogged` | `unit` | Persistence failure logged as warning when SessionMetadataStore write fails. | Temporary test directory created; mock SessionMetadataStore configured to return write error simulating disk full; all file operations occur within test fixtures | Valid RuntimeError raised | In-memory status updated to "failed"; persistence fails; warning logged matching `/failed.*persist.*RuntimeError/i`; session remains failed in memory |
| `TestRuntimeError_SessionDeletion` | `unit` | RuntimeError removed when session is deleted. | Temporary test directory created; session exists with recorded RuntimeError in test directory; all file operations occur within test fixtures | Delete session | RuntimeError file removed from filesystem; subsequent error queries match error `/session.*not found/i` |

### Invariants — FailingState Consistency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_FailingStateMatchesCurrentState` | `unit` | FailingState must match CurrentState at error time. | Session with `CurrentState="processing"` | RuntimeError with `FailingState="processing"` | RuntimeError recorded; `FailingState="processing"` matches session `CurrentState` |
| `TestRuntimeError_CurrentStateUnchanged` | `unit` | CurrentState does not change when error occurs. | Session with `CurrentState="review"` | RuntimeError raised | Session `CurrentState` remains "review"; `FailingState="review"` |

### Happy Path — Terminal Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_StatusPermanentlyFailed` | `unit` | Session Status remains failed and cannot transition. | Session transitioned to `Status="failed"` by RuntimeError | Attempt any status transition | Status remains "failed"; transition rejected |
| `TestRuntimeError_NoAutomaticRetry` | `unit` | Runtime does not automatically retry failed session. | Session with `Status="failed"` due to RuntimeError; mock clock advanced by 5 seconds | Query session status after time advancement | Session remains failed; no retry attempted |
| `TestRuntimeError_ManualRecoveryRejected` | `unit` | Recovery requests for failed session are rejected. | Session with `Status="failed"` | Request session recovery | Returns error; error message matches `/recovery not supported|cannot recover/i` and mentions creating new session |

### Happy Path — In-Memory Status Priority

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_InMemoryStatusAuthoritative` | `unit` | In-memory session status is authoritative even if persistence fails. | Mock SessionMetadataStore configured to return write error | RuntimeError raised | In-memory status updated to "failed" immediately; runtime behavior reflects failed status; subsequent operations see failed status |
| `TestRuntimeError_RuntimeBehaviorConsistentAfterPersistenceFailure` | `unit` | Runtime behavior remains correct after persistence failure. | Mock SessionMetadataStore configured to return write error | RuntimeError raised; subsequent session query | Session query returns `Status="failed"` from memory; runtime correctly rejects new events for this session with error matching `/session.*terminated|failed/i` |

### Happy Path — Sensitive Data

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_SensitiveDataPersisted` | `unit` | Detail with sensitive info is persisted as-is (issuer responsibility to sanitize). | | RuntimeError with `Detail={"file_path": "/home/user/secrets.txt"}` | Detail persisted exactly as provided; no sanitization by runtime |

### Happy Path — Panic Recovery

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_PanicInMessageRouter` | `unit` | RuntimeError raised when MessageRouter panics. | Mock MessageRouter configured to panic with "index out of bounds" during message processing | Trigger message processing that causes panic | RuntimeError created with `Issuer="MessageRouter"`; `Detail` contains panic message and stack trace; session transitions to failed |
| `TestRuntimeError_PanicStackTraceInDetail` | `unit` | Panic stack trace included in Detail field. | Mock component configured to panic during processing | Trigger panic condition | RuntimeError `Detail` field contains "panic" key with message; "stack" key with trace |

### Happy Path — Socket Creation Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_SocketCreationFailure` | `unit` | RuntimeError raised when socket creation fails during initialization. | Temporary test directory created; socket file created programmatically at target path within test directory before session initialization; all file operations occur within test fixtures | Initialize session | RuntimeError with `Issuer="Session"` or `"RuntimeSocketManager"`; `Detail` contains error details matching `/socket.*exists|file.*exists/i`; session `Status="failed"` |
| `TestRuntimeError_SocketPermissionDenied` | `unit` | RuntimeError raised when socket creation fails due to permissions. | Temporary test directory created; target socket directory created within test directory with permissions set to read-only (0444); all file operations occur within test fixtures | Initialize session | RuntimeError with `Detail` containing error matching `/permission denied|EACCES/i`; session initialization aborted; `Status="failed"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_UnderlyingErrorWrapped` | `unit` | Underlying system error is wrapped in Detail field. | Mock system operation configured to return specific errno (e.g., ENOENT) | Trigger failing operation | RuntimeError created; `Detail` contains original error details with errno; error message provides context |

### Boundary Values — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_LargeMessage` | `unit` | Accepts RuntimeError with very large message string. | | `Issuer="MessageRouter"`, `Message=<1MB string>`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError; message stored correctly |
| `TestRuntimeError_UnicodeMessage` | `unit` | Accepts RuntimeError with Unicode characters in message. | | `Issuer="MessageRouter"`, `Message="错误: 消息路由失败 ⚠️"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError; Unicode preserved correctly |

### Boundary Values — Detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_LargeDetail` | `unit` | Accepts RuntimeError with very large Detail JSON object. | | `Issuer="MessageRouter"`, `Message="error"`, `Detail=<10MB JSON object>`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError; Detail stored correctly |
| `TestRuntimeError_DeepNestedDetail` | `unit` | Accepts RuntimeError with deeply nested JSON in Detail. | | `Issuer="MessageRouter"`, `Message="error"`, `Detail=<JSON nested 100 levels deep>`, `SessionID=<valid-uuid>`, `FailingState="processing"`, `OccurredAt=1714147200` | Returns valid RuntimeError; nested structure preserved |

### Happy Path — Human Notification

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_HumanNotified` | `e2e` | Human is notified when RuntimeError occurs. | Temporary test directory created; runtime running with session files in test directory; session active; all file operations occur within test fixtures | Trigger RuntimeError condition | RuntimeError logged to error log file within test directory; console output notifies human with error details |
| `TestRuntimeError_ErrorLogWritten` | `unit` | RuntimeError details written to session error log. | Temporary test directory created; session files placed within this directory; all file operations occur within test fixtures | RuntimeError raised | Error details written to session's error log file within test directory with timestamp and full context |
