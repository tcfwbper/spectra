# Test Specification: `session_metadata.go`

## Source File Under Test
`entities/session_metadata.go`

## Test File
`entities/session_metadata_test.go`

---

**Fixture Isolation**: All tests create SessionMetadata instances in memory using programmatic construction. No external files or directories are required. SessionMetadata is a plain data structure with no methods of its own.

---

## `SessionMetadata`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_ValidFields` | `unit` | SessionMetadata constructed with all required fields. | | SessionMetadata with `ID=<uuid>`, `WorkflowName="TestFlow"`, `Status="initializing"`, `CreatedAt=1234567890`, `UpdatedAt=1234567890`, `CurrentState="start"`, `SessionData={}`, `Error=nil` | All fields accessible; values match input |
| `TestSessionMetadata_EmptySessionData` | `unit` | SessionData initialized as empty map. | | SessionMetadata with `SessionData=map[string]any{}` | `SessionData` is empty map (length 0); not nil |
| `TestSessionMetadata_NilError` | `unit` | Error field can be nil. | | SessionMetadata with `Error=nil` | `Error` field is nil |

### Happy Path — Field Access

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_AccessID` | `unit` | ID field accessible directly. | SessionMetadata with `ID="test-uuid-123"` | Access `metadata.ID` | Returns `"test-uuid-123"` |
| `TestSessionMetadata_AccessWorkflowName` | `unit` | WorkflowName field accessible directly. | SessionMetadata with `WorkflowName="MyWorkflow"` | Access `metadata.WorkflowName` | Returns `"MyWorkflow"` |
| `TestSessionMetadata_AccessStatus` | `unit` | Status field accessible directly. | SessionMetadata with `Status="running"` | Access `metadata.Status` | Returns `"running"` |
| `TestSessionMetadata_AccessTimestamps` | `unit` | Timestamp fields accessible directly. | SessionMetadata with `CreatedAt=1000`, `UpdatedAt=2000` | Access `metadata.CreatedAt` and `metadata.UpdatedAt` | Returns `1000` and `2000` respectively |
| `TestSessionMetadata_AccessCurrentState` | `unit` | CurrentState field accessible directly. | SessionMetadata with `CurrentState="processing"` | Access `metadata.CurrentState` | Returns `"processing"` |
| `TestSessionMetadata_AccessSessionData` | `unit` | SessionData field accessible directly. | SessionMetadata with `SessionData={"key": "value"}` | Access `metadata.SessionData` | Returns map with `"key": "value"` |
| `TestSessionMetadata_AccessError` | `unit` | Error field accessible directly. | SessionMetadata with `Error=*AgentError{...}` | Access `metadata.Error` | Returns `*AgentError` instance |

### Happy Path — JSON Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_JSONMarshalAllFields` | `unit` | All fields serialized to JSON. | | SessionMetadata with all fields populated (except Error=nil) | JSON contains `id`, `workflowName`, `status`, `createdAt`, `updatedAt`, `currentState`, `sessionData`; no `error` field |
| `TestSessionMetadata_JSONMarshalWithError` | `unit` | Error field serialized when non-nil. | | SessionMetadata with `Error=*AgentError{Issuer:"Agent", Message:"test error"}` | JSON contains `error` field with nested structure |
| `TestSessionMetadata_JSONMarshalErrorOmitEmpty` | `unit` | Error field omitted when nil. | | SessionMetadata with `Error=nil` | JSON does not contain `error` field; no `"error": null` |
| `TestSessionMetadata_JSONMarshalNestedSessionData` | `unit` | SessionData with nested structures serialized correctly. | | SessionMetadata with `SessionData={"key": {"nested": [1, 2, 3]}}` | JSON contains nested structure; deserializes back to same structure |
| `TestSessionMetadata_JSONUnmarshalAllFields` | `unit` | All fields deserialized from JSON. | | Valid JSON with all fields | SessionMetadata populated with correct values; all fields match JSON |
| `TestSessionMetadata_JSONUnmarshalErrorFieldPresent` | `unit` | Error field deserialized when present in JSON. | | JSON with `error` field containing AgentError structure | SessionMetadata `Error` field populated with `*AgentError` |
| `TestSessionMetadata_JSONUnmarshalErrorFieldAbsent` | `unit` | Error field nil when absent from JSON. | | JSON without `error` field | SessionMetadata `Error` field is nil |

