# Test Specification: `run.go`

## Source File Under Test
`cmd/spectra/run.go`

## Test File
`cmd/spectra/run_test.go`

---

## `RunCommand`

### Happy Path — Successful Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_SuccessfulWorkflow` | `unit` | Exits with code 0 when Runtime.Run returns nil. | Mock Runtime that returns `nil` | `--workflow MyWorkflow` | Exit code 0; no output to stderr; no additional output to stdout |
| `TestRunCommand_RuntimeInvokedWithWorkflowName` | `unit` | Passes workflow name to Runtime.Run exactly as provided. | Mock Runtime that captures input | `--workflow TestWorkflow123` | Runtime.Run called once with `"TestWorkflow123"`; exit code 0 |
| `TestRunCommand_SingleRuntimeInvocation` | `unit` | Invokes Runtime.Run exactly once per command execution. | Mock Runtime with invocation counter | `--workflow MyWorkflow` | Runtime.Run called exactly once; exit code 0 |

### Happy Path — Help and Usage

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_HelpFlag` | `unit` | Displays usage information when invoked with --help. | None | `--help` | Prints usage information to stdout including --workflow flag description; exit code 0 |
| `TestRunCommand_UsageIncludesWorkflowFlag` | `unit` | Usage information documents the required --workflow flag. | None | `--help` | stdout contains "--workflow" flag with description indicating it specifies workflow name; marked as required |

### Validation Failures — Missing or Empty Workflow Flag

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_MissingWorkflowFlag` | `unit` | Exits with code 1 when --workflow flag not provided. | None | No arguments | Prints `"Error: required flag --workflow not provided"` to stderr; exit code 1 |
| `TestRunCommand_EmptyWorkflowString` | `unit` | Exits with code 1 when --workflow value is empty string. | None | `--workflow ""` | Prints `"Error: --workflow flag cannot be empty"` to stderr; exit code 1 |
| `TestRunCommand_WorkflowFlagWithoutValue` | `unit` | Treats missing flag value same as missing flag. | None | `--workflow` (no value) | Prints `"Error: required flag --workflow not provided"` to stderr; exit code 1 |

### Validation Failures — Positional Arguments

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_PositionalArgument` | `unit` | Rejects positional argument and suggests using --workflow flag. | None | `MyWorkflow` | Prints `"Error: unexpected argument 'MyWorkflow'. Use --workflow flag to specify workflow name."` to stderr; exit code 1 |
| `TestRunCommand_MultiplePositionalArguments` | `unit` | Reports error for first positional argument. | None | `arg1 arg2` | Prints `"Error: unexpected argument 'arg1'. Use --workflow flag to specify workflow name."` to stderr; exit code 1 |
| `TestRunCommand_PositionalWithWorkflowFlag` | `unit` | Rejects positional argument even when --workflow flag is valid. | None | `--workflow Valid extra-arg` | Prints `"Error: unexpected argument 'extra-arg'. Use --workflow flag to specify workflow name."` to stderr; exit code 1 |

### Happy Path — No Workflow Validation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_NoWorkflowNameValidation` | `unit` | Does not validate workflow name format before invoking Runtime. | Mock Runtime that captures input | `--workflow ../../../etc/passwd` | Runtime.Run called with `"../../../etc/passwd"`; no pre-validation error from run command |
| `TestRunCommand_SpecialCharactersInWorkflowName` | `unit` | Accepts workflow names with special characters. | Mock Runtime that captures input | `--workflow "workflow-with-special!@#$%"` | Runtime.Run called with workflow name as provided; no validation error |

