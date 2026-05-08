# Test Specification: `event_test.go`

## Source File Under Test
`entities/event.go`

## Test File
`entities/event_test.go`

---

## `Event`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_ValidInputs` | `unit` | Constructs an Event with all valid fields. | | `id="550e8400-e29b-41d4-a716-446655440000"`, `eventType="ReviewCompleted"`, `message="done"`, `payload=json.RawMessage('{"score":95}')`, `emittedBy="ReviewNode"`, `emittedAt=1700000000`, `sessionID="660e8400-e29b-41d4-a716-446655440000"` | Returns no error; all getters return the provided values |
| `TestNewEvent_EmptyMessage` | `unit` | Accepts empty string Message as valid. | | `message=""` with all other fields valid | Returns no error; Message getter returns `""` |
| `TestNewEvent_EmptyObjectPayload` | `unit` | Accepts empty JSON object payload. | | `payload=json.RawMessage('{}')` with all other fields valid | Returns no error; Payload getter returns `{}` |

### Validation Failures — ID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_InvalidID` | `unit` | Rejects ID that is not a valid UUID format. | | `id="not-a-uuid"` with other fields valid | Returns validation error |
| `TestNewEvent_EmptyID` | `unit` | Rejects empty string ID. | | `id=""` with other fields valid | Returns validation error |

### Validation Failures — Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_EmptyType` | `unit` | Rejects empty string Type. | | `eventType=""` with other fields valid | Returns validation error |
| `TestNewEvent_TypeStartsLowercase` | `unit` | Rejects Type starting with a lowercase letter. | | `eventType="reviewNeeded"` with other fields valid | Returns validation error |
| `TestNewEvent_TypeContainsHyphen` | `unit` | Rejects Type with non-alphanumeric characters (hyphen). | | `eventType="Review-Needed"` with other fields valid | Returns validation error |
| `TestNewEvent_TypeContainsUnderscore` | `unit` | Rejects Type with non-alphanumeric characters (underscore). | | `eventType="Review_Needed"` with other fields valid | Returns validation error |
| `TestNewEvent_TypeContainsSpace` | `unit` | Rejects Type with space characters. | | `eventType="Review Needed"` with other fields valid | Returns validation error |

### Validation Failures — Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_NilPayload` | `unit` | Rejects nil Payload. | | `payload=nil` with other fields valid | Returns validation error |
| `TestNewEvent_PayloadIsArray` | `unit` | Rejects Payload that is a JSON array. | | `payload=json.RawMessage('[1,2,3]')` with other fields valid | Returns validation error |
| `TestNewEvent_PayloadIsPrimitiveString` | `unit` | Rejects Payload that is a JSON string primitive. | | `payload=json.RawMessage('"hello"')` with other fields valid | Returns validation error |
| `TestNewEvent_PayloadIsPrimitiveNumber` | `unit` | Rejects Payload that is a JSON number primitive. | | `payload=json.RawMessage('42')` with other fields valid | Returns validation error |
| `TestNewEvent_PayloadIsPrimitiveBoolean` | `unit` | Rejects Payload that is a JSON boolean. | | `payload=json.RawMessage('true')` with other fields valid | Returns validation error |
| `TestNewEvent_PayloadIsPrimitiveNull` | `unit` | Rejects Payload that is JSON null literal. | | `payload=json.RawMessage('null')` with other fields valid | Returns validation error |
| `TestNewEvent_PayloadIsInvalidJSON` | `unit` | Rejects Payload that is invalid JSON bytes. | | `payload=json.RawMessage('{broken')` with other fields valid | Returns validation error |

### Validation Failures — EmittedBy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_EmptyEmittedBy` | `unit` | Rejects empty string EmittedBy. | | `emittedBy=""` with other fields valid | Returns validation error |

### Validation Failures — EmittedAt

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_EmittedAtZero` | `unit` | Rejects EmittedAt value of zero. | | `emittedAt=0` with other fields valid | Returns validation error |
| `TestNewEvent_EmittedAtNegative` | `unit` | Rejects negative EmittedAt value. | | `emittedAt=-1` with other fields valid | Returns validation error |

### Validation Failures — SessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEvent_InvalidSessionID` | `unit` | Rejects SessionID that is not a valid UUID format. | | `sessionID="invalid"` with other fields valid | Returns validation error |
| `TestNewEvent_EmptySessionID` | `unit` | Rejects empty string SessionID. | | `sessionID=""` with other fields valid | Returns validation error |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvent_Immutability` | `unit` | All fields remain unchanged after construction. | Construct a valid Event | Verify all getter values after construction | All getter values remain identical to construction inputs |