### Not Immutable

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_FieldsMutable` | `unit` | SessionMetadata fields can be modified. | SessionMetadata with `Status="initializing"` | Assign `metadata.Status = "running"` | Field updated; no error or panic |
| `TestSessionMetadata_SessionDataMutable` | `unit` | SessionData map can be modified. | SessionMetadata with `SessionData={}` | Assign `metadata.SessionData["key"] = "value"` | Map updated; value stored |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_CopyIndependent` | `unit` | Copying SessionMetadata creates independent instance. | SessionMetadata with `Status="running"`, `SessionData={"key": "value"}` | Create copy `copy := metadata`; modify `copy.Status = "completed"` | Original `Status` remains `"running"`; copy has `"completed"` |
| `TestSessionMetadata_SessionDataShallowCopy` | `unit` | SessionData map reference shared in shallow copy. | SessionMetadata with `SessionData={"key": "value"}` | Create copy `copy := metadata`; modify `copy.SessionData["key"] = "modified"` | Both original and copy `SessionData` see `"modified"` (map is reference type) |

### Invariants — Type Integrity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_NoEmbeddedLocks` | `unit` | SessionMetadata contains no locks or channels. | | SessionMetadata instance | Struct has no `sync.Mutex`, `sync.RWMutex`, or channel fields; safe to copy |
| `TestSessionMetadata_ValueType` | `unit` | SessionMetadata is a value type (struct). | | SessionMetadata instance | Can be copied by value; not a pointer type |

### Invariants — Status Enumeration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_StatusInitializing` | `unit` | Status can be "initializing". | | SessionMetadata with `Status="initializing"` | `Status` field equals `"initializing"` |
| `TestSessionMetadata_StatusRunning` | `unit` | Status can be "running". | | SessionMetadata with `Status="running"` | `Status` field equals `"running"` |
| `TestSessionMetadata_StatusCompleted` | `unit` | Status can be "completed". | | SessionMetadata with `Status="completed"` | `Status` field equals `"completed"` |
| `TestSessionMetadata_StatusFailed` | `unit` | Status can be "failed". | | SessionMetadata with `Status="failed"` | `Status` field equals `"failed"` |

### Invariants — Timestamp Ordering

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_CreatedAtPositive` | `unit` | CreatedAt must be positive. | | SessionMetadata with `CreatedAt=1234567890` | `CreatedAt > 0` |
| `TestSessionMetadata_UpdatedAtGreaterOrEqual` | `unit` | UpdatedAt must be >= CreatedAt. | | SessionMetadata with `CreatedAt=1000`, `UpdatedAt=2000` | `UpdatedAt >= CreatedAt` |
| `TestSessionMetadata_TimestampsEqual` | `unit` | UpdatedAt can equal CreatedAt. | | SessionMetadata with `CreatedAt=1000`, `UpdatedAt=1000` | `CreatedAt == UpdatedAt` |

### Invariants — Error Correlation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_ErrorNilWhenNotFailed` | `unit` | Error is nil when Status is not "failed". | | SessionMetadata with `Status="running"`, `Error=nil` | `Error` is nil |
| `TestSessionMetadata_ErrorNonNilWhenFailed` | `unit` | Error is non-nil when Status is "failed". | | SessionMetadata with `Status="failed"`, `Error=*AgentError{...}` | `Error` is non-nil |

### Invariants — Non-Empty Fields

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_IDNonEmpty` | `unit` | ID must be non-empty. | | SessionMetadata with `ID="test-uuid"` | `ID` is non-empty string |
| `TestSessionMetadata_WorkflowNameNonEmpty` | `unit` | WorkflowName must be non-empty. | | SessionMetadata with `WorkflowName="TestFlow"` | `WorkflowName` is non-empty string |
| `TestSessionMetadata_CurrentStateNonEmpty` | `unit` | CurrentState must be non-empty. | | SessionMetadata with `CurrentState="start"` | `CurrentState` is non-empty string |

### Invariants — SessionData Never Nil

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_SessionDataNotNil` | `unit` | SessionData must not be nil. | | SessionMetadata with `SessionData=map[string]any{}` | `SessionData` is not nil; can be empty map |

