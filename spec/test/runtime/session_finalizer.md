# Test Specification: `session_finalizer.go`

## Source File Under Test
`runtime/session_finalizer.go`

## Test File
`runtime/session_finalizer_test.go`

---

## `SessionFinalizer`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionFinalizer_New` | `unit` | Constructs SessionFinalizer with valid dependencies. | Test fixture; mock RuntimeSocketManager | `RuntimeSocketManager=<mock>` | Returns SessionFinalizer instance; no error |

### Happy Path — Finalize (Completed Session)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_CompletedSession_StdoutOutput` | `unit` | Prints success message to stdout for completed session. | Test fixture; mock Session with `ID="abc-123"`, `Status="completed"`, `WorkflowName="TestWorkflow"`; mock RuntimeSocketManager; capture stdout | `Session=<mock>` | RuntimeSocketManager.DeleteSocket() called; stdout contains: `"Session abc-123 completed successfully. Workflow: TestWorkflow"`; no stderr output; no error returned |
| `TestFinalize_CompletedSession_SocketDeleted` | `unit` | Deletes runtime socket for completed session. | Test fixture; mock Session with `Status="completed"`; mock RuntimeSocketManager tracks DeleteSocket() call | `Session=<mock>` | RuntimeSocketManager.DeleteSocket() called exactly once |

### Happy Path — Finalize (Failed Session with AgentError)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_FailedSession_AgentError_FullDetail` | `unit` | Prints error details to stderr for failed session with AgentError and non-empty detail. | Test fixture; mock Session with `ID="def-456"`, `Status="failed"`, `WorkflowName="ReviewWorkflow"`, `Error=AgentError{Message:"validation failed", AgentRole:"reviewer", FailingState:"review_node", Detail:{"code":400,"context":"missing field"}}`; capture stderr | `Session=<mock>` | Stderr contains: `"Session def-456 failed. Workflow: ReviewWorkflow"`, `"Error: validation failed"`, `"Agent: reviewer"`, `"State: review_node"`, `"Detail: {\"code\":400,\"context\":\"missing field\"}"` (compact JSON); no stdout output |
| `TestFinalize_FailedSession_AgentError_EmptyDetail` | `unit` | Omits Detail line when AgentError.Detail is empty. | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{Message:"agent error", AgentRole:"architect", FailingState:"design_node", Detail:{}}`; capture stderr | `Session=<mock>` | Stderr contains: `"Session <id> failed. Workflow: <name>"`, `"Error: agent error"`, `"Agent: architect"`, `"State: design_node"`; Detail line NOT present |
| `TestFinalize_FailedSession_AgentError_NullDetail` | `unit` | Treats null detail as empty object. | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{Message:"error", AgentRole:"reviewer", FailingState:"node1", Detail:nil}`; capture stderr | `Session=<mock>` | Stderr contains error message, agent, state; Detail line NOT present |

### Happy Path — Finalize (Failed Session with RuntimeError)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_FailedSession_RuntimeError_FullDetail` | `unit` | Prints error details to stderr for failed session with RuntimeError and non-empty detail. | Test fixture; mock Session with `ID="ghi-789"`, `Status="failed"`, `WorkflowName="TestWorkflow"`, `Error=RuntimeError{Message:"timeout exceeded", Issuer:"SessionInitializer", FailingState:"start", Detail:{"duration":"30s"}}`; capture stderr | `Session=<mock>` | Stderr contains: `"Session ghi-789 failed. Workflow: TestWorkflow"`, `"Error: timeout exceeded"`, `"Issuer: SessionInitializer"`, `"State: start"`, `"Detail: {\"duration\":\"30s\"}"` (compact JSON) |
| `TestFinalize_FailedSession_RuntimeError_EmptyDetail` | `unit` | Omits Detail line when RuntimeError.Detail is empty. | Test fixture; mock Session with `Status="failed"`, `Error=RuntimeError{Message:"runtime error", Issuer="MessageRouter", FailingState:"node2", Detail:{}}`; capture stderr | `Session=<mock>` | Stderr contains: `"Error: runtime error"`, `"Issuer: MessageRouter"`, `"State: node2"`; Detail line NOT present |

