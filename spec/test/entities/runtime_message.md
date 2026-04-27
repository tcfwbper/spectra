# Test Specification: `runtime_message.go`

## Source File Under Test
`entities/runtime_message.go`

## Test File
`entities/runtime_message_test.go`

---

## `RuntimeMessage`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_ValidEventMessage` | `unit` | Creates RuntimeMessage with type=event and all valid fields. | | `Type="event"`, `ClaudeSessionID="550e8400-e29b-41d4-a716-446655440000"`, `Payload={"eventType": "DraftCompleted", "message": "ready", "payload": {"count": 3}}` | Returns valid RuntimeMessage; all fields match input |
| `TestRuntimeMessage_ValidErrorMessage` | `unit` | Creates RuntimeMessage with type=error and all valid fields. | | `Type="error"`, `ClaudeSessionID="550e8400-e29b-41d4-a716-446655440000"`, `Payload={"message": "Failed to load", "detail": {"error": "not found"}}` | Returns valid RuntimeMessage; all fields match input |
| `TestRuntimeMessage_EmptyClaudeSessionID` | `unit` | Creates RuntimeMessage with empty claudeSessionID (valid for human nodes). | | `Type="event"`, `ClaudeSessionID=""`, `Payload={"eventType": "RequirementProvided", "message": "", "payload": {}}` | Returns valid RuntimeMessage; `ClaudeSessionID=""` |
| `TestRuntimeMessage_OmittedClaudeSessionID` | `unit` | Creates RuntimeMessage with claudeSessionID field omitted (defaults to empty). | | `Type="event"`, `Payload={"eventType": "RequirementProvided", "message": "", "payload": {}}`, claudeSessionID field not included in JSON | Parsed RuntimeMessage has `ClaudeSessionID=""` |

### Happy Path — Event Payload Defaults

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_EventPayload_MessageOmitted` | `unit` | Accepts event payload with message field omitted (defaults to empty string). | | `Type="event"`, `Payload={"eventType": "Started", "payload": {}}`, message field not included | Parsed event payload has `message=""` |
| `TestRuntimeMessage_EventPayload_PayloadOmitted` | `unit` | Accepts event payload with payload field omitted (defaults to empty object). | | `Type="event"`, `Payload={"eventType": "Started", "message": "begin"}`, payload field not included | Parsed event payload has `payload={}` |
| `TestRuntimeMessage_EventPayload_BothOmitted` | `unit` | Accepts event payload with both message and payload fields omitted. | | `Type="event"`, `Payload={"eventType": "Continue"}`, message and payload fields not included | Parsed event payload has `message=""` and `payload={}` |

### Happy Path — Error Payload Defaults

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_ErrorPayload_DetailOmitted` | `unit` | Accepts error payload with detail field omitted (defaults to empty object). | | `Type="error"`, `Payload={"message": "Failed to process"}`, detail field not included | Parsed error payload has `detail={}` |

### Happy Path — JSON Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_SerializesToJSON` | `unit` | RuntimeMessage serializes to valid JSON. | | Valid RuntimeMessage instance | Serializes to valid JSON string; terminates with newline `\n` when transmitted |
| `TestRuntimeMessage_DeserializesFromJSON` | `unit` | RuntimeMessage deserializes from valid JSON. | | Valid JSON string: `{"type": "event", "claudeSessionID": "test-id", "payload": {"eventType": "Test", "message": "", "payload": {}}}` | Parses to valid RuntimeMessage struct; all fields match |

### Validation Failures — Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_MissingType` | `unit` | Rejects message with missing type field. | | JSON: `{"payload": {"eventType": "Test"}}` | Returns error; error message matches `/missing required field 'type'/i`; sends error response; closes connection |
| `TestRuntimeMessage_EmptyType` | `unit` | Rejects message with empty type field. | | `Type=""`, `Payload={"eventType": "Test"}` | Returns error; error message matches `/type field must not be empty/i`; sends error response; closes connection |
| `TestRuntimeMessage_UnrecognizedType` | `unit` | Rejects message with unrecognized type value. | | `Type="unknown"`, `Payload={"eventType": "Test"}` | Returns error; error message matches `/invalid message type 'unknown'/i`; sends error response; closes connection |
| `TestRuntimeMessage_LegacyType` | `unit` | Rejects message with legacy type value. | | `Type="legacy"`, `Payload={"eventType": "Test"}` | Returns error; error message matches `/invalid message type 'legacy'/i`; sends error response; closes connection |

