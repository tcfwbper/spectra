# SessionMetadataStore

## Overview

SessionMetadataStore manages persistent storage of SessionMetadata for a single session. It reads and writes the `session.json` file in pretty-printed JSON format, ensuring metadata is written with file-level write locks. SessionMetadataStore serializes the SessionMetadata fields (ID, WorkflowName, Pid, Status, CreatedAt, UpdatedAt, CurrentState, SessionData, Error) and does not serialize EventHistory, which is persisted separately by EventStore. SessionMetadataStore does not manage the parent session directory; it expects the directory to exist before writing metadata.

**Persistence Role**: SessionMetadataStore provides best-effort, last-write-wins persistence for user inspection and debugging. The Read operation is intended for external tools only (e.g., `spectra clear`, `spectra status`). The running Runtime must never read from SessionMetadataStore to determine behavior; the in-memory Session entity is the authoritative source of truth.

## Boundaries

- Owns: write persistence of SessionMetadata snapshot to `session.json` (truncate-and-write).
- Owns: read and parse of SessionMetadata from `session.json`.
- Owns: file-level locking (exclusive for writes, shared for reads).
- Owns: enforcement of `MaxPayloadSize` per-write size limit.
- Owns: serialization of the Error field and type recovery on read via mutually exclusive fields (`agentRole` for AgentError, `issuer` for RuntimeError).
- Delegates: path composition to StorageLayout.
- Delegates: file existence check and preparation callback to FileAccessor.
- Delegates: session directory creation to SessionDirectoryManager (called before store usage).
- Delegates: all field mutations and timestamp updates to the Session entity.
- Must not: create the session directory (`.spectra/sessions/<session-uuid>/`).
- Must not: modify, validate, or overwrite the `UpdatedAt` field — it uses the value from the provided snapshot as-is.
- Must not: validate session state transitions (e.g., "initializing" → "running").
- Must not: cache metadata in memory between calls.
- Must not: serialize EventHistory.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `StorageLayout` | Path composition | `GetSessionMetadataPath` | Must not bypass for path construction |
| `FileAccessor` | File preparation | Call with path and preparation callback | Must not use for content read/write |
| `AgentError` | Error field source | Read fields via exported getter methods for serialization; reconstruct via `NewAgentError` on read | Must not call entity MarshalJSON; must not rely on entity-level JSON tags for the error sub-object |
| `RuntimeError` | Error field source | Read fields via exported getter methods for serialization; reconstruct via `NewRuntimeError` on read | Must not call entity MarshalJSON; must not rely on entity-level JSON tags for the error sub-object |

Construction constraint: Must be constructed via `NewSessionMetadataStore(projectRoot, sessionUUID string)`. Direct struct literal is forbidden. Constructor only composes the file path via StorageLayout — no I/O is performed at construction time.

## Behavior

1. `NewSessionMetadataStore(projectRoot, sessionUUID)` composes the `session.json` path via `StorageLayout.GetSessionMetadataPath` and stores it internally. No I/O is performed.
2. When writing session metadata, SessionMetadataStore first calls FileAccessor with the file path and a preparation callback.
3. The preparation callback checks if the parent directory (`.spectra/sessions/<session-uuid>/`) exists. If it does not exist, the callback returns an error: `"session directory does not exist: <path>"`.
4. If the parent directory exists, the callback creates an empty `session.json` file with permissions `0644`.
5. After FileAccessor confirms the file exists, SessionMetadataStore acquires an exclusive file-level lock.
6. SessionMetadataStore serializes the session metadata to pretty-printed JSON format with 2-space indentation.
7. The `Error` field is serialized with the `omitempty` semantics. When `Error` is nil, the field is omitted from the JSON output. When `Error` is set, SessionMetadataStore reads the error entity's fields via its exported getter methods and constructs the JSON sub-object directly (it does not call the entity's MarshalJSON). No explicit discriminator field is injected; the two error types are distinguished on read by mutually exclusive fields: `"agentRole"` is present only for AgentError, `"issuer"` is present only for RuntimeError.
8. The `UpdatedAt` field is serialized as-is from the provided snapshot. SessionMetadataStore does not modify it.
9. SessionMetadataStore checks the serialized metadata size against `MaxPayloadSize`. If exceeded, returns an error without writing.
10. SessionMetadataStore writes the JSON content to the file, replacing any existing content (truncate and write).
11. SessionMetadataStore flushes the file buffer and releases the lock after each write.
12. When reading session metadata, SessionMetadataStore opens the file with a shared read lock.
13. SessionMetadataStore reads the entire file content and parses the JSON into a SessionMetadata structure.
14. During deserialization of the `Error` field, SessionMetadataStore inspects the parsed JSON object for mutually exclusive fields to determine the concrete type: if `"agentRole"` is present, it reconstructs via `NewAgentError`; if `"issuer"` is present, it reconstructs via `NewRuntimeError`. If both or neither field is present, it returns a reconstruction error.
15. If the file does not exist, SessionMetadataStore returns an error: `"session metadata file does not exist: <path>"`.
16. If the file contains invalid JSON or missing required fields, SessionMetadataStore returns a parsing error with details.
17. SessionMetadataStore does not cache metadata in memory. Each read operation reads from disk.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| sessionUUID | string | UUID v4 format string | Yes |

