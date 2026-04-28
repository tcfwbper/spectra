# Test Specification: `session.go`

## Source File Under Test
`entities/session/session.go`

## Test File
`entities/session/session_test.go`

---

**Fixture Isolation**: All tests create Session instances in memory using programmatic construction. No external files or directories are required unless explicitly stated in the Setup column. Mock dependencies (SessionMetadataStore, EventStore, etc.) are created within each test.

**SessionMetadata Embedding**: Session embeds SessionMetadata as an anonymous field. All SessionMetadata fields (ID, WorkflowName, Status, CreatedAt, UpdatedAt, CurrentState, SessionData, Error) are directly accessible on Session instances (e.g., `session.ID`, `session.Status`).

---

## `Session`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_ValidConstruction` | `unit` | Creates Session with all required fields initialized correctly. | Mock workflow with `EntryNode="start"`, `WorkflowName="TestWorkflow"` | Session constructed via SessionInitializer | Returns valid Session; SessionMetadata fields embedded and accessible: `ID` is UUID; `Status="initializing"`; `CreatedAt=UpdatedAt` (both POSIX timestamps > 0); `CurrentState="start"`; `SessionData={}`; `Error=nil`; Runtime-only fields: `EventHistory=[]`; `mu` initialized; `terminationNotifier` provided |
| `TestSession_TimestampInitialization` | `unit` | CreatedAt and UpdatedAt are set to same timestamp at construction. | Mock workflow | Session constructed | SessionMetadata embedded: `CreatedAt == UpdatedAt`; both are POSIX timestamps > 0 |
| `TestSession_EmptyEventHistory` | `unit` | EventHistory initialized as empty slice. | Mock workflow | Session constructed | Runtime-only field: `EventHistory` is empty slice (length 0); not nil |
| `TestSession_EmptySessionData` | `unit` | SessionData initialized as empty map. | Mock workflow | Session constructed | SessionMetadata embedded: `SessionData` is empty map (length 0); not nil |
| `TestSession_SessionMetadataEmbedded` | `unit` | SessionMetadata fields directly accessible on Session. | Mock workflow | Session constructed with `ID="test-uuid"`, `WorkflowName="TestFlow"` | Can access `session.ID`, `session.WorkflowName`, `session.Status`, etc. directly without `.Metadata` accessor |

### Happy Path — Field Access via Getters

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_GetStatusSafe` | `unit` | GetStatusSafe returns current status. | Session with `Status="running"` | Call `GetStatusSafe()` | Returns `"running"` |
| `TestSession_GetCurrentStateSafe` | `unit` | GetCurrentStateSafe returns current state. | Session with `CurrentState="processing"` | Call `GetCurrentStateSafe()` | Returns `"processing"` |
| `TestSession_GetErrorSafeNil` | `unit` | GetErrorSafe returns nil when no error set. | Session with `Status="running"`, `Error=nil` | Call `GetErrorSafe()` | Returns `nil` |
| `TestSession_GetErrorSafeAgentError` | `unit` | GetErrorSafe returns AgentError when set. | Session with `Status="failed"`, `Error=*AgentError` | Call `GetErrorSafe()` | Returns `*AgentError` matching stored error |
| `TestSession_GetErrorSafeRuntimeError` | `unit` | GetErrorSafe returns RuntimeError when set. | Session with `Status="failed"`, `Error=*RuntimeError` | Call `GetErrorSafe()` | Returns `*RuntimeError` matching stored error |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_InitializingToRunning` | `unit` | Session transitions from initializing to running. | Session with `Status="initializing"` | Call `Run(terminationNotifier)` | Returns `nil`; `Status="running"`; `UpdatedAt` refreshed |
| `TestSession_RunningToCompleted` | `unit` | Session transitions from running to completed. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)` | Returns `nil`; `Status="completed"`; `UpdatedAt` refreshed; notification sent on channel |
| `TestSession_InitializingToFailed` | `unit` | Session transitions from initializing to failed. | Session with `Status="initializing"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error` set; `UpdatedAt` refreshed; notification sent on channel |
| `TestSession_RunningToFailed` | `unit` | Session transitions from running to failed. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*AgentError, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error` set; `UpdatedAt` refreshed; notification sent on channel |