### Validation Failures — Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_MissingPayload` | `unit` | Rejects message with missing payload field. | | JSON: `{"type": "event"}` | Returns error; error message matches `/missing required field 'payload'/i`; sends error response; closes connection |
| `TestRuntimeMessage_PayloadPrimitiveString` | `unit` | Rejects message with payload as JSON primitive string. | | `Type="event"`, `Payload="string"` (JSON primitive) | Returns error; error message matches `/payload must be a JSON object/i`; sends error response; closes connection |
| `TestRuntimeMessage_PayloadPrimitiveNumber` | `unit` | Rejects message with payload as JSON primitive number. | | `Type="event"`, `Payload=123` (JSON primitive) | Returns error; error message matches `/payload must be a JSON object/i`; sends error response; closes connection |
| `TestRuntimeMessage_PayloadPrimitiveBoolean` | `unit` | Rejects message with payload as JSON primitive boolean. | | `Type="event"`, `Payload=true` (JSON primitive) | Returns error; error message matches `/payload must be a JSON object/i`; sends error response; closes connection |
| `TestRuntimeMessage_PayloadNull` | `unit` | Rejects message with payload as null. | | `Type="event"`, `Payload=null` | Returns error; error message matches `/payload must be a JSON object/i`; sends error response; closes connection |
| `TestRuntimeMessage_PayloadArray` | `unit` | Rejects message with payload as JSON array. | | `Type="event"`, `Payload=[1, 2, 3]` (JSON array) | Returns error; error message matches `/payload must be a JSON object/i`; sends error response; closes connection |
| `TestRuntimeMessage_PayloadEmptyArray` | `unit` | Rejects message with payload as empty JSON array. | | `Type="event"`, `Payload=[]` (empty JSON array) | Returns error; error message matches `/payload must be a JSON object/i`; sends error response; closes connection |

### Validation Failures — Event Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_EventPayload_MissingEventType` | `unit` | Rejects event message with missing eventType field. | | `Type="event"`, `Payload={"message": "test", "payload": {}}`, eventType field not included | Returns error; error message matches `/event payload missing required field 'eventType'/i`; sends error response; closes connection |
| `TestRuntimeMessage_EventPayload_EmptyEventType` | `unit` | Rejects event message with empty eventType field. | | `Type="event"`, `Payload={"eventType": "", "message": "test", "payload": {}}` | Returns error; error message matches `/eventType must not be empty/i`; sends error response; closes connection |

### Validation Failures — Error Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_ErrorPayload_MissingMessage` | `unit` | Rejects error message with missing message field. | | `Type="error"`, `Payload={"detail": {}}`, message field not included | Returns error; error message matches `/error payload missing required field 'message'/i`; sends error response; closes connection |
| `TestRuntimeMessage_ErrorPayload_EmptyMessage` | `unit` | Rejects error message with empty message field. | | `Type="error"`, `Payload={"message": "", "detail": {}}` | Returns error; error message matches `/error payload missing required field 'message'/i`; sends error response; closes connection |

### Validation Failures — ClaudeSessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_ClaudeSessionID_Null` | `unit` | Rejects message with claudeSessionID as null. | | `Type="event"`, `ClaudeSessionID=null`, `Payload={"eventType": "Test"}` | Returns error; error message matches `/claudeSessionID must be a string/i`; sends error response; closes connection |
| `TestRuntimeMessage_ClaudeSessionID_Number` | `unit` | Rejects message with claudeSessionID as number. | | `Type="event"`, `ClaudeSessionID=123`, `Payload={"eventType": "Test"}` | Returns error; error message matches `/claudeSessionID must be a string/i`; sends error response; closes connection |
| `TestRuntimeMessage_ClaudeSessionID_Object` | `unit` | Rejects message with claudeSessionID as object. | | `Type="event"`, `ClaudeSessionID={"id": "test"}`, `Payload={"eventType": "Test"}` | Returns error; error message matches `/claudeSessionID must be a string/i`; sends error response; closes connection |

