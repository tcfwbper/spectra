# Test Specification: `runtime_error_test.go`

## Source File Under Test
`entities/runtime_error.go`

## Test File
`entities/runtime_error_test.go`

---

## `RuntimeError`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeError_ValidInputs` | `unit` | Constructs a RuntimeError with all valid fields. | | `issuer="MessageRouter"`, `message="socket creation failed"`, `detail=json.RawMessage('{"err":"ECONNREFUSED"}')`, `occurredAt=1700000000`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Connecting"` | Returns no error; all getters return the provided values |
| `TestNewRuntimeError_ArbitraryIssuerName` | `unit` | Accepts any non-whitespace issuer string without verifying against known components. | | `issuer="UnknownComponent99"`, `message="err"`, `detail=nil`, `occurredAt=1`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Init"` | Returns no error; Issuer getter returns `"UnknownComponent99"` |

### Validation Failures — Issuer

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeError_EmptyIssuer` | `unit` | Rejects empty string Issuer. | | `issuer=""`, other fields valid | Returns validation error |
| `TestNewRuntimeError_WhitespaceOnlyIssuer` | `unit` | Rejects whitespace-only Issuer. | | `issuer="   "`, other fields valid | Returns validation error |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewRuntimeError_PropagatesMessageError` | `unit` | Propagates validation error when Message is invalid. | | `issuer="Router"`, `message=""`, other fields valid | Returns validation error from SessionError |
| `TestNewRuntimeError_PropagatesSessionIDError` | `unit` | Propagates validation error when SessionID is invalid. | | `issuer="Router"`, `sessionID="bad"`, other fields valid | Returns validation error from SessionError |
| `TestNewRuntimeError_PropagatesOccurredAtError` | `unit` | Propagates validation error when OccurredAt is invalid. | | `issuer="Router"`, `occurredAt=-1`, other fields valid | Returns validation error from SessionError |
| `TestNewRuntimeError_PropagatesDetailError` | `unit` | Propagates validation error when Detail is a JSON array. | | `issuer="Router"`, `detail=json.RawMessage('[1]')`, other fields valid | Returns validation error from SessionError |
| `TestNewRuntimeError_PropagatesFailingStateError` | `unit` | Propagates validation error when FailingState is empty. | | `issuer="Router"`, `failingState=""`, other fields valid | Returns validation error from SessionError |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeError_Immutability` | `unit` | All fields including Issuer remain unchanged after construction. | Construct a valid RuntimeError | Verify all getter values after construction | All getter values remain identical to construction inputs |
