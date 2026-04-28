# Test Specification: `session_metadata_store.go`

## Source File Under Test
`storage/session_metadata_store.go`

## Test File
`storage/session_metadata_store_test.go`

---

## `SessionMetadataStore`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_New` | `unit` | Constructs SessionMetadataStore with valid inputs. | Temporary test directory created programmatically within test fixture with `.spectra/sessions/` subdirectory | `ProjectRoot=<test-fixture>`, `SessionUUID=<valid-uuid-v4>` | Returns SessionMetadataStore instance; no error |

### Happy Path — Write (First Write)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteFirst` | `unit` | Writes metadata to non-existent file. | Temporary test directory created programmatically within test fixture with session directory `.spectra/sessions/<uuid>/` created; `session.json` does not exist | SessionMetadata with all required fields (ID, WorkflowName, Status, CreatedAt, UpdatedAt, CurrentState, SessionData) | Returns nil; `session.json` created with permissions `0644`; contains pretty-printed JSON with 2-space indentation |
| `TestSessionMetadataStore_WriteCreatesFile` | `unit` | FileAccessor callback creates file on first write. | Temporary test directory created programmatically within test fixture with session directory created; `session.json` does not exist | SessionMetadata | Returns nil; file created with correct permissions |

### Happy Path — Write (Subsequent Writes)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteOverwrites` | `unit` | Second write replaces entire file content. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with initial metadata created inside test fixture | SessionMetadata with Status changed from "initializing" to "running" | Returns nil; file content fully replaced; new Status reflected; old content not present |
| `TestSessionMetadataStore_WriteMultipleTimes` | `unit` | Multiple writes succeed with last-write-wins. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Write metadata 3 times with different Status values | Each write succeeds; final file contains last written Status |

### Happy Path — Pretty-Printed JSON

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_PrettyPrinted2SpaceIndent` | `unit` | Serializes as pretty-printed JSON with 2-space indentation. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with nested SessionData | File contains multi-line JSON; uses 2-space indentation; human-readable format |
| `TestSessionMetadataStore_EmptySessionData` | `unit` | Serializes empty SessionData as empty object. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `SessionData={}` | JSON contains `"sessionData": {}` or equivalent pretty-printed form |

### Happy Path — EventHistory Exclusion

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_EventHistoryNotSerialized` | `unit` | EventHistory field excluded from JSON output (not part of SessionMetadata struct). | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata struct (does not include EventHistory field) | Returns nil; file does not contain "eventHistory" field; only SessionMetadata fields present |
| `TestSessionMetadataStore_EventHistoryFieldIgnored` | `unit` | Pre-existing eventHistory field in JSON ignored on read. | Temporary test directory created programmatically within test fixture; session directory and `session.json` manually created with "eventHistory" array inside test fixture | | Returns SessionMetadata; EventHistory not populated (field does not exist in SessionMetadata struct); other fields read correctly |

### Happy Path — Error Field Omitempty

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ErrorFieldOmittedWhenNil` | `unit` | Error field omitted when nil. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `Error=nil` | File does not contain "error" field; no `"error": null` line |
| `TestSessionMetadataStore_ErrorFieldPresentWhenSet` | `unit` | Error field serialized when set. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `Error=<AgentError-instance>` | File contains "error" field with nested AgentError JSON structure |

### Happy Path — UpdatedAt Auto-Update

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_UpdatedAtAutoUpdated` | `unit` | UpdatedAt automatically set to current timestamp. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; record timestamp before write | SessionMetadata with `UpdatedAt=<old-timestamp>` | Returns nil; file contains `UpdatedAt` matching current POSIX timestamp (within 1 second); caller-provided value overwritten |
| `TestSessionMetadataStore_UpdatedAtChangesOnEachWrite` | `unit` | UpdatedAt updated on each write. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Write metadata; wait 2 seconds; write again with same data | First write has timestamp T1; second write has timestamp T2; T2 > T1 by ~2 seconds |

### Happy Path — Read (Existing File)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadValidFile` | `unit` | Reads metadata from valid JSON file. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with valid metadata created inside test fixture | | Returns SessionMetadata; all fields match file content; EventHistory is nil/empty |
| `TestSessionMetadataStore_ReadComplexSessionData` | `unit` | Reads metadata with complex nested SessionData. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with nested SessionData (arrays, objects) created inside test fixture | | Returns SessionMetadata; SessionData structure preserved; all nested fields correct |
| `TestSessionMetadataStore_ReadErrorFieldPresent` | `unit` | Reads metadata with Error field set. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with Error field created inside test fixture | | Returns SessionMetadata; Error field populated with correct AgentError structure |

