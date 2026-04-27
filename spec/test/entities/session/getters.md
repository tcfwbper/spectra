# Test Specification: `getters.go`

## Source File Under Test
`entities/session/getters.go`

## Test File
`entities/session/getters_test.go`

---

**Fixture Isolation**: All tests create Session instances in memory using programmatic construction. No external files or directories are required unless explicitly stated in the Setup column.

---

## `Session` Safe Getters

### Happy Path — GetStatusSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetStatusSafe_Initializing` | `unit` | Returns status when session is initializing. | Session with `Status="initializing"` | Call `GetStatusSafe()` | Returns `"initializing"` |
| `TestGetStatusSafe_Running` | `unit` | Returns status when session is running. | Session with `Status="running"` | Call `GetStatusSafe()` | Returns `"running"` |
| `TestGetStatusSafe_Completed` | `unit` | Returns status when session is completed. | Session with `Status="completed"` | Call `GetStatusSafe()` | Returns `"completed"` |
| `TestGetStatusSafe_Failed` | `unit` | Returns status when session is failed. | Session with `Status="failed"`, `Error=*RuntimeError` | Call `GetStatusSafe()` | Returns `"failed"` |

### Happy Path — GetCurrentStateSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetCurrentStateSafe_EntryNode` | `unit` | Returns entry node for newly initialized session. | Session with `CurrentState="start"` (entry node) | Call `GetCurrentStateSafe()` | Returns `"start"` |
| `TestGetCurrentStateSafe_IntermediateNode` | `unit` | Returns current intermediate node. | Session with `CurrentState="processing"` | Call `GetCurrentStateSafe()` | Returns `"processing"` |
| `TestGetCurrentStateSafe_ExitNode` | `unit` | Returns exit node for completed session. | Session with `Status="completed"`, `CurrentState="exit"` | Call `GetCurrentStateSafe()` | Returns `"exit"` |

### Happy Path — GetErrorSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetErrorSafe_NilWhenNoError` | `unit` | Returns nil when Error is nil. | Session with `Status="running"`, `Error=nil` | Call `GetErrorSafe()` | Returns `nil` |
| `TestGetErrorSafe_AgentError` | `unit` | Returns AgentError when set. | Session with `Status="failed"`, `Error=*AgentError{NodeName: "agent", Message: "agent failed"}` | Call `GetErrorSafe()` | Returns `*AgentError` with matching fields |
| `TestGetErrorSafe_RuntimeError` | `unit` | Returns RuntimeError when set. | Session with `Status="failed"`, `Error=*RuntimeError{Issuer: "runtime", Message: "runtime failed"}` | Call `GetErrorSafe()` | Returns `*RuntimeError` with matching fields |
| `TestGetErrorSafe_NilBeforeFail` | `unit` | Returns nil when session is running (before Fail called). | Session with `Status="running"`, `Error=nil` | Call `GetErrorSafe()` | Returns `nil` |
| `TestGetErrorSafe_NonNilAfterFail` | `unit` | Returns error after Fail is called. | Session with `Status="running"`; call `Fail(*AgentError, terminationNotifier)` | Call `GetErrorSafe()` | Returns `*AgentError` |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetStatusSafe_Idempotent` | `unit` | Repeated calls return same result. | Session with `Status="running"` | Call `GetStatusSafe()` multiple times | All calls return `"running"`; no mutations |
| `TestGetCurrentStateSafe_Idempotent` | `unit` | Repeated calls return same result. | Session with `CurrentState="processing"` | Call `GetCurrentStateSafe()` multiple times | All calls return `"processing"`; no mutations |
| `TestGetErrorSafe_Idempotent` | `unit` | Repeated calls return same error pointer. | Session with `Status="failed"`, `Error=*AgentError` | Call `GetErrorSafe()` multiple times | All calls return same `*AgentError` pointer |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetErrorSafe_ReturnsImmutableError` | `unit` | Returned error shares allocation but is immutable by convention. | Session with `Status="failed"`, `Error=*AgentError{Message: "original"}` | Call `GetErrorSafe()`; attempt to modify returned error via public API | AgentError has no exported mutators; original error fields unchanged |

