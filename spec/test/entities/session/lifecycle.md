# Test Specification: `lifecycle.go`

## Source File Under Test
`entities/session/lifecycle.go`

## Test File
`entities/session/lifecycle_test.go`

---

**Fixture Isolation**: All tests create Session instances in memory using programmatic construction. No external files or directories are required unless explicitly stated in the Setup column. Mock dependencies (SessionMetadataStore, loggers, etc.) are created within each test.

---

## `Session` Lifecycle Methods

### Happy Path — Run

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_InitializingToRunning` | `unit` | Transitions session from initializing to running. | Session with `Status="initializing"`, `UpdatedAt=T0`; buffered `terminationNotifier` channel (capacity 2) | Call `Run(terminationNotifier)` | Returns `nil`; `Status="running"`; `UpdatedAt > T0` |
| `TestRun_PersistsToStore` | `unit` | Persists updated session metadata to store. | Mock SessionMetadataStore; session with `Status="initializing"` | Call `Run(terminationNotifier)` | Returns `nil`; store write called with updated session; in-memory `Status="running"` |
| `TestRun_NoNotificationSent` | `unit` | Run does not send notification on terminationNotifier. | Session with `Status="initializing"`; buffered `terminationNotifier` channel (capacity 2, empty) | Call `Run(terminationNotifier)` | Returns `nil`; channel remains empty (length 0) |

### Happy Path — Done

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDone_RunningToCompleted` | `unit` | Transitions session from running to completed. | Session with `Status="running"`, `UpdatedAt=T0`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)` | Returns `nil`; `Status="completed"`; `UpdatedAt > T0`; notification sent on channel |
| `TestDone_PersistsToStore` | `unit` | Persists updated session metadata to store. | Mock SessionMetadataStore; session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)` | Returns `nil`; store write called with updated session; in-memory `Status="completed"` |
| `TestDone_SendsNotification` | `unit` | Done sends exactly one notification. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2, empty) | Call `Done(terminationNotifier)` | Returns `nil`; channel length increases by 1; single `struct{}` value sent |
| `TestDone_NonBlockingSend` | `unit` | Notification send does not block with buffered channel. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)` | Returns immediately (< 100ms); notification sent |

### Happy Path — Fail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_InitializingToFailed` | `unit` | Transitions session from initializing to failed. | Session with `Status="initializing"`, `Error=nil`, `UpdatedAt=T0`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError{Issuer="Test", Message="error"}, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error` set to provided RuntimeError; `UpdatedAt > T0`; notification sent |
| `TestFail_RunningToFailed` | `unit` | Transitions session from running to failed. | Session with `Status="running"`, `Error=nil`, `UpdatedAt=T0`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*AgentError{NodeName="agent", Message="error"}, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error` set to provided AgentError; `UpdatedAt > T0`; notification sent |
| `TestFail_PersistsToStore` | `unit` | Persists updated session metadata to store. | Mock SessionMetadataStore; session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError, terminationNotifier)` | Returns `nil`; store write called with updated session including Error; in-memory `Status="failed"` |
| `TestFail_SendsNotification` | `unit` | Fail sends exactly one notification. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2, empty) | Call `Fail(*AgentError, terminationNotifier)` | Returns `nil`; channel length increases by 1; single `struct{}` value sent |
| `TestFail_AcceptsAgentError` | `unit` | Fail accepts AgentError type. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*AgentError{NodeName="test", Message="agent failed"}, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error` is `*AgentError` |
| `TestFail_AcceptsRuntimeError` | `unit` | Fail accepts RuntimeError type. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError{Issuer="runtime", Message="runtime failed"}, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error` is `*RuntimeError` |

### Validation Failures — Run Preconditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_RejectsRunning` | `unit` | Run returns error when status is already running. | Session with `Status="running"` | Call `Run(terminationNotifier)` | Returns error matching `/cannot run.*status is 'running'.*expected 'initializing'/i`; status unchanged |
| `TestRun_RejectsCompleted` | `unit` | Run returns error when status is completed. | Session with `Status="completed"` | Call `Run(terminationNotifier)` | Returns error matching `/cannot run.*status is 'completed'.*expected 'initializing'/i`; status unchanged |
| `TestRun_RejectsFailed` | `unit` | Run returns error when status is failed. | Session with `Status="failed"`, `Error=*RuntimeError` | Call `Run(terminationNotifier)` | Returns error matching `/cannot run.*status is 'failed'.*expected 'initializing'/i`; status unchanged; `Error` preserved |

### Validation Failures — Done Preconditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDone_RejectsInitializing` | `unit` | Done returns error when status is initializing. | Session with `Status="initializing"` | Call `Done(terminationNotifier)` | Returns error matching `/cannot complete.*status is 'initializing'.*expected 'running'/i`; status unchanged |
| `TestDone_RejectsCompleted` | `unit` | Done returns error when status is already completed. | Session with `Status="completed"` | Call `Done(terminationNotifier)` | Returns error matching `/cannot complete.*status is 'completed'.*expected 'running'/i`; status unchanged |
| `TestDone_RejectsFailed` | `unit` | Done returns error when status is failed. | Session with `Status="failed"`, `Error=*AgentError` | Call `Done(terminationNotifier)` | Returns error matching `/cannot complete.*status is 'failed'.*expected 'running'/i`; status unchanged; `Error` preserved |

