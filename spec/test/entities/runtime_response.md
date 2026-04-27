# Test Specification: `runtime_response.go`

## Source File Under Test
`entities/runtime_response.go`

## Test File
`entities/runtime_response_test.go`

---

## `RuntimeResponse`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_ValidSuccessResponse` | `unit` | Creates RuntimeResponse with status=success and message. | | `Status="success"`, `Message="Event 'DraftCompleted' recorded successfully"` | Returns valid RuntimeResponse; all fields match input |
| `TestRuntimeResponse_ValidErrorResponse` | `unit` | Creates RuntimeResponse with status=error and message. | | `Status="error"`, `Message="session not ready: status is 'initializing'"` | Returns valid RuntimeResponse; all fields match input |
| `TestRuntimeResponse_SuccessWithEmptyMessage` | `unit` | Creates RuntimeResponse with status=success and empty message. | | `Status="success"`, `Message=""` | Returns valid RuntimeResponse; `Message=""`; valid |
| `TestRuntimeResponse_ErrorWithEmptyMessage` | `unit` | Creates RuntimeResponse with status=error and empty message. | | `Status="error"`, `Message=""` | Returns valid RuntimeResponse; `Message=""`; valid |

### Happy Path — JSON Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_SerializesToJSON_Success` | `unit` | RuntimeResponse serializes to valid JSON. | | RuntimeResponse with `Status="success"`, `Message="Event recorded"` | Serializes to valid JSON: `{"status": "success", "message": "Event recorded"}`; terminates with newline `\n` when transmitted |
| `TestRuntimeResponse_SerializesToJSON_Error` | `unit` | RuntimeResponse serializes to valid JSON for error status. | | RuntimeResponse with `Status="error"`, `Message="no transition found"` | Serializes to valid JSON: `{"status": "error", "message": "no transition found"}`; terminates with newline `\n` when transmitted |
| `TestRuntimeResponse_DeserializesFromJSON` | `unit` | RuntimeResponse deserializes from valid JSON. | | Valid JSON string: `{"status": "success", "message": "Event recorded"}` | Parses to valid RuntimeResponse struct; all fields match |

### Validation Failures — Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_ValidStatusSuccess` | `unit` | Accepts status=success as valid. | | `Status="success"`, `Message="test"` | Returns valid RuntimeResponse; status accepted |
| `TestRuntimeResponse_ValidStatusError` | `unit` | Accepts status=error as valid. | | `Status="error"`, `Message="test"` | Returns valid RuntimeResponse; status accepted |

### Boundary Values — Message Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_MessageWithNewlines` | `unit` | Serializes message containing newline characters correctly. | | `Status="error"`, `Message="line1\nline2\nline3"` | Serializes to valid JSON with escaped newlines: `{"status": "error", "message": "line1\nline2\nline3"}`; newlines inside message do not interfere with terminator newline |
| `TestRuntimeResponse_LargeMessage` | `unit` | Accepts RuntimeResponse with very large message string. | | `Status="error"`, `Message=<1MB string>` | Returns valid RuntimeResponse; message serialized correctly; size under 10 MB limit |
| `TestRuntimeResponse_UnicodeMessage` | `unit` | Accepts RuntimeResponse with Unicode characters in message. | | `Status="success"`, `Message="通知: 成功 🎉"` | Returns valid RuntimeResponse; Unicode preserved correctly in JSON |

### Invariants — Protocol Constraints

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_OnlyStatusAndMessageFields` | `unit` | RuntimeResponse contains only status and message fields (no additional fields). | | Valid RuntimeResponse instance | Serialized JSON contains exactly two fields: `status` and `message`; no additional fields present |

### Boundary Values — Response Size

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_SizeLimit_JustUnder10MB` | `unit` | Accepts response with serialized JSON just under 10 MB. | Uses mock serializer to simulate size without allocating gigabytes | RuntimeResponse with message field totaling 10 MB - 100 bytes | Returns valid RuntimeResponse; serialized successfully |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_RepeatedSerialization` | `unit` | Repeated serialization of same RuntimeResponse produces identical results. | | Same valid RuntimeResponse instance serialized multiple times | All serialization results are identical; no mutations |

### Happy Path — Socket Communication

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeResponse_SocketTransmission_Success` | `e2e` | RuntimeResponse successfully transmitted over Unix domain socket. | Temporary test directory created; RuntimeSocketManager listening on test socket in test directory; spectra-agent client connected; all file operations occur within test fixtures | MessageHandler returns success RuntimeResponse | RuntimeSocketManager serializes response, sends over socket with newline terminator, and closes connection; client receives complete response |
| `TestRuntimeResponse_SocketTransmission_Error` | `e2e` | RuntimeResponse with error status successfully transmitted. | Temporary test directory created; RuntimeSocketManager listening on test socket in test directory; spectra-agent client connected; all file operations occur within test fixtures | MessageHandler returns error RuntimeResponse | RuntimeSocketManager serializes response, sends over socket with newline terminator, and closes connection; client receives complete response |