### Not Immutable

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetters_ReturnValueNotDeepCopy` | `unit` | Getters return scalar copies, not deep copies (errors share allocation). | Session with `Status="failed"`, `Error=*AgentError` | Call `GetErrorSafe()` twice | Both calls return same pointer (aliased); not separate allocations |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetters_ConcurrentGetStatusSafe` | `race` | Multiple concurrent GetStatusSafe calls succeed. | Session with `Status="running"` | 100 goroutines call `GetStatusSafe()` simultaneously | All return `"running"`; no race conditions |
| `TestGetters_ConcurrentGetCurrentStateSafe` | `race` | Multiple concurrent GetCurrentStateSafe calls succeed. | Session with `CurrentState="processing"` | 100 goroutines call `GetCurrentStateSafe()` simultaneously | All return `"processing"`; no race conditions |
| `TestGetters_ConcurrentGetErrorSafe` | `race` | Multiple concurrent GetErrorSafe calls succeed. | Session with `Status="failed"`, `Error=*AgentError` | 100 goroutines call `GetErrorSafe()` simultaneously | All return same `*AgentError`; no race conditions |
| `TestGetters_ConcurrentGettersAndWriter` | `race` | Getters block during write lock; observe consistent state. | Session with `Status="running"` | 50 goroutines call `GetStatusSafe()`; 1 goroutine calls `Run()` | No race conditions; getters see either "running" or new status consistently after transition |
| `TestGetters_ConcurrentMixedGetters` | `race` | Different getters called concurrently. | Session with `Status="running"`, `CurrentState="processing"` | 100 goroutines call mix of `GetStatusSafe()`, `GetCurrentStateSafe()`, `GetErrorSafe()` | All succeed; no race conditions; consistent snapshots |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetters_ReleasesReadLock` | `unit` | Read lock released after each getter call. | Session with `Status="running"` | Call `GetStatusSafe()` then `GetCurrentStateSafe()` | Both succeed without deadlock |
| `TestGetters_MultipleReadersDoNotBlock` | `unit` | Multiple concurrent readers do not block each other. | Session with `Status="running"` | 10 goroutines call `GetStatusSafe()` concurrently | All complete quickly (< 100ms total); no serialization between readers |

### Invariants — No Mutation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetStatusSafe_NoMutation` | `unit` | GetStatusSafe does not modify any Session field. | Session with `Status="running"`, `UpdatedAt=T0` | Call `GetStatusSafe()` | Returns `"running"`; `UpdatedAt` remains T0 (not refreshed) |
| `TestGetCurrentStateSafe_NoMutation` | `unit` | GetCurrentStateSafe does not modify any Session field. | Session with `CurrentState="node1"`, `UpdatedAt=T0` | Call `GetCurrentStateSafe()` | Returns `"node1"`; `UpdatedAt` remains T0 |
| `TestGetErrorSafe_NoMutation` | `unit` | GetErrorSafe does not modify any Session field. | Session with `Status="failed"`, `Error=*AgentError`, `UpdatedAt=T0` | Call `GetErrorSafe()` | Returns `*AgentError`; `UpdatedAt` remains T0 |

### Invariants — Snapshot Consistency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetters_SnapshotNotAtomic` | `unit` | Two separate getter calls may observe values from different points in time. | Session with `Status="running"`, `CurrentState="node1"`; concurrent goroutine calling `UpdateCurrentStateSafe("node2")` | Call `GetStatusSafe()` then `GetCurrentStateSafe()` sequentially | May observe `Status="running"` with `CurrentState="node2"` (not atomic snapshot); both calls individually consistent |
| `TestGetters_EachGetterAtomicByItself` | `unit` | Each individual getter call observes consistent snapshot of its field. | Session with `Status="running"`; concurrent goroutine calling `Run()` (transition to "running" or error) | Call `GetStatusSafe()` | Returns either "initializing" or "running" (never partial or corrupted value) |

### Invariants — Error Aliasing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetErrorSafe_ReturnsSharedPointer` | `unit` | GetErrorSafe returns same pointer stored in Session.Error. | Session with `Status="failed"`, `Error=*AgentError` at address X | Call `GetErrorSafe()` | Returns pointer to same address X |
| `TestGetErrorSafe_CallerMustNotMutate` | `unit` | Caller must treat returned error as immutable (enforced by API, not reflection). | Session with `Status="failed"`, `Error=*AgentError` | Call `GetErrorSafe()`; check AgentError public API | AgentError has no exported setters; immutability enforced by convention |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetStatusSafe_BeforeRun` | `unit` | GetStatusSafe returns "initializing" before Run called. | Session with `Status="initializing"` (just constructed) | Call `GetStatusSafe()` | Returns `"initializing"` |
| `TestGetCurrentStateSafe_NeverEmpty` | `unit` | GetCurrentStateSafe never returns empty string for successfully initialized session. | Session constructed with entry node "start" | Call `GetCurrentStateSafe()` | Returns `"start"` (non-empty) |
| `TestGetErrorSafe_NilBeforeFailure` | `unit` | GetErrorSafe returns nil before any failure. | Session with `Status="initializing"` or `"running"`, `Error=nil` | Call `GetErrorSafe()` | Returns `nil` |
| `TestGetErrorSafe_SetOnlyByFail` | `unit` | Error can only be set via Fail method (not directly). | Session with `Status="running"`, `Error=nil` | Attempt to access/modify `Error` field directly from outside package | Field is unexported or protected; only `Fail` method can set it |
| `TestGetters_CalledDuringTransition` | `unit` | Getters block while write lock held during transition. | Session with `Status="initializing"`; goroutine calling `Run()` holds write lock | Call `GetStatusSafe()` concurrently | Blocks until write lock released; then returns updated "running" status |
