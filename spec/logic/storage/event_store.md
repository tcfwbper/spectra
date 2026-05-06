# EventStore

## Overview

EventStore manages persistent storage of event history for a single session. It appends events to a JSONL (JSON Lines) file in compact format, ensuring events are written atomically with file-level write locks. EventStore reads the entire event history when requested, logging warnings for malformed lines via the injected Logger. It is append-only: events cannot be modified or deleted once written. EventStore does not manage the parent session directory; it expects the directory to exist before writing events.

**Persistence Role**: The Read operation is provided for external tools and debugging only (e.g., post-session analysis). During active Runtime execution, the in-memory Session.EventHistory is the authoritative event list. The Runtime does not read from EventStore to make decisions or restore state during normal operation.

## Boundaries

- Owns: append-only persistence of Event entities to `events.jsonl` in compact JSONL format.
- Owns: custom JSON serialization of Event for persistence, including field ordering (Message last). This serialization logic is independent of the Event entity's struct field layout.
- Owns: reading and parsing events from `events.jsonl`, logging warnings for malformed lines via Logger.
- Owns: file-level locking (exclusive for writes, shared for reads).
- Owns: enforcement of `MaxPayloadSize` per-event size limit.
- Delegates: path composition to StorageLayout.
- Delegates: file existence check and preparation callback to FileAccessor.
- Delegates: session directory creation to SessionDirectoryManager (called before store usage).
- Must not: create the session directory (`.spectra/sessions/<session-uuid>/`).
- Must not: modify or delete existing lines in the events file.
- Must not: cache events in memory between calls.
- Must not: validate event semantics (session existence, event type defined in workflow).
- Must not: update any Session entity state — it is a passive persistence sink.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `StorageLayout` | Path composition | `GetEventHistoryPath` | Must not bypass for path construction |
| `FileAccessor` | File preparation | Call with path and preparation callback | Must not use for content read/write |
| `Event` | Data source for serialization | Read fields via exported getter methods | Must not rely on Event's struct field order or any entity-level MarshalJSON |
| `logger.Logger` | Structured logging | `Warn(msg string, args ...any)` | Must not use Logger for structured return values; must not use stdlib log directly |

Construction constraint: Must be constructed via `NewEventStore(projectRoot, sessionUUID string, logger logger.Logger)`. Direct struct literal is forbidden. Constructor only composes the file path via StorageLayout — no I/O is performed at construction time.

## Behavior

1. `NewEventStore(projectRoot, sessionUUID, logger)` composes the `events.jsonl` path via `StorageLayout.GetEventHistoryPath` and stores it internally. Stores the `logger` reference for warning output. No I/O is performed.
2. When appending an event, EventStore first calls FileAccessor with the file path and a preparation callback.
3. The preparation callback checks if the parent directory (`.spectra/sessions/<session-uuid>/`) exists. If it does not exist, the callback returns an error: `"session directory does not exist: <path>"`.
4. If the parent directory exists, the callback creates an empty `events.jsonl` file with permissions `0644`.
5. After FileAccessor confirms the file exists, EventStore acquires an exclusive file-level lock (using `flock` or equivalent) to prevent concurrent writes.
6. EventStore checks the serialized event size against `MaxPayloadSize`. If exceeded, returns an error without writing.
7. EventStore serializes the event to compact JSON format (single-line, no unnecessary whitespace) using custom serialization logic that reads Event fields via getter methods and constructs the JSON output directly. EventStore controls the field ordering to ensure `Message` appears last in the JSON object, improving readability for streaming readers (long message text does not obscure other fields). This serialization is independent of the Event entity's struct field layout.
8. EventStore appends the compact JSON line followed by a newline (`\n`) to the file.
9. EventStore flushes the file buffer and releases the lock after each write.
10. When reading events, EventStore opens the file with a shared read lock.
11. EventStore reads the file line-by-line. For each line, it attempts to parse the JSON into an Event entity via the constructor.
12. If a line is malformed (invalid JSON or missing required fields), EventStore logs a warning via Logger with the line number and description (e.g., `"skipping malformed event line"` with args: line number, parse error), and skips the line.
13. EventStore returns all successfully parsed events in the order they appear in the file.
14. If the file does not exist (and no events have been written yet), EventStore returns an empty list of events and no warnings, without error.
15. EventStore does not cache events in memory. Each read operation reads from disk.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| sessionUUID | string | UUID v4 format string | Yes |
| logger | logger.Logger | Non-nil Logger interface implementation from the logger package | Yes |

### For Append Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| event | Event entity | Valid Event as defined in `logic/entities/event.md` | Yes |

### For Read Operation

No inputs beyond the instance state (file path).

## Outputs

### For Construction

| Field | Type | Description |
|-------|------|-------------|
| eventStore | *EventStore | Configured instance holding the file path |

No error — constructor does not perform I/O.

### For Append Operation

