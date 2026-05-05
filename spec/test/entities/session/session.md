# Test Specification: `session_test.go`

## Source File Under Test
`entities/session/session.go`

## Test File
`entities/session/session_test.go`

---

## `NewSession`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSession_ValidInputs` | `unit` | Constructs a session with all valid inputs and verifies initial field values. | None | `id="550e8400-e29b-41d4-a716-446655440000"`, `workflowName="my-workflow"`, `entryNode="start"`, `createdAt=1700000000` | Returns non-nil `*Session`, `nil` error; `ID` equals input id; `WorkflowName` equals `"my-workflow"`; `Status == "initializing"`; `CreatedAt == 1700000000`; `UpdatedAt == CreatedAt`; `CurrentState == "start"`; `SessionData` is empty non-nil map; `Error == nil`; `EventHistory` is empty slice |
| `TestNewSession_MinimalCreatedAt` | `unit` | Accepts `createdAt=1` as the minimum valid positive timestamp. | None | `id` valid UUID, `workflowName="w"`, `entryNode="n"`, `createdAt=1` | Returns non-nil `*Session`, `nil` error; `CreatedAt == 1` |

### Validation Failures — id

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSession_InvalidUUID_Empty` | `unit` | Rejects empty string as session ID. | None | `id=""`, other params valid | Returns `nil` session and error with message `"invalid session ID: must be a valid UUID"` |
| `TestNewSession_InvalidUUID_Malformed` | `unit` | Rejects a malformed UUID string. | None | `id="not-a-uuid"`, other params valid | Returns `nil` session and error with message `"invalid session ID: must be a valid UUID"` |
| `TestNewSession_InvalidUUID_TooShort` | `unit` | Rejects a UUID with missing segments. | None | `id="550e8400-e29b-41d4"`, other params valid | Returns `nil` session and error with message `"invalid session ID: must be a valid UUID"` |

### Validation Failures — workflowName

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSession_EmptyWorkflowName` | `unit` | Rejects empty workflow name. | None | `id` valid UUID, `workflowName=""`, `entryNode="start"`, `createdAt=1700000000` | Returns `nil` session and error with message `"workflow name cannot be empty"` |

### Validation Failures — entryNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSession_EmptyEntryNode` | `unit` | Rejects empty entry node. | None | `id` valid UUID, `workflowName="w"`, `entryNode=""`, `createdAt=1700000000` | Returns `nil` session and error with message `"entry node cannot be empty"` |

### Validation Failures — createdAt

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSession_CreatedAtZero` | `unit` | Rejects zero timestamp. | None | `id` valid UUID, `workflowName="w"`, `entryNode="n"`, `createdAt=0` | Returns `nil` session and error with message `"createdAt must be a positive POSIX timestamp"` |
| `TestNewSession_CreatedAtNegative` | `unit` | Rejects negative timestamp. | None | `id` valid UUID, `workflowName="w"`, `entryNode="n"`, `createdAt=-1` | Returns `nil` session and error with message `"createdAt must be a positive POSIX timestamp"` |