### Validation Failures — JSON Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_JSONMarshalUnserializableSessionData` | `unit` | JSON marshaling fails for un-serializable SessionData values. | | SessionMetadata with `SessionData={"ch": make(chan int)}` | `json.Marshal()` returns error matching `/json: unsupported type/i` |

### Validation Failures — JSON Deserialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_JSONUnmarshalInvalidJSON` | `unit` | Deserialization fails for malformed JSON. | | JSON string `{"id": "test"` (missing closing brace) | `json.Unmarshal()` returns error matching `/unexpected end of JSON/i` |
| `TestSessionMetadata_JSONUnmarshalWrongType` | `unit` | Deserialization fails when field has wrong type. | | JSON with `"createdAt": "not-a-number"` | `json.Unmarshal()` returns error matching `/cannot unmarshal string into.*int64/i` |

### Boundary Values — UUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_ValidUUIDv4` | `unit` | ID can be valid UUID v4. | | SessionMetadata with `ID="550e8400-e29b-41d4-a716-446655440000"` | `ID` stored correctly |
| `TestSessionMetadata_EmptyID` | `unit` | ID can be empty string (validation not enforced by struct). | | SessionMetadata with `ID=""` | `ID` is empty string (note: violates invariant but struct does not enforce) |

### Boundary Values — Timestamps

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_ZeroTimestamps` | `unit` | Timestamps can be zero (violates invariant but struct allows). | | SessionMetadata with `CreatedAt=0`, `UpdatedAt=0` | Timestamps stored as `0` (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_NegativeTimestamps` | `unit` | Timestamps can be negative (violates invariant but struct allows). | | SessionMetadata with `CreatedAt=-1`, `UpdatedAt=-1` | Timestamps stored as `-1` (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_UpdatedAtLessThanCreatedAt` | `unit` | UpdatedAt can be less than CreatedAt (violates invariant but struct allows). | | SessionMetadata with `CreatedAt=2000`, `UpdatedAt=1000` | Timestamps stored (note: violates invariant but struct does not enforce) |

### Boundary Values — Error Types

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_AgentError` | `unit` | Error can be AgentError type. | | SessionMetadata with `Error=*AgentError{Issuer:"Agent", Message:"test"}` | `Error` field holds `*AgentError` |
| `TestSessionMetadata_RuntimeError` | `unit` | Error can be RuntimeError type. | | SessionMetadata with `Error=*RuntimeError{Issuer:"Runtime", Message:"test"}` | `Error` field holds `*RuntimeError` |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata_EmptyWorkflowName` | `unit` | WorkflowName can be empty (violates invariant but struct allows). | | SessionMetadata with `WorkflowName=""` | `WorkflowName` stored as empty string (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_EmptyCurrentState` | `unit` | CurrentState can be empty (violates invariant but struct allows). | | SessionMetadata with `CurrentState=""` | `CurrentState` stored as empty string (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_InvalidStatus` | `unit` | Status can be invalid value (violates invariant but struct allows). | | SessionMetadata with `Status="invalid"` | `Status` stored as `"invalid"` (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_NilSessionData` | `unit` | SessionData can be nil (violates invariant but struct allows). | | SessionMetadata with `SessionData=nil` | `SessionData` is nil (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_ErrorWithoutFailed` | `unit` | Error can be set without Status="failed" (violates invariant but struct allows). | | SessionMetadata with `Status="running"`, `Error=*AgentError{...}` | Both fields stored (note: violates invariant but struct does not enforce) |
| `TestSessionMetadata_FailedWithoutError` | `unit` | Status can be "failed" without Error set (violates invariant but struct allows). | | SessionMetadata with `Status="failed"`, `Error=nil` | Both fields stored (note: violates invariant but struct does not enforce) |