### Happy Path — Read (SessionMetadata Structure)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadata Store_ReadOnlySessionMetadataFields` | `unit` | Read returns SessionMetadata struct with all persistable fields. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with complete metadata created inside test fixture | | Returns SessionMetadata with ID, WorkflowName, Status, CreatedAt, UpdatedAt, CurrentState, SessionData, Error populated correctly |

### Happy Path — File Locking (Write)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteAcquiresExclusiveLock` | `unit` | Acquires exclusive lock during write. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata | Lock acquired before write; lock released after write completes |
| `TestSessionMetadataStore_WriteReleasesLockOnError` | `unit` | Releases lock when write fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock FileAccessor that fails during write | SessionMetadata | Returns error; lock released before function returns |

### Happy Path — File Locking (Read)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadAcquiresSharedLock` | `unit` | Acquires shared read lock during read. | Temporary test directory created programmatically within test fixture; session directory and `session.json` created inside test fixture | | Shared lock acquired before read; lock released after read completes |
| `TestSessionMetadataStore_ReadReleasesLockOnError` | `unit` | Releases lock when read fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; file exists but fails during read | | Returns error; lock released before function returns |

### Validation Failures — Write (Parent Directory)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteParentDirDoesNotExist` | `unit` | Returns error when session directory missing. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/<uuid>/` directory does not exist | SessionMetadata | Returns error matching `/session directory does not exist:.*\.spectra\/sessions\/.*<uuid>/i`; file not created |

### Validation Failures — Write (Serialization)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteSerializationFails` | `unit` | Returns error when JSON marshaling fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with un-serializable field in SessionData (e.g., channel) | Returns error matching `/failed to serialize session metadata:/i`; file not modified |

### Validation Failures — Write (Write Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteWriteFails` | `unit` | Returns error when file write fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock FileAccessor that fails write | SessionMetadata | Returns error matching `/failed to write session metadata:/i` |
| `TestSessionMetadataStore_WriteLockFails` | `unit` | Returns error when lock acquisition fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock FileAccessor that fails lock | SessionMetadata | Returns error matching `/failed to acquire write lock:/i` |

### Validation Failures — Write (Size Limit)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteExceeds10MBLimit` | `unit` | Rejects metadata exceeding 10 MB serialized size. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with SessionData containing 11 MB of nested data | Returns error matching `/session metadata size exceeds 10 MB limit:.*bytes/i`; file not modified |
| `TestSessionMetadataStore_WriteExactly10MB` | `unit` | Accepts metadata at exactly 10 MB limit. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with SessionData totaling exactly 10 MB serialized (including indentation) | Returns nil; metadata written successfully |

### Validation Failures — Read (File Does Not Exist)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadFileDoesNotExist` | `unit` | Returns error when file missing. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `session.json` does not exist | | Returns error matching `/session metadata file does not exist:.*session\.json/i` |

### Validation Failures — Read (Lock Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadLockFails` | `unit` | Returns error when lock acquisition fails. | Temporary test directory created programmatically within test fixture; session directory and file created inside test fixture; mock FileAccessor that fails lock | | Returns error matching `/failed to acquire read lock:/i` |

### Validation Failures — Read (Parse Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadInvalidJSON` | `unit` | Returns error when JSON is malformed. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with invalid JSON (missing `}`) created inside test fixture | | Returns error matching `/failed to parse session metadata:.*unexpected end of JSON/i` |
| `TestSessionMetadataStore_ReadMissingRequiredField` | `unit` | Returns error when required field missing. | Temporary test directory created programmatically within test fixture; session directory and `session.json` with valid JSON but missing `ID` field created inside test fixture | | Returns error matching `/failed to parse session metadata:.*missing required field.*ID/i` |

### Validation Failures — Read (File Errors)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadFileReadFails` | `unit` | Returns error when file read operation fails. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; file exists but becomes unreadable after lock acquired | | Returns error matching `/failed to read session metadata file:/i` |
| `TestSessionMetadataStore_ReadPermissionDenied` | `unit` | Returns error when file permissions deny read. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `session.json` with permissions `0000` created inside test fixture | | Returns error matching `/permission denied/i` |

### Idempotency — Read

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadIdempotent` | `unit` | Multiple reads return identical results. | Temporary test directory created programmatically within test fixture; session directory and `session.json` created inside test fixture | | First read returns metadata; second read returns identical metadata; no caching; both reads access disk |

### Idempotency — Write

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteIdempotent` | `unit` | Writing same metadata twice produces consistent file. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Write same SessionMetadata twice (except UpdatedAt will differ) | Both writes succeed; second write overwrites first; file contains expected content |

