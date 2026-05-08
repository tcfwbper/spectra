# Test Specification: `run_test.go`

## Source File Under Test

`internal/cmd/spectra/run.go`

## Test File

`internal/cmd/spectra/run_test.go`

---

## `RunCommand`

### Happy Path — Run

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_Success_ExitsZero` | `unit` | Exits with code 0 when Runtime returns success. | Mock `Runtime.Run()` to return `(0, nil)`. | `--workflow "MyWorkflow"` | Exit code 0; no output to stderr |
| `TestRun_PassesWorkflowNameToRuntime` | `unit` | Passes the workflow name from --workflow flag to Runtime.Run. | Mock `Runtime.Run()` to capture arguments and return `(0, nil)`. | `--workflow "deploy-prod"` | `Runtime.Run()` called with `workflowName="deploy-prod"` |
| `TestRun_PassesLoggerToRuntime` | `unit` | Passes a non-nil Logger to Runtime.Run. | Mock `Runtime.Run()` to capture arguments and return `(0, nil)`. | `--workflow "MyWorkflow"` | `Runtime.Run()` called with a non-nil `logger.Logger` |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_MissingWorkflowFlag_ExitsOne` | `unit` | Prints error and exits with code 1 when --workflow is not provided. | No mocks needed (command fails before Runtime invocation). | No flags | stderr contains `"Error: required flag --workflow not provided"`; exit code 1 |
| `TestRun_EmptyWorkflowFlag_ExitsOne` | `unit` | Prints error and exits with code 1 when --workflow is empty string. | No mocks needed (command fails before Runtime invocation). | `--workflow ""` | stderr contains `"Error: --workflow flag cannot be empty"`; exit code 1 |
| `TestRun_PositionalArgs_ExitsOne` | `unit` | Prints error and exits with code 1 when positional arguments are provided. | No mocks needed (command fails before Runtime invocation). | `"MyWorkflow"` (positional) | stderr contains `"Error: unexpected argument 'MyWorkflow'. Use --workflow flag to specify workflow name."`; exit code 1 |

### Happy Path — Exit Code Mapping

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SignalInterrupt_ExitsCode130` | `unit` | Maps signal interrupt error to exit code 130. | Mock `Runtime.Run()` to return `(1, errors.New("session terminated by signal interrupt"))`. | `--workflow "MyWorkflow"` | stderr contains `"Error: session terminated by signal interrupt"`; exit code 130 |
| `TestRun_SignalTerminated_ExitsCode143` | `unit` | Maps signal terminated error to exit code 143. | Mock `Runtime.Run()` to return `(1, errors.New("session terminated by signal terminated"))`. | `--workflow "MyWorkflow"` | stderr contains `"Error: session terminated by signal terminated"`; exit code 143 |
| `TestRun_RuntimeError_ExitsWithRuntimeCode` | `unit` | Exits with Runtime-returned exit code for non-signal errors. | Mock `Runtime.Run()` to return `(1, errors.New("failed to initialize session: workflow file not found: MyWorkflow"))`. | `--workflow "MyWorkflow"` | stderr contains `"Error: failed to initialize session: workflow file not found: MyWorkflow"`; exit code 1 |
| `TestRun_SignalSubstring_ExitsCode130` | `unit` | Detects signal interrupt as substring within a larger error message. | Mock `Runtime.Run()` to return `(1, errors.New("runtime failure: session terminated by signal interrupt during cleanup"))`. | `--workflow "MyWorkflow"` | Exit code 130 |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_RuntimeCleanupTimeout_ExitsOne` | `unit` | Propagates cleanup timeout error with exit code 1. | Mock `Runtime.Run()` to return `(1, errors.New("cleanup timeout"))`. | `--workflow "MyWorkflow"` | stderr contains `"Error: cleanup timeout"`; exit code 1 |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_InvokesRuntimeExactlyOnce` | `unit` | Invokes Runtime.Run exactly once per command execution. | Mock `Runtime.Run()` to count invocations and return `(0, nil)`. | `--workflow "MyWorkflow"` | `Runtime.Run()` invocation count is 1 |
| `TestRun_DoesNotInvokeRuntimeOnValidationFailure` | `unit` | Does not invoke Runtime.Run when flag validation fails. | Mock `Runtime.Run()` to count invocations. | No flags | `Runtime.Run()` invocation count is 0 |