### Error Handling — Generic Runtime Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeInitializationError` | `unit` | Exits with code 1 when Runtime returns initialization error. | Mock Runtime that returns `errors.New("failed to initialize session: failed to load workflow definition: workflow file not found: MyWorkflow")` | `--workflow MyWorkflow` | Prints `"Error: failed to initialize session: failed to load workflow definition: workflow file not found: MyWorkflow"` to stderr; exit code 1 |
| `TestRunCommand_RuntimeSessionFailureError` | `unit` | Exits with code 1 when Runtime returns session failure error. | Mock Runtime that returns `errors.New("session failed: agent execution error: ArchitectAgent failed to generate specifications")` | `--workflow MyWorkflow` | Prints `"Error: session failed: agent execution error: ArchitectAgent failed to generate specifications"` to stderr; exit code 1 |
| `TestRunCommand_RuntimeProjectNotInitializedError` | `unit` | Exits with code 1 when Runtime returns project not initialized error. | Mock Runtime that returns `errors.New("failed to locate project root: spectra not initialized")` | `--workflow MyWorkflow` | Prints `"Error: failed to locate project root: spectra not initialized"` to stderr; exit code 1 |
| `TestRunCommand_RuntimeUnexpectedError` | `unit` | Exits with code 1 for any non-signal Runtime error. | Mock Runtime that returns `errors.New("unexpected error from runtime")` | `--workflow MyWorkflow` | Prints `"Error: unexpected error from runtime"` to stderr; exit code 1 |

### Error Handling — SIGINT Termination

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_SIGINTExactMatch` | `unit` | Exits with code 130 when Runtime error contains exact SIGINT message. | Mock Runtime that returns `errors.New("session terminated by signal SIGINT")` | `--workflow MyWorkflow` | Prints `"Error: session terminated by signal SIGINT"` to stderr; exit code 130 |
| `TestRunCommand_SIGINTSubstringMatch` | `unit` | Exits with code 130 when error contains SIGINT substring. | Mock Runtime that returns `errors.New("cleanup failed after session terminated by signal SIGINT")` | `--workflow MyWorkflow` | Prints error to stderr; exit code 130 (substring match by design) |
| `TestRunCommand_SIGINTWithContextError` | `unit` | Exits with code 130 when error wraps SIGINT message. | Mock Runtime that returns `fmt.Errorf("workflow execution interrupted: session terminated by signal SIGINT: context")` | `--workflow MyWorkflow` | Prints error to stderr; exit code 130 |

### Error Handling — SIGTERM Termination

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_SIGTERMExactMatch` | `unit` | Exits with code 143 when Runtime error contains exact SIGTERM message. | Mock Runtime that returns `errors.New("session terminated by signal SIGTERM")` | `--workflow MyWorkflow` | Prints `"Error: session terminated by signal SIGTERM"` to stderr; exit code 143 |
| `TestRunCommand_SIGTERMSubstringMatch` | `unit` | Exits with code 143 when error contains SIGTERM substring. | Mock Runtime that returns `errors.New("cleanup failed after session terminated by signal SIGTERM")` | `--workflow MyWorkflow` | Prints error to stderr; exit code 143 (substring match by design) |
| `TestRunCommand_SIGTERMWithContextError` | `unit` | Exits with code 143 when error wraps SIGTERM message. | Mock Runtime that returns `fmt.Errorf("workflow execution interrupted: session terminated by signal SIGTERM: context")` | `--workflow MyWorkflow` | Prints error to stderr; exit code 143 |

