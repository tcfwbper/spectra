# spectra run

## Overview

The `spectra run` command is a CLI subcommand that starts workflow execution. It accepts a `--workflow` flag to specify the workflow name, invokes the Runtime to execute the workflow, handles Runtime errors, maps errors to appropriate exit codes based on signal type or failure reason, and exits with the determined exit code. This command is the user-facing entry point for running Spectra workflows.

## Behavior

### Command-Line Interface

1. `spectra run` is invoked as a subcommand of the `spectra` CLI.
2. The command accepts one required flag:
   - `--workflow <WorkflowName>`: Specifies the name of the workflow to execute. This value is passed directly to Runtime as the `workflowName` input.
3. The `--workflow` flag is required. If not provided, the command prints an error message: `"Error: required flag --workflow not provided"` to stderr and exits with code 1.
4. The `<WorkflowName>` must be a non-empty string. If an empty string is provided, the command prints an error message: `"Error: --workflow flag cannot be empty"` to stderr and exits with code 1.
5. The command does not accept any positional arguments. If positional arguments are provided, the command prints an error message: `"Error: unexpected argument '<argument>'. Use --workflow flag to specify workflow name."` to stderr and exits with code 1.
6. The command does not perform any validation on the workflow name format or existence. Workflow validation is delegated to Runtime (specifically, WorkflowDefinitionLoader invoked by SessionInitializer).

### Runtime Invocation

7. After validating command-line arguments, `spectra run` invokes `Runtime.Run(workflowName)` with the provided workflow name.
8. `Runtime.Run()` is a blocking call that manages the entire workflow execution lifecycle: session initialization, socket management, event processing, and cleanup.
9. `spectra run` waits for `Runtime.Run()` to return. The return value is a Go `error` type (or `nil` on success).

### Error Handling and Exit Code Mapping

10. After `Runtime.Run()` returns, `spectra run` examines the returned error to determine the appropriate exit code.
11. **Case 1: Success** (`error == nil`):
    - Runtime returned `nil`, indicating the session completed successfully (Session.Status == "completed").
    - `spectra run` exits with code **0** (success).
    - No additional output is printed by `spectra run` (SessionFinalizer already printed the success message to stdout).
12. **Case 2: SIGINT Termination** (error message contains `"session terminated by signal SIGINT"`):
    - Runtime was terminated by SIGINT (Ctrl+C on Unix/Linux/macOS/Windows).
    - `spectra run` prints the error message to stderr: `"Error: <error-message>"`.
    - `spectra run` exits with code **130** (128 + 2, standard Unix convention for SIGINT).
13. **Case 3: SIGTERM Termination** (error message contains `"session terminated by signal SIGTERM"`):
    - Runtime was terminated by SIGTERM (Unix/Linux/macOS only; Windows does not support SIGTERM).
    - `spectra run` prints the error message to stderr: `"Error: <error-message>"`.
    - `spectra run` exits with code **143** (128 + 15, standard Unix convention for SIGTERM).
14. **Case 4: Other Errors** (all other non-nil errors):
    - Runtime returned an error due to initialization failure, session failure, listener error, or other runtime error.
    - `spectra run` prints the error message to stderr: `"Error: <error-message>"`.
    - `spectra run` exits with code **1** (generic failure).
15. Error message matching is performed using string substring matching:
    - If `error.Error()` contains the substring `"session terminated by signal SIGINT"`, exit with code 130.
    - Else if `error.Error()` contains the substring `"session terminated by signal SIGTERM"`, exit with code 143.
    - Else, exit with code 1.
16. The error message returned by Runtime already includes sufficient context (session ID, workflow name, error details). `spectra run` does not add additional context to the error message before printing.

### Output and Logging

17. `spectra run` does not print any additional success messages. SessionFinalizer (invoked by Runtime) is responsible for printing session status to stdout (success) or stderr (failure/termination).
18. `spectra run` only prints error messages to stderr in the following cases:
    - Required flag `--workflow` not provided or empty
    - Unexpected positional arguments
    - Runtime returned a non-nil error (as described in steps 12-14)
19. All error messages printed by `spectra run` are prefixed with `"Error: "` for consistency.
20. `spectra run` does not write to any log files. All logging is handled by Runtime and its dependencies.

### Platform Compatibility

21. Exit codes 130 and 143 follow standard Unix conventions (128 + signal number). These conventions are widely recognized on Unix-like systems (Linux, macOS, BSD) and are also compatible with Windows (which supports exit codes 0-255).
22. On Windows, SIGTERM is not available. `spectra run` will never encounter a `"session terminated by signal SIGTERM"` error message on Windows. Only SIGINT (Ctrl+C) is supported.
23. On Unix-like systems, both SIGINT (Ctrl+C, terminal interrupt) and SIGTERM (kill command default, process termination) are supported.

