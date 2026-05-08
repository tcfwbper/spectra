# Test Specification: `runtime_message_test.go`

## Source File Under Test

`entities/runtime_message.go`

## Test File

`entities/runtime_message_test.go`

---

## `RuntimeMessage`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeMessage_EventType` | `unit` | Constructs a valid RuntimeMessage with type "event". | | `msgType="event"`, `payload=json.RawMessage('{"eventType":"ReviewNeeded"}')`, `claudeSessionID="sess-123"` | Returns no error; `Type()` returns `"event"`; `Payload()` returns the given JSON object; `ClaudeSessionID()` returns `"sess-123"` |
| `TestNewRuntimeMessage_ErrorType` | `unit` | Constructs a valid RuntimeMessage with type "error". | | `msgType="error"`, `payload=json.RawMessage('{"detail":"something failed"}')`, `claudeSessionID="sess-456"` | Returns no error; `Type()` returns `"error"`; `Payload()` returns the given JSON object; `ClaudeSessionID()` returns `"sess-456"` |
| `TestNewRuntimeMessage_EmptyClaudeSessionID` | `unit` | Constructs a valid RuntimeMessage with empty ClaudeSessionID. | | `msgType="event"`, `payload=json.RawMessage('{"key":"value"}')`, `claudeSessionID=""` | Returns no error; `ClaudeSessionID()` returns `""` |
| `TestNewRuntimeMessage_EmptyPayloadObject` | `unit` | Constructs a valid RuntimeMessage with an empty JSON object payload. | | `msgType="event"`, `payload=json.RawMessage('{}')`, `claudeSessionID=""` | Returns no error; `Payload()` returns `{}` |

### Validation Failures — Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeMessage_EmptyType` | `unit` | Returns validation error when type is empty string. | | `msgType=""`, `payload=json.RawMessage('{}')`, `claudeSessionID=""` | Returns a validation error; no RuntimeMessage is created |
| `TestNewRuntimeMessage_UnrecognizedType` | `unit` | Returns validation error when type is not "event" or "error". | | `msgType="unknown"`, `payload=json.RawMessage('{}')`, `claudeSessionID=""` | Returns a validation error indicating the type is not recognized |
| `TestNewRuntimeMessage_WarningType` | `unit` | Returns validation error for "warning" type. | | `msgType="warning"`, `payload=json.RawMessage('{}')`, `claudeSessionID=""` | Returns a validation error indicating the type is not recognized |

### Validation Failures — Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeMessage_NilPayload` | `unit` | Returns validation error when payload is nil. | | `msgType="event"`, `payload=nil`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsArray` | `unit` | Returns validation error when payload is a JSON array. | | `msgType="event"`, `payload=json.RawMessage('[1,2,3]')`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsEmptyArray` | `unit` | Returns validation error when payload is an empty JSON array. | | `msgType="event"`, `payload=json.RawMessage('[]')`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsString` | `unit` | Returns validation error when payload is a JSON string primitive. | | `msgType="event"`, `payload=json.RawMessage('"hello"')`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsNumber` | `unit` | Returns validation error when payload is a JSON number primitive. | | `msgType="event"`, `payload=json.RawMessage('123')`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsBoolean` | `unit` | Returns validation error when payload is a JSON boolean primitive. | | `msgType="event"`, `payload=json.RawMessage('true')`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsNull` | `unit` | Returns validation error when payload is JSON null. | | `msgType="event"`, `payload=json.RawMessage('null')`, `claudeSessionID=""` | Returns a validation error indicating payload must be a JSON object |
| `TestNewRuntimeMessage_PayloadIsInvalidJSON` | `unit` | Returns validation error when payload is invalid JSON. | | `msgType="event"`, `payload=json.RawMessage('{invalid')`, `claudeSessionID=""` | Returns a validation error |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_PayloadImmutability` | `unit` | Modifying the original payload slice after construction does not affect the stored value. | Construct a RuntimeMessage with a known payload; keep reference to original slice | Mutate the original `json.RawMessage` slice after construction | `Payload()` returns the original unmodified JSON object |
| `TestRuntimeMessage_GetterReturnImmutability` | `unit` | Modifying the slice returned by Payload() does not affect the internal state. | Construct a valid RuntimeMessage | Call `Payload()`, mutate the returned slice, call `Payload()` again | Second `Payload()` call returns the original unmodified JSON object |