### For Write Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| metadata | SessionMetadata struct | Valid SessionMetadata snapshot (obtained via `GetMetadataSnapshotSafe`) | Yes |

### For Read Operation

No inputs beyond the instance state (file path).

## Outputs

### For Construction

| Field | Type | Description |
|-------|------|-------------|
| store | *SessionMetadataStore | Configured instance holding the file path |

No error — constructor does not perform I/O.

### For Write Operation

**Success Case**: nil error.

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"session directory does not exist: <path>"` | Parent directory does not exist |
| `"failed to prepare file <path>: <callback error>"` | FileAccessor preparation failed |
| `"failed to acquire write lock: <error>"` | Unable to acquire exclusive file lock |
| `"session metadata size exceeds limit: <actual-size> bytes (max <MaxPayloadSize> bytes)"` | Serialized metadata exceeds MaxPayloadSize |
| `"failed to serialize session metadata: <error>"` | JSON marshaling failed |
| `"failed to write session metadata: <error>"` | File write or flush operation failed |

### For Read Operation

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| metadata | SessionMetadata struct | SessionMetadata populated from the file (EventHistory is not included) |

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"session metadata file does not exist: <path>"` | The `session.json` file does not exist |
| `"failed to acquire read lock: <error>"` | Unable to acquire shared file lock |
| `"failed to read session metadata file: <error>"` | File read operation failed |
| `"failed to parse session metadata: <error>"` | JSON parsing failed or required fields are missing |
| `"failed to reconstruct error: <error>"` | Error field present but type determination (via mutual exclusion) or constructor reconstruction failed |

## Invariants

1. **Pretty-Printed JSON Format**: SessionMetadataStore must serialize session metadata as pretty-printed JSON with 2-space indentation for human readability.

2. **EventHistory Exclusion**: SessionMetadataStore must only serialize SessionMetadata fields. EventHistory is persisted separately by EventStore.

3. **Pid Field Serialization**: The `pid` field is always serialized as a JSON number (no omitempty). It is always > 0 for newly created sessions. When reading legacy `session.json` files that lack the `pid` field, SessionMetadataStore defaults `Pid` to 0 without returning an error (backward compatibility).

4. **Error Field Omitempty**: When `Error` is nil, the field is omitted from the JSON output. When set, the full error structure is serialized without an explicit discriminator.

5. **Error Type Via Mutual Exclusion**: When `Error` is non-nil, the concrete type is determined by mutually exclusive fields: `"agentRole"` (present only for AgentError) and `"issuer"` (present only for RuntimeError). No `"errorType"` discriminator field is written or expected. The `"agentRole"` key must always be written for AgentError regardless of whether the value is an empty string (i.e., no omitempty on this field during serialization).

6. **File-Level Locking**: Write operations must acquire an exclusive lock. Read operations must acquire a shared lock. Locks must be released before the function returns (including error paths).

7. **Truncate-and-Write**: Each metadata write replaces the entire file content. Partial updates are not supported.

8. **File Permissions**: Newly created `session.json` files must have permissions `0644`.

9. **No In-Memory Cache**: SessionMetadataStore does not cache metadata in memory. Each read operation reads from disk.

10. **UpdatedAt Pass-Through**: SessionMetadataStore serializes `UpdatedAt` as-is from the provided snapshot. It must not overwrite or modify this value.

11. **Parent Directory Requirement**: SessionMetadataStore must not create the session directory. It returns an error if the directory does not exist.

12. **Idempotent Read**: Reading session metadata multiple times returns the same result if the file has not been modified.

13. **No State Transition Validation**: SessionMetadataStore does not validate session status transitions. It persists whatever snapshot it receives.

14. **Concurrent Access Safety**: File-level locks support both single-process (multiple goroutines) and multi-process concurrent access. Last write wins.

15. **Size Limit**: Each serialized metadata payload must not exceed `MaxPayloadSize` (defined as a package-level constant in the storage package).

16. **No Constructor I/O**: The constructor performs only path composition. No filesystem access occurs until Write or Read is called.