## Inputs

### Command-Line Arguments

| Argument | Type | Constraints | Required | Description |
|----------|------|-------------|----------|-------------|
| `--workflow` | string (flag value) | Non-empty, no format validation by `spectra run` | Yes | Name of the workflow to execute. Passed directly to Runtime. |

### From Runtime

| Field | Type | Description |
|-------|------|-------------|
| error | Go error type or nil | Returned by `Runtime.Run()`. `nil` indicates success, non-nil indicates failure or signal termination. |

## Outputs

### Exit Codes

| Exit Code | Condition | Description |
|-----------|-----------|-------------|
| 0 | Runtime returned `nil` | Session completed successfully |
| 1 | Runtime returned non-nil error (not signal-related) | Initialization failure, session failure, listener error, or other runtime error |
| 130 | Runtime returned error containing `"session terminated by signal SIGINT"` | Workflow terminated by SIGINT (Ctrl+C) |
| 143 | Runtime returned error containing `"session terminated by signal SIGTERM"` | Workflow terminated by SIGTERM (Unix/Linux/macOS only) |

### Standard Error Output

`spectra run` prints to stderr only in error cases:

| Error Case | Format |
|------------|--------|
| Missing `--workflow` flag | `Error: required flag --workflow not provided` |
| Empty `--workflow` value | `Error: --workflow flag cannot be empty` |
| Unexpected positional arguments | `Error: unexpected argument '<argument>'. Use --workflow flag to specify workflow name.` |
| Runtime error | `Error: <error-message-from-runtime>` |

### Standard Output

`spectra run` does not print to stdout. SessionFinalizer (invoked by Runtime) prints session completion messages to stdout.

## Invariants

1. **Single Runtime Invocation**: `spectra run` invokes `Runtime.Run()` exactly once per command execution. Runtime manages the entire session lifecycle.

2. **No Workflow Validation**: `spectra run` does not validate the workflow name format or existence. Validation is delegated to Runtime (WorkflowDefinitionLoader).

3. **Error Message Substring Matching**: Exit code determination is based on substring matching of the error message returned by Runtime. The exact error message format is defined by Runtime's specification.

4. **No Additional Output on Success**: On success (exit code 0), `spectra run` does not print any messages. SessionFinalizer already printed the success message.

5. **Error Prefix Consistency**: All error messages printed by `spectra run` are prefixed with `"Error: "`.

6. **Blocking Runtime Call**: `spectra run` blocks on `Runtime.Run()` until the session completes, fails, or is terminated. The command does not return control to the user until Runtime returns.

7. **Exit Code Priority**: Signal-based exit codes (130, 143) take precedence over generic failure exit code (1). If Runtime returns an error indicating signal termination, `spectra run` exits with the signal-specific exit code, not 1.

8. **No Retry Logic**: `spectra run` does not retry Runtime invocation on failure. Each invocation creates a new session. To retry, the user must manually re-run the command.

9. **No Signal Handling in spectra run**: `spectra run` does not register its own signal handlers. Signal handling is performed by Runtime. `spectra run` only examines the error message returned by Runtime to determine if a signal caused termination.

10. **Platform-Agnostic Error Handling**: Exit code mapping logic is identical on all platforms. Platform differences (e.g., Windows lacking SIGTERM) are handled by Runtime, not by `spectra run`.

## Edge Cases

- **Condition**: User provides `--workflow` flag but no value (e.g., `spectra run --workflow`).
  **Expected**: The CLI framework (e.g., cobra, flag) returns an error indicating missing flag value. `spectra run` treats this as a missing `--workflow` flag, prints `"Error: required flag --workflow not provided"` to stderr, and exits with code 1.

- **Condition**: User provides `--workflow` with an empty string (e.g., `spectra run --workflow ""`).
  **Expected**: `spectra run` detects the empty string, prints `"Error: --workflow flag cannot be empty"` to stderr, and exits with code 1.

- **Condition**: User provides an invalid workflow name (e.g., workflow file does not exist).
  **Expected**: `spectra run` passes the invalid name to Runtime. Runtime's WorkflowDefinitionLoader fails to load the workflow and returns an error (e.g., `"failed to initialize session: failed to load workflow definition: workflow file not found: <name>"`). `spectra run` prints the error to stderr and exits with code 1.

