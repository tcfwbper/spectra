# Test Specification: `session_finalizer_test.go`

## Source File Under Test

`runtime/session_finalizer.go`

## Test File

`runtime/session_finalizer_test.go`

---

## `SessionFinalizer`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionFinalizer_ValidLogger` | `unit` | Constructs SessionFinalizer with a valid Logger. | Create a mock Logger. | `NewSessionFinalizer(logger)` | Returns non-nil `*SessionFinalizer`; no panic |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionFinalizer_NilLogger` | `unit` | Panics or returns error when logger is nil. | No dependencies. | `NewSessionFinalizer(nil)` | Panics with message indicating nil logger |

### Happy Path — Finalize

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_Finalize_Completed` | `unit` | Logs info and returns exit code 0 for a completed session. | Mock PersistentSession: `GetStatusSafe()` returns `"completed"`, `ID` returns `"sess-1"`, `WorkflowName` returns `"wf-1"`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `0`; `Logger.Info` called with `"session completed"`, `"sessionID"`, `"sess-1"`, `"workflow"`, `"wf-1"` |
| `TestSessionFinalizer_Finalize_FailedWithAgentError` | `unit` | Logs error with agent details and returns exit code 1 for AgentError. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*AgentError` with `Message()="agent broke"`, `AgentRole()="parser"`, `FailingState()="node_3"`, `Detail=map["key":"val"]`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called with `"session failed"`, includes `"agent"`, `"parser"`, `"state"`, `"node_3"`, `"detail"` with compact JSON `{"key":"val"}` |
| `TestSessionFinalizer_Finalize_FailedWithRuntimeError` | `unit` | Logs error with runtime details and returns exit code 1 for RuntimeError. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*RuntimeError` with `Message()="timeout"`, `Issuer()="SessionInitializer"`, `FailingState()="entry"`, `Detail=map["elapsed":"30s"]`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called with `"session failed"`, includes `"issuer"`, `"SessionInitializer"`, `"state"`, `"entry"`, `"detail"` with compact JSON `{"elapsed":"30s"}` |
| `TestSessionFinalizer_Finalize_FailedWithAgentError_EmptyDetail` | `unit` | Omits detail key-value when AgentError.Detail is empty. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*AgentError` with `Message()="err"`, `AgentRole()="runner"`, `FailingState()="s1"`, `Detail=map{}` (empty). Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called; log args do not contain `"detail"` key-value pair |
| `TestSessionFinalizer_Finalize_FailedWithAgentError_NilDetail` | `unit` | Omits detail key-value when AgentError.Detail is nil. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*AgentError` with `Message()="err"`, `AgentRole()="runner"`, `FailingState()="s1"`, `Detail=nil`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called; log args do not contain `"detail"` key-value pair |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_Finalize_NilSession` | `unit` | Logs error and returns exit code 1 for nil session. | Mock Logger. | `sf.Finalize(nil)` | Returns exit code `1`; `Logger.Error` called with `"SessionFinalizer called with nil session"` |
| `TestSessionFinalizer_Finalize_FailedWithNilError` | `unit` | Logs unknown error when session failed but error is nil. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `nil`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called with `"session failed"`, includes `"error"`, `"unknown error"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_Finalize_FailedWithUnexpectedErrorType` | `unit` | Falls back to error.Error() for unexpected error types. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns a plain `errors.New("something went wrong")`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called with `"session failed"`, includes `"error"`, `"something went wrong"` |
| `TestSessionFinalizer_Finalize_FailedWithDetailSerializationError` | `unit` | Logs fallback string when detail JSON serialization fails. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*AgentError` with Detail containing a non-serializable value (e.g., `math.Inf(1)` or a channel). Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called with `"detail"` value `"<failed to serialize detail>"` |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_Finalize_NonTerminalInitializing` | `unit` | Logs warning for non-terminal status "initializing". | Mock PersistentSession: `GetStatusSafe()` returns `"initializing"`, `GetErrorSafe()` returns `nil`, `ID` returns `"sess-2"`, `WorkflowName` returns `"wf-2"`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Warn` called with `"session terminated with non-terminal status"`, includes `"status"`, `"initializing"` |
| `TestSessionFinalizer_Finalize_NonTerminalRunning` | `unit` | Logs warning for non-terminal status "running". | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetErrorSafe()` returns `nil`, `ID` returns `"sess-3"`, `WorkflowName` returns `"wf-3"`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Warn` called with `"session terminated with non-terminal status"`, includes `"status"`, `"running"` |
| `TestSessionFinalizer_Finalize_NonTerminalWithError` | `unit` | Logs warning and error details for non-terminal status with error. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetErrorSafe()` returns `*AgentError` with `Message()="interrupted"`, `AgentRole()="worker"`, `FailingState()="node_5"`, `Detail=nil`. Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Warn` called with non-terminal status message; `Logger.Error` called with error details |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_Finalize_CalledTwice` | `unit` | Returns same exit code and logs same output on repeated calls. | Mock PersistentSession: `GetStatusSafe()` returns `"completed"`, `ID` returns `"sess-4"`, `WorkflowName` returns `"wf-4"`. Mock Logger that records calls. | Call `sf.Finalize(session)` twice. | Both calls return exit code `0`; Logger.Info called twice with identical arguments |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_Finalize_NopLogger` | `unit` | Returns correct exit code even when Logger silently drops messages. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*RuntimeError`. Use a no-op Logger. | `sf.Finalize(session)` | Returns exit code `1`; no panic |
| `TestSessionFinalizer_Finalize_RuntimeError_EmptyDetail` | `unit` | Omits detail key-value when RuntimeError.Detail is empty. | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`, `GetErrorSafe()` returns `*RuntimeError` with `Detail=map{}` (empty). Mock Logger. | `sf.Finalize(session)` | Returns exit code `1`; `Logger.Error` called; log args do not contain `"detail"` key-value pair |
