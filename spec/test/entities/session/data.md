# Test Specification: `data.go`

## Source File Under Test
`entities/session/data.go`

## Test File
`entities/session/data_test.go`

---

**Fixture Isolation**: All tests create Session instances in memory using programmatic construction. No external files or directories are required unless explicitly stated in the Setup column. Mock dependencies (SessionMetadataStore, loggers, etc.) are created within each test.

---

## `Session` Data Methods

### Happy Path — UpdateSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_NewKey` | `unit` | Stores new key-value pair in SessionData. | Session with empty `SessionData`, `UpdatedAt=T0` | Call `UpdateSessionDataSafe("key1", "value1")` | Returns `nil`; `SessionData["key1"]="value1"`; `UpdatedAt > T0` |
| `TestUpdateSessionDataSafe_OverwriteExisting` | `unit` | Overwrites existing key with new value. | Session with `SessionData={"key1": "old"}`, `UpdatedAt=T0` | Call `UpdateSessionDataSafe("key1", "new")` | Returns `nil`; `SessionData["key1"]="new"`; `UpdatedAt > T0` |
| `TestUpdateSessionDataSafe_MultipleKeys` | `unit` | Stores multiple distinct keys. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("key1", "val1")`, then `UpdateSessionDataSafe("key2", "val2")` | Both return `nil`; `SessionData={"key1": "val1", "key2": "val2"}` |
| `TestUpdateSessionDataSafe_NilValueForNonClaudeSessionID` | `unit` | Accepts nil value for keys not matching ClaudeSessionID pattern. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("logicSpec.result", nil)` | Returns `nil`; `SessionData["logicSpec.result"]=nil`; GetSessionDataSafe returns `(nil, true)` |
| `TestUpdateSessionDataSafe_PersistsToStore` | `unit` | Persists updated SessionData to SessionMetadataStore. | Mock SessionMetadataStore; session with empty `SessionData` | Call `UpdateSessionDataSafe("key", "value")` | Returns `nil`; store write called with updated session |

### Happy Path — GetSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDataSafe_ExistingKey` | `unit` | Returns value and true for existing key. | Session with `SessionData={"key1": "value1"}` | Call `GetSessionDataSafe("key1")` | Returns `("value1", true)` |
| `TestGetSessionDataSafe_MissingKey` | `unit` | Returns nil and false for missing key. | Session with `SessionData={"key1": "value1"}` | Call `GetSessionDataSafe("key2")` | Returns `(nil, false)` |
| `TestGetSessionDataSafe_NilValue` | `unit` | Distinguishes nil value from missing key. | Session with `SessionData={"key1": nil}` | Call `GetSessionDataSafe("key1")` | Returns `(nil, true)` (key exists with nil value) |
| `TestGetSessionDataSafe_EmptyKey` | `unit` | Returns nil and false for empty key. | Session with non-empty `SessionData` | Call `GetSessionDataSafe("")` | Returns `(nil, false)` |
| `TestGetSessionDataSafe_DoesNotModifyUpdatedAt` | `unit` | GetSessionDataSafe is a pure read; does not refresh UpdatedAt. | Session with `SessionData={"key": "val"}`, `UpdatedAt=T0` | Call `GetSessionDataSafe("key")` | Returns `("val", true)`; `UpdatedAt` remains `T0` |

### Happy Path — ClaudeSessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_ClaudeSessionIDString` | `unit` | Accepts string value for ClaudeSessionID key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("ProcessNode.ClaudeSessionID", "session-abc-123")` | Returns `nil`; `SessionData["ProcessNode.ClaudeSessionID"]="session-abc-123"` |
| `TestUpdateSessionDataSafe_ClaudeSessionIDEmptyString` | `unit` | Accepts empty string for ClaudeSessionID key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("ReviewNode.ClaudeSessionID", "")` | Returns `nil`; `SessionData["ReviewNode.ClaudeSessionID"]=""` |
| `TestUpdateSessionDataSafe_ClaudeSessionIDOverwrite` | `unit` | Overwrites existing ClaudeSessionID value. | Session with `SessionData={"Node1.ClaudeSessionID": "old-session"}` | Call `UpdateSessionDataSafe("Node1.ClaudeSessionID", "new-session")` | Returns `nil`; `SessionData["Node1.ClaudeSessionID"]="new-session"` |
| `TestGetSessionDataSafe_ClaudeSessionID` | `unit` | Retrieves ClaudeSessionID value. | Session with `SessionData={"AgentNode.ClaudeSessionID": "sess-456"}` | Call `GetSessionDataSafe("AgentNode.ClaudeSessionID")` | Returns `("sess-456", true)` |

### Happy Path — Namespace Conventions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_LogicSpecNamespace` | `unit` | Accepts keys with logicSpec. prefix. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("logicSpec.output", map[string]any{"result": "done"})` | Returns `nil`; `SessionData["logicSpec.output"]` stored |
| `TestUpdateSessionDataSafe_ArbitraryNamespace` | `unit` | Accepts keys without recognized namespace. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("custom.key", 123)` | Returns `nil`; `SessionData["custom.key"]=123` |

