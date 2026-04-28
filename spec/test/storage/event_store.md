# Test Specification: `event_store.go`

## Source File Under Test
`storage/event_store.go`

## Test File
`storage/event_store_test.go`

---

## `EventStore`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_New` | `unit` | Constructs EventStore with valid inputs. | Temporary test directory created programmatically within test fixture with `.spectra/sessions/` subdirectory | `ProjectRoot=<test-fixture>`, `SessionUUID=<valid-uuid-v4>` | Returns EventStore instance; no error |

### Happy Path — Append (First Write)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendFirstEvent` | `unit` | Appends first event to non-existent file. | Temporary test directory created programmatically within test fixture with session directory `.spectra/sessions/<uuid>/` created; `events.jsonl` does not exist | Valid Event with all required fields | Returns nil; `events.jsonl` created with permissions `0644`; file contains one compact JSON line ending with `\n` |
| `TestEventStore_AppendCreatesFile` | `unit` | FileAccessor callback creates file on first write. | Temporary test directory created programmatically within test fixture with session directory `.spectra/sessions/<uuid>/` created; `events.jsonl` does not exist | Valid Event | Returns nil; file created with correct permissions |

### Happy Path — Append (Subsequent Writes)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendSecondEvent` | `unit` | Appends second event to existing file. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with one existing event created inside test fixture | Valid Event (different from first) | Returns nil; file contains two lines; second line is compact JSON with newline |
| `TestEventStore_AppendMultipleEvents` | `unit` | Appends sequence of events. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `events.jsonl` does not exist | Sequence of 5 valid Events | Returns nil for each append; file contains 5 lines in append order |

### Happy Path — Compact JSON Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_CompactJSONNoWhitespace` | `unit` | Serializes event as single-line compact JSON. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with multiple fields | Serialized JSON is single line; no newlines within JSON; single space after `:` in key-value pairs; line ends with `\n` |
| `TestEventStore_MessageFieldLast` | `unit` | Message field appears last in serialized JSON. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with `Message="long text"` | Serialized JSON has `"Message":"long text"}` at the end before newline; Message is last key in object |
| `TestEventStore_EmptyPayload` | `unit` | Serializes empty Payload as empty JSON object. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with `Payload={}` | Serialized JSON contains `"Payload":{}` |

### Happy Path — Read (Empty File)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadFileDoesNotExist` | `unit` | Returns empty list when file does not exist. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `events.jsonl` does not exist | | Returns empty slice `[]Event`; no error |

### Happy Path — Read (Existing Events)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadSingleEvent` | `unit` | Reads single event from file. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with one valid event line created inside test fixture | | Returns slice with one Event; fields match written event |
| `TestEventStore_ReadMultipleEvents` | `unit` | Reads multiple events in chronological order. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 5 valid event lines created inside test fixture | | Returns slice with 5 Events in file order; all fields parsed correctly |
| `TestEventStore_ReadLongMessage` | `unit` | Reads event with very long Message field (1 MB). | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with event containing 1 MB Message created inside test fixture | | Returns Event with full Message content; no truncation |

### Happy Path — Read (Malformed Lines)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadSkipsMalformedJSON` | `unit` | Skips line with invalid JSON and logs warning. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 3 lines: valid event, malformed JSON (missing `}`), valid event created inside test fixture | | Returns slice with 2 Events (line 1 and 3); logs warning matching `/skipping malformed event at line 2:.*unexpected end of JSON/i` |
| `TestEventStore_ReadSkipsMissingRequiredField` | `unit` | Skips line missing required Event field. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 2 lines: valid event, JSON missing `ID` field created inside test fixture | | Returns slice with 1 Event; logs warning matching `/skipping malformed event at line 2:.*missing required field.*ID/i` |
| `TestEventStore_ReadSkipsBlankLine` | `unit` | Skips blank line and logs warning. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 3 lines: valid event, blank line, valid event created inside test fixture | | Returns slice with 2 Events; logs warning for line 2 |

### Happy Path — File Locking (Write)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendAcquiresExclusiveLock` | `unit` | Acquires exclusive lock during append. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Valid Event | Lock acquired before write; lock released after write completes |
| `TestEventStore_AppendReleasesLockOnError` | `unit` | Releases lock when write fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock FileAccessor that fails during write | Valid Event | Returns error; lock released before function returns |

### Happy Path — File Locking (Read)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadAcquiresSharedLock` | `unit` | Acquires shared read lock during read. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with events created inside test fixture | | Shared lock acquired before read; lock released after read completes |
| `TestEventStore_ReadReleasesLockOnError` | `unit` | Releases lock when read fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; file exists but permission denied after lock | | Returns error; lock released before function returns |

### Validation Failures — Append (Parent Directory)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendParentDirDoesNotExist` | `unit` | Returns error when session directory missing. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory does not exist | Valid Event | Returns error matching `/session directory does not exist:.*\.spectra\/sessions\/.*<uuid>/i`; file not created |

### Validation Failures — Append (Serialization)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendSerializationFails` | `unit` | Returns error when JSON marshaling fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with un-serializable field (e.g., channel) | Returns error matching `/failed to serialize event:/i`; file not modified |