### Error Handling — Exit Code Priority

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_SIGINTPriorityOverGenericError` | `unit` | SIGINT exit code takes precedence when substring detected. | Mock Runtime that returns `errors.New("multiple errors: session terminated by signal SIGINT")` | `--workflow MyWorkflow` | Exit code 130 (not 1) |
| `TestRunCommand_SIGTERMPriorityOverGenericError` | `unit` | SIGTERM exit code takes precedence when substring detected. | Mock Runtime that returns `errors.New("multiple errors: session terminated by signal SIGTERM")` | `--workflow MyWorkflow` | Exit code 143 (not 1) |
| `TestRunCommand_SIGINTBeforeSIGTERM` | `unit` | SIGINT detected first when both substrings present. | Mock Runtime that returns `errors.New("session terminated by signal SIGINT then session terminated by signal SIGTERM")` | `--workflow MyWorkflow` | Exit code 130 (SIGINT checked first) |

### Error Output Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_ErrorPrefixConsistency` | `unit` | All error messages prefixed with "Error: ". | Mock Runtime that returns `errors.New("test error")` | `--workflow MyWorkflow` | stderr output starts with `"Error: test error"` |
| `TestRunCommand_ErrorToStderr` | `unit` | Error messages printed to stderr, not stdout. | Mock Runtime that returns `errors.New("test error")` | `--workflow MyWorkflow` | Error appears on stderr; stdout is empty; exit code 1 |
| `TestRunCommand_NoAdditionalContextAdded` | `unit` | Does not add additional context to Runtime error message. | Mock Runtime that returns `errors.New("original error message")` | `--workflow MyWorkflow` | stderr contains exactly `"Error: original error message"` (no additional wrapping) |

### Output and Logging — Success Case

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_NoOutputOnSuccess` | `unit` | Does not print to stdout or stderr on successful execution. | Mock Runtime that returns `nil` | `--workflow MyWorkflow` | No output to stderr; no output to stdout from run command (SessionFinalizer handles success message); exit code 0 |
| `TestRunCommand_SessionFinalizerHandlesSuccessMessage` | `unit` | Run command does not duplicate SessionFinalizer success output. | Mock Runtime that returns `nil` and mock SessionFinalizer that prints to stdout | `--workflow MyWorkflow` | Only SessionFinalizer output appears; run command adds no additional output; exit code 0 |

### Blocking Behavior

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_BlocksOnRuntimeRun` | `unit` | Does not return until Runtime.Run completes. | Mock Runtime that sleeps for 100ms then returns `nil` | `--workflow MyWorkflow` | Command does not return before 100ms; exit code 0 after Runtime.Run completes |
| `TestRunCommand_ReturnsImmediatelyOnRuntimeError` | `unit` | Returns as soon as Runtime.Run returns error. | Mock Runtime that immediately returns error | `--workflow MyWorkflow` | Command returns immediately with error; exit code 1 |

### No Retry Logic

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_NoRetryOnFailure` | `unit` | Does not retry Runtime.Run on failure. | Mock Runtime that tracks invocation count and always returns error | `--workflow MyWorkflow` | Runtime.Run called exactly once; exit code 1; no retry attempts |

### Edge Cases — Panic Recovery

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimePanicWithRecovery` | `unit` | Recovers from Runtime panic and exits with code 1 if panic recovery implemented. | Mock Runtime that panics with `"nil pointer dereference"` | `--workflow MyWorkflow` | Prints `"Error: runtime panic: nil pointer dereference"` to stderr; exit code 1 (if panic recovery implemented) |
| `TestRunCommand_RuntimePanicWithoutRecovery` | `unit` | Allows panic to propagate if panic recovery not implemented. | Mock Runtime that panics with `"nil pointer dereference"` | `--workflow MyWorkflow` | Panic propagates; Go runtime prints stack trace to stderr; exit code 2 (Go default) (if panic recovery not implemented) |

### Edge Cases — Parallel Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_ParallelInvocationsIndependent` | `unit` | Multiple parallel invocations execute independently. | Test fixture with temporary directory created programmatically; two mock Runtimes with unique session IDs | Two simultaneous invocations with different workflow names | Both Runtime.Run calls execute independently; no shared state; both complete successfully |