- **Condition**: User provides positional arguments (e.g., `spectra run MyWorkflow` instead of `spectra run --workflow MyWorkflow`).
  **Expected**: `spectra run` detects the positional argument, prints `"Error: unexpected argument 'MyWorkflow'. Use --workflow flag to specify workflow name."` to stderr, and exits with code 1.

- **Condition**: Runtime returns `nil` (session completed successfully).
  **Expected**: `spectra run` exits with code 0. SessionFinalizer already printed the success message to stdout.

- **Condition**: Runtime returns error `"session failed: agent execution error: ArchitectAgent failed to generate specifications"`.
  **Expected**: `spectra run` prints `"Error: session failed: agent execution error: ArchitectAgent failed to generate specifications"` to stderr and exits with code 1.

- **Condition**: Runtime returns error `"session terminated by signal SIGINT"` (user pressed Ctrl+C).
  **Expected**: `spectra run` prints `"Error: session terminated by signal SIGINT"` to stderr and exits with code 130.

- **Condition**: Runtime returns error `"session terminated by signal SIGTERM"` (Unix/Linux/macOS only).
  **Expected**: `spectra run` prints `"Error: session terminated by signal SIGTERM"` to stderr and exits with code 143.

- **Condition**: Runtime returns error `"failed to locate project root: spectra not initialized"` (project not initialized).
  **Expected**: `spectra run` prints `"Error: failed to locate project root: spectra not initialized"` to stderr and exits with code 1.

- **Condition**: Runtime panics during execution (unhandled panic, unexpected nil pointer).
  **Expected**: The panic propagates to `spectra run`. If `spectra run` does not implement panic recovery, the Go runtime prints the panic stack trace to stderr and exits with code 2 (Go default for panics). If `spectra run` implements panic recovery (deferred recovery), it logs the panic, prints an error message `"Error: runtime panic: <panic-value>"` to stderr, and exits with code 1.

- **Condition**: Runtime returns error message containing `"session terminated by signal SIGINT"` as a substring within a larger error message (e.g., `"cleanup failed after session terminated by signal SIGINT"`).
  **Expected**: `spectra run` detects the substring `"session terminated by signal SIGINT"` and exits with code 130. This behavior is by design: if Runtime includes signal information in the error, it indicates the session was terminated by that signal.

- **Condition**: Runtime returns error with unexpected format (not matching any known error message pattern).
  **Expected**: `spectra run` applies the default case: prints `"Error: <error-message>"` to stderr and exits with code 1.

- **Condition**: User runs `spectra run --workflow MyWorkflow` twice in parallel in the same project directory.
  **Expected**: Each invocation creates a separate Runtime and session (unique session UUID). Both sessions execute independently. No conflict occurs unless system resources are exhausted. Both commands may complete successfully or fail independently based on their own workflow execution.

- **Condition**: User interrupts `spectra run` with Ctrl+C during Runtime initialization (before socket listener starts).
  **Expected**: Runtime receives SIGINT, stores it in `receivedSignal`, does not complete initialization, proceeds to cleanup (if session was constructed), and returns error `"session terminated by signal SIGINT"`. `spectra run` prints the error to stderr and exits with code 130.

- **Condition**: User sends SIGTERM to the `spectra run` process on Windows.
  **Expected**: Windows does not support SIGTERM. The signal is ignored or treated as an unknown signal by the OS. Runtime continues execution. No error is returned. This edge case is platform-specific and cannot be triggered on Windows.

- **Condition**: SessionFinalizer fails to print output (stdout/stderr closed, permission denied).
  **Expected**: SessionFinalizer is best-effort and does not return errors. Runtime proceeds to return the appropriate error (nil, session failed, or signal terminated). `spectra run` handles the error as usual. The user may not see SessionFinalizer's output, but will see `spectra run`'s error message (if any).

- **Condition**: Runtime returns an error during cleanup phase (e.g., `"cleanup exceeded 5 second grace period, forcing exit"`).
  **Expected**: Runtime logs the warning and exits immediately with exit code 1 (implemented by Runtime, not returned as error to `spectra run`). `spectra run` never receives this error because Runtime exits directly. This case is handled entirely within Runtime.

- **Condition**: User presses Ctrl+C twice (second signal during Runtime's grace period).
  **Expected**: Runtime receives the second signal, logs `"received second signal, forcing exit"`, and exits immediately with exit code 1 (implemented by Runtime, not returned as error to `spectra run`). `spectra run` never receives control back because Runtime exits directly.

## Related

- [Runtime](../../runtime/runtime.md) - Main workflow execution orchestrator invoked by `spectra run`
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - Framework architecture and CLI overview
- [root.md](./root.md) - Root command for `spectra` CLI