### Validation Failures — UpdateSessionDataSafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_EmptyKey` | `unit` | Rejects empty key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("", "value")` | Returns error matching `/session data key cannot be empty/i`; `SessionData` unchanged |
| `TestUpdateSessionDataSafe_ClaudeSessionIDNonString` | `unit` | Rejects non-string value for ClaudeSessionID key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("Node.ClaudeSessionID", 123)` | Returns error matching `/ClaudeSessionID value must be a string.*got.*int/i`; `SessionData` unchanged |
| `TestUpdateSessionDataSafe_ClaudeSessionIDNil` | `unit` | Rejects nil value for ClaudeSessionID key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("AgentNode.ClaudeSessionID", nil)` | Returns error matching `/ClaudeSessionID value must be a string/i`; `SessionData` unchanged |
| `TestUpdateSessionDataSafe_ClaudeSessionIDMap` | `unit` | Rejects map value for ClaudeSessionID key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("ProcessNode.ClaudeSessionID", map[string]any{"id": "sess"})` | Returns error matching `/ClaudeSessionID value must be a string.*got.*map/i`; `SessionData` unchanged |
| `TestUpdateSessionDataSafe_ClaudeSessionIDSlice` | `unit` | Rejects slice value for ClaudeSessionID key. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("Node.ClaudeSessionID", []string{"sess1", "sess2"})` | Returns error matching `/ClaudeSessionID value must be a string.*got.*slice/i`; `SessionData` unchanged |

### Validation Failures — Type Validation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_ClaudeSessionIDInterface` | `unit` | Rejects interface value that is not string for ClaudeSessionID. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("Node.ClaudeSessionID", interface{}(123))` | Returns error matching `/ClaudeSessionID value must be a string/i`; `SessionData` unchanged |
| `TestUpdateSessionDataSafe_ClaudeSessionIDStringer` | `unit` | Rejects fmt.Stringer implementation for ClaudeSessionID (requires actual string type). | Session with empty `SessionData`; custom type implementing String() method | Call `UpdateSessionDataSafe("Node.ClaudeSessionID", customStringer)` | Returns error matching `/ClaudeSessionID value must be a string/i`; `SessionData` unchanged |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_Idempotent` | `unit` | Repeated writes with same value are accepted. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("key", "val")` twice | Both return `nil`; `SessionData["key"]="val"` |
| `TestGetSessionDataSafe_Idempotent` | `unit` | Repeated reads return same result. | Session with `SessionData={"key": "val"}` | Call `GetSessionDataSafe("key")` multiple times | All calls return `("val", true)`; no mutations |