### Validation Failures — JSON Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_MalformedJSON_MissingBrace` | `unit` | Rejects message with malformed JSON (missing closing brace). | | Raw JSON string: `{"type": "event", "payload": {"eventType": "Test"` (missing `}`) | Returns error; error message matches `/JSON parse error|unmarshal/i`; sends error response if possible; closes connection |
| `TestRuntimeMessage_MalformedJSON_InvalidEscape` | `unit` | Rejects message with malformed JSON (invalid escape sequence). | | Raw JSON string with invalid escape: `{"type": "event\x", "payload": {}}` | Returns error; error message matches `/JSON parse error|unmarshal/i`; sends error response if possible; closes connection |

### Boundary Values — Message Size

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_SizeLimit_JustUnder10MB` | `unit` | Accepts message with serialized JSON just under 10 MB. | | RuntimeMessage with total serialized size 10 MB - 1 byte | Returns valid RuntimeMessage; message processed successfully |
| `TestRuntimeMessage_SizeLimit_Exceeds10MB` | `unit` | Rejects message with serialized JSON exceeding 10 MB. | | RuntimeMessage with total serialized size 10 MB + 1 byte | RuntimeSocketManager detects size violation before fully reading; returns error; error message matches `/size limit/i`; sends error response if possible; closes connection |

### Boundary Values — Payload Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_LargeEventPayload` | `unit` | Accepts event message with very large payload object. | | `Type="event"`, `Payload={"eventType": "Test", "message": "", "payload": <9MB JSON object>}` | Returns valid RuntimeMessage; large payload stored correctly |
| `TestRuntimeMessage_DeepNestedPayload` | `unit` | Accepts message with deeply nested JSON in payload. | | `Type="event"`, `Payload={"eventType": "Test", "message": "", "payload": <JSON nested 100 levels deep>}` | Returns valid RuntimeMessage; nested structure preserved |
| `TestRuntimeMessage_UnicodeInPayload` | `unit` | Accepts message with Unicode characters in payload fields. | | `Type="event"`, `Payload={"eventType": "Test", "message": "通知: Process complete 🎉", "payload": {"emoji": "🚀"}}` | Returns valid RuntimeMessage; Unicode preserved correctly in all fields |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_RepeatedDeserialization` | `unit` | Repeated deserialization of same JSON produces identical results. | | Same valid JSON string deserialized multiple times | All deserialization results are identical; no mutations |

### Happy Path — Socket Communication

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_SocketTransmission_EventMessage` | `e2e` | RuntimeMessage successfully transmitted over Unix domain socket. | Temporary test directory created; RuntimeSocketManager listening on test socket in test directory; all file operations occur within test fixtures | Send valid event RuntimeMessage JSON with newline terminator over socket | RuntimeSocketManager receives, parses, and processes message; sends RuntimeResponse; closes connection |
| `TestRuntimeMessage_SocketTransmission_ErrorMessage` | `e2e` | RuntimeMessage with error type successfully transmitted. | Temporary test directory created; RuntimeSocketManager listening on test socket in test directory; all file operations occur within test fixtures | Send valid error RuntimeMessage JSON with newline terminator over socket | RuntimeSocketManager receives, parses, and processes message; sends RuntimeResponse; closes connection |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_ConcurrentConnections` | `race` | Multiple concurrent client connections handled correctly. | Temporary test directory created; RuntimeSocketManager listening on test socket in test directory; all file operations occur within test fixtures | 10 clients simultaneously connect and send valid RuntimeMessages | All messages processed successfully; each receives RuntimeResponse; connections closed; no data races detected |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntimeMessage_SocketClosed_AfterValidation` | `unit` | Connection closed after validation failure. | Mock socket connection | Send invalid RuntimeMessage (missing type field) | RuntimeSocketManager sends error response and closes connection; connection is closed |
| `TestRuntimeMessage_SocketClosed_AfterMalformedJSON` | `unit` | Connection closed after JSON parse error. | Mock socket connection | Send malformed JSON | RuntimeSocketManager attempts to send error response and closes connection; connection is closed |