### Happy Path — Socket Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_SocketDeleteIdempotent` | `unit` | RuntimeSocketManager.DeleteSocket() called even if socket already deleted. | Test fixture; mock RuntimeSocketManager.DeleteSocket() returns nil (idempotent); mock Session with `Status="completed"` | `Session=<mock>` | RuntimeSocketManager.DeleteSocket() called; no error; success message printed to stdout |
| `TestFinalize_SocketDeleteWarning_ContinuesFinalization` | `unit` | Continues finalization when socket deletion logs warning. | Test fixture; mock RuntimeSocketManager.DeleteSocket() logs warning "socket file not found" but returns nil; mock Session with `Status="completed"`; capture stdout and logs | `Session=<mock>` | RuntimeSocketManager logs warning; SessionFinalizer continues; success message printed to stdout; no error returned |

### Validation Failures — Session Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_NonTerminalStatus_Initializing` | `unit` | Logs warning when Status is "initializing" but proceeds with finalization. | Test fixture; mock Session with `Status="initializing"`, `WorkflowName="TestWorkflow"`; capture logs and stdout | `Session=<mock>` | Log contains: `"SessionFinalizer called with non-terminal session status 'initializing'. This may indicate a programming error."`; RuntimeSocketManager.DeleteSocket() called; stdout contains: `"Session <id> initializing. Workflow: TestWorkflow"` (status as-is) |
| `TestFinalize_NonTerminalStatus_Running` | `unit` | Logs warning when Status is "running" but proceeds with finalization. | Test fixture; mock Session with `Status="running"`; capture logs | `Session=<mock>` | Log contains: `"SessionFinalizer called with non-terminal session status 'running'. This may indicate a programming error."`; finalization proceeds |

### Error Propagation — Nil Error

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_FailedSession_NilError` | `unit` | Handles nil Error gracefully for failed session. | Test fixture; mock Session with `ID="xyz-999"`, `Status="failed"`, `WorkflowName="TestWorkflow"`, `Error=nil`; capture stderr | `Session=<mock>` | Stderr contains: `"Session xyz-999 failed. Workflow: TestWorkflow"`, `"Error: <unknown error>"`; no Agent/Issuer/State/Detail lines |

### Error Propagation — Unknown Error Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_FailedSession_UnknownErrorType` | `unit` | Handles error that is neither AgentError nor RuntimeError. | Test fixture; mock Session with `Status="failed"`, `Error=errors.New("generic error")`; capture stderr | `Session=<mock>` | Stderr contains: `"Session <id> failed. Workflow: <name>"`, `"Error: generic error"` (error.Error() string); no Agent/Issuer/State/Detail lines |

### Error Propagation — Detail Serialization Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_DetailSerializationFails` | `unit` | Handles non-serializable detail gracefully. | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{Detail:map[string]any{"func":func(){}}}` (Go function, not JSON-serializable); capture stderr and logs | `Session=<mock>` | Stderr contains: `"Detail: <failed to serialize detail>"`; logs contain serialization error |

### Boundary Values — Large Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_VeryLargeErrorMessage` | `unit` | Handles very large error message (10 KB). | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{Message:<10KB-string>}`; capture stderr | `Session=<mock>` | Entire error message printed to stderr; no truncation |
| `TestFinalize_VeryLargeDetail` | `unit` | Handles very large detail structure (1 MB). | Test fixture; mock Session with `Status="failed"`, `Error=RuntimeError{Detail:map[string]any{"trace":<1MB-string>}}`; capture stderr | `Session=<mock>` | Entire detail JSON printed to stderr; no size limit enforced |

### Boundary Values — Special Characters

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_SessionIDWithSpecialChars` | `unit` | Handles session ID with special characters. | Test fixture; mock Session with `ID="abc\n123"` (contains newline), `Status="completed"`; capture stdout | `Session=<mock>` | Stdout contains: `"Session abc\n123 completed successfully"` (printed as-is, no escaping) |
| `TestFinalize_WorkflowNameWithSpecialChars` | `unit` | Handles workflow name with special characters. | Test fixture; mock Session with `WorkflowName="Test\tWorkflow"` (contains tab), `Status="completed"`; capture stdout | `Session=<mock>` | Stdout contains: `"Workflow: Test\tWorkflow"` (printed as-is) |
| `TestFinalize_ErrorMessageWithUnicode` | `unit` | Handles Unicode characters in error message. | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{Message:"错误: emoji 🚨"}`; capture stderr | `Session=<mock>` | Stderr contains: `"Error: 错误: emoji 🚨"` (Unicode preserved) |

