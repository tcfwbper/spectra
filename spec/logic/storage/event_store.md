# EventStore

## Overview

EventStore manages persistent storage of event history for a single session. It appends events to a JSONL (JSON Lines) file in compact format, ensuring events are written atomically with file-level write locks. EventStore reads the entire event history when requested, skipping malformed lines with warnings. It is append-only: events cannot be modified or deleted once written. EventStore does not manage the parent session directory; it expects the directory to exist before writing events.

**Persistence Role**: The Read operation is provided for external tools and debugging only (e.g., post-session analysis). During active Runtime execution, the in-memory Session.EventHistory is the authoritative event list. The Runtime does not read from EventStore to make decisions or restore state during normal operation.

## Behavior

1. EventStore is initialized with a session UUID and uses StorageLayout to determine the path to `events.jsonl`.
2. EventStore uses FileAccessor to access the `events.jsonl` file, providing a callback that creates the file on first write.
3. The preparation callback checks if the parent directory (`.spectra/sessions/<session-uuid>/`) exists. If it does not exist, the callback returns an error: "session directory does not exist: <path>".
4. If the parent directory exists, the callback creates an empty `events.jsonl` file with permissions `0644`.
5. When appending an event, EventStore acquires an exclusive file-level lock (using `flock` or equivalent) to prevent concurrent writes.
6. EventStore serializes the event to compact JSON format: single-line JSON with all whitespace removed, except for a single space after `:` in key-value pairs.
7. EventStore ensures the `Message` field is serialized last in the JSON object to improve readability (long text does not obscure other fields).
8. EventStore appends the compact JSON line followed by a newline (`\n`) to the file.
9. EventStore flushes the file buffer and releases the lock after each write.
10. When reading events, EventStore opens the file with a shared read lock.
11. EventStore reads the file line-by-line. For each line, it attempts to parse the JSON into an Event struct.
12. If a line is malformed (invalid JSON or missing required fields), EventStore logs a warning with the line number and skips the line.
13. EventStore returns all successfully parsed events in the order they appear in the file.
14. If the file does not exist (and no events have been written yet), EventStore returns an empty list of events without error.
15. EventStore does not cache events in memory. Each read operation reads from disk.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| SessionUUID | string (UUID) | Valid UUID v4 format | Yes |

### For Append Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Event | Event struct | Valid Event as defined in `logic/entities/event.md` | Yes |

## Outputs

### For Append Operation

**Success Case**: No return value (void / nil error in Go).

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"session directory does not exist: <path>"` | Parent directory `.spectra/sessions/<session-uuid>/` does not exist |
| `"failed to acquire write lock: <error>"` | Unable to acquire exclusive file lock |
| `"failed to serialize event: <error>"` | JSON marshaling failed |
| `"failed to write event: <error>"` | File write operation failed |

### For Read Operation

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| Events | []Event | List of events in chronological order (order of appearance in file) |

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"failed to acquire read lock: <error>"` | Unable to acquire shared file lock |
| `"failed to read event file: <error>"` | File read operation failed (not including "file does not exist") |

**Warnings** (logged, not returned as errors):

| Warning Message Format | Description |
|----------------------|-------------|
| `"skipping malformed event at line <N>: <error>"` | JSON parsing failed for a specific line |

## Invariants

1. **Append-Only**: EventStore must never modify or delete existing lines in the `events.jsonl` file. All writes must append to the end.

2. **Compact JSON Format**: Each event must be serialized as a single-line JSON string with no newlines or unnecessary whitespace within the JSON object.

3. **Message Field Last**: The `Message` field must appear as the last key in the JSON object serialization order.

4. **File-Level Locking**: Write operations must acquire an exclusive lock. Read operations must acquire a shared lock. Locks must be released before the function returns (including error paths).

5. **Atomic Append**: Each event write must be atomic at the filesystem level. If a write fails, the file must not be corrupted (partially written lines).

6. **File Permissions**: Newly created `events.jsonl` files must have permissions `0644` (owner read/write, group/others read-only).

7. **No In-Memory Cache**: EventStore does not cache events in memory. Each read operation reads from disk. Note: Read is intended for external inspection and post-session analysis only. The running Runtime does not read from EventStore to determine behavior; the in-memory Session.EventHistory is the authoritative source of truth.