**Success Case**: nil error.

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"session directory does not exist: <path>"` | Parent directory does not exist |
| `"failed to prepare file <path>: <callback error>"` | FileAccessor preparation failed |
| `"failed to acquire write lock: <error>"` | Unable to acquire exclusive file lock |
| `"event size exceeds limit: <actual-size> bytes (max <MaxPayloadSize> bytes)"` | Serialized event exceeds MaxPayloadSize |
| `"failed to serialize event: <error>"` | JSON marshaling failed |
| `"failed to write event: <error>"` | File write or flush operation failed |

### For Read Operation

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| events | []Event | List of events in chronological order (order of appearance in file) |

Malformed lines are logged as warnings via Logger and skipped. They do not appear in the return value or as errors.

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"failed to acquire read lock: <error>"` | Unable to acquire shared file lock |
| `"failed to read event file: <error>"` | File read operation failed (not including "file does not exist") |

## Invariants

1. **Append-Only**: EventStore must never modify or delete existing lines in the `events.jsonl` file. All writes must append to the end.

2. **Compact JSON Format**: Each event must be serialized as a single-line JSON string with no newlines or unnecessary whitespace within the JSON object.

3. **Message Field Last**: The `Message` field must appear last in the JSON object serialization. This is achieved by EventStore's own custom serialization logic, not by relying on the Event entity's struct field order.

4. **File-Level Locking**: Write operations must acquire an exclusive lock. Read operations must acquire a shared lock. Locks must be released before the function returns (including error paths).

5. **Atomic Append**: Each event write is a best-effort atomic append. If a write fails mid-operation, the file may contain a partially written line.

6. **File Permissions**: Newly created `events.jsonl` files must have permissions `0644`.

7. **No In-Memory Cache**: EventStore does not cache events in memory. Each read operation reads from disk.

8. **Idempotent Read**: Reading events multiple times returns the same event list if the file has not been modified. Warnings are logged again on each read.

9. **Parent Directory Requirement**: EventStore must not create the session directory. It returns an error if the directory does not exist.

10. **Line-Oriented Format**: Each event must be terminated with exactly one newline character (`\n`).

11. **Size Limit**: Each serialized event must not exceed `MaxPayloadSize` (defined as a package-level constant in the storage package). Events exceeding this size are rejected with an error and not appended to the file.

12. **No Constructor I/O**: The constructor performs only path composition. No filesystem access occurs until Append or Read is called.

## Edge Cases

- **Condition**: `events.jsonl` does not exist, and this is the first append operation.
  **Expected**: FileAccessor callback creates the file. EventStore writes the first event to the new file.

- **Condition**: `events.jsonl` does not exist, and a read operation is requested.
  **Expected**: EventStore returns an empty event list `[]` and empty warnings `[]` without error.

- **Condition**: Parent directory `.spectra/sessions/<session-uuid>/` does not exist when appending.
  **Expected**: FileAccessor callback returns an error: `"session directory does not exist: <path>"`. EventStore propagates the error.

- **Condition**: File contains a malformed JSON line (e.g., missing closing brace).
  **Expected**: EventStore logs a warning via Logger (e.g., line number and "invalid JSON: unexpected end of input") and continues reading subsequent lines.

- **Condition**: File contains a blank line.
  **Expected**: EventStore logs a warning via Logger (e.g., line number and "empty line") and skips it.

- **Condition**: File contains a line with valid JSON but missing required Event fields (e.g., no `ID` field).
  **Expected**: EventStore logs a warning via Logger (e.g., line number and "invalid event: <validation error>") and skips the line.

- **Condition**: Two goroutines attempt to append events simultaneously to the same file.
  **Expected**: The file-level exclusive lock serializes the writes. Both events are appended successfully without corruption.

- **Condition**: A goroutine attempts to read events while another is appending.
  **Expected**: The read operation acquires a shared lock and blocks until the write completes and releases the exclusive lock.

- **Condition**: `Message` field contains newline characters or Unicode characters.
  **Expected**: JSON encoding escapes these characters. The serialized line remains a single line.

- **Condition**: Total serialized event size exceeds `MaxPayloadSize`.
  **Expected**: EventStore rejects the write with error: `"event size exceeds limit: <actual-size> bytes (max <MaxPayloadSize> bytes)"`. The event is not appended.

- **Condition**: File write fails mid-write due to disk full or I/O error.
  **Expected**: EventStore returns error `"failed to write event: <error>"`. The file may contain a partially written line. The next read will report that line as a warning.

- **Condition**: Event `Payload` is an empty JSON object `{}`.
  **Expected**: EventStore serializes it as `"Payload":{}` in the compact JSON.

## Related

- [Event](../entities/event.md) — Defines the Event entity structure and constructor
- [FileAccessor](./file_accessor.md) — Used to ensure `events.jsonl` exists with preparation callback
- [StorageLayout](./storage_layout.md) — Provides the path to `events.jsonl`
- [SessionDirectoryManager](./session_directory_manager.md) — Creates session directories before store usage
- [Logger](../logger/logger.md) — Structured logging interface used for malformed line warnings