17. **Error Reconstruction Via Constructors**: When reading the Error field, SessionMetadataStore must reconstruct `*AgentError` or `*RuntimeError` via their respective constructors (`NewAgentError`, `NewRuntimeError`), not via direct struct literal assignment.

18. **Pid Backward Compatibility**: When reading a `session.json` that lacks the `"pid"` field, `Pid` defaults to 0. No error is raised. This allows external tools to handle legacy sessions gracefully.

## Edge Cases

- **Condition**: `session.json` does not exist, and this is the first write operation.
  **Expected**: FileAccessor callback creates the file. SessionMetadataStore writes the initial metadata.

- **Condition**: `session.json` does not exist, and a read operation is requested.
  **Expected**: Returns error: `"session metadata file does not exist: <path>"`.

- **Condition**: Parent directory `.spectra/sessions/<session-uuid>/` does not exist when writing.
  **Expected**: FileAccessor callback returns error: `"session directory does not exist: <path>"`. SessionMetadataStore propagates the error.

- **Condition**: File contains invalid JSON (e.g., missing closing brace).
  **Expected**: Returns error: `"failed to parse session metadata: unexpected end of JSON input"`.

- **Condition**: File contains valid JSON but missing required fields (e.g., no `ID` field).
  **Expected**: Returns error: `"failed to parse session metadata: missing required field 'ID'"`.

- **Condition**: Two goroutines or processes attempt to write session metadata simultaneously.
  **Expected**: The file-level exclusive lock serializes the writes. The second write overwrites the first (last write wins).

- **Condition**: A goroutine attempts to read while another is writing.
  **Expected**: The read blocks until the write completes and releases the exclusive lock.

- **Condition**: `SessionData` contains values that cannot be serialized to JSON (e.g., Go channels, functions).
  **Expected**: JSON marshaling fails. Returns error: `"failed to serialize session metadata: json: unsupported type: <type>"`.

- **Condition**: `Error` field is nil (session is not failed).
  **Expected**: The `"error"` key is omitted from the JSON output entirely.

- **Condition**: `Error` field is set with an `*AgentError`.
  **Expected**: Serialized as `"error": {"agentRole": "...", "message": "...", ...}` (no `errorType` discriminator).

- **Condition**: `Error` field is set with a `*RuntimeError`.
  **Expected**: Serialized as `"error": {"issuer": "...", "message": "...", ...}` (no `errorType` discriminator).

- **Condition**: Read encounters an `"error"` field containing both `"agentRole"` and `"issuer"`.
  **Expected**: Returns error: `"failed to reconstruct error: ambiguous error object contains both 'agentRole' and 'issuer'"`.

- **Condition**: Read encounters an `"error"` field containing neither `"agentRole"` nor `"issuer"`.
  **Expected**: Returns error: `"failed to reconstruct error: cannot determine error type — missing 'agentRole' or 'issuer' field"`.

- **Condition**: Read encounters an `"error"` field with valid type-discriminating field but invalid fields for the constructor.
  **Expected**: Returns error: `"failed to reconstruct error: <constructor validation error>"`.

- **Condition**: File write fails mid-write due to disk full or I/O error.
  **Expected**: Returns error: `"failed to write session metadata: <error>"`. The file may be corrupted (partially written). Subsequent reads may fail with a parse error.

- **Condition**: Pretty-printed JSON exceeds `MaxPayloadSize`.
  **Expected**: Returns error: `"session metadata size exceeds limit: <actual-size> bytes (max <MaxPayloadSize> bytes)"`. The file is not modified.

- **Condition**: JSON file contains an `"eventHistory"` field from a manual edit.
  **Expected**: The field is ignored during parsing. The returned SessionMetadata does not include EventHistory.

- **Condition**: JSON file was created by an older version and lacks the `"pid"` field.
  **Expected**: `Pid` defaults to 0 in the returned SessionMetadata. No error is raised.

## Related

- [SessionMetadata](../entities/session/session_metadata.md) — Defines the SessionMetadata structure that is persisted
- [Session](../entities/session/session.md) — Provides `GetMetadataSnapshotSafe` for obtaining the snapshot to write
- [AgentError](../entities/agent_error.md) — Possible concrete type for the Error field
- [RuntimeError](../entities/runtime_error.md) — Possible concrete type for the Error field
- [EventStore](./event_store.md) — Manages EventHistory separately from SessionMetadata
- [FileAccessor](./file_accessor.md) — Used to ensure `session.json` exists with preparation callback
- [StorageLayout](./storage_layout.md) — Provides the path to `session.json`
- [SessionDirectoryManager](./session_directory_manager.md) — Creates session directories before store usage
