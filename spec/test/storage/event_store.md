# Test Specification: `event_store_test.go`

## Source File Under Test
`storage/event_store.go`

## Test File
`storage/event_store_test.go`

---

## `EventStore`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewEventStore_ValidInputs` | `unit` | Constructs an EventStore with valid projectRoot, sessionUUID, and logger. | Stub StorageLayout to return a known path. Provide a no-op Logger. | `projectRoot="/tmp/project"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"`, `logger=<mock logger>` | Returns non-nil `*EventStore`; no I/O performed; no error |
| `TestNewEventStore_NoFileSystemAccess` | `unit` | Constructor does not touch the filesystem. | Provide a non-existent projectRoot. Stub StorageLayout. Provide a no-op Logger. | `projectRoot="/nonexistent"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"`, `logger=<mock logger>` | Returns non-nil `*EventStore`; no panic; no file created |

### Happy Path — Append

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Append_FirstEvent` | `unit` | Appends the first event to a new events.jsonl file. | Create a temp directory simulating the session directory. Stub FileAccessor to invoke the preparation callback (creating the file). Construct a valid Event via `NewEvent`. | Valid Event entity | Returns nil error; file contains one JSON line terminated with `\n` |
| `TestEventStore_Append_MultipleEvents` | `unit` | Appends multiple events sequentially; file grows with one line per event. | Create a temp directory simulating the session directory. Stub FileAccessor. Construct two valid Events. | Two valid Event entities appended sequentially | Returns nil error for both; file contains exactly two lines |
| `TestEventStore_Append_CompactJSON` | `unit` | Serialized event line contains no unnecessary whitespace or newlines within the JSON object. | Create session directory fixture. Stub FileAccessor. Construct a valid Event with a multi-field payload. | Valid Event | Returns nil; the written line contains no indentation or trailing spaces within the JSON |
| `TestEventStore_Append_MessageFieldLast` | `unit` | The Message field appears as the last key in the serialized JSON object. | Create session directory fixture. Stub FileAccessor. Construct a valid Event with a non-empty message. | Valid Event with `message="hello world"` | Returns nil; parsing the written JSON line confirms `"Message"` is the last key in the object |
| `TestEventStore_Append_EscapesNewlinesInMessage` | `unit` | Newline characters in Message are JSON-escaped so the serialized line remains a single line. | Create session directory fixture. Stub FileAccessor. Construct a valid Event with message containing `\n`. | Event with `message="line1\nline2"` | Returns nil; file contains exactly one line for this event; the raw line contains `\\n` (escaped) |
| `TestEventStore_Append_EmptyPayload` | `unit` | Event with empty JSON object payload serializes correctly. | Create session directory fixture. Stub FileAccessor. | Event with `payload=json.RawMessage('{}')` | Returns nil; written line includes `"Payload":{}` |

### Happy Path — Read

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Read_MultipleEvents` | `unit` | Reads all events from a file with multiple valid JSONL lines. | Create a temp file with three valid event JSON lines (programmatically written). Construct EventStore pointing to this file. Provide a mock Logger. | | Returns slice of 3 Events in file order; no error |
| `TestEventStore_Read_FileNotExists` | `unit` | Returns empty list when events.jsonl does not exist. | Create a temp directory without an events.jsonl file. Construct EventStore pointing to this directory. Provide a mock Logger. | | Returns empty `[]Event` slice; no error; no warnings logged |
| `TestEventStore_Read_PreservesOrder` | `unit` | Events are returned in the order they appear in the file. | Create a temp file with events having distinct IDs in a known order. | | Returned events have IDs matching the file order |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Append_SessionDirNotExists` | `unit` | Returns error when session directory does not exist. | Create a temp directory without the session subdirectory. Stub FileAccessor so the preparation callback detects missing directory. | Valid Event | Returns error containing `"session directory does not exist:"` |
| `TestEventStore_Append_FileAccessorError` | `unit` | Propagates error when FileAccessor preparation fails. | Stub FileAccessor to return an error from the preparation callback. | Valid Event | Returns error containing `"failed to prepare file"` |
| `TestEventStore_Append_ExceedsMaxPayloadSize` | `unit` | Returns size limit error when serialized event exceeds MaxPayloadSize. | Create session directory fixture. Stub FileAccessor. Construct an Event with a very large message (> 10 MB). | Event with oversized message | Returns error containing `"event size exceeds limit:"` and `"bytes (max"` |
| `TestEventStore_Append_ExceedsMaxPayloadSize_NoWrite` | `unit` | File is not modified when event exceeds size limit. | Create session directory fixture with an existing events.jsonl containing one event. Stub FileAccessor. Construct an oversized Event. | Oversized Event | Returns error; file content remains unchanged (still one line) |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Read_MalformedJSON` | `unit` | Skips line with invalid JSON and logs a warning. | Create a temp file with one valid line and one malformed line (missing closing brace). Provide a mock Logger that records calls. | | Returns slice with 1 Event (the valid one); Logger.Warn called with line number and error description |
| `TestEventStore_Read_BlankLine` | `unit` | Skips blank lines and logs a warning. | Create a temp file with a valid line, a blank line, and another valid line. Provide a mock Logger. | | Returns slice with 2 Events; Logger.Warn called once for the blank line |
| `TestEventStore_Read_MissingRequiredFields` | `unit` | Skips line with valid JSON but missing required Event fields and logs warning. | Create a temp file with a JSON line missing the `ID` field. Provide a mock Logger. | | Returns empty slice; Logger.Warn called with line number and validation error description |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Append_CallsFileAccessor` | `unit` | FileAccessor is called with the correct file path and a non-nil preparation callback. | Stub FileAccessor to record the call arguments. Create session directory fixture. | Valid Event | FileAccessor was called exactly once with the events.jsonl path |
| `TestEventStore_Append_ReadsEventViaGetters` | `unit` | Serialization uses Event getter methods, not struct field access. | Create session directory fixture. Stub FileAccessor. Construct a valid Event. | Valid Event | Written JSON contains all expected field values matching getter return values |
| `TestEventStore_Read_LogsWarningForEachMalformedLine` | `unit` | Logger.Warn is called once per malformed line with line number. | Create a temp file with two malformed lines among valid lines. Provide a mock Logger. | | Logger.Warn called exactly twice; each call includes the respective line number |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Read_IdempotentReads` | `unit` | Two consecutive reads return the same event list. | Create a temp file with two valid event lines. Construct EventStore. Provide a mock Logger. | | First and second read return identical slices; Logger.Warn call count is the same for both reads |

### Boundary Values — MaxPayloadSize

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_Append_ExactlyAtMaxPayloadSize` | `unit` | Event whose serialized size equals exactly MaxPayloadSize is accepted. | Create session directory fixture. Stub FileAccessor. Construct an Event whose serialized JSON is exactly MaxPayloadSize bytes. | Event at size boundary | Returns nil error; event is written to file |
| `TestEventStore_Append_OneByteOverMaxPayloadSize` | `unit` | Event whose serialized size is MaxPayloadSize + 1 is rejected. | Create session directory fixture. Stub FileAccessor. Construct an Event whose serialized JSON is MaxPayloadSize + 1 bytes. | Event one byte over limit | Returns error containing `"event size exceeds limit:"` |
