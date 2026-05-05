# Test Specification: `current_state_test.go`

## Source File Under Test
`entities/session/current_state.go`

## Test File
`entities/session/current_state_test.go`

---

## `UpdateCurrentStateSafe`

### Happy Path — UpdateCurrentStateSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_ValidState` | `unit` | Updates CurrentState to a new non-empty value. | Construct session via `NewSession` with `entryNode="start"` | `newState="processing"` | Returns `nil`; `GetCurrentStateSafe()` returns `"processing"` |
| `TestUpdateCurrentStateSafe_SelfTransition` | `unit` | Accepts self-transition (same value as current state). | Construct session with `entryNode="start"` | `newState="start"` | Returns `nil`; `GetCurrentStateSafe()` returns `"start"`; `UpdatedAt` advances |
| `TestUpdateCurrentStateSafe_UpdatesUpdatedAt` | `unit` | Successful update advances UpdatedAt. | Construct session; record initial `UpdatedAt` | `newState="next"` | Returns `nil`; `GetMetadataSnapshotSafe().UpdatedAt >= initial UpdatedAt` |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_EmptyState` | `unit` | Rejects empty string without acquiring lock or mutating state. | Construct session with `entryNode="start"` | `newState=""` | Returns error with message `"current state cannot be empty"`; `GetCurrentStateSafe()` still returns `"start"` |

