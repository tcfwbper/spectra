# Test Specification: `session_metadata_store_test.go`

## Source File Under Test
`storage/session_metadata_store.go`

## Test File
`storage/session_metadata_store_test.go`

---

## `SessionMetadataStore`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionMetadataStore_ValidInputs` | `unit` | Constructs a SessionMetadataStore with valid projectRoot and sessionUUID. | Stub StorageLayout to return a known path. | `projectRoot="/tmp/project"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns non-nil `*SessionMetadataStore`; no error |
| `TestNewSessionMetadataStore_NoFileSystemAccess` | `unit` | Constructor does not touch the filesystem. | Provide a non-existent projectRoot. Stub StorageLayout. | `projectRoot="/nonexistent"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns non-nil `*SessionMetadataStore`; no panic; no file created |

### Happy Path — Write

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_Write_ValidMetadata` | `unit` | Writes a valid SessionMetadata snapshot to session.json. | Create a temp directory simulating the session directory. Stub FileAccessor to invoke the preparation callback. Construct a valid SessionMetadata snapshot with all fields populated (Error=nil). | Valid SessionMetadata snapshot | Returns nil error; file contains pretty-printed JSON |
| `TestSessionMetadataStore_Write_PrettyPrintedJSON` | `unit` | Written JSON is pretty-printed with 2-space indentation. | Create session directory fixture. Stub FileAccessor. Construct valid metadata. | Valid SessionMetadata | Returns nil; file content contains newlines and 2-space indentation |
| `TestSessionMetadataStore_Write_OmitsErrorWhenNil` | `unit` | The "error" key is omitted from JSON when Error field is nil. | Create session directory fixture. Stub FileAccessor. Construct metadata with Error=nil. | Metadata with nil Error | Returns nil; parsing file JSON confirms no `"error"` key present |
| `TestSessionMetadataStore_Write_WithAgentError` | `unit` | Serializes Error field correctly when set to an AgentError. | Create session directory fixture. Stub FileAccessor. Construct metadata with Error set to a valid `*AgentError` (agentRole="Reviewer"). | Metadata with AgentError | Returns nil; JSON `"error"` object contains `"agentRole":"Reviewer"` and no `"errorType"` discriminator |
| `TestSessionMetadataStore_Write_WithRuntimeError` | `unit` | Serializes Error field correctly when set to a RuntimeError. | Create session directory fixture. Stub FileAccessor. Construct metadata with Error set to a valid `*RuntimeError` (issuer="system"). | Metadata with RuntimeError | Returns nil; JSON `"error"` object contains `"issuer":"system"` and no `"errorType"` discriminator |
| `TestSessionMetadataStore_Write_AgentErrorEmptyRole` | `unit` | The "agentRole" key is always written for AgentError even when the value is empty string. | Create session directory fixture. Stub FileAccessor. Construct metadata with AgentError where agentRole="". | Metadata with AgentError (empty role) | Returns nil; JSON `"error"` object contains `"agentRole":""` |
| `TestSessionMetadataStore_Write_PidAlwaysSerialized` | `unit` | The "pid" field is always present in written JSON (no omitempty). | Create session directory fixture. Stub FileAccessor. Construct metadata with `Pid=1234`. | Valid SessionMetadata with `Pid=1234` | Returns nil; JSON contains `"pid":1234` |
| `TestSessionMetadataStore_Write_UpdatedAtPassThrough` | `unit` | UpdatedAt is serialized as-is from the snapshot without modification. | Create session directory fixture. Stub FileAccessor. Construct metadata with a specific UpdatedAt value (e.g., 1700000000). | Metadata with `updatedAt=1700000000` | Returns nil; JSON contains `"updatedAt":1700000000` (exact value) |
| `TestSessionMetadataStore_Write_ExcludesEventHistory` | `unit` | EventHistory is not included in the serialized JSON. | Create session directory fixture. Stub FileAccessor. Construct metadata (EventHistory is not part of the write input). | Valid SessionMetadata | Returns nil; parsing file JSON confirms no `"eventHistory"` key |
| `TestSessionMetadataStore_Write_TruncatesExistingFile` | `unit` | Writing overwrites any previous file content entirely. | Create session directory fixture with an existing session.json containing old data. Stub FileAccessor. Construct new metadata. | New SessionMetadata | Returns nil; file contains only the new metadata JSON |

