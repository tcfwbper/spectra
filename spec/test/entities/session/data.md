# Test Specification: `data_test.go`

## Source File Under Test
`entities/session/data.go`

## Test File
`entities/session/data_test.go`

---

## `UpdateSessionDataSafe`

### Happy Path — UpdateSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_StringValue` | `unit` | Stores a string value under a key. | Construct session via `NewSession` | `key="logicSpec.output"`, `value="hello"` | Returns `nil`; `GetSessionDataSafe("logicSpec.output")` returns `("hello", true)` |
| `TestUpdateSessionDataSafe_NilValue` | `unit` | Stores nil value for a non-ClaudeSessionID key. | Construct session | `key="someKey"`, `value=nil` | Returns `nil`; `GetSessionDataSafe("someKey")` returns `(nil, true)` |
| `TestUpdateSessionDataSafe_OverwriteExisting` | `unit` | Overwrites an existing key with a new value. | Construct session; call `UpdateSessionDataSafe("k", "v1")` | `key="k"`, `value="v2"` | Returns `nil`; `GetSessionDataSafe("k")` returns `("v2", true)` |
| `TestUpdateSessionDataSafe_ClaudeSessionID_ValidString` | `unit` | Accepts a string value for a ClaudeSessionID key. | Construct session | `key="nodeA.ClaudeSessionID"`, `value="sess-123"` | Returns `nil`; `GetSessionDataSafe("nodeA.ClaudeSessionID")` returns `("sess-123", true)` |
| `TestUpdateSessionDataSafe_ClaudeSessionID_EmptyString` | `unit` | Accepts empty string for ClaudeSessionID key. | Construct session | `key="nodeA.ClaudeSessionID"`, `value=""` | Returns `nil`; `GetSessionDataSafe("nodeA.ClaudeSessionID")` returns `("", true)` |
| `TestUpdateSessionDataSafe_UpdatesUpdatedAt` | `unit` | Successful write advances UpdatedAt. | Construct session; record initial `UpdatedAt` | `key="k"`, `value="v"` | Returns `nil`; `GetMetadataSnapshotSafe().UpdatedAt >= initial UpdatedAt` |

### Validation Failures — key

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_EmptyKey` | `unit` | Rejects empty key without mutating state. | Construct session | `key=""`, `value="x"` | Returns error with message `"session data key cannot be empty"` |

### Validation Failures — ClaudeSessionID type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_ClaudeSessionID_IntValue` | `unit` | Rejects non-string value for ClaudeSessionID key. | Construct session | `key="nodeA.ClaudeSessionID"`, `value=42` | Returns error with message `"ClaudeSessionID value must be a string, got int"` |
| `TestUpdateSessionDataSafe_ClaudeSessionID_NilValue` | `unit` | Rejects nil value for ClaudeSessionID key. | Construct session | `key="nodeA.ClaudeSessionID"`, `value=nil` | Returns error containing `"ClaudeSessionID value must be a string"` |
| `TestUpdateSessionDataSafe_ClaudeSessionID_Stringer` | `unit` | Rejects fmt.Stringer that is not dynamic type string. | Construct session | `key="nodeA.ClaudeSessionID"`, `value=bytes.NewBufferString("x")` | Returns error containing `"ClaudeSessionID value must be a string"` |

---

## `GetSessionDataSafe`

### Happy Path — GetSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDataSafe_ExistingKey` | `unit` | Returns stored value and true for an existing key. | Construct session; `UpdateSessionDataSafe("k", "v")` | `key="k"` | Returns `("v", true)` |
| `TestGetSessionDataSafe_MissingKey` | `unit` | Returns nil and false for a key that does not exist. | Construct session | `key="nonexistent"` | Returns `(nil, false)` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDataSafe_EmptyKey` | `unit` | Returns (nil, false) for empty key via map semantics. | Construct session | `key=""` | Returns `(nil, false)` |

