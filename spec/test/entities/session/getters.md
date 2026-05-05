# Test Specification: `getters_test.go`

## Source File Under Test
`entities/session/getters.go`

## Test File
`entities/session/getters_test.go`

---

## `GetStatusSafe`

### Happy Path — GetStatusSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetStatusSafe_Initializing` | `unit` | Returns "initializing" for a newly constructed session. | Construct session via `NewSession` | Call `GetStatusSafe()` | Returns `"initializing"` |
| `TestGetStatusSafe_Running` | `unit` | Returns "running" after Run() succeeds. | Construct session; call `Run()` | Call `GetStatusSafe()` | Returns `"running"` |
| `TestGetStatusSafe_Completed` | `unit` | Returns "completed" after Done() succeeds. | Construct session; call `Run()`; call `Done(ch)` | Call `GetStatusSafe()` | Returns `"completed"` |
| `TestGetStatusSafe_Failed` | `unit` | Returns "failed" after Fail() succeeds. | Construct session; call `Fail(agentErr, ch)` | Call `GetStatusSafe()` | Returns `"failed"` |

---

## `GetCurrentStateSafe`

### Happy Path — GetCurrentStateSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetCurrentStateSafe_Initial` | `unit` | Returns the entry node after construction. | Construct session with `entryNode="start"` | Call `GetCurrentStateSafe()` | Returns `"start"` |
| `TestGetCurrentStateSafe_AfterUpdate` | `unit` | Returns the updated state after UpdateCurrentStateSafe. | Construct session; call `UpdateCurrentStateSafe("processing")` | Call `GetCurrentStateSafe()` | Returns `"processing"` |

---

## `GetErrorSafe`

### Happy Path — GetErrorSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetErrorSafe_NoError` | `unit` | Returns nil when session has not failed. | Construct session | Call `GetErrorSafe()` | Returns `nil` |
| `TestGetErrorSafe_AgentError` | `unit` | Returns the stored *AgentError after Fail. | Construct session; call `Fail(agentErr, ch)` | Call `GetErrorSafe()` | Returns the same `*AgentError` instance |
| `TestGetErrorSafe_RuntimeError` | `unit` | Returns the stored *RuntimeError after Fail. | Construct session; call `Run()`; call `Fail(runtimeErr, ch)` | Call `GetErrorSafe()` | Returns the same `*RuntimeError` instance |

---

## `GetMetadataSnapshotSafe`

### Happy Path — GetMetadataSnapshotSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetMetadataSnapshotSafe_ReturnsAllFields` | `unit` | Returns a complete snapshot of all metadata fields. | Construct session with known inputs; call `Run()` | Call `GetMetadataSnapshotSafe()` | Returns `SessionMetadata` with `ID`, `WorkflowName`, `Status=="running"`, `CreatedAt`, `UpdatedAt`, `CurrentState`, `SessionData` (non-nil empty map), `Error==nil` matching session state |
| `TestGetMetadataSnapshotSafe_EmptySessionData` | `unit` | Returns non-nil empty map when no data has been set. | Construct session | Call `GetMetadataSnapshotSafe()` | `snapshot.SessionData` is non-nil and length 0 |
| `TestGetMetadataSnapshotSafe_WithSessionData` | `unit` | Returns shallow copy of populated SessionData. | Construct session; call `UpdateSessionDataSafe("k", "v")` | Call `GetMetadataSnapshotSafe()` | `snapshot.SessionData["k"] == "v"` |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetMetadataSnapshotSafe_MapIsolation` | `unit` | Mutating the returned map does not affect the session. | Construct session; call `UpdateSessionDataSafe("k", "v")`; obtain snapshot | Mutate `snapshot.SessionData["k"] = "modified"` | `GetSessionDataSafe("k")` still returns `("v", true)` |
| `TestGetMetadataSnapshotSafe_InsertionIsolation` | `unit` | Inserting into the returned map does not affect the session. | Construct session; obtain snapshot | Insert `snapshot.SessionData["new"] = "x"` | `GetSessionDataSafe("new")` returns `(nil, false)` |