### Concurrent Behaviour — Multiple Writers

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ConcurrentWriteSameFile` | `race` | Multiple goroutines write to same file safely. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `session.json` does not exist | 10 goroutines each write different metadata | All writes succeed; file contains metadata from last writer to acquire lock (last-write-wins); no corruption |
| `TestSessionMetadataStore_ConcurrentWriteSerializes` | `race` | File lock serializes concurrent writes. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | 5 goroutines write simultaneously with distinct Status values | Writes serialized by lock; final file contains valid complete JSON; content matches one of the written values |

### Concurrent Behaviour — Read During Write

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ReadBlocksDuringWrite` | `race` | Read waits for write to complete. | Temporary test directory created programmatically within test fixture; session directory and `session.json` created inside test fixture | Goroutine 1 writes (holds lock); Goroutine 2 reads (blocked) | Read blocked until write completes; read returns newly written metadata |
| `TestSessionMetadataStore_WriteBlocksDuringRead` | `race` | Write waits for read to complete. | Temporary test directory created programmatically within test fixture; session directory and `session.json` created inside test fixture | Goroutine 1 reads (holds lock); Goroutine 2 writes (blocked) | Write blocked until read completes; write succeeds after read releases lock |

### Concurrent Behaviour — Same Process

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_ConcurrentSameProcess` | `race` | Multiple goroutines in same process write safely. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | 5 goroutines in same process write concurrently | File lock works within process; writes serialized; no data races; final file valid |

### Boundary Values — File Permissions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_NewFilePermissions0644` | `unit` | Creates new file with correct permissions. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; `session.json` does not exist | SessionMetadata | File created with permissions `0644` (owner rw, group/others r) |

### Boundary Values — Invalid Session UUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_InvalidSessionUUID` | `unit` | Fails with malformed UUID. | Temporary test directory created programmatically within test fixture | `ProjectRoot=<test-fixture>`, `SessionUUID="not-a-uuid"` | Subsequent write/read operations fail with filesystem errors (malformed path) |
| `TestSessionMetadataStore_EmptySessionUUID` | `unit` | Fails with empty UUID. | Temporary test directory created programmatically within test fixture | `ProjectRoot=<test-fixture>`, `SessionUUID=""` | Subsequent operations fail with filesystem errors |

### Boundary Values — SessionData Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_SessionDataWithNamespacedKeys` | `unit` | Handles SessionData with namespaced keys. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `SessionData` containing key `"NodeA.ClaudeSessionID": "session-123"` | Returns nil; file contains key with correct value; read returns identical structure |
| `TestSessionMetadataStore_SessionDataNonStringValue` | `unit` | Serializes SessionData with non-string values as-is. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `SessionData` containing `"NodeA.ClaudeSessionID": 12345` (number) | Returns nil; number serialized; no type validation by store |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_FileAccessorErrorPropagated` | `unit` | Propagates FileAccessor callback error. | Temporary test directory created programmatically within test fixture; mock FileAccessor that returns callback error | SessionMetadata | Returns error containing callback error details; error wrapped with context |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_LocksReleasedOnPanic` | `unit` | Locks released if panic occurs during operation. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture; mock that panics during write | SessionMetadata | Panic propagates; file lock released (verified by subsequent successful operation) |

### Happy Path — No Caching

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_NoCaching` | `unit` | Read always accesses disk, no in-memory cache. | Temporary test directory created programmatically within test fixture; session directory and `session.json` created inside test fixture | Read metadata; externally modify file with new Status; read again | First read returns original Status; second read returns modified Status (detects external change) |

### Happy Path — State Transition (No Validation)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_NoStateValidation` | `unit` | Store does not validate state transitions. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | Write metadata with `Status="running"`; write again with `Status="initializing"` (invalid transition) | Both writes succeed; store does not enforce state machine logic |

### Happy Path — Write Truncates File

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_WriteTruncates` | `unit` | Write replaces all previous content. | Temporary test directory created programmatically within test fixture; session directory and large `session.json` (5 KB) created inside test fixture | SessionMetadata with minimal SessionData (resulting in 1 KB file) | File truncated to new size; no remnants of old content |

### Boundary Values — Edge Case JSON Values

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_NullValuesInSessionData` | `unit` | Handles null values in SessionData. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `SessionData` containing null value for a key | Null value serialized as `null` in JSON; read returns nil for that key |
| `TestSessionMetadataStore_EmptyStringsInFields` | `unit` | Handles empty string values. | Temporary test directory created programmatically within test fixture; session directory created inside test fixture | SessionMetadata with `WorkflowName=""` | Empty string serialized as `""`; read returns empty string |