### Validation Failures — Status Preconditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_RunOnRunning` | `unit` | Run rejects session already running. | Session with `Status="running"` | Call `Run(terminationNotifier)` | Returns error matching `/cannot run.*status is 'running'/i`; status unchanged |
| `TestSession_RunOnCompleted` | `unit` | Run rejects session already completed. | Session with `Status="completed"` | Call `Run(terminationNotifier)` | Returns error matching `/cannot run.*status is 'completed'/i`; status unchanged |
| `TestSession_RunOnFailed` | `unit` | Run rejects session already failed. | Session with `Status="failed"` | Call `Run(terminationNotifier)` | Returns error matching `/cannot run.*status is 'failed'/i`; status unchanged |
| `TestSession_DoneOnInitializing` | `unit` | Done rejects session still initializing. | Session with `Status="initializing"` | Call `Done(terminationNotifier)` | Returns error matching `/cannot complete.*status is 'initializing'/i`; status unchanged |
| `TestSession_DoneOnCompleted` | `unit` | Done rejects session already completed. | Session with `Status="completed"` | Call `Done(terminationNotifier)` | Returns error matching `/cannot complete.*status is 'completed'/i`; status unchanged |
| `TestSession_DoneOnFailed` | `unit` | Done rejects session already failed. | Session with `Status="failed"` | Call `Done(terminationNotifier)` | Returns error matching `/cannot complete.*status is 'failed'/i`; status unchanged |

### Validation Failures — Terminal State Finality

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_FailOnCompleted` | `unit` | Fail rejects session already completed. | Session with `Status="completed"` | Call `Fail(*RuntimeError, terminationNotifier)` | Returns error matching `/cannot fail.*status is 'completed'/i`; status unchanged; original `Error` remains `nil` |
| `TestSession_FailOnFailed` | `unit` | Fail rejects session already failed (first error wins). | Session with `Status="failed"`, `Error=*AgentError("first error")` | Call `Fail(*RuntimeError("second error"), terminationNotifier)` | Returns error matching `/session already failed/i`; `Error` remains first error |

### Read-Only Convention

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_SessionMetadataFieldsExported` | `unit` | SessionMetadata fields are exported (accessible) but should not be modified directly. | Session constructed with `ID=<uuid>`, `WorkflowName="TestFlow"`, `CreatedAt=<timestamp>` | Access `session.ID`, `session.WorkflowName`, `session.CreatedAt` | All fields accessible; modification not prevented by compiler (read-only by convention only) |
| `TestSession_MutationThroughMethodsOnly` | `unit` | Session mutations should use provided methods, not direct field assignment. | Session with `Status="initializing"` | Call `session.Run(terminationNotifier)` instead of `session.Status = "running"` | Status updated correctly via method; direct assignment possible but not recommended |

### Atomic Replacement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_ErrorReplacementAtomic` | `unit` | Error field updated atomically with Status. | Session with `Status="running"` | Call `Fail(*AgentError, terminationNotifier)` under write lock | Both `Status` and `Error` updated in same critical section; no intermediate state visible to readers |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_ConcurrentGetters` | `race` | Multiple concurrent GetStatusSafe calls succeed. | Session with `Status="running"` | 100 concurrent goroutines call `GetStatusSafe()` | All calls return `"running"`; no race conditions |
| `TestSession_ConcurrentMixedAccess` | `race` | Concurrent reads and writes are serialized correctly. | Session with `Status="running"` | 50 goroutines call `GetStatusSafe()` concurrently with 1 goroutine calling `UpdateCurrentStateSafe()` | No race conditions; readers see either old or new value consistently |
| `TestSession_ConcurrentRunFail` | `race` | Concurrent Run and Fail are serialized by write lock. | Session with `Status="initializing"`; buffered `terminationNotifier` channel (capacity 2) | Goroutine 1 calls `Run(terminationNotifier)`; Goroutine 2 calls `Fail(*RuntimeError, terminationNotifier)` simultaneously | Whichever acquires lock first wins; loser gets precondition error; status consistent |

