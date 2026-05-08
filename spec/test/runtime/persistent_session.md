# Test Specification: `persistent_session_test.go`

## Source File Under Test

`runtime/persistent_session.go`

## Test File

`runtime/persistent_session_test.go`

---

## `PersistentSession`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewPersistentSession_ValidDeps` | `unit` | Constructs PersistentSession with all valid dependencies. | Create a mock Session, mock SessionMetadataStore, mock EventStore, and mock Logger. | `NewPersistentSession(session, metadataStore, eventStore, logger)` | Returns non-nil `*PersistentSession`; no panic |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewPersistentSession_NilSession` | `unit` | Panics when session is nil. | Create mock SessionMetadataStore, mock EventStore, and mock Logger. | `NewPersistentSession(nil, metadataStore, eventStore, logger)` | Panics with `"NewPersistentSession: session must not be nil"` |
| `TestNewPersistentSession_NilMetadataStore` | `unit` | Panics when metadataStore is nil. | Create mock Session, mock EventStore, and mock Logger. | `NewPersistentSession(session, nil, eventStore, logger)` | Panics with `"NewPersistentSession: metadataStore must not be nil"` |
| `TestNewPersistentSession_NilEventStore` | `unit` | Panics when eventStore is nil. | Create mock Session, mock SessionMetadataStore, and mock Logger. | `NewPersistentSession(session, metadataStore, nil, logger)` | Panics with `"NewPersistentSession: eventStore must not be nil"` |
| `TestNewPersistentSession_NilLogger` | `unit` | Panics when logger is nil. | Create mock Session, mock SessionMetadataStore, and mock EventStore. | `NewPersistentSession(session, metadataStore, eventStore, nil)` | Panics with `"NewPersistentSession: logger must not be nil"` |

### Happy Path — Run

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_Run_Success` | `unit` | Run delegates to session and persists metadata on success. | Mock Session.Run() returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.Run()` | Returns `nil`; `session.Run()` called once; `metadataStore.Write()` called once with the metadata snapshot |

### Happy Path — Done

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_Done_Success` | `unit` | Done delegates to session and persists metadata on success. | Mock Session.Done(notifier) returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.Done(notifier)` | Returns `nil`; `session.Done(notifier)` called once; `metadataStore.Write()` called once with the metadata snapshot |

### Happy Path — Fail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_Fail_Success` | `unit` | Fail delegates to session and persists metadata on success. | Mock Session.Fail(err, notifier) returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.Fail(someErr, notifier)` | Returns `nil`; `session.Fail(someErr, notifier)` called once; `metadataStore.Write()` called once with the metadata snapshot |

### Happy Path — UpdateCurrentStateSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_UpdateCurrentStateSafe_Success` | `unit` | UpdateCurrentStateSafe delegates to session and persists metadata on success. | Mock Session.UpdateCurrentStateSafe("node_2") returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.UpdateCurrentStateSafe("node_2")` | Returns `nil`; `session.UpdateCurrentStateSafe("node_2")` called once; `metadataStore.Write()` called once with the metadata snapshot |

### Happy Path — UpdateSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_UpdateSessionDataSafe_Success` | `unit` | UpdateSessionDataSafe delegates to session and persists metadata on success. | Mock Session.UpdateSessionDataSafe("key1", "val1") returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.UpdateSessionDataSafe("key1", "val1")` | Returns `nil`; `session.UpdateSessionDataSafe("key1", "val1")` called once; `metadataStore.Write()` called once with the metadata snapshot |

### Happy Path — UpdateEventHistorySafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_UpdateEventHistorySafe_Success` | `unit` | UpdateEventHistorySafe delegates to session, appends event, and persists metadata. | Mock Session.UpdateEventHistorySafe(event) returns nil. Mock eventStore.Append(event) returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.UpdateEventHistorySafe(event)` | Returns `nil`; `session.UpdateEventHistorySafe(event)` called once; `eventStore.Append(event)` called once; `metadataStore.Write()` called once with the metadata snapshot |

