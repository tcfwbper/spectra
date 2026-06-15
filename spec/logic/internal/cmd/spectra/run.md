# spectra run

## Overview

The `spectra run` command starts workflow execution. It accepts a `--workflow` flag and an optional `--session-id` flag, constructs a Logger, invokes `Runtime.Run(workflowName, sessionID, logger)`, examines the returned exit code and error, maps signal-terminated errors to exit codes 130/143, and exits with the determined exit code. This command is the user-facing entry point for running Spectra workflows.

## Boundaries

- Owns: `--workflow` flag parsing and validation (required, non-empty).
- Owns: `--session-id` flag parsing (optional).
- Owns: positional argument rejection.
- Owns: Logger construction (slog-based, output to stderr).
- Owns: exit code mapping from Runtime result to process exit code (signal detection).
- Owns: printing error messages to stderr on failure.
- Delegates: all workflow execution logic to Runtime.
- Delegates: error message formatting to `cmdutil.ErrorFormatter`.
- Must not: register signal handlers (Runtime owns signal handling).
- Must not: validate workflow name format or existence (Runtime's responsibility).
- Must not: validate session-id UUID format (SessionInitializer's responsibility).
- Must not: retry on failure.
- Must not: print success messages (SessionFinalizer logs via Logger).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `Runtime` | Workflow execution | `Run(workflowName, sessionID, logger) (int, error)` | Must not access internals |
| `cmdutil.ErrorFormatter` | Error formatting | `FormatError(msg)` | — |
| `cmdutil.SignalExitCodes` | Exit code constants | Read `ExitSignalINT`, `ExitSignalTERM` | — |
| `logger.NewSlogLogger` | Logger construction | Construct slog-based Logger for stderr | — |

Construction constraint: Registered as a Cobra subcommand of the root command. The `RunRunCommand` function accepts a `RunCommandOptions` struct with injectable dependencies (`Runtime`, `Workflow`, `WorkflowProvided`, `SessionID`, `Args`, `Stdout`, `Stderr`, `Logger`). Production wiring constructs a `productionRuntime` adapter and passes Cobra's captured flag/args state.

## Behavior

### Command-Line Interface

1. `spectra run` is invoked as a subcommand of `spectra`.
2. Accepts one required flag: `--workflow <WorkflowName>` (non-empty string).
3. Accepts one optional flag: `--session-id <UUID>` (string, default empty).
4. If `--workflow` is not provided, prints `"Error: required flag --workflow not provided"` to stderr and exits with code 1.
5. If `--workflow` is an empty string, prints `"Error: --workflow flag cannot be empty"` to stderr and exits with code 1.
6. If positional arguments are provided, prints `"Error: unexpected argument '<argument>'. Use --workflow flag to specify workflow name."` to stderr and exits with code 1.
7. Does not validate workflow name format or existence. Validation is delegated to Runtime.
8. Does not validate `--session-id` UUID format. Validation is delegated to SessionInitializer via Runtime. If not provided, the empty string is passed to Runtime which signals auto-generation.

### Logger Construction

7. Constructs a `*slog.Logger` that outputs to stderr with text format.
8. Wraps it via `logger.NewSlogLogger(slogger)` to produce a `logger.Logger` interface.
9. This Logger is passed to Runtime and used for all structured logging during execution.

### Runtime Invocation

10. Calls `Runtime.Run(workflowName, sessionID, logger)` which returns `(exitCode int, err error)`.
11. `Runtime.Run()` is a blocking call that manages the entire workflow execution lifecycle.

### Exit Code Determination

12. If `err == nil`: exits with the exit code returned by Runtime (expected to be 0).
13. If `err != nil`: examines the error message for signal indicators.
14. If `err.Error()` contains substring `"session terminated by signal interrupt"`: prints the error to stderr and exits with code 130 (`ExitSignalINT`).
15. If `err.Error()` contains substring `"session terminated by signal terminated"`: prints the error to stderr and exits with code 143 (`ExitSignalTERM`).
16. Otherwise: prints the error to stderr and exits with the exit code returned by Runtime (expected to be 1).

### Output

17. On success (exit code 0), `spectra run` does not print any additional messages. SessionFinalizer already logged via Logger.
18. On failure, prints `"Error: <error-message>"` to stderr.
19. All error messages are formatted via `cmdutil.FormatError`.

## Inputs

### Command-Line Arguments

| Argument | Type | Constraints | Required | Description |
|----------|------|-------------|----------|-------------|
| `--workflow` | string | Non-empty | Yes | Name of the workflow to execute |
| `--session-id` | string | Valid UUID format (validated downstream by SessionInitializer) | No | User-specified session UUID. If not provided, SessionInitializer auto-generates one. |

### From Runtime

| Field | Type | Description |
|-------|------|-------------|
| exitCode | int | 0 on success, 1 on failure |
| error | error | nil on success, descriptive error on failure |

## Outputs

### Exit Codes

| Exit Code | Condition | Description |
|-----------|-----------|-------------|
| 0 | Runtime returned (0, nil) | Session completed successfully |
| 1 | Runtime returned non-zero exit code with non-signal error | Initialization failure, session failure, or other runtime error |
| 130 | Runtime error contains `"session terminated by signal interrupt"` | SIGINT termination |
| 143 | Runtime error contains `"session terminated by signal terminated"` | SIGTERM termination (Unix/Linux/macOS only) |

### Standard Error Output

| Error Case | Format |
|------------|--------|
| Missing `--workflow` flag | `Error: required flag --workflow not provided` |
| Empty `--workflow` value | `Error: --workflow flag cannot be empty` |
| Unexpected positional arguments | `Error: unexpected argument '<argument>'. Use --workflow flag to specify workflow name.` |
| Runtime error | `Error: <error-message-from-runtime>` |

## Invariants

1. **Single Runtime Invocation**: Invokes `Runtime.Run()` exactly once per command execution.
2. **No Workflow Validation**: Does not validate workflow name format or existence.
3. **No Session ID Validation**: Does not validate `--session-id` UUID format. Passes the raw string value (or empty string if not provided) to Runtime.
4. **Signal Detection via Substring**: Exit code 130/143 is determined by substring matching of the error message.
5. **Signal Exit Code Priority**: Signal-based exit codes (130, 143) take precedence over the Runtime-returned exit code.
6. **No Additional Output on Success**: On success, no messages are printed (Logger output handled by Runtime internals).
7. **No Signal Handling**: Does not register signal handlers. Signal handling is Runtime's responsibility.
8. **No Retry**: Does not retry Runtime invocation on failure.
9. **Error Prefix Consistency**: All error messages are prefixed with `"Error: "` via ErrorFormatter.
10. **Logger to stderr**: The constructed Logger outputs to stderr, not stdout.

## Edge Cases

- Condition: `--workflow` flag not provided.
  Expected: Prints error to stderr, exits with code 1.

- Condition: `--workflow ""` (empty string).
  Expected: Prints error to stderr, exits with code 1.

- Condition: Positional arguments provided (e.g., `spectra run MyWorkflow`).
  Expected: Prints error to stderr, exits with code 1.

- Condition: Runtime returns (0, nil).
  Expected: Exits with code 0. No additional output.

- Condition: Runtime returns (1, error) with message `"session terminated by signal interrupt"`.
  Expected: Prints `"Error: session terminated by signal interrupt"` to stderr, exits with code 130.

- Condition: Runtime returns (1, error) with message `"session terminated by signal terminated"`.
  Expected: Prints error to stderr, exits with code 143.

- Condition: Runtime returns (1, error) with message `"failed to initialize session: workflow file not found: MyWorkflow"`.
  Expected: Prints error to stderr, exits with code 1.

- Condition: Runtime returns (1, error) with message `"cleanup timeout"`.
  Expected: Prints error to stderr, exits with code 1.

- Condition: Runtime returns (1, error) with message containing `"session terminated by signal interrupt"` as substring within larger message.
  Expected: Exits with code 130 (substring match).

## Related

- [Runtime](../../../../runtime/runtime.md) - Workflow execution orchestrator invoked by `spectra run`
- [root](./root.md) - Parent command
- [SignalExitCodes](../../cmdutil/signal_exit_codes.md) - Signal exit code constants
- [ErrorFormatter](../../cmdutil/error_formatter.md) - Error message formatting
- [SlogLogger](../../../../logger/slog_logger.md) - Logger implementation used for Runtime
