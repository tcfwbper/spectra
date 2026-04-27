# SessionMetadataStore

## Overview

SessionMetadataStore manages persistent storage of session metadata for a single session. It reads and writes the `session.json` file in pretty-printed JSON format, ensuring metadata is written atomically with file-level write locks. SessionMetadataStore handles only the session metadata fields (ID, WorkflowName, Status, CreatedAt, UpdatedAt, CurrentState, SessionData, Error); it does not manage EventHistory, which is maintained in memory by the Runtime. SessionMetadataStore does not manage the parent session directory; it expects the directory to exist before writing metadata.

**Persistence Role**: SessionMetadataStore provides best-effort, last-write-wins persistence for user inspection and debugging. The Read operation is intended for external tools only (e.g., `spectra clear`, `spectra status`). The running Runtime must never read from SessionMetadataStore to determine behavior; the in-memory Session entity is the authoritative source of truth.

## Behavior

1. SessionMetadataStore is initialized with a session UUID and uses StorageLayout to determine the path to `session.json`.
2. SessionMetadataStore uses FileAccessor to access the `session.json` file, providing a callback that creates the file on first write.
3. The preparation callback checks if the parent directory (`.spectra/sessions/<session-uuid>/`) exists. If it does not exist, the callback returns an error: "session directory does not exist: <path>".
4. If the parent directory exists, the callback creates an empty `session.json` file with permissions `0644`.
5. When writing session metadata, SessionMetadataStore acquires an exclusive file-level lock (using `flock` or equivalent) to prevent concurrent writes from multiple goroutines or processes.
6. SessionMetadataStore serializes the session metadata to pretty-printed JSON format with 2-space indentation.
7. The `Error` field is serialized with the `omitempty` JSON tag. When `Error` is nil, the field is omitted from the JSON output. When `Error` is set, it is serialized as a nested JSON object.
8. The `EventHistory` field is excluded from serialization entirely. SessionMetadataStore only persists metadata fields defined in the session structure minus EventHistory.
9. SessionMetadataStore writes the JSON content to the file, replacing any existing content (truncate and write).
10. SessionMetadataStore updates the `UpdatedAt` timestamp to the current POSIX timestamp before each write operation.
11. SessionMetadataStore flushes the file buffer and releases the lock after each write.
12. When reading session metadata, SessionMetadataStore opens the file with a shared read lock.
13. SessionMetadataStore reads the entire file content and parses the JSON into a session metadata structure.
14. If the file does not exist, SessionMetadataStore returns an error: "session metadata file does not exist: <path>".
15. If the file contains invalid JSON or missing required fields, SessionMetadataStore returns a parsing error with details.
16. SessionMetadataStore does not cache metadata in memory. Each read operation reads from disk.
17. After a successful read, the `EventHistory` field is empty (nil or empty slice). It is the Runtime's responsibility to populate EventHistory from EventStore.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| SessionUUID | string (UUID) | Valid UUID v4 format | Yes |

### For Write Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionMetadata | SessionMetadata struct | Valid session metadata structure (ID, WorkflowName, Status, CreatedAt, UpdatedAt, CurrentState, SessionData, Error). EventHistory is ignored if present. | Yes |

## Outputs

### For Write Operation

**Success Case**: No return value (void / nil error in Go).

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"session directory does not exist: <path>"` | Parent directory `.spectra/sessions/<session-uuid>/` does not exist |
| `"failed to acquire write lock: <error>"` | Unable to acquire exclusive file lock |
| `"failed to serialize session metadata: <error>"` | JSON marshaling failed |
| `"failed to write session metadata: <error>"` | File write operation failed |

### For Read Operation

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| SessionMetadata | SessionMetadata struct | Session metadata structure with all fields except EventHistory populated from the file. EventHistory is empty. |

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"session metadata file does not exist: <path>"` | The `session.json` file does not exist |
| `"failed to acquire read lock: <error>"` | Unable to acquire shared file lock |
| `"failed to read session metadata file: <error>"` | File read operation failed |
| `"failed to parse session metadata: <error>"` | JSON parsing failed or required fields are missing |

## Invariants

1. **Pretty-Printed JSON Format**: SessionMetadataStore must serialize session metadata as pretty-printed JSON with 2-space indentation for human readability.

2. **EventHistory Exclusion**: The `EventHistory` field must never be serialized to `session.json`. SessionMetadataStore must exclude this field from JSON output.

3. **Error Field Omitempty**: The `Error` field must use the `omitempty` JSON tag. When nil, the field is omitted from the JSON. When set, the full AgentError structure is serialized.

4. **File-Level Locking**: Write operations must acquire an exclusive lock. Read operations must acquire a shared lock. Locks must be released before the function returns (including error paths).

5. **Atomic Write**: Each metadata write must replace the entire file content atomically. Partial updates are not supported.

6. **File Permissions**: Newly created `session.json` files must have permissions `0644` (owner read/write, group/others read-only).

7. **No In-Memory Cache**: SessionMetadataStore does not cache metadata in memory. Each read operation reads from disk. Note: Read is intended for external inspection and debugging tools only. The running Runtime must never read from SessionMetadataStore to determine behavior; the in-memory Session entity is the authoritative source of truth.

8. **UpdatedAt Auto-Update**: The `UpdatedAt` field must be automatically updated to the current POSIX timestamp before each write operation, ensuring it reflects the last modification time.

9. **Parent Directory Requirement**: SessionMetadataStore must not create the session directory. It must return an error if the directory does not exist.

10. **Idempotent Read**: Reading session metadata multiple times must return the same results if the file has not been modified.

