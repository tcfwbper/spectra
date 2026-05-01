# spectra run Command

## Overview

The `spectra run` command starts a workflow execution by invoking the Runtime module. It serves as the CLI entry point for workflow execution and is responsible for locating the project root, validating the workflow name argument, and delegating to the Runtime. The command forwards all Runtime output (stdout and stderr) to the user and exits with the Runtime's exit code.

## Behavior

### Execution Flow

1. The `run` command is invoked as `spectra run --workflow <WorkflowName>` or `spectra run <WorkflowName>` (positional argument).
2. The command validates that exactly one workflow name argument is provided. If missing or multiple arguments are provided, the command exits with code 1 and prints: `"Error: workflow name is required"` or `"Error: too many arguments"`.
3. The command validates that the workflow name is non-empty after trimming whitespace. If empty, the command exits with code 1 and prints: `"Error: workflow name cannot be empty"`.
4. The command invokes the Runtime module by calling `Runtime.Run(workflowName)`.
5. The Runtime module handles all workflow execution logic as defined in `logic/runtime/runtime.md`, including locating the project root.
6. The `run` command forwards all output from Runtime's stdout to its own stdout (visible to the user).
7. The `run` command forwards all output from Runtime's stderr to its own stderr (visible to the user).
8. When Runtime exits, the `run` command exits with the same exit code as Runtime.
9. Runtime exit codes are:
    - `0`: Workflow completed successfully (Session status = "completed")
    - `1`: Workflow failed or initialization error (Session status = "failed" or initialization failed)
10. The `run` command does not perform any additional output or processing after Runtime exits. It simply propagates the exit code.

### Command Syntax

The command accepts the workflow name as a positional argument or via the `--workflow` flag:

```
spectra run <WorkflowName>
spectra run --workflow <WorkflowName>
```

Both forms are equivalent. If both are provided, the flag takes precedence.

### Usage Information

When invoked with `--help`:

```
Run a workflow

Usage:
  spectra run [flags] <WorkflowName>
  spectra run --workflow <WorkflowName>

Flags:
  --workflow string   Workflow name to execute (alternative to positional argument)
  --help              Show help information

Examples:
  spectra run SimpleSdd
  spectra run --workflow SimpleSdd
```

### Error Output

When workflow name is missing:

```
Error: workflow name is required
```

When workflow name is empty:

```
Error: workflow name cannot be empty
```

When too many arguments are provided:

```
Error: too many arguments
```

When `.spectra` is not found:

```
Error: .spectra directory not found. Run 'spectra init' to initialize a project.
```

All other errors (workflow not found, invalid workflow definition, runtime errors) are reported by the Runtime module and forwarded to stderr.

## Inputs

### Command-line Arguments

| Argument | Type | Constraints | Required |
|----------|------|-------------|----------|
| WorkflowName | string (positional) | Non-empty, corresponds to `<WorkflowName>.yaml` in `.spectra/workflows/` | Yes (unless `--workflow` flag is provided) |

### Flags

| Flag | Type | Constraints | Required | Default |
|------|------|-------------|----------|---------|
| `--workflow` | string | Non-empty, corresponds to `<WorkflowName>.yaml` in `.spectra/workflows/` | No (if positional argument is provided) | None |
| `--help` | boolean | N/A | No | false |

### Environment

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Current Working Directory | string | Process environment | Yes (implicit, used by Runtime to locate project root) |

## Outputs

### stdout

- All output from Runtime's stdout (workflow execution logs, agent output, completion messages)

### stderr

- Command invocation errors (missing workflow name, `.spectra` not found)
- All output from Runtime's stderr (workflow errors, runtime errors, failure messages)

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | Runtime completed successfully (workflow reached "completed" status) |
| 1 | Error | Command invocation error (missing/invalid workflow name, `.spectra` not found), or Runtime execution error (workflow failed, initialization error) |

## Invariants

1. **Runtime Delegation**: After validating the workflow name, the command must delegate all execution logic to the Runtime module. It must not implement any workflow execution logic itself, including project root location logic.

2. **Output Forwarding**: The command must forward all Runtime output (stdout and stderr) to its own stdout and stderr without modification, buffering, or filtering.

3. **Exit Code Propagation**: The command must exit with the same exit code returned by the Runtime. It must not modify or interpret the exit code.

4. **No Project Root Resolution**: The command must not resolve the project root or use SpectraFinder. Project root resolution is handled by Runtime.

