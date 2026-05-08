# Test Specification: `event_emit_test.go`

## Source File Under Test
`internal/cmd/spectra_agent/event_emit.go`

## Test File
`internal/cmd/spectra_agent/event_emit_test.go`

---

## `EventEmitCmd`

### Happy Path — EventEmitCmd

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventEmitCmd_Success` | `unit` | Sends event and prints success text on exit code 0. | Mock `cmdutil.SendAndHandle` to return exit code 0. Provide valid sessionID and projectRoot via command context. | args: `["event", "emit", "ReviewNeeded"]`, flags: default | Exit code 0; stdout contains `"Event emitted successfully"` |
| `TestEventEmitCmd_WithMessage` | `unit` | Includes --message value in wire message payload. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "BuildDone", "--message", "build completed"]` | Exit code 0; message payload message equals `"build completed"` |
| `TestEventEmitCmd_WithPayloadObject` | `unit` | Accepts valid JSON object in --payload flag. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "Deploy", "--payload", "{\"env\":\"prod\",\"version\":\"1.2\"}"]` | Exit code 0; message payload payload equals `{"env":"prod","version":"1.2"}` |
| `TestEventEmitCmd_WithClaudeSessionID` | `unit` | Includes claude-session-id in wire message. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "Ping", "--claude-session-id", "claude-xyz"]` | Exit code 0; message `claudeSessionID` equals `"claude-xyz"` |
| `TestEventEmitCmd_EmptyMessageAccepted` | `unit` | Accepts explicit empty --message without error. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "Notify", "--message", ""]` | Exit code 0; message payload message equals `""` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventEmitCmd_MissingType` | `unit` | Returns exit code 1 when no positional argument is provided. | Provide valid sessionID and projectRoot via command context. | args: `["event", "emit"]` (no type) | Exit code 1; stderr contains `"event type is required"` |
| `TestEventEmitCmd_EmptyType` | `unit` | Returns exit code 1 when positional argument is an empty string. | Provide valid sessionID and projectRoot via command context. | args: `["event", "emit", ""]` | Exit code 1; stderr contains `"event type is required"` |

### Validation Failures — payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventEmitCmd_PayloadPrimitive` | `unit` | Rejects JSON string primitive in --payload. | Provide valid context. | args: `["event", "emit", "Evt", "--payload", "\"hello\""]` | Exit code 1; stderr contains `"--payload must be a valid JSON object, e.g., {}"` |
| `TestEventEmitCmd_PayloadArray` | `unit` | Rejects JSON array in --payload. | Provide valid context. | args: `["event", "emit", "Evt", "--payload", "[1,2,3]"]` | Exit code 1; stderr contains `"--payload must be a valid JSON object, e.g., {}"` |
| `TestEventEmitCmd_PayloadInvalidJSON` | `unit` | Rejects malformed JSON in --payload. | Provide valid context. | args: `["event", "emit", "Evt", "--payload", "{invalid}"]` | Exit code 1; stderr contains `"--payload must be a valid JSON object, e.g., {}"` |
| `TestEventEmitCmd_PayloadNull` | `unit` | Rejects JSON null in --payload (must be object). | Provide valid context. | args: `["event", "emit", "Evt", "--payload", "null"]` | Exit code 1; stderr contains `"--payload must be a valid JSON object, e.g., {}"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventEmitCmd_ConstructsWireFormat` | `unit` | Constructs RuntimeMessage with type "event" and correct payload structure. | Mock `cmdutil.SendAndHandle` to capture message argument and return exit code 0. Provide sessionID=`"sess-1"`, projectRoot=`"/tmp/proj"` via context. | args: `["event", "emit", "ReviewNeeded", "--message", "changes staged", "--payload", "{\"pr\":42}", "--claude-session-id", "c-1"]` | `SendAndHandle` called with message struct: `{type:"event", claudeSessionID:"c-1", payload:{eventType:"ReviewNeeded", message:"changes staged", payload:{"pr":42}}}` |
| `TestEventEmitCmd_DefaultPayload` | `unit` | Uses empty object as default payload when --payload is omitted. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "Ping"]` | Message payload payload equals `{}` |
| `TestEventEmitCmd_DefaultMessage` | `unit` | Uses empty string as default message when --message is omitted. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "Ping"]` | Message payload message equals `""` |
| `TestEventEmitCmd_DefaultClaudeSessionID` | `unit` | Uses empty string as default claude-session-id when flag is omitted. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["event", "emit", "Ping"]` | Message `claudeSessionID` equals `""` |
| `TestEventEmitCmd_PassesCorrectSuccessText` | `unit` | Passes "Event emitted successfully" as successText to SendAndHandle. | Mock `cmdutil.SendAndHandle` to capture successText argument and return exit code 0. Provide valid context. | args: `["event", "emit", "Ping"]` | `SendAndHandle` called with `successText="Event emitted successfully"` |