8. **Idempotent Read**: Reading events multiple times must return the same results if the file has not been modified.

9. **Parent Directory Requirement**: EventStore must not create the session directory. It must return an error if the directory does not exist.

10. **Line-Oriented Format**: Each event must be terminated with exactly one newline character (`\n`). No blank lines are allowed.

11. **Size Limit**: Each serialized event must not exceed 10 MB. This matches the wire-level limit enforced by RuntimeSocketManager, ensuring any event accepted from the socket can also be persisted.

## Edge Cases

- **Condition**: `events.jsonl` does not exist, and this is the first append operation.
  **Expected**: FileAccessor callback creates the file. EventStore writes the first event to the new file.

- **Condition**: `events.jsonl` does not exist, and a read operation is requested.
  **Expected**: EventStore returns an empty list `[]` without error.

- **Condition**: Parent directory `.spectra/sessions/<session-uuid>/` does not exist when appending.
  **Expected**: FileAccessor callback returns an error: `"session directory does not exist: <path>"`. EventStore propagates the error.

- **Condition**: File contains a malformed JSON line (e.g., missing closing brace).
  **Expected**: EventStore logs a warning: `"skipping malformed event at line <N>: unexpected end of JSON input"` and continues reading subsequent lines.

- **Condition**: File contains a blank line.
  **Expected**: EventStore attempts to parse the blank line as JSON, fails, logs a warning, and skips the line.

- **Condition**: File contains a line with valid JSON but missing required Event fields (e.g., no `ID` field).
  **Expected**: EventStore logs a warning: `"skipping malformed event at line <N>: missing required field 'ID'"` and skips the line.

- **Condition**: Two goroutines attempt to append events simultaneously to the same `events.jsonl` file.
  **Expected**: The file-level exclusive lock serializes the writes. One goroutine acquires the lock, writes, releases. The other waits, then writes. Both events are appended successfully without corruption.

- **Condition**: A goroutine attempts to read events while another goroutine is appending.
  **Expected**: The read operation acquires a shared lock and blocks until the write operation completes and releases the exclusive lock. The read then proceeds.

- **Condition**: `Message` field contains a very long string (e.g., 1 MB of text).
  **Expected**: EventStore serializes the entire message as a single-line compact JSON string. The line may be very long, but it is still valid JSONL.

- **Condition**: Total serialized event size exceeds 10 MB.
  **Expected**: EventStore enforces a 10 MB per-event size limit (matching RuntimeSocketManager's wire-level limit). Events exceeding this size are rejected with an error: `"event size exceeds 10 MB limit: <actual-size> bytes"`. The event is not appended to the file.

- **Condition**: `Message` field contains newline characters (`\n`) or Unicode characters.
  **Expected**: JSON encoding escapes these characters (e.g., `\n` becomes `\\n`). The serialized line remains a single line.

- **Condition**: Event `Payload` is an empty JSON object `{}`.
  **Expected**: EventStore serializes it as `"Payload":{}` in the compact JSON.

- **Condition**: File write operation fails mid-write due to disk full or I/O error.
  **Expected**: EventStore returns an error: `"failed to write event: <error>"`. The file may be partially written. The next append operation continues from the end of the file, potentially leaving a corrupted line. (Note: true atomic append requires OS-level support; this is a best-effort design.)

- **Condition**: File read operation encounters a permission error after acquiring the lock.
  **Expected**: EventStore returns an error: `"failed to read event file: permission denied"`.

- **Condition**: `SessionUUID` is invalid (e.g., empty string or malformed UUID).
  **Expected**: StorageLayout produces a malformed path. FileAccessor and subsequent operations will fail with filesystem errors.

- **Condition**: `Event.Message` field is very long and appears in the middle of the JSON object serialization.
  **Expected**: The long `Message` still obscures other fields. This violates the design goal. Implementation must ensure `Message` is serialized last.

## Related

- [Event](../entities/event.md) - Defines the Event structure
- [FileAccessor](./file_accessor.md) - Used to access `events.jsonl` with preparation callback
- [StorageLayout](./storage_layout.md) - Provides the path to `events.jsonl`
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Event-driven workflow architecture
