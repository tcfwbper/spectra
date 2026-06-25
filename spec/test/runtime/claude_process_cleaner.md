# Test Specification: `claude_process_cleaner_test.go`

## Source File Under Test

`runtime/claude_process_cleaner.go`

## Test File

`runtime/claude_process_cleaner_test.go`

---

## `ClaudeProcessCleaner`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewClaudeProcessCleaner_ValidDeps` | `unit` | Constructs ClaudeProcessCleaner with valid PersistentSession and Logger. | Create mock PersistentSession and mock Logger. | `NewClaudeProcessCleaner(mockSession, mockLogger)` | Returns non-nil `*ClaudeProcessCleaner`; no panic |

### Happy Path — Clean

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClean_TerminatesRunningClaudeProcess` | `unit` | Sends SIGTERM to a running claude process discovered from session data. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command contains "claude". Stub signal sender: SIGTERM succeeds. Stub process waiter: process exits within 2 seconds. | `cleaner.Clean()` | SIGTERM sent to PID 1234; Logger.Info called with `"sent SIGTERM to 1 claude process(es)"`; Logger.Info called with `"claude process cleanup complete"` |
| `TestClean_MultipleClaudeProcesses` | `unit` | Terminates multiple running claude processes from different nodes. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1111`, `"NodeB.PID": 2222`. Stub process inspector: both PIDs running, commands contain "claude". Stub signal sender: SIGTERM succeeds for both. Stub process waiter: both exit within 2 seconds. | `cleaner.Clean()` | SIGTERM sent to both PIDs; Logger.Info logged with count 2 |
| `TestClean_NoPIDKeys` | `unit` | Returns immediately when no .PID keys exist in session data. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"logicSpec.output": "data"`, `"NodeA.ClaudeSessionID": "uuid"`. | `cleaner.Clean()` | No signals sent; no error; returns immediately |
| `TestClean_AllPIDsAlreadyExited` | `unit` | Logs info and returns when all PIDs have already exited. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is not running. | `cleaner.Clean()` | No signals sent; Logger.Info called with `"no active claude processes to terminate"` |
| `TestClean_DeduplicatesSamePID` | `unit` | Sends SIGTERM only once when same PID appears under multiple keys. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 5555`, `"NodeB.PID": 5555`. Stub process inspector: PID 5555 is running, command contains "claude". Stub signal sender: SIGTERM succeeds. Stub process waiter: process exits within 2 seconds. | `cleaner.Clean()` | SIGTERM sent to PID 5555 exactly once |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClean_EscalatesToSIGKILL` | `unit` | Sends SIGKILL after 2-second SIGTERM wait for surviving process. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command contains "claude". Stub signal sender: SIGTERM succeeds. Use fake timer/clock seam: process does not exit within 2 seconds. Stub signal sender: SIGKILL succeeds. | `cleaner.Clean()` | SIGTERM sent first; after 2-second simulated wait, SIGKILL sent to PID 1234; Logger.Warn called with `"escalating to SIGKILL for PID 1234"` |
| `TestClean_ProcessExitsBeforeSIGKILL` | `unit` | Does not send SIGKILL when process exits during the 2-second wait. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command contains "claude". Stub signal sender: SIGTERM succeeds. Use fake timer/clock seam: process exits after 500ms (simulated). | `cleaner.Clean()` | SIGTERM sent; SIGKILL not sent; Logger.Info with cleanup complete |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClean_SIGTERMFailsPermissionDenied` | `unit` | Logs warning and skips SIGKILL when SIGTERM delivery fails. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command contains "claude". Stub signal sender: SIGTERM returns permission denied error. | `cleaner.Clean()` | Logger.Warn called about SIGTERM failure; SIGKILL not attempted for PID 1234; Clean() does not panic or return error |
| `TestClean_SIGKILLFailsPermissionDenied` | `unit` | Logs warning and continues when SIGKILL fails. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command contains "claude". Stub signal sender: SIGTERM succeeds. Use fake timer/clock seam: process survives 2 seconds. Stub signal sender: SIGKILL returns permission denied error. | `cleaner.Clean()` | Logger.Warn called with `"failed to kill claude process"` containing PID 1234; Clean() does not panic or return error |
| `TestClean_NonIntegerPIDValue` | `unit` | Logs warning and skips non-integer PID values. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": "not-an-int"`. | `cleaner.Clean()` | Logger.Warn called with `"skipping non-integer PID value"` containing key and type; no signals sent |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClean_SkipsNonClaudeProcess` | `unit` | Does not signal a PID that belongs to a non-claude process. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command is "python my_script.py". | `cleaner.Clean()` | Logger.Warn called with `"PID does not belong to a claude process, skipping"` containing PID 1234 and observed command; no signals sent; Logger.Info called with `"no active claude processes to terminate"` |
| `TestClean_CommandMatchCaseInsensitive` | `unit` | Matches process command containing "Claude" (mixed case). | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is running, command is "/usr/local/bin/Claude". Stub signal sender: SIGTERM succeeds. Stub process waiter: process exits within 2 seconds. | `cleaner.Clean()` | SIGTERM sent to PID 1234 (command match is case-insensitive) |
| `TestClean_ReadOnlySessionAccess` | `unit` | Only calls GetMetadataSnapshotSafe on PersistentSession, no mutation methods. | Mock PersistentSession with call tracking for all methods. Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is not running. | `cleaner.Clean()` | PersistentSession.GetMetadataSnapshotSafe called once; no mutation methods called (UpdateSessionDataSafe, Fail, Run, Done never called) |
| `TestClean_OnlyExtractsPIDSuffixKeys` | `unit` | Ignores keys without .PID suffix even if values are integers. | Mock PersistentSession.GetMetadataSnapshotSafe() returns SessionData with `"NodeA.ClaudeSessionID": "uuid-val"`, `"logicSpec.count": 42`, `"NodeA.PID": 1234`. Stub process inspector: PID 1234 is not running. | `cleaner.Clean()` | Only PID 1234 is inspected; value 42 under `"logicSpec.count"` is ignored |