### Validation Failures — Fail Preconditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_RejectsNilError` | `unit` | Fail returns error when err parameter is nil. | Session with `Status="running"` | Call `Fail(nil, terminationNotifier)` | Returns error matching `/error cannot be nil/i`; status unchanged; lock not acquired |
| `TestFail_RejectsInvalidErrorType` | `unit` | Fail returns error when err is not AgentError or RuntimeError. | Session with `Status="running"` | Call `Fail(errors.New("generic error"), terminationNotifier)` | Returns error matching `/invalid error type.*must be.*AgentError.*RuntimeError/i`; status unchanged; lock not acquired |
| `TestFail_RejectsCompleted` | `unit` | Fail returns error when status is completed (terminal state finality). | Session with `Status="completed"` | Call `Fail(*RuntimeError, terminationNotifier)` | Returns error matching `/cannot fail.*status is 'completed'.*workflow already terminated/i`; status unchanged; no notification sent |
| `TestFail_RejectsFailed` | `unit` | Fail returns error when status is already failed (first error wins). | Session with `Status="failed"`, `Error=*AgentError{NodeName="first", Message="first error"}` | Call `Fail(*RuntimeError{Issuer="second", Message="second error"}, terminationNotifier)` | Returns error matching `/session already failed/i`; status unchanged; `Error` remains first error; no notification sent |

### Validation Failures — Error Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_RejectsStandardError` | `unit` | Fail rejects standard Go error type. | Session with `Status="running"` | Call `Fail(fmt.Errorf("wrapped: %w", errors.New("base")), terminationNotifier)` | Returns error matching `/invalid error type/i`; status unchanged |
| `TestFail_RejectsNilAgentError` | `unit` | Fail rejects nil pointer typed as AgentError. | Session with `Status="running"` | Call `Fail((*AgentError)(nil), terminationNotifier)` | Returns error matching `/error cannot be nil/i`; status unchanged |
| `TestFail_RejectsNilRuntimeError` | `unit` | Fail rejects nil pointer typed as RuntimeError. | Session with `Status="running"` | Call `Fail((*RuntimeError)(nil), terminationNotifier)` | Returns error matching `/error cannot be nil/i`; status unchanged |

### Atomic Replacement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_AtomicStatusAndErrorUpdate` | `unit` | Status and Error updated atomically in same critical section. | Session with `Status="running"`; concurrent goroutine reading `GetStatusSafe()` and `GetErrorSafe()` in loop; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*AgentError, terminationNotifier)` | Concurrent reader never observes `Status="failed"` with `Error=nil` or vice versa; both updated together |
| `TestLifecycle_AtomicStatusAndTimestampUpdate` | `unit` | Status and UpdatedAt updated atomically. | Session with `Status="initializing"`, `UpdatedAt=T0`; concurrent goroutine reading `GetStatusSafe()` in loop | Call `Run(terminationNotifier)` | Concurrent reader never observes intermediate state; UpdatedAt refreshed when Status changes |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_NotIdempotent` | `unit` | Second Run call returns error (not idempotent). | Session with `Status="initializing"` | Call `Run(terminationNotifier)` twice | First returns `nil`; second returns error; status remains `"running"` |
| `TestDone_NotIdempotent` | `unit` | Second Done call returns error (not idempotent). | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)` twice | First returns `nil`; second returns error; status remains `"completed"`; only one notification sent |
| `TestFail_NotIdempotent` | `unit` | Second Fail call returns error (first error wins). | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*AgentError{Message="first"}, terminationNotifier)` then `Fail(*RuntimeError{Message="second"}, terminationNotifier)` | First returns `nil`; second returns error matching `/already failed/i`; `Error` remains first; only one notification sent |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestLifecycle_ReleasesLockAfterRun` | `unit` | Write lock released after Run completes. | Session with `Status="initializing"` | Call `Run(terminationNotifier)`; then call `GetStatusSafe()` | Both succeed without deadlock |
| `TestLifecycle_ReleasesLockAfterDone` | `unit` | Write lock released after Done completes. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Done(terminationNotifier)`; then call `GetStatusSafe()` | Both succeed without deadlock |
| `TestLifecycle_ReleasesLockAfterFail` | `unit` | Write lock released after Fail completes. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError, terminationNotifier)`; then call `GetErrorSafe()` | Both succeed without deadlock |
| `TestLifecycle_ReleasesLockOnValidationFailure` | `unit` | Write lock not held when validation fails before lock acquisition. | Session with `Status="running"` | Call `Fail(nil, terminationNotifier)` (validation fails); simultaneously call `GetStatusSafe()` | Validation error returned immediately; getter not blocked |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestLifecycle_ConcurrentRunFail` | `race` | Concurrent Run and Fail are serialized correctly. | Session with `Status="initializing"`; buffered `terminationNotifier` channel (capacity 2) | Goroutine 1 calls `Run(terminationNotifier)`; Goroutine 2 calls `Fail(*RuntimeError, terminationNotifier)` simultaneously | Whichever acquires lock first succeeds; loser gets precondition error; final status is either `"running"` or `"failed"` consistently; at most one notification sent |
| `TestLifecycle_ConcurrentDoneFail` | `race` | Concurrent Done and Fail are serialized correctly. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Goroutine 1 calls `Done(terminationNotifier)`; Goroutine 2 calls `Fail(*AgentError, terminationNotifier)` simultaneously | Whichever acquires lock first succeeds; loser gets precondition error; final status is either `"completed"` or `"failed"` consistently; exactly one notification sent |
| `TestLifecycle_ConcurrentFailCalls` | `race` | Multiple concurrent Fail calls result in first error wins. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | 10 goroutines call `Fail()` simultaneously with different errors | Exactly one succeeds (returns `nil`); others return error matching `/already failed/i`; single error stored; exactly one notification sent |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestLifecycle_PersistenceFailureLoggedNotReturned` | `unit` | Persistence failures logged but do not return error. | Mock SessionMetadataStore that returns error on write; session with `Status="initializing"`; mock logger | Call `Run(terminationNotifier)` | Returns `nil`; in-memory `Status="running"`; warning logged matching `/persistence failed/i` or store error message |
| `TestLifecycle_PersistenceFailureInDone` | `unit` | Done persistence failure logged but returns nil. | Mock SessionMetadataStore that fails; session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2); mock logger | Call `Done(terminationNotifier)` | Returns `nil`; in-memory `Status="completed"`; warning logged; notification sent |
| `TestLifecycle_PersistenceFailureInFail` | `unit` | Fail persistence failure logged but returns nil. | Mock SessionMetadataStore that fails; session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2); mock logger | Call `Fail(*RuntimeError, terminationNotifier)` | Returns `nil`; in-memory `Status="failed"`, `Error` set; warning logged; notification sent |