### Happy Path — GetStatusSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_GetStatusSafe` | `unit` | GetStatusSafe passes through to session without persisting. | Mock Session.GetStatusSafe() returns `"running"`. | `ps.GetStatusSafe()` | Returns `"running"`; no calls to metadataStore.Write() or eventStore.Append() |

### Happy Path — GetCurrentStateSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_GetCurrentStateSafe` | `unit` | GetCurrentStateSafe passes through to session without persisting. | Mock Session.GetCurrentStateSafe() returns `"node_1"`. | `ps.GetCurrentStateSafe()` | Returns `"node_1"`; no calls to metadataStore.Write() or eventStore.Append() |

### Happy Path — GetErrorSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_GetErrorSafe` | `unit` | GetErrorSafe passes through to session without persisting. | Mock Session.GetErrorSafe() returns `someErr`. | `ps.GetErrorSafe()` | Returns `someErr`; no calls to metadataStore.Write() or eventStore.Append() |

### Happy Path — GetMetadataSnapshotSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_GetMetadataSnapshotSafe` | `unit` | GetMetadataSnapshotSafe passes through to session without persisting. | Mock Session.GetMetadataSnapshotSafe() returns a SessionMetadata value. | `ps.GetMetadataSnapshotSafe()` | Returns the same SessionMetadata value; no calls to metadataStore.Write() or eventStore.Append() |

