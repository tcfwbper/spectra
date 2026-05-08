# Test Specification: `runtime_response_test.go`

## Source File Under Test

`entities/runtime_response.go`

## Test File

`entities/runtime_response_test.go`

---

## `RuntimeResponse`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSuccessResponse_WithMessage` | `unit` | Creates a success response with a non-empty message. | | `message="operation completed"` | `Status()` returns `"success"`; `Message()` returns `"operation completed"` |
| `TestSuccessResponse_EmptyMessage` | `unit` | Creates a success response with an empty message. | | `message=""` | `Status()` returns `"success"`; `Message()` returns `""` |
| `TestErrorResponse_WithMessage` | `unit` | Creates an error response with a non-empty message. | | `message="something failed"` | `Status()` returns `"error"`; `Message()` returns `"something failed"` |
| `TestErrorResponse_EmptyMessage` | `unit` | Creates an error response with an empty message. | | `message=""` | `Status()` returns `"error"`; `Message()` returns `""` |
| `TestSuccessResponse_MessageWithNewlines` | `unit` | Creates a success response with message containing newline characters. | | `message="line1\nline2\nline3"` | `Status()` returns `"success"`; `Message()` returns `"line1\nline2\nline3"` |
| `TestErrorResponse_MessageWithNewlines` | `unit` | Creates an error response with message containing newline characters. | | `message="error\ndetails"` | `Status()` returns `"error"`; `Message()` returns `"error\ndetails"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_StatusImmutable` | `unit` | Status cannot be modified after construction. | Construct a `SuccessResponse` | Attempt to assign to the Status field directly (struct literal) | Compilation fails or field is unexported; getter always returns `"success"` |
| `TestRuntimeResponse_MessageImmutable` | `unit` | Message cannot be modified after construction. | Construct an `ErrorResponse` with `message="original"` | Attempt to assign to the Message field directly (struct literal) | Compilation fails or field is unexported; getter always returns `"original"` |