### Invariants — Status-Error Correlation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_ErrorNilWhenNotFailed` | `unit` | Error is nil when Status is not failed. | Session with `Status="running"` | Check `Error` field via `GetErrorSafe()` | Returns `nil` |
| `TestSession_ErrorNonNilWhenFailed` | `unit` | Error is non-nil when Status is failed. | Session with `Status="failed"`, `Error=*AgentError` | Check `Error` field via `GetErrorSafe()` | Returns non-nil `*AgentError` |

### Invariants — Timestamp Ordering

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_UpdatedAtRefreshedOnMutation` | `unit` | UpdatedAt refreshed after Run. | Session with `Status="initializing"`, `UpdatedAt=T0` | Wait 1 second; call `Run(terminationNotifier)` | `UpdatedAt > T0`; `UpdatedAt >= CreatedAt` |
| `TestSession_TimestampOrderingMaintained` | `unit` | CreatedAt <= UpdatedAt always holds. | Session constructed at T0 | Perform multiple mutations (`Run`, `UpdateSessionDataSafe`, etc.) | `CreatedAt <= UpdatedAt` after each mutation |

### Invariants — In-Memory Authority

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_InMemoryStateAuthoritative` | `unit` | In-memory state updated even when persistence fails. | Mock SessionMetadataStore that returns error on write; session with `Status="initializing"` | Call `Run(terminationNotifier)` | Returns `nil`; in-memory `Status="running"`; persistence failure logged as warning |
| `TestSession_PersistenceFailureLogged` | `unit` | Persistence failures logged but do not error. | Mock SessionMetadataStore that fails; session with `Status="running"`; mock logger | Call `UpdateSessionDataSafe("key", "value")` | Returns `nil`; warning logged matching `/persistence failed/i`; in-memory state updated |

### Happy Path — SessionMetadata Persistence Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_SessionMetadataExtraction` | `unit` | SessionMetadata can be extracted from Session for persistence. | Session with populated fields | Access embedded SessionMetadata fields directly or copy struct | SessionMetadata fields (ID, WorkflowName, Status, CreatedAt, UpdatedAt, CurrentState, SessionData, Error) accessible; runtime-only fields (EventHistory, mu, terminationNotifier) not included in SessionMetadata |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSession_InitializationTimeout` | `unit` | Session fails on initialization timeout. | Session with `Status="initializing"`; mock timer set to fire after 10ms (simulating 30s timeout); buffered `terminationNotifier` channel (capacity 2) | Timer handler calls `Fail(*RuntimeError{Issuer="SessionInitializer", Message="session initialization timeout exceeded 30 seconds"}, terminationNotifier)` | `Status="failed"`; `Error` set; notification sent |
| `TestSession_RuntimeSocketLossMidSession` | `unit` | Session fails when runtime socket is lost. | Session with `Status="running"`; mock RuntimeSocketManager that immediately reports loss; buffered `terminationNotifier` channel (capacity 2) | Runtime calls `Fail(*RuntimeError, terminationNotifier)` | `Status="failed"`; `Error` set; notification sent |
| `TestSession_SignalInterrupt` | `unit` | Session status unchanged on OS signal (no Fail called). | Session with `Status="running"`; SIGINT received | Runtime stops listener without calling `Fail` | `Status` remains `"running"`; no error set; session persisted as last snapshot |
| `TestSession_TerminationNotifierCapacity` | `unit` | Termination notifier has capacity >= 2. | Session constructed with buffered channel | Check channel capacity | Channel capacity >= 2 |
| `TestSession_SingleNotificationOnDone` | `unit` | Done sends exactly one notification. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)` | Exactly one value sent on channel; channel length increases by 1 |
| `TestSession_SingleNotificationOnFail` | `unit` | Fail sends exactly one notification. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError, terminationNotifier)` | Exactly one value sent on channel; channel length increases by 1 |
| `TestSession_NoDuplicateNotifications` | `unit` | Multiple Done/Fail calls do not send duplicate notifications. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)`; attempt second `Done(terminationNotifier)` | First call sends notification; second call returns error without sending |