### Invariants — UpdatedAt Refresh

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_RefreshesUpdatedAt` | `unit` | Run refreshes UpdatedAt timestamp. | Session with `Status="initializing"`, `UpdatedAt=T0` | Wait 1 second; call `Run(terminationNotifier)` | Returns `nil`; `UpdatedAt > T0` |
| `TestDone_RefreshesUpdatedAt` | `unit` | Done refreshes UpdatedAt timestamp. | Session with `Status="running"`, `UpdatedAt=T0`; buffered `terminationNotifier` channel (capacity 2) | Wait 1 second; call `Done(terminationNotifier)` | Returns `nil`; `UpdatedAt > T0` |
| `TestFail_RefreshesUpdatedAt` | `unit` | Fail refreshes UpdatedAt timestamp. | Session with `Status="running"`, `UpdatedAt=T0`; buffered `terminationNotifier` channel (capacity 2) | Wait 1 second; call `Fail(*RuntimeError, terminationNotifier)` | Returns `nil`; `UpdatedAt > T0` |

### Invariants — Terminal State Finality

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestLifecycle_CompletedIsFinal` | `unit` | Completed status cannot transition elsewhere. | Session with `Status="completed"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError, terminationNotifier)` | Returns error; status remains `"completed"`; no notification sent |
| `TestLifecycle_FailedIsFinal` | `unit` | Failed status cannot transition elsewhere. | Session with `Status="failed"`, `Error=*AgentError` | Call `Run(terminationNotifier)` | Returns error; status remains `"failed"`; `Error` preserved |
| `TestLifecycle_FirstErrorWins` | `unit` | First Fail call wins; subsequent calls rejected. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*AgentError{Message="first"}, terminationNotifier)`; then `Fail(*RuntimeError{Message="second"}, terminationNotifier)` | First succeeds; second returns error; `Error` remains first |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestLifecycle_NotificationAfterLockRelease` | `unit` | Notification sent after write lock released. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2); mock blocking channel send | Call `Done(terminationNotifier)` | Lock released before send; concurrent `GetStatusSafe()` does not block on slow send |
| `TestLifecycle_ValidationBeforeLockAcquisition` | `unit` | Fail validation occurs before lock acquisition. | Session with `Status="running"`; concurrent goroutine holding write lock indefinitely | Call `Fail(nil, terminationNotifier)` | Returns error immediately without blocking on lock |
| `TestFail_EmptyErrorMessage` | `unit` | Fail accepts error with empty message. | Session with `Status="running"`; buffered `terminationNotifier` channel (capacity 2) | Call `Fail(*RuntimeError{Issuer="test", Message=""}, terminationNotifier)` | Returns `nil`; `Status="failed"`; `Error.Message=""` |