### Edge Cases — Unknown Flags

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_UnknownFlag` | `unit` | Returns error for unknown flag. | None | `--unknown-flag value` | Prints error to stderr mentioning unknown flag; exit code 1 |
| `TestRunCommand_ValidAndInvalidFlagsCombined` | `unit` | Returns error when unknown flag provided alongside valid flag. | None | `--workflow MyWorkflow --unknown-flag` | Prints error to stderr mentioning unknown flag; exit code 1; Runtime not invoked |

### Platform Compatibility

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_ExitCodeRangeCompatibility` | `unit` | Exit codes 0, 1, 130, 143 are valid on all platforms (0-255 range). | None | Test all four exit code paths | All exit codes within valid range (0-255) for Unix/Windows compatibility |
| `TestRunCommand_SIGTERMNotAvailableOnWindows` | `unit` | SIGTERM error never returned by Runtime on Windows. | Windows OS check (skip on non-Windows); Mock Runtime that returns `errors.New("session terminated by signal SIGTERM")` | `--workflow MyWorkflow` | Test documents Windows limitation: SIGTERM not supported; this test simulates the behavior if Runtime incorrectly returned SIGTERM error on Windows |

### Cobra Framework Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_UsesCobraFramework` | `unit` | Run subcommand implemented using Cobra library. | None | Code structure inspection | Run command registered as Cobra subcommand; uses Cobra flag parsing |
| `TestRunCommand_RegisteredAsSubcommand` | `unit` | Run command registered with root command. | Mock root command | Invoke via root command: `spectra run --workflow MyWorkflow` | Root command delegates to run subcommand; Runtime.Run invoked |

### Stateless Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_StatelessInvocations` | `unit` | Multiple sequential invocations are independent with no shared state. | Test fixture with temporary directory created programmatically | First invocation: `--workflow W1` returns error; second invocation: `--workflow W2` returns nil | First exits with code 1; second exits with code 0; no state leakage between invocations |

### Boundary Values — Workflow Name Length

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_SingleCharacterWorkflowName` | `unit` | Accepts single-character workflow name. | Mock Runtime that captures input | `--workflow a` | Runtime.Run called with `"a"`; exit code 0 |
| `TestRunCommand_VeryLongWorkflowName` | `unit` | Accepts very long workflow name without length restriction. | Mock Runtime that captures input | `--workflow` with 10000-character string | Runtime.Run called with full 10000-character string; exit code 0 (no truncation by run command) |

### Boundary Values — Whitespace in Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_WorkflowNameWithLeadingWhitespace` | `unit` | Passes workflow name with leading whitespace as-is. | Mock Runtime that captures input | `--workflow " LeadingSpace"` | Runtime.Run called with `" LeadingSpace"` (leading space preserved) |
| `TestRunCommand_WorkflowNameWithTrailingWhitespace` | `unit` | Passes workflow name with trailing whitespace as-is. | Mock Runtime that captures input | `--workflow "TrailingSpace "` | Runtime.Run called with `"TrailingSpace "` (trailing space preserved) |
| `TestRunCommand_WorkflowNameAllWhitespace` | `unit` | Passes workflow name that is only whitespace (not empty string). | Mock Runtime that captures input | `--workflow "   "` | Runtime.Run called with `"   "` (whitespace-only string is not empty, validation passes) |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeRunCalledWithCorrectParameter` | `unit` | Runtime.Run receives workflow name exactly as specified by --workflow flag. | Mock Runtime with parameter capture | `--workflow MyTestWorkflow` | Mock verifies `Runtime.Run("MyTestWorkflow")` called with correct parameter |
| `TestRunCommand_RuntimeNotInvokedOnValidationFailure` | `unit` | Runtime.Run not called when command-line validation fails. | Mock Runtime with invocation counter | `--workflow ""` (empty) | Runtime.Run never called; exit code 1; error printed to stderr |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_PropagatesRuntimeErrorMessage` | `unit` | Prints Runtime error message to stderr without modification (except "Error: " prefix). | Mock Runtime that returns `errors.New("detailed runtime error with session ID abc-123 and context")` | `--workflow MyWorkflow` | stderr contains `"Error: detailed runtime error with session ID abc-123 and context"` exactly; exit code 1 |

### Integration — Runtime Context

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeReceivesNoAdditionalContext` | `unit` | Run command does not pass additional context beyond workflow name to Runtime.Run. | Mock Runtime that captures all parameters | `--workflow MyWorkflow` | Runtime.Run called with single parameter: workflow name string; no context, no options, no additional parameters |
