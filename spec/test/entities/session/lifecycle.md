# Test Specification: `lifecycle_test.go`

## Source File Under Test
`entities/session/lifecycle.go`

## Test File
`entities/session/lifecycle_test.go`

---

## `Run`

### Happy Path — Run

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_FromInitializing` | `unit` | Transitions status from "initializing" to "running". | Construct session via `NewSession` (status is "initializing") | Call `session.Run()` | Returns `nil`; `GetStatusSafe()` returns `"running"`; `UpdatedAt >= CreatedAt` |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_FromRunning_ReturnsError` | `unit` | Rejects Run when status is already "running". | Construct session; call `Run()` successfully | Call `session.Run()` again | Returns error with message `"cannot run session: status is 'running', expected 'initializing'"` |
| `TestRun_FromCompleted_ReturnsError` | `unit` | Rejects Run when status is "completed". | Construct session; call `Run()` then `Done(ch)` | Call `session.Run()` | Returns error with message `"cannot run session: status is 'completed', expected 'initializing'"` |
| `TestRun_FromFailed_ReturnsError` | `unit` | Rejects Run when status is "failed". | Construct session; call `Fail(agentErr, ch)` | Call `session.Run()` | Returns error with message `"cannot run session: status is 'failed', expected 'initializing'"` |

---

## `Done`

### Happy Path — Done

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDone_FromRunning` | `unit` | Transitions status from "running" to "completed" and sends notification. | Construct session; call `Run()`; create buffered channel `ch` with capacity 2 | Call `session.Done(ch)` | Returns `nil`; `GetStatusSafe()` returns `"completed"`; `ch` receives exactly one `struct{}`; `UpdatedAt` advanced |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDone_FromInitializing_ReturnsError` | `unit` | Rejects Done when status is "initializing". | Construct session (no `Run()` called); create buffered channel `ch` | Call `session.Done(ch)` | Returns error with message `"cannot complete session: status is 'initializing', expected 'running'"` |
| `TestDone_FromCompleted_ReturnsError` | `unit` | Rejects Done when already completed. | Construct session; `Run()`; `Done(ch)` | Call `session.Done(ch)` again | Returns error with message `"cannot complete session: status is 'completed', expected 'running'"` |
| `TestDone_FromFailed_ReturnsError` | `unit` | Rejects Done when status is "failed". | Construct session; call `Fail(agentErr, ch)` | Call `session.Done(ch)` | Returns error with message `"cannot complete session: status is 'failed', expected 'running'"` |

---

## `Fail`

### Happy Path — Fail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_FromInitializing_WithAgentError` | `unit` | Transitions from "initializing" to "failed" with AgentError. | Construct session; create buffered channel `ch` with capacity 2; create `*AgentError` | Call `session.Fail(agentErr, ch)` | Returns `nil`; `GetStatusSafe()` returns `"failed"`; `GetErrorSafe()` returns the same `*AgentError`; `ch` receives exactly one `struct{}` |
| `TestFail_FromRunning_WithRuntimeError` | `unit` | Transitions from "running" to "failed" with RuntimeError. | Construct session; call `Run()`; create buffered channel `ch`; create `*RuntimeError` | Call `session.Fail(runtimeErr, ch)` | Returns `nil`; `GetStatusSafe()` returns `"failed"`; `GetErrorSafe()` returns the same `*RuntimeError`; `ch` receives one value |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_NilError` | `unit` | Rejects nil error without acquiring the lock. | Construct session; create buffered channel `ch` | Call `session.Fail(nil, ch)` | Returns error with message `"error cannot be nil"`; status unchanged |
| `TestFail_InvalidErrorType` | `unit` | Rejects error that is not *AgentError or *RuntimeError. | Construct session; create buffered channel `ch` | Call `session.Fail(errors.New("plain"), ch)` | Returns error with message `"invalid error type: must be *AgentError or *RuntimeError"`; status unchanged |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFail_FromFailed_ReturnsError` | `unit` | Rejects second Fail call preserving first error. | Construct session; `Fail(agentErr1, ch)` succeeds | Call `session.Fail(agentErr2, ch)` | Returns error with message `"session already failed"`; `GetErrorSafe()` still returns `agentErr1` |
| `TestFail_FromCompleted_ReturnsError` | `unit` | Rejects Fail when already completed. | Construct session; `Run()`; `Done(ch)` | Call `session.Fail(agentErr, ch)` | Returns error with message `"cannot fail session: status is 'completed', workflow already terminated"` |

