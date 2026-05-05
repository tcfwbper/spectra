# Test Specification: `session_error_test.go`

## Source File Under Test
`entities/session_error.go`

## Test File
`entities/session_error_test.go`

---

## `SessionError`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionError_ValidInputs` | `unit` | Constructs a SessionError with all valid fields. | | `message="connection lost"`, `detail=json.RawMessage('{"code":500}')`, `occurredAt=1700000000`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Processing"` | Returns no error; all getters return the provided values |
| `TestNewSessionError_NilDetail` | `unit` | Constructs a SessionError with nil Detail (represents JSON null). | | `message="timeout"`, `detail=nil`, `occurredAt=1700000000`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Waiting"` | Returns no error; Detail getter returns nil |
| `TestNewSessionError_EmptyObjectDetail` | `unit` | Constructs a SessionError with empty JSON object Detail. | | `message="err"`, `detail=json.RawMessage('{}')`, `occurredAt=1`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Init"` | Returns no error; Detail getter returns `{}` |

### Validation Failures — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionError_EmptyMessage` | `unit` | Rejects empty string message. | | `message=""` with other fields valid | Returns validation error |
| `TestNewSessionError_WhitespaceOnlyMessage` | `unit` | Rejects whitespace-only message. | | `message="   "` with other fields valid | Returns validation error |

### Validation Failures — Detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionError_DetailIsArray` | `unit` | Rejects Detail that is a JSON array. | | `detail=json.RawMessage('[1,2,3]')` with other fields valid | Returns validation error |
| `TestNewSessionError_DetailIsPrimitive` | `unit` | Rejects Detail that is a JSON primitive string. | | `detail=json.RawMessage('"hello"')` with other fields valid | Returns validation error |
| `TestNewSessionError_DetailIsNumber` | `unit` | Rejects Detail that is a JSON number primitive. | | `detail=json.RawMessage('123')` with other fields valid | Returns validation error |
| `TestNewSessionError_DetailIsBooleanTrue` | `unit` | Rejects Detail that is a JSON boolean. | | `detail=json.RawMessage('true')` with other fields valid | Returns validation error |
| `TestNewSessionError_DetailIsInvalidJSON` | `unit` | Rejects Detail that is invalid JSON bytes. | | `detail=json.RawMessage('{broken')` with other fields valid | Returns validation error |

### Validation Failures — OccurredAt

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionError_OccurredAtZero` | `unit` | Rejects OccurredAt value of zero. | | `occurredAt=0` with other fields valid | Returns validation error |
| `TestNewSessionError_OccurredAtNegative` | `unit` | Rejects negative OccurredAt value. | | `occurredAt=-1` with other fields valid | Returns validation error |

### Validation Failures — SessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionError_InvalidSessionID` | `unit` | Rejects SessionID that is not a valid UUID format. | | `sessionID="not-a-uuid"` with other fields valid | Returns validation error |
| `TestNewSessionError_EmptySessionID` | `unit` | Rejects empty string SessionID. | | `sessionID=""` with other fields valid | Returns validation error |

### Validation Failures — FailingState

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionError_EmptyFailingState` | `unit` | Rejects empty string FailingState. | | `failingState=""` with other fields valid | Returns validation error |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionError_Immutability` | `unit` | All fields remain unchanged after construction; no setter methods exist. | Construct a valid SessionError | Attempt to verify no exported setter methods or field assignments are possible | All getter values remain identical to construction inputs |
