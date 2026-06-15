# Test Specification: `session_metadata_test.go`

## Source File Under Test
`entities/session/session_metadata.go`

## Test File
`entities/session/session_metadata_test.go`

---

## `SessionMetadata`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_AccessViaSession` | `unit` | SessionMetadata fields are accessible directly on the Session (embedded struct). | Construct session via `NewSession` with known inputs (including `pid=42`) | Access `session.ID`, `session.WorkflowName`, `session.Pid`, `session.Status`, `session.CreatedAt`, `session.UpdatedAt`, `session.CurrentState`, `session.SessionData`, `session.Error` | All fields match the values established by `NewSession`; `session.Pid == 42` |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_SnapshotIsDetachedCopy` | `unit` | The snapshot returned by GetMetadataSnapshotSafe is a detached value copy. | Construct session; call `GetMetadataSnapshotSafe()` to get snapshot; then call `Run()` on session | Compare snapshot.Status with `GetStatusSafe()` | `snapshot.Status == "initializing"` while `GetStatusSafe() == "running"` — snapshot is detached |
