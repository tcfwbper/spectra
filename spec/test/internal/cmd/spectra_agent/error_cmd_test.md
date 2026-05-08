# Test Specification: `error_cmd_test.go`

## Source File Under Test
`internal/cmd/spectra_agent/error_cmd.go`

## Test File
`internal/cmd/spectra_agent/error_cmd_test.go`

---

## `ErrorCmd`

### Happy Path — ErrorCmd

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCmd_Success` | `unit` | Sends error message and prints success text on exit code 0. | Mock `cmdutil.SendAndHandle` to return exit code 0. Provide valid sessionID and projectRoot via command context. | args: `["error", "something broke"]`, flags: default | Exit code 0; stdout contains `"Error reported successfully"` |
| `TestErrorCmd_WithDetailObject` | `unit` | Accepts valid JSON object in --detail flag. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["error", "crash", "--detail", "{\"stack\":\"trace\",\"code\":500}"]` | Exit code 0; message payload detail equals `{"stack":"trace","code":500}` |
| `TestErrorCmd_WithDetailNull` | `unit` | Accepts JSON null in --detail flag. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["error", "crash", "--detail", "null"]` | Exit code 0; message payload detail is `null` |
| `TestErrorCmd_WithClaudeSessionID` | `unit` | Includes claude-session-id in wire message. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["error", "oops", "--claude-session-id", "claude-abc"]` | Exit code 0; message `claudeSessionID` equals `"claude-abc"` |
| `TestErrorCmd_WhitespaceOnlyMessageAccepted` | `unit` | Accepts whitespace-only message without error. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["error", "   "]` | Exit code 0; message payload message equals `"   "` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCmd_MissingMessage` | `unit` | Returns exit code 1 when no positional argument is provided. | Provide valid sessionID and projectRoot via command context. | args: `["error"]` (no message) | Exit code 1; stderr contains `"error message is required"` |
| `TestErrorCmd_EmptyMessage` | `unit` | Returns exit code 1 when positional argument is an empty string. | Provide valid sessionID and projectRoot via command context. | args: `["error", ""]` | Exit code 1; stderr contains `"error message is required"` |

### Validation Failures — detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCmd_DetailPrimitive` | `unit` | Rejects JSON string primitive in --detail. | Provide valid context. | args: `["error", "msg", "--detail", "\"string\""]` | Exit code 1; stderr contains `"--detail must be a JSON object or null"` |
| `TestErrorCmd_DetailNumber` | `unit` | Rejects JSON number primitive in --detail. | Provide valid context. | args: `["error", "msg", "--detail", "42"]` | Exit code 1; stderr contains `"--detail must be a JSON object or null"` |
| `TestErrorCmd_DetailBoolean` | `unit` | Rejects JSON boolean primitive in --detail. | Provide valid context. | args: `["error", "msg", "--detail", "true"]` | Exit code 1; stderr contains `"--detail must be a JSON object or null"` |
| `TestErrorCmd_DetailArray` | `unit` | Rejects JSON array in --detail. | Provide valid context. | args: `["error", "msg", "--detail", "[1,2,3]"]` | Exit code 1; stderr contains `"--detail must be a JSON object or null"` |
| `TestErrorCmd_DetailInvalidJSON` | `unit` | Rejects malformed JSON in --detail. | Provide valid context. | args: `["error", "msg", "--detail", "{invalid}"]` | Exit code 1; stderr contains `"--detail must be a JSON object or null"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorCmd_ConstructsWireFormat` | `unit` | Constructs RuntimeMessage with type "error" and correct payload structure. | Mock `cmdutil.SendAndHandle` to capture message argument and return exit code 0. Provide sessionID=`"sess-1"`, projectRoot=`"/tmp/proj"` via context. | args: `["error", "disk full", "--detail", "{\"disk\":\"/dev/sda\"}", "--claude-session-id", "c-99"]` | `SendAndHandle` called with message struct: `{type:"error", claudeSessionID:"c-99", payload:{message:"disk full", detail:{"disk":"/dev/sda"}}}` |
| `TestErrorCmd_DefaultDetail` | `unit` | Uses empty object as default detail when --detail is omitted. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["error", "msg"]` | Message payload detail equals `{}` |
| `TestErrorCmd_DefaultClaudeSessionID` | `unit` | Uses empty string as default claude-session-id when flag is omitted. | Mock `cmdutil.SendAndHandle` to capture message and return exit code 0. Provide valid context. | args: `["error", "msg"]` | Message `claudeSessionID` equals `""` |
| `TestErrorCmd_PassesCorrectSuccessText` | `unit` | Passes "Error reported successfully" as successText to SendAndHandle. | Mock `cmdutil.SendAndHandle` to capture successText argument and return exit code 0. Provide valid context. | args: `["error", "msg"]` | `SendAndHandle` called with `successText="Error reported successfully"` |