### Happy Path — GetSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_GetSessionDataSafe` | `unit` | GetSessionDataSafe passes through to session without persisting. | Mock Session.GetSessionDataSafe("key1") returns `("val1", true)`. | `ps.GetSessionDataSafe("key1")` | Returns `("val1", true)`; no calls to metadataStore.Write() or eventStore.Append() |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_Run_SessionError` | `unit` | Returns session error without persisting when Run fails. | Mock Session.Run() returns an error (e.g., precondition failure). | `ps.Run()` | Returns the session error; `metadataStore.Write()` not called |
| `TestPersistentSession_Done_SessionError` | `unit` | Returns session error without persisting when Done fails. | Mock Session.Done(notifier) returns an error. | `ps.Done(notifier)` | Returns the session error; `metadataStore.Write()` not called |
| `TestPersistentSession_Fail_SessionError` | `unit` | Returns session error without persisting when Fail fails. | Mock Session.Fail(err, notifier) returns an error (e.g., already failed). | `ps.Fail(someErr, notifier)` | Returns the session error; `metadataStore.Write()` not called |
| `TestPersistentSession_UpdateCurrentStateSafe_SessionError` | `unit` | Returns session error without persisting when UpdateCurrentStateSafe fails. | Mock Session.UpdateCurrentStateSafe("x") returns an error. | `ps.UpdateCurrentStateSafe("x")` | Returns the session error; `metadataStore.Write()` not called |
| `TestPersistentSession_UpdateSessionDataSafe_SessionError` | `unit` | Returns session error without persisting when UpdateSessionDataSafe fails. | Mock Session.UpdateSessionDataSafe("k", "v") returns an error. | `ps.UpdateSessionDataSafe("k", "v")` | Returns the session error; `metadataStore.Write()` not called |
| `TestPersistentSession_UpdateEventHistorySafe_SessionError` | `unit` | Returns session error without persisting when UpdateEventHistorySafe fails. | Mock Session.UpdateEventHistorySafe(event) returns an error. | `ps.UpdateEventHistorySafe(event)` | Returns the session error; `eventStore.Append()` not called; `metadataStore.Write()` not called |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_Run_MetadataWriteFails_LogsError` | `unit` | Logs error and returns nil when metadata persist fails after Run. | Mock Session.Run() returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock Session.ID returns `"sess-123"`. Mock metadataStore.Write() returns an error. | `ps.Run()` | Returns `nil`; logger.Error called with message `"failed to persist session metadata after Run"`, args include `"error"` and `"sessionID"="sess-123"` |
| `TestPersistentSession_Done_MetadataWriteFails_LogsError` | `unit` | Logs error and returns nil when metadata persist fails after Done. | Mock Session.Done(notifier) returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock Session.ID returns `"sess-456"`. Mock metadataStore.Write() returns an error. | `ps.Done(notifier)` | Returns `nil`; logger.Error called with message `"failed to persist session metadata after Done"`, args include `"error"` and `"sessionID"="sess-456"` |
| `TestPersistentSession_Fail_MetadataWriteFails_LogsError` | `unit` | Logs error and returns nil when metadata persist fails after Fail. | Mock Session.Fail(err, notifier) returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock Session.ID returns `"sess-789"`. Mock metadataStore.Write() returns an error. | `ps.Fail(someErr, notifier)` | Returns `nil`; logger.Error called with message `"failed to persist session metadata after Fail"`, args include `"error"` and `"sessionID"="sess-789"` |
| `TestPersistentSession_UpdateCurrentStateSafe_MetadataWriteFails_LogsError` | `unit` | Logs error and returns nil when metadata persist fails after UpdateCurrentStateSafe. | Mock Session.UpdateCurrentStateSafe("node_x") returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock Session.ID returns `"sess-100"`. Mock metadataStore.Write() returns an error. | `ps.UpdateCurrentStateSafe("node_x")` | Returns `nil`; logger.Error called with message `"failed to persist session metadata after UpdateCurrentStateSafe"`, args include `"error"` and `"sessionID"="sess-100"` |
| `TestPersistentSession_UpdateSessionDataSafe_MetadataWriteFails_LogsError` | `unit` | Logs error and returns nil when metadata persist fails after UpdateSessionDataSafe. | Mock Session.UpdateSessionDataSafe("myKey", "myVal") returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock Session.ID returns `"sess-200"`. Mock metadataStore.Write() returns an error. | `ps.UpdateSessionDataSafe("myKey", "myVal")` | Returns `nil`; logger.Error called with message `"failed to persist session metadata after UpdateSessionDataSafe"`, args include `"error"`, `"sessionID"="sess-200"`, and `"key"="myKey"` |
| `TestPersistentSession_UpdateEventHistorySafe_AppendFails_LogsError` | `unit` | Logs error when event append fails but still attempts metadata persist. | Mock Session.UpdateEventHistorySafe(event) returns nil. Mock event.ID() returns `"evt-1"`. Mock Session.ID returns `"sess-300"`. Mock eventStore.Append(event) returns an error. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns nil. | `ps.UpdateEventHistorySafe(event)` | Returns `nil`; logger.Error called with message `"failed to persist event"`, args include `"error"`, `"sessionID"="sess-300"`, `"eventID"="evt-1"`; `metadataStore.Write()` still called once |
| `TestPersistentSession_UpdateEventHistorySafe_MetadataWriteFails_LogsError` | `unit` | Logs error when metadata persist fails after successful event append. | Mock Session.UpdateEventHistorySafe(event) returns nil. Mock eventStore.Append(event) returns nil. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock Session.ID returns `"sess-400"`. Mock metadataStore.Write() returns an error. | `ps.UpdateEventHistorySafe(event)` | Returns `nil`; logger.Error called with message `"failed to persist session metadata after UpdateEventHistorySafe"`, args include `"error"` and `"sessionID"="sess-400"` |
| `TestPersistentSession_UpdateEventHistorySafe_BothFail_LogsBothErrors` | `unit` | Logs both errors independently when both append and metadata persist fail. | Mock Session.UpdateEventHistorySafe(event) returns nil. Mock event.ID() returns `"evt-2"`. Mock Session.ID returns `"sess-500"`. Mock eventStore.Append(event) returns errAppend. Mock Session.GetMetadataSnapshotSafe() returns valid metadata. Mock metadataStore.Write() returns errWrite. | `ps.UpdateEventHistorySafe(event)` | Returns `nil`; logger.Error called twice: once with `"failed to persist event"` and once with `"failed to persist session metadata after UpdateEventHistorySafe"` |

### Happy Path — ID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_ID` | `unit` | ID delegates to session.ID. | Mock Session with ID `"sess-abc"`. | `ps.ID` | Returns `"sess-abc"` |

### Happy Path — WorkflowName

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestPersistentSession_WorkflowName` | `unit` | WorkflowName delegates to session.WorkflowName. | Mock Session with WorkflowName `"my-workflow"`. | `ps.WorkflowName` | Returns `"my-workflow"` |
