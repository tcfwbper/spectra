# spectra Root Command

## Overview

The root command for the `spectra` CLI binary. It provides the Cobra root command definition, global flag registration (`--version`), usage/help text, and dispatches to subcommands (`init`, `run`, `clear`). It does not perform any business logic, I/O operations, or state management.

## Boundaries

- Owns: Cobra root command definition and subcommand registration.
- Owns: `--version` flag handling and version string output.
- Owns: usage/help text rendering.
- Owns: exit code propagation from subcommands to the process.
- Delegates: all subcommand logic to their respective handlers.
- Must not: perform filesystem operations.
- Must not: call SpectraFinder (subcommands that need it call it themselves).
- Must not: maintain state between invocations.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `cobra` (spf13/cobra) | CLI framework | Define commands, register flags, execute | — |
| `runtime` | Workflow execution (via adapter) | `Run(workflowName, logger)` | Must not access internals |
| `storage` | Path composition (via adapters) | `GetWorkflowPath`, `GetAgentPath`, `GetSessionDir`, `GetSessionsDir`, `FindSpectraRoot` | — |
| `internal/builtin` | Embedded FS sources (via `builtin_resources.go`) | Read `Workflows`, `Agents`, `SpecFiles` | Must not modify |

Construction constraint: Exposes a function (e.g., `Execute() int`) that builds and runs the Cobra command tree. Returns the process exit code. The caller (`main.go`) calls `os.Exit` with this value.

Production adapter wiring: The root command file defines unexported production adapter types that implement the interfaces expected by subcommands (`RunRuntime`, `StorageLayoutInterface`, `ClearSpectraFinder`, `ClearStorageLayout`). Each adapter is a thin wrapper that delegates to the real package-level function (e.g., `runtime.Run`, `storage.GetSessionDir`). These adapters contain no logic beyond delegation. A separate `builtin_resources.go` file assigns `internal/builtin` embed.FS variables to package-level `fs.FS` variables for use by `BuiltinResourceCopier`.

## Behavior

1. Defines a Cobra root command `spectra` with usage `"spectra [command]"`.
2. Registers three subcommands: `init`, `run`, `clear`.
3. If invoked without a subcommand, prints usage information and exits with code 0.
4. If invoked with `--help`, prints usage information and exits with code 0.
5. If invoked with `--version`, prints the version string in format `"spectra version <version-string>"` and exits with code 0. Version string follows semantic versioning (e.g., `v0.1.0`).
6. If invoked with an unknown subcommand, prints `"Error: unknown command \"<subcommand>\" for \"spectra\""` to stderr and exits with code 1.
7. After successful initialization, delegates execution to the appropriate subcommand handler.
8. Propagates subcommand exit codes unchanged.

## Inputs

### Global Flags

| Flag | Type | Constraints | Required | Default |
|------|------|-------------|----------|---------|
| `--help` | boolean | N/A | No | false |
| `--version` | boolean | N/A | No | false |

## Outputs

### stdout

- Usage/help information when invoked with `--help` or without subcommand.
- Version information when invoked with `--version`.
- Subcommand outputs (delegated to subcommand handlers).

### stderr

- Error messages for invocation errors (prefixed with `"Error: "`).

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | Subcommand completed successfully, or help/version displayed |
| 1 | Error | Unknown subcommand or subcommand execution error (default) |
| 130 | SIGINT | Propagated from `run` subcommand |
| 143 | SIGTERM | Propagated from `run` subcommand |

## Invariants

1. **Stateless Execution**: Each invocation is independent. No state between runs.
2. **Error Prefix**: All error messages printed to stderr are prefixed with `"Error: "`.
3. **Subcommand Delegation**: Root command's sole responsibility after initialization is dispatch to the appropriate subcommand handler.
4. **Exit Code Propagation**: Subcommand exit codes are propagated unchanged (including 130, 143 from `run`).
5. **Cobra Consistency**: Uses the same CLI framework (Cobra) as `spectra-agent` for consistency.

## Edge Cases

- Condition: User invokes `spectra` without arguments.
  Expected: Print usage to stdout, exit with code 0.

- Condition: User invokes `spectra --version`.
  Expected: Print `"spectra version v0.1.0"` to stdout, exit with code 0.

- Condition: User invokes `spectra foo` (unknown subcommand).
  Expected: Exit with code 1, stderr `"Error: unknown command \"foo\" for \"spectra\""`.

- Condition: Subcommand returns exit code 130.
  Expected: Root command propagates exit code 130.

## Related

- [init](./init.md) - Initialize subcommand
- [run](./run.md) - Run subcommand
- [clear](./clear.md) - Clear subcommand
- [ARCHITECTURE.md](../../../../../ARCHITECTURE.md) - System architecture overview