### Atomic Replacement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_AtomicUpdate` | `unit` | SessionData and UpdatedAt updated atomically. | Session with `SessionData={"key": "old"}`, `UpdatedAt=T0`; concurrent goroutine reading `GetSessionDataSafe("key")` | Call `UpdateSessionDataSafe("key", "new")` | Concurrent reader never observes intermediate state; both value and timestamp updated together |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionData_ConcurrentWrites` | `race` | Concurrent writes to different keys are serialized. | Session with empty `SessionData` | 10 goroutines call `UpdateSessionDataSafe()` with different keys simultaneously | All succeed; all keys present in `SessionData`; no race conditions |
| `TestSessionData_ConcurrentWritesSameKey` | `race` | Concurrent writes to same key are serialized; last write wins. | Session with `SessionData={"key": "initial"}` | 10 goroutines call `UpdateSessionDataSafe("key", <unique-value>)` simultaneously | All calls return `nil`; final value is one of the written values; no race conditions |
| `TestSessionData_ConcurrentReads` | `race` | Multiple concurrent reads succeed. | Session with `SessionData={"key": "value"}` | 100 goroutines call `GetSessionDataSafe("key")` simultaneously | All return `("value", true)`; no race conditions |
| `TestSessionData_ConcurrentReadWrite` | `race` | Concurrent reads and writes are serialized correctly. | Session with `SessionData={"key": "old"}` | 50 goroutines call `GetSessionDataSafe("key")`; 1 goroutine calls `UpdateSessionDataSafe("key", "new")` | Readers see either "old" or "new" consistently; no race conditions; no partial state |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_ReleasesLock` | `unit` | Write lock released after update. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("key", "val")`; then call `GetSessionDataSafe("key")` | Both succeed without deadlock |
| `TestGetSessionDataSafe_ReleasesLock` | `unit` | Read lock released after read. | Session with `SessionData={"key": "val"}` | Call `GetSessionDataSafe("key")` multiple times | All succeed without deadlock |
| `TestSessionData_ValidationBeforeLock` | `unit` | Validation occurs before lock acquisition. | Session with empty `SessionData`; concurrent goroutine holding write lock | Call `UpdateSessionDataSafe("", "value")` (empty key) | Returns error immediately without blocking on lock |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_PersistenceFailureLogged` | `unit` | Persistence failures logged but not returned. | Mock SessionMetadataStore that returns error on write; session with empty `SessionData`; mock logger | Call `UpdateSessionDataSafe("key", "value")` | Returns `nil`; in-memory `SessionData["key"]="value"`; warning logged matching `/persistence failed/i` |
| `TestUpdateSessionDataSafe_PersistenceFailureDoesNotRevert` | `unit` | In-memory state authoritative even when persistence fails. | Mock SessionMetadataStore that fails; session with empty `SessionData` | Call `UpdateSessionDataSafe("key", "value")` | Returns `nil`; `SessionData["key"]="value"` persists in memory |

### Invariants — UpdatedAt Refresh

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_RefreshesUpdatedAt` | `unit` | UpdateSessionDataSafe refreshes UpdatedAt. | Session with empty `SessionData`, `UpdatedAt=T0` | Wait 1 second; call `UpdateSessionDataSafe("key", "value")` | Returns `nil`; `UpdatedAt > T0` |
| `TestUpdateSessionDataSafe_UpdatedAtInSameCriticalSection` | `unit` | UpdatedAt refreshed in same critical section as data write. | Session with empty `SessionData`, `UpdatedAt=T0`; concurrent goroutine reading `GetSessionDataSafe()` in loop | Call `UpdateSessionDataSafe("key", "value")` | Concurrent reader observes consistent snapshot; never sees new key with old `UpdatedAt` |

### Invariants — Memory Authority

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionData_InMemoryAuthoritative` | `unit` | In-memory SessionData is source of truth. | Mock SessionMetadataStore that fails; session with empty `SessionData` | Call `UpdateSessionDataSafe("key", "value")`; then `GetSessionDataSafe("key")` | Update returns `nil`; get returns `("value", true)` from in-memory map |

### Invariants — ClaudeSessionID Validation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionData_ClaudeSessionIDSuffixCaseSensitive` | `unit` | ClaudeSessionID suffix match is case-sensitive. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("Node.claudesessionid", 123)` (lowercase suffix) | Returns `nil`; value accepted (suffix does not match `.ClaudeSessionID` exactly) |
| `TestSessionData_ClaudeSessionIDNoNodeNameValidation` | `unit` | Node name prefix in ClaudeSessionID key is not validated. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("NonExistentNode.ClaudeSessionID", "session-id")` | Returns `nil`; value stored (no workflow validation at this layer) |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_LargeValue` | `unit` | Accepts large value in SessionData. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("key", <10MB byte slice>)` | Returns `nil`; value stored |
| `TestUpdateSessionDataSafe_ComplexNestedValue` | `unit` | Accepts deeply nested map value. | Session with empty `SessionData` | Call `UpdateSessionDataSafe("key", map[string]any{"a": map[string]any{"b": map[string]any{"c": "deep"}}})` | Returns `nil`; nested structure stored |
| `TestGetSessionDataSafe_TypeAssertion` | `unit` | Caller responsible for type assertion of returned value. | Session with `SessionData={"key": 123}` | Call `GetSessionDataSafe("key")`; type-assert to int | Returns `(123, true)`; caller can assert to `int` successfully |
| `TestSessionData_KeyWithDotNotation` | `unit` | Keys with dot notation treated as literal strings (no path traversal). | Session with empty `SessionData` | Call `UpdateSessionDataSafe("a.b.c", "value")` | Returns `nil`; `SessionData["a.b.c"]="value"` (single key, not nested) |
| `TestUpdateSessionDataSafe_OverwriteDifferentType` | `unit` | Overwrites value with different type. | Session with `SessionData={"key": "string"}` | Call `UpdateSessionDataSafe("key", 123)` | Returns `nil`; `SessionData["key"]=123` (int replaces string) |