### Happy Path — Read

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_Read_ValidFile` | `unit` | Reads and parses a valid session.json file. | Programmatically create a temp session.json with valid pretty-printed JSON containing all fields (Error=nil). Construct SessionMetadataStore pointing to this file. | | Returns SessionMetadata with all fields matching file content; no error |
| `TestSessionMetadataStore_Read_WithAgentError` | `unit` | Reconstructs AgentError from JSON containing "agentRole" field. | Create a temp session.json with an `"error"` object containing `"agentRole"` and valid fields. | | Returns SessionMetadata with Error field set to a reconstructed `*AgentError`; AgentRole getter returns the expected value |
| `TestSessionMetadataStore_Read_WithRuntimeError` | `unit` | Reconstructs RuntimeError from JSON containing "issuer" field. | Create a temp session.json with an `"error"` object containing `"issuer"` and valid fields. | | Returns SessionMetadata with Error field set to a reconstructed `*RuntimeError`; Issuer getter returns the expected value |
| `TestSessionMetadataStore_Read_IgnoresEventHistoryField` | `unit` | Ignores an "eventHistory" field if present in the JSON file. | Create a temp session.json that includes an extra `"eventHistory"` key with array value. | | Returns SessionMetadata without EventHistory; no error |
| `TestSessionMetadataStore_Read_LegacyFileMissingPid` | `unit` | Defaults Pid to 0 when reading a legacy session.json that lacks the "pid" field. | Create a temp session.json with valid JSON containing all required fields except `"pid"`. | | Returns SessionMetadata with `Pid == 0`; no error |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_Write_SessionDirNotExists` | `unit` | Returns error when session directory does not exist. | Create a temp directory without the session subdirectory. Stub FileAccessor so the preparation callback detects missing directory. | Valid SessionMetadata | Returns error containing `"session directory does not exist:"` |
| `TestSessionMetadataStore_Write_FileAccessorError` | `unit` | Propagates error when FileAccessor preparation fails. | Stub FileAccessor to return an error from the preparation callback. | Valid SessionMetadata | Returns error containing `"failed to prepare file"` |
| `TestSessionMetadataStore_Write_ExceedsMaxPayloadSize` | `unit` | Returns size limit error when serialized metadata exceeds MaxPayloadSize. | Create session directory fixture. Stub FileAccessor. Construct metadata with very large SessionData (> 10 MB). | Oversized SessionMetadata | Returns error containing `"session metadata size exceeds limit:"` and `"bytes (max"` |
| `TestSessionMetadataStore_Write_ExceedsMaxPayloadSize_NoWrite` | `unit` | File is not modified when metadata exceeds size limit. | Create session directory fixture with an existing session.json containing old data. Stub FileAccessor. Construct oversized metadata. | Oversized SessionMetadata | Returns error; file content remains unchanged |
| `TestSessionMetadataStore_Write_UnserializableSessionData` | `unit` | Returns serialization error when SessionData contains non-JSON-serializable values. | Create session directory fixture. Stub FileAccessor. Construct metadata with SessionData containing a Go channel value. | Metadata with `sessionData=map[string]any{"ch": make(chan int)}` | Returns error containing `"failed to serialize session metadata:"` and `"unsupported type"` |
| `TestSessionMetadataStore_Read_FileNotExists` | `unit` | Returns error when session.json does not exist. | Create a temp directory without session.json. Construct SessionMetadataStore. | | Returns error containing `"session metadata file does not exist:"` |
| `TestSessionMetadataStore_Read_InvalidJSON` | `unit` | Returns parse error when file contains invalid JSON. | Create a temp session.json with malformed JSON (missing closing brace). | | Returns error containing `"failed to parse session metadata:"` |
| `TestSessionMetadataStore_Read_MissingRequiredField` | `unit` | Returns parse error when file is missing a required field. | Create a temp session.json with valid JSON but missing the `"id"` field. | | Returns error containing `"failed to parse session metadata:"` and `"missing required field"` |
| `TestSessionMetadataStore_Read_ErrorBothFields` | `unit` | Returns reconstruction error when error object contains both "agentRole" and "issuer". | Create a temp session.json with an `"error"` object containing both `"agentRole"` and `"issuer"` keys. | | Returns error containing `"failed to reconstruct error: ambiguous error object contains both 'agentRole' and 'issuer'"` |
| `TestSessionMetadataStore_Read_ErrorNeitherField` | `unit` | Returns reconstruction error when error object contains neither "agentRole" nor "issuer". | Create a temp session.json with an `"error"` object containing neither key. | | Returns error containing `"failed to reconstruct error: cannot determine error type"` |
| `TestSessionMetadataStore_Read_ErrorInvalidConstructorFields` | `unit` | Returns reconstruction error when error type is determined but constructor fields are invalid. | Create a temp session.json with an `"error"` object containing `"agentRole"` but with invalid fields (e.g., empty message) that cause `NewAgentError` to fail. | | Returns error containing `"failed to reconstruct error:"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_Write_CallsFileAccessor` | `unit` | FileAccessor is called with the correct file path and a non-nil preparation callback. | Stub FileAccessor to record the call arguments. Create session directory fixture. | Valid SessionMetadata | FileAccessor was called exactly once with the session.json path |
| `TestSessionMetadataStore_Write_ReadsErrorViaGetters` | `unit` | Serialization reads AgentError fields via getter methods, not struct access. | Create session directory fixture. Stub FileAccessor. Construct metadata with a valid AgentError. | Metadata with AgentError | Written JSON `"error"` object field values match getter return values |
| `TestSessionMetadataStore_Read_ReconstructsViaConstructor` | `unit` | AgentError is reconstructed via NewAgentError, not struct literal. | Create a temp session.json with a valid AgentError `"error"` object. | | Returned metadata Error field is a valid `*AgentError` whose getters return expected values |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_Read_IdempotentReads` | `unit` | Two consecutive reads return the same metadata. | Create a temp session.json with valid content. Construct SessionMetadataStore. | | First and second read return identical SessionMetadata values |
| `TestSessionMetadataStore_Write_IdempotentWrites` | `unit` | Writing the same metadata twice results in identical file content. | Create session directory fixture. Stub FileAccessor. Construct valid metadata. | Same metadata written twice | File content after first write equals file content after second write |

### Boundary Values — MaxPayloadSize

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionMetadataStore_Write_ExactlyAtMaxPayloadSize` | `unit` | Metadata whose serialized size equals exactly MaxPayloadSize is accepted. | Create session directory fixture. Stub FileAccessor. Construct metadata whose pretty-printed JSON is exactly MaxPayloadSize bytes. | Metadata at size boundary | Returns nil error; metadata is written to file |
| `TestSessionMetadataStore_Write_OneByteOverMaxPayloadSize` | `unit` | Metadata whose serialized size is MaxPayloadSize + 1 is rejected. | Create session directory fixture. Stub FileAccessor. Construct metadata whose pretty-printed JSON is MaxPayloadSize + 1 bytes. | Metadata one byte over limit | Returns error containing `"session metadata size exceeds limit:"` |
