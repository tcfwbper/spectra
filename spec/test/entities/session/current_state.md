# Test Specification: `current_state.go`

## Source File Under Test
`entities/session/current_state.go`

## Test File
`entities/session/current_state_test.go`

---

**Fixture Isolation**: All tests create Session instances in memory using programmatic construction. No external files or directories are required unless explicitly stated in the Setup column. Mock dependencies (SessionMetadataStore, loggers, etc.) are created within each test.

---

## `Session` Current State Method

### Happy Path — UpdateCurrentStateSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_UpdatesState` | `unit` | Updates CurrentState to new value. | Session with `CurrentState="node1"`, `UpdatedAt=T0` | Call `UpdateCurrentStateSafe("node2")` | Returns `nil`; `CurrentState="node2"`; `UpdatedAt > T0` |
| `TestUpdateCurrentStateSafe_MultipleUpdates` | `unit` | Multiple updates each change CurrentState. | Session with `CurrentState="start"` | Call `UpdateCurrentStateSafe("processing")`, then `UpdateCurrentStateSafe("review")`, then `UpdateCurrentStateSafe("complete")` | All return `nil`; final `CurrentState="complete"` |
| `TestUpdateCurrentStateSafe_PersistsToStore` | `unit` | Persists updated session metadata to store. | Mock SessionMetadataStore; session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node2")` | Returns `nil`; store write called with updated session; in-memory `CurrentState="node2"` |
| `TestUpdateCurrentStateSafe_AcceptsAnyNonEmptyString` | `unit` | Accepts any non-empty string as newState (no workflow validation). | Session with `CurrentState="start"` | Call `UpdateCurrentStateSafe("UnknownNodeName")` | Returns `nil`; `CurrentState="UnknownNodeName"` (validation is caller's responsibility) |

### Happy Path — Self-Transition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_SelfTransition` | `unit` | Accepts newState equal to current CurrentState (idempotent at in-memory level). | Session with `CurrentState="processing"`, `UpdatedAt=T0` | Call `UpdateCurrentStateSafe("processing")` | Returns `nil`; `CurrentState` remains "processing"; `UpdatedAt` advances (> T0) |

### Validation Failures — Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_EmptyNewState` | `unit` | Logs warning and returns nil for empty newState; state unchanged. | Session with `CurrentState="node1"`, `UpdatedAt=T0`; mock logger | Call `UpdateCurrentStateSafe("")` | Returns `nil`; `CurrentState` remains "node1"; `UpdatedAt` unchanged (T0); warning logged matching `/UpdateCurrentStateSafe called with empty newState.*in-memory state unchanged/i`; lock not acquired |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_Idempotent` | `unit` | Repeated calls with same newState accepted (UpdatedAt advances each time). | Session with `CurrentState="node1"`, `UpdatedAt=T0` | Call `UpdateCurrentStateSafe("node2")` twice | Both return `nil`; `CurrentState="node2"`; `UpdatedAt` advances after each call |

### Atomic Replacement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_AtomicUpdate` | `unit` | CurrentState and UpdatedAt updated atomically. | Session with `CurrentState="old"`, `UpdatedAt=T0`; concurrent goroutine reading `GetCurrentStateSafe()` in loop | Call `UpdateCurrentStateSafe("new")` | Concurrent reader never observes intermediate state; both CurrentState and UpdatedAt updated together |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCurrentState_ConcurrentUpdates` | `race` | Concurrent updates serialized by write lock; last write wins. | Session with `CurrentState="start"` | 10 goroutines call `UpdateCurrentStateSafe()` with different values simultaneously | All return `nil`; final `CurrentState` is one of the written values; no race conditions |
| `TestCurrentState_ConcurrentReadWrite` | `race` | Concurrent reads and writes are serialized correctly. | Session with `CurrentState="node1"` | 50 goroutines call `GetCurrentStateSafe()` concurrently with 1 goroutine calling `UpdateCurrentStateSafe("node2")` | No race conditions; readers see either "node1" or "node2" consistently |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_ReleasesLock` | `unit` | Write lock released after update. | Session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node2")`; then call `GetCurrentStateSafe()` | Both succeed without deadlock |
| `TestUpdateCurrentStateSafe_EmptyInputNoLockAcquisition` | `unit` | Empty newState does not acquire lock. | Session with `CurrentState="node1"`; concurrent goroutine holding write lock indefinitely | Call `UpdateCurrentStateSafe("")` | Returns `nil` immediately without blocking; warning logged |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_PersistenceFailureLogged` | `unit` | Persistence failures logged but not returned. | Mock SessionMetadataStore that returns error on write; session with `CurrentState="node1"`; mock logger | Call `UpdateCurrentStateSafe("node2")` | Returns `nil`; in-memory `CurrentState="node2"`; warning logged matching `/UpdateCurrentStateSafe persistence failed/i` or error message |
| `TestUpdateCurrentStateSafe_PersistenceFailureDoesNotRevert` | `unit` | In-memory state authoritative even when persistence fails. | Mock SessionMetadataStore that fails; session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node2")` | Returns `nil`; in-memory `CurrentState="node2"` persists |

### Invariants — Always Returns Nil

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_AlwaysReturnsNil` | `unit` | UpdateCurrentStateSafe never returns non-nil error. | Session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node2")` | Returns `nil` |
| `TestUpdateCurrentStateSafe_AlwaysReturnsNilOnEmpty` | `unit` | Returns nil even for empty newState. | Session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("")` | Returns `nil` (not an error; warning logged) |
| `TestUpdateCurrentStateSafe_AlwaysReturnsNilOnPersistenceFailure` | `unit` | Returns nil even when persistence fails. | Mock SessionMetadataStore that fails; session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node2")` | Returns `nil` |

### Invariants — UpdatedAt Refresh

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_RefreshesUpdatedAt` | `unit` | UpdateCurrentStateSafe refreshes UpdatedAt. | Session with `CurrentState="node1"`, `UpdatedAt=T0` | Wait 1 second; call `UpdateCurrentStateSafe("node2")` | Returns `nil`; `UpdatedAt > T0` |
| `TestUpdateCurrentStateSafe_UpdatedAtInSameCriticalSection` | `unit` | UpdatedAt refreshed in same critical section as state write. | Session with `CurrentState="old"`, `UpdatedAt=T0`; concurrent goroutine reading `GetCurrentStateSafe()` in loop | Call `UpdateCurrentStateSafe("new")` | Concurrent reader observes consistent snapshot; never sees new CurrentState with old UpdatedAt |
| `TestUpdateCurrentStateSafe_EmptyInputDoesNotRefreshUpdatedAt` | `unit` | Empty newState does not mutate UpdatedAt. | Session with `CurrentState="node1"`, `UpdatedAt=T0` | Call `UpdateCurrentStateSafe("")` | Returns `nil`; `UpdatedAt` remains T0 (no mutation) |

### Invariants — Write Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCurrentState_WriteSerializationLastWriteWins` | `unit` | Concurrent writes serialized; last write wins. | Session with `CurrentState="start"` | Call `UpdateCurrentStateSafe("node1")` and `UpdateCurrentStateSafe("node2")` concurrently | Both return `nil`; final `CurrentState` is either "node1" or "node2" (last write wins) |

### Invariants — Memory Authority

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCurrentState_InMemoryAuthoritative` | `unit` | In-memory CurrentState is source of truth. | Mock SessionMetadataStore that fails; session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node2")`; then `GetCurrentStateSafe()` | Update returns `nil`; get returns "node2" from in-memory state |

### Invariants — No Workflow Validation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_NoWorkflowValidation` | `unit` | Method does not validate newState against workflow definition. | Session with `CurrentState="validNode"` | Call `UpdateCurrentStateSafe("NonExistentNode")` | Returns `nil`; `CurrentState="NonExistentNode"` (caller responsible for validation) |
| `TestUpdateCurrentStateSafe_AcceptsInvalidNodeName` | `unit` | Accepts newState that is not a valid workflow node (caller responsibility). | Session with `CurrentState="start"` | Call `UpdateCurrentStateSafe("!!!invalid!!!")` | Returns `nil`; `CurrentState="!!!invalid!!!"` |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_WhitespaceOnlyState` | `unit` | Accepts whitespace-only newState (non-empty string). | Session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("   ")` (spaces) | Returns `nil`; `CurrentState="   "` (whitespace not trimmed; caller responsibility) |
| `TestUpdateCurrentStateSafe_VeryLongStateName` | `unit` | Accepts very long state name. | Session with `CurrentState="start"` | Call `UpdateCurrentStateSafe(<10KB string>)` | Returns `nil`; `CurrentState` set to long string |
| `TestUpdateCurrentStateSafe_UnicodeStateName` | `unit` | Accepts Unicode characters in state name. | Session with `CurrentState="start"` | Call `UpdateCurrentStateSafe("处理节点")` | Returns `nil`; `CurrentState="处理节点"`; Unicode preserved |
| `TestUpdateCurrentStateSafe_SpecialCharactersInStateName` | `unit` | Accepts special characters in state name (no sanitization). | Session with `CurrentState="node1"` | Call `UpdateCurrentStateSafe("node-2_final.state")` | Returns `nil`; `CurrentState="node-2_final.state"` |
| `TestUpdateCurrentStateSafe_SelfLoopNotBlockedAtThisLayer` | `unit` | Self-transition accepted even though workflow definitions reject self-loops (workflow validation happens elsewhere). | Session with `CurrentState="processing"` | Call `UpdateCurrentStateSafe("processing")` | Returns `nil`; `CurrentState` remains "processing"; UpdatedAt advances |
