# spectra Root Command

## Overview

spectra is the main command-line interface for the Spectra framework. It serves as the root command for all spectra subcommands and handles global initialization, flag parsing, version information, and dispatches to subcommands. The root command provides three subcommands: `init`, `run`, and `clear`.

## Behavior

### Root Command Responsibilities

1. spectra is invoked as `spectra <subcommand> [flags]`.
2. The root command provides three subcommands: `init`, `run`, and `clear`.
3. If invoked without a subcommand, the root command prints usage information and exits with code 0.
4. If invoked with `--help`, the root command prints usage information and exits with code 0.
5. If invoked with `--version`, the root command prints the version string and exits with code 0.
6. If invoked with an unknown subcommand, the root command exits with code 1 and prints: `"Error: unknown command \"<subcommand>\" for \"spectra\""`
7. The root command uses the Cobra library for command structure, consistent with spectra-agent CLI.
8. After successful initialization, the root command delegates execution to the appropriate subcommand handler.
9. The root command does not maintain any state between invocations. Each invocation is independent.

### Version Information

When invoked with `--version`:

```
spectra version <version-string>
```

The version string follows semantic versioning (e.g., `v0.1.0`, `v1.2.3`).

### Usage Information

When invoked without a subcommand or with `--help`:

```
spectra - Framework for defining and executing flexible AI agent workflows

Usage:
  spectra [command]

Available Commands:
  init        Initialize a new Spectra project
  run         Run a workflow
  clear       Clear session data

Flags:
  --help      Show help information
  --version   Show version information

Use "spectra [command] --help" for more information about a command.
```

### Error Output Format

All error messages are printed to stderr with the prefix `"Error: "`.

## Inputs

### Global Flags

| Flag | Type | Constraints | Required | Default |
|------|------|-------------|----------|---------|
| `--help` | boolean | N/A | No | false |
| `--version` | boolean | N/A | No | false |

## Outputs

### stdout

- Usage/help information when invoked with `--help` or without subcommand
- Version information when invoked with `--version`
- Subcommand outputs (delegated to subcommand handlers)

### stderr

- Error messages for invocation errors
- Error messages propagated from subcommands

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | Subcommand completed successfully, or help/version displayed |
| 1 | Error | Unknown subcommand, subcommand invocation error, or subcommand execution error |

## Invariants

1. **Stateless Execution**: The root command must not maintain any state between invocations. Each invocation is independent.

2. **Error Prefix**: All error messages must be prefixed with `"Error: "` when printed to stderr.

3. **Subcommand Delegation**: The root command's sole responsibility after initialization is to dispatch to the appropriate subcommand handler.

4. **Exit Code Consistency**: The root command propagates subcommand exit codes unchanged.

5. **Cobra Consistency**: The root command must use the same command-line framework (Cobra) as spectra-agent for consistency.

## Edge Cases

- **Condition**: User invokes `spectra` without any arguments.
  **Expected**: Print usage information to stdout and exit with code 0.

- **Condition**: User invokes `spectra --help`.
  **Expected**: Print usage information to stdout and exit with code 0.

- **Condition**: User invokes `spectra --version`.
  **Expected**: Print version information to stdout and exit with code 0.

- **Condition**: User invokes `spectra foo` (unknown subcommand).
  **Expected**: Exit with code 1, print `"Error: unknown command \"foo\" for \"spectra\""` to stderr.

- **Condition**: User invokes `spectra init --invalid-flag`.
  **Expected**: The `init` subcommand handles flag parsing. If the flag is invalid, the subcommand returns an error, which is propagated by the root command.

- **Condition**: Subcommand returns exit code 0.
  **Expected**: Root command propagates exit code 0.

- **Condition**: Subcommand returns exit code 1.
  **Expected**: Root command propagates exit code 1.

## Related

- [init Subcommand](./init.md) - Initialize a new Spectra project
- [run Subcommand](./run.md) - Run a workflow
- [clear Subcommand](./clear.md) - Clear session data
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - System architecture overview