5. **No Workflow Validation**: The command must not validate the workflow definition (YAML syntax, structure, etc.). This is handled by the Runtime via WorkflowDefinitionLoader.

6. **Single Workflow Execution**: Each invocation of `spectra run` executes exactly one workflow. Concurrent workflow executions require multiple invocations.

7. **Non-interactive**: The command does not prompt the user for input. All required inputs must be provided via command-line arguments or flags.

## Edge Cases

- **Condition**: User invokes `spectra run` without a workflow name argument.
  **Expected**: Exit with code 1, print `"Error: workflow name is required"` to stderr.

- **Condition**: User invokes `spectra run ""` (empty string as workflow name).
  **Expected**: Exit with code 1, print `"Error: workflow name cannot be empty"` to stderr.

- **Condition**: User invokes `spectra run SimpleSdd ExtraArg` (too many arguments).
  **Expected**: Exit with code 1, print `"Error: too many arguments"` to stderr.

- **Condition**: User invokes `spectra run --workflow SimpleSdd SimpleSdd` (both flag and positional argument).
  **Expected**: The flag takes precedence. Runtime is invoked with workflow name `SimpleSdd`.

- **Condition**: User invokes `spectra run NonExistentWorkflow` (workflow file does not exist).
  **Expected**: Runtime initialization fails (WorkflowDefinitionLoader returns "workflow definition not found"). Runtime prints error to stderr and exits with code 1. `run` command propagates exit code 1.

- **Condition**: User invokes `spectra run SimpleSdd` (workflow file exists but has invalid YAML syntax).
  **Expected**: Runtime initialization fails (WorkflowDefinitionLoader returns parse error). Runtime prints error to stderr and exits with code 1. `run` command propagates exit code 1.

- **Condition**: Runtime prints to stdout during workflow execution (e.g., agent output, transition logs).
  **Expected**: `run` command forwards all stdout to the user in real-time.

- **Condition**: Runtime prints to stderr during workflow execution (e.g., agent errors, runtime errors).
  **Expected**: `run` command forwards all stderr to the user in real-time.

- **Condition**: User terminates the `run` command with Ctrl+C (SIGINT).
  **Expected**: The signal is propagated to Runtime. Runtime handles graceful shutdown as defined in `logic/runtime/runtime.md`. `run` command exits when Runtime exits.

- **Condition**: User invokes `spectra run SimpleSdd` while another workflow execution is in progress (same or different workflow).
  **Expected**: The command proceeds to invoke Runtime. Runtime handles concurrent execution detection (session lock, socket binding). If a session is already running, Runtime reports an error and exits with code 1.

- **Condition**: `.spectra/workflows/SimpleSdd.yaml` exists but is not readable (permission denied).
  **Expected**: Runtime initialization fails (WorkflowDefinitionLoader returns read error). Runtime prints error to stderr and exits with code 1. `run` command propagates exit code 1.

- **Condition**: Workflow completes successfully after 10 minutes of execution.
  **Expected**: Runtime exits with code 0. `run` command propagates exit code 0.

- **Condition**: Workflow fails due to an agent error after 5 minutes of execution.
  **Expected**: Runtime exits with code 1. `run` command propagates exit code 1.

- **Condition**: Runtime crashes or panics unexpectedly (programming error).
  **Expected**: The panic is propagated to the `run` command. The Go runtime prints a stack trace to stderr and exits with code 2 (standard panic exit code). The `run` command does not handle panics explicitly; the Go runtime handles them.

- **Condition**: User invokes `spectra run --help`.
  **Expected**: Print usage information to stdout and exit with code 0. Do not invoke Runtime.

- **Condition**: User invokes `spectra run` with a workflow name containing path separators (e.g., `../malicious/workflow`).
  **Expected**: Runtime passes the name to WorkflowDefinitionLoader. StorageLayout composes a potentially malicious path. File access may succeed or fail depending on filesystem structure. This is not the `run` command's responsibility; WorkflowDefinitionLoader handles path composition via StorageLayout.

## Related

- [Runtime](../../runtime/runtime.md) - Main workflow execution engine
- [WorkflowDefinitionLoader](../../storage/workflow_definition_loader.md) - Loads and validates workflow definitions (used by Runtime)
- [init Subcommand](./init.md) - Initialize a Spectra project
- [clear Subcommand](./clear.md) - Clear session data
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - System architecture overview