### Boundary Values — Empty Fields

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_EmptySessionID` | `unit` | Handles empty session ID. | Test fixture; mock Session with `ID=""`, `Status="completed"`; capture stdout | `Session=<mock>` | Stdout contains: `"Session  completed successfully. Workflow: <name>"` (empty ID printed as-is) |
| `TestFinalize_EmptyWorkflowName` | `unit` | Handles empty workflow name. | Test fixture; mock Session with `WorkflowName=""`, `Status="completed"`; capture stdout | `Session=<mock>` | Stdout contains: `"Workflow: "` (empty name printed as-is) |
| `TestFinalize_EmptyAgentRole` | `unit` | Handles empty AgentRole in AgentError. | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{AgentRole:""}`; capture stderr | `Session=<mock>` | Stderr contains: `"Agent: "` (empty role printed as-is) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_CalledMultipleTimes` | `unit` | Subsequent calls produce same output without errors. | Test fixture; mock Session with `Status="completed"`; capture stdout; call Finalize 3 times | `Session=<mock>` (called 3 times) | Each call prints success message to stdout; RuntimeSocketManager.DeleteSocket() called 3 times (idempotent); no errors |

### Resource Cleanup — Session Files Retained

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_SessionFilesNotDeleted` | `unit` | Verifies session directory and files are NOT deleted. | Test fixture creates temporary session directory with `session.json` and `events.jsonl`; mock Session with `Status="completed"` | `Session=<mock>` | SessionFinalizer completes; session directory still exists; `session.json` and `events.jsonl` files still exist; only socket deleted |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_NoReturnError` | `unit` | Verifies SessionFinalizer never returns error. | Test fixture; mock RuntimeSocketManager.DeleteSocket() logs warning; mock Session with `Status="failed"`, `Error=nil` | `Session=<mock>` | Finalize completes; no error returned (void/nil return) |
| `TestFinalize_OutputStreamClosed` | `unit` | Handles closed stdout/stderr gracefully. | Test fixture redirects stdout and stderr to closed pipe; mock Session with `Status="completed"` | `Session=<mock>` | Print operations may fail silently; SessionFinalizer does not check for print errors; no panic |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_ConcurrentCalls` | `race` | Multiple concurrent Finalize calls on same session are safe. | Test fixture; mock Session with `Status="completed"`; spawn 5 goroutines calling Finalize | `Session=<mock>` (called concurrently 5 times) | All calls complete without panic; success message may print 5 times; RuntimeSocketManager.DeleteSocket() called 5 times (idempotent) |
| `TestFinalize_ConcurrentSocketDeletion` | `race` | Multiple concurrent calls to RuntimeSocketManager.DeleteSocket are safe. | Test fixture creates temporary session directory with socket file; spawn 10 goroutines calling Finalize on same Session | `Session=<mock>` (called concurrently 10 times) | All calls complete without panic; RuntimeSocketManager.DeleteSocket() handles concurrent deletion safely (first deletes, rest are no-ops); no race conditions on socket file access |

### State Transitions — Terminal Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_CompletedStatus_NoErrorField` | `unit` | Completed session has nil Error field. | Test fixture; mock Session with `Status="completed"`, `Error=nil`; capture stdout | `Session=<mock>` | Stdout contains success message; no error details printed |
| `TestFinalize_FailedStatus_ErrorFieldPresent` | `unit` | Failed session has non-nil Error field. | Test fixture; mock Session with `Status="failed"`, `Error=AgentError{Message:"test"}`; capture stderr | `Session=<mock>` | Stderr contains error details |

### Mock / Dependency Interaction — RuntimeSocketManager

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFinalize_RuntimeSocketManagerDeleteSocketCalled` | `unit` | Verifies RuntimeSocketManager.DeleteSocket() is called during finalization. | Test fixture; mock RuntimeSocketManager with call tracking; mock Session with `Status="completed"`; capture stdout | `Session=<mock>` | RuntimeSocketManager.DeleteSocket() called; stdout contains success message; both operations complete successfully |