11. **Session Status Consistency**: SessionMetadataStore does not validate session state transitions (e.g., "initializing" -> "running"). It is the Runtime's responsibility to enforce state machine invariants before calling write operations.

12. **Concurrent Access Safety**: SessionMetadataStore must handle concurrent read and write operations safely using file-level locks, supporting both single-process (multiple goroutines) and multi-process concurrent access.

13. **Concurrent Write Resolution**: When multiple writers attempt to modify session metadata concurrently, the file-level lock serializes writes. The last write to acquire the lock overwrites all previous changes (last write wins). The Runtime never performs read-modify-write against SessionMetadataStore. All writes originate from the in-memory Session entity, which is the authoritative source. Persistence is best-effort and uses last-write-wins semantics.

14. **Size Limit**: Each serialized metadata payload must not exceed 10 MB. This matches the wire-level limit enforced by RuntimeSocketManager.

## Edge Cases

- **Condition**: `session.json` does not exist, and this is the first write operation.
  **Expected**: FileAccessor callback creates the file. SessionMetadataStore writes the initial metadata to the new file.

- **Condition**: `session.json` does not exist, and a read operation is requested.
  **Expected**: SessionMetadataStore returns an error: `"session metadata file does not exist: <path>"`.

- **Condition**: Parent directory `.spectra/sessions/<session-uuid>/` does not exist when writing.
  **Expected**: FileAccessor callback returns an error: `"session directory does not exist: <path>"`. SessionMetadataStore propagates the error.

- **Condition**: File contains invalid JSON (e.g., missing closing brace, syntax error).
  **Expected**: SessionMetadataStore returns an error: `"failed to parse session metadata: unexpected end of JSON input"`.

- **Condition**: File contains valid JSON but missing required fields (e.g., no `ID` field).
  **Expected**: SessionMetadataStore returns an error: `"failed to parse session metadata: missing required field 'ID'"`.

- **Condition**: Two goroutines or processes attempt to write session metadata simultaneously.
  **Expected**: The file-level exclusive lock serializes the writes. One acquires the lock, writes, releases. The other waits, then writes. The second write overwrites the first (last write wins).

- **Condition**: A goroutine attempts to read session metadata while another goroutine or process is writing.
  **Expected**: The read operation acquires a shared lock and blocks until the write operation completes and releases the exclusive lock. The read then proceeds with the updated content.

- **Condition**: `SessionData` contains a very large nested JSON object (e.g., 10 MB).
  **Expected**: SessionMetadataStore serializes the entire object as pretty-printed JSON. The file may be very large. Performance may degrade. Consider implementing a size warning or limit in the Runtime layer.

- **Condition**: `SessionData` contains values that cannot be serialized to JSON (e.g., Go channels, functions).
  **Expected**: JSON marshaling fails. SessionMetadataStore returns an error: `"failed to serialize session metadata: json: unsupported type: chan int"`.

- **Condition**: `Error` field is nil (session is not failed).
  **Expected**: The `error` field is omitted from the JSON output entirely due to the `omitempty` tag. The file does not contain an `"error": null` line.

- **Condition**: `Error` field is set (session is failed).
  **Expected**: The full `AgentError` structure is serialized as a nested JSON object under the `"error"` key.

- **Condition**: File write operation fails mid-write due to disk full or I/O error.
  **Expected**: SessionMetadataStore returns an error: `"failed to write session metadata: <error>"`. The file may be corrupted (partially written). Subsequent reads may fail. The Runtime should handle this as a critical error.

- **Condition**: File read operation encounters a permission error after acquiring the lock.
  **Expected**: SessionMetadataStore returns an error: `"failed to read session metadata file: permission denied"`.

- **Condition**: `SessionUUID` is invalid (e.g., empty string or malformed UUID).
  **Expected**: StorageLayout produces a malformed path. FileAccessor and subsequent operations will fail with filesystem errors.

- **Condition**: Multiple goroutines in the same process write to `session.json` concurrently.
  **Expected**: The file-level lock serializes writes correctly, even within the same process. No data races occur.

- **Condition**: `UpdatedAt` is manually set by the caller before write operation.
  **Expected**: SessionMetadataStore overwrites the `UpdatedAt` field with the current timestamp. The caller-provided value is ignored.

- **Condition**: `EventHistory` is present in the input SessionMetadata structure passed to write operation.
  **Expected**: SessionMetadataStore ignores the `EventHistory` field entirely. It is not serialized to JSON. The Runtime must ensure EventHistory is populated separately by querying EventStore.

- **Condition**: JSON file contains an `"eventHistory"` field from a previous version or manual edit.
  **Expected**: SessionMetadataStore reads the field but does not populate it into the returned structure (assuming the Go struct tags exclude it). The field is effectively ignored.

- **Condition**: `SessionData` contains a key with namespace prefix `<NodeName>.ClaudeSessionID` with a non-string value.
  **Expected**: SessionMetadataStore serializes the value as-is (e.g., number, object). It does not validate the type. The Runtime is responsible for validating Claude session ID values are strings before writing.

- **Condition**: Pretty-printed JSON exceeds 10 MB.
  **Expected**: SessionMetadataStore rejects the write with an error: `"session metadata size exceeds 10 MB limit: <actual-size> bytes"`. This matches the wire-level limit enforced by RuntimeSocketManager.

## Related

- [Session](../entities/session/session.md) - Defines the Session structure and lifecycle
- [EventStore](./event_store.md) - Manages EventHistory separately from session metadata
- [FileAccessor](./file_accessor.md) - Used to access `session.json` with preparation callback
- [StorageLayout](./storage_layout.md) - Provides the path to `session.json`
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Session lifecycle and workflow runtime