### Validation Failures — Append (Write Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendWriteFails` | `unit` | Returns error when file write fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock FileAccessor that fails write | Valid Event | Returns error matching `/failed to write event:/i` |
| `TestEventStore_AppendLockFails` | `unit` | Returns error when lock acquisition fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock FileAccessor that fails lock | Valid Event | Returns error matching `/failed to acquire write lock:/i` |

### Validation Failures — Append (Size Limit)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendExceeds10MBLimit` | `unit` | Rejects event exceeding 10 MB serialized size. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with Message field containing 11 MB of text | Returns error matching `/event size exceeds 10 MB limit:.*bytes/i`; file not modified |
| `TestEventStore_AppendExactly10MB` | `unit` | Accepts event at exactly 10 MB limit. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with Message field totaling exactly 10 MB serialized | Returns nil; event written successfully |

### Validation Failures — Read (Lock Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadLockFails` | `unit` | Returns error when lock acquisition fails. | Temporary test directory created programmatically within test fixture; session directory and file created inside test fixture; mock FileAccessor that fails lock | | Returns error matching `/failed to acquire read lock:/i` |

### Validation Failures — Read (File Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadFileReadFails` | `unit` | Returns error when file read operation fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; file exists but becomes unreadable after lock acquired | | Returns error matching `/failed to read event file:/i` |
| `TestEventStore_ReadPermissionDenied` | `unit` | Returns error when file permissions deny read. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `events.jsonl` with permissions `0000` created inside test fixture | | Returns error matching `/permission denied/i` |

### Boundary Values — Special Characters

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_MessageWithNewlines` | `unit` | Escapes newline characters in Message field. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with `Message="line1\nline2\nline3"` | Serialized JSON contains `\\n` instead of literal newlines; remains single line |
| `TestEventStore_MessageWithUnicode` | `unit` | Handles Unicode characters in Message field. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with `Message="emoji: 🎉, CJK: 中文"` | Serialized JSON preserves Unicode characters; remains single line |
| `TestEventStore_PayloadWithComplexJSON` | `unit` | Serializes complex nested Payload. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Event with deeply nested Payload (arrays, objects, various types) | Serialized JSON is valid; Payload correctly serialized; read operation returns identical structure |

### Idempotency — Read

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadIdempotent` | `unit` | Multiple reads return identical results. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 3 events created inside test fixture | | First read returns 3 events; second read returns identical 3 events; no caching; both reads access disk |

### Idempotency — Append

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_AppendDoesNotModifyExisting` | `unit` | Append never modifies existing lines. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 2 events created inside test fixture; record original file hash | Valid Event | New event appended; first 2 lines unchanged (verified by byte comparison); third line added |

### Concurrent Behaviour — Multiple Writers

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ConcurrentAppendSameFile` | `race` | Multiple goroutines append to same file safely. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `events.jsonl` does not exist | 10 goroutines each append 5 events | All 50 events written successfully; no corruption; file contains exactly 50 valid lines; all locks properly acquired and released |
| `TestEventStore_ConcurrentAppendSerializes` | `race` | File lock serializes concurrent writes. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | 5 goroutines append simultaneously | Writes serialized by lock; each complete event written atomically; no partial lines |

### Concurrent Behaviour — Read During Write

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_ReadBlocksDuringWrite` | `race` | Read waits for write to complete. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 1 event created inside test fixture | Goroutine 1 appends (holds lock); Goroutine 2 reads (blocked) | Read blocked until write completes; read returns all events including newly written |
| `TestEventStore_WriteBlocksDuringRead` | `race` | Write waits for read to complete. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with events created inside test fixture | Goroutine 1 reads (holds lock); Goroutine 2 appends (blocked) | Append blocked until read completes; append succeeds after read releases lock |

### Boundary Values — File Permissions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_NewFilePermissions0644` | `unit` | Creates new file with correct permissions. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `events.jsonl` does not exist | Valid Event | File created with permissions `0644` (owner rw, group/others r) |

### Boundary Values — Invalid Session UUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_InvalidSessionUUID` | `unit` | Fails with malformed UUID. | Temporary test directory created programmatically within test fixture | `ProjectRoot=<test-fixture>`, `SessionUUID="not-a-uuid"` | Subsequent append/read operations fail with filesystem errors (malformed path) |
| `TestEventStore_EmptySessionUUID` | `unit` | Fails with empty UUID. | Temporary test directory created programmatically within test fixture | `ProjectRoot=<test-fixture>`, `SessionUUID=""` | Subsequent operations fail with filesystem errors |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_FileAccessorErrorPropagated` | `unit` | Propagates FileAccessor callback error. | Temporary test directory created programmatically within test fixture; mock FileAccessor that returns callback error | Valid Event | Returns error containing callback error details; error wrapped with context |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_LocksReleasedOnPanic` | `unit` | Locks released if panic occurs during operation. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock that panics during write | Valid Event | Panic propagates; file lock released (verified by subsequent successful operation) |

### Happy Path — No Caching

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventStore_NoCaching` | `unit` | Read always accesses disk, no in-memory cache. | Temporary test directory created programmatically within test fixture; session directory and `events.jsonl` with 2 events created inside test fixture | Read events; externally append third event to file; read again | First read returns 2 events; second read returns 3 events (detects external change) |
