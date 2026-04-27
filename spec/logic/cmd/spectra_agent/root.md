# spectra-agent Root Command

## Overview

spectra-agent is a lightweight, stateless command-line tool that provides the interface for AI agents and humans to interact with the workflow runtime. It serves as the root command for all spectra-agent subcommands. The root command handles global initialization, flag parsing, exit code management, and dispatches to subcommands. All workflow state is managed by the Runtime; spectra-agent acts purely as a synchronous transport client.

## Behavior

### Root Command Responsibilities

1. spectra-agent is invoked as `spectra-agent <subcommand> [flags]`.
2. The root command parses global flags and validates their presence before dispatching to subcommands.
3. The root command provides two subcommands: `event` (with nested `emit` subcommand) and `error`.
4. All subcommands require the global flag `--session-id <UUID>` to identify the target session.
5. The root command does not read the `SPECTRA_SESSION_ID` environment variable. The `--session-id` flag is mandatory for all subcommands.
6. If `--session-id` is missing or empty, the root command exits with code 1 and prints: `"Error: --session-id flag is required"`
7. The root command uses `SpectraFinder` to locate the project root directory from the current working directory.
8. If `SpectraFinder` cannot find the project root (no `.spectra/` directory in any ancestor), the root command exits with code 1 and prints: `"Error: .spectra directory not found. Are you in a Spectra project?"`
9. After successful initialization, the root command delegates execution to the appropriate subcommand handler.
10. The root command does not perform any socket operations directly. Socket communication is handled by subcommands via the shared `SocketClient` component.
11. If the user invokes `spectra-agent` without a subcommand, it prints usage information to stdout and exits with code 0.
12. If the user invokes an unknown subcommand, it prints an error message and exits with code 1.

### Exit Code Management

The root command enforces a consistent exit code scheme across all subcommands:

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | Runtime responded with success, operation completed |
| 1 | Invocation Error | Missing required argument/flag, invalid flag value, invalid JSON format, `.spectra` directory not found, unknown subcommand |
| 2 | Transport Error | Socket file not found, connection refused, connection timeout (30s), I/O error during send/receive |
| 3 | Runtime Execution Error | Runtime responded with error status, malformed response JSON, response missing required fields |

### Global Flags

| Flag | Type | Constraints | Required | Default |
|------|------|-------------|----------|---------|
| `--session-id` | string (UUID) | Valid UUID v4 format | Yes | None |

### Error Output Format

All error messages are printed to stderr with the prefix `"Error: "`.

**Exit Code 1 Examples**:
```
Error: --session-id flag is required
Error: .spectra directory not found. Are you in a Spectra project?
Error: unknown command "foo" for "spectra-agent"
```

**Exit Code 2 and 3**: Handled by subcommands, propagated through the root command.

### Usage Information

When invoked without a subcommand or with `--help`:

```
spectra-agent - Interact with the Spectra workflow runtime

Usage:
  spectra-agent [command]

Available Commands:
  event       Emit events to the workflow runtime
  error       Report unrecoverable errors to the workflow runtime

Flags:
  --session-id string   Session UUID (required)

Use "spectra-agent [command] --help" for more information about a command.
```

## Inputs

### Global Initialization

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Session ID | string (UUID) | `--session-id` flag | Yes |
| Current Working Directory | string | Process environment | Yes (implicit) |

## Outputs

### stdout

- Usage/help information when invoked with `--help` or without subcommand

### stderr

- Error messages for invocation errors (exit code 1)
- Error messages propagated from subcommands (exit codes 2, 3)

### Exit Codes

See "Exit Code Management" section above.

## Caller Responsibility (Expected Usage Pattern)

spectra-agent does **not** read environment variables. However, agent processes spawned by AgentInvoker receive two environment variables: `SPECTRA_SESSION_ID` and `SPECTRA_CLAUDE_SESSION_ID`. The agent's system prompt is responsible for instructing the agent to forward these values to spectra-agent invocations as flags:

```
spectra-agent event emit <type> --session-id "$SPECTRA_SESSION_ID" --claude-session-id "$SPECTRA_CLAUDE_SESSION_ID" [--message ...] [--payload ...]
spectra-agent error <message> --session-id "$SPECTRA_SESSION_ID" --claude-session-id "$SPECTRA_CLAUDE_SESSION_ID" [--detail ...]
```

When invoked from a human node (no Claude agent), the human (or human-facing wrapper script) must supply `--session-id` explicitly and either omit `--claude-session-id` or pass an empty string `""`. Agent nodes always require a non-empty `--claude-session-id` matching the value stored in SessionData (see EventProcessor / ErrorProcessor validation rules).

This split (CLI flags only, no env reads) keeps the CLI deterministic and testable; environment-to-flag forwarding is the responsibility of the agent's system prompt or the human wrapper.

## Invariants

1. **Mandatory Session ID**: The `--session-id` flag must always be provided. The root command must not read from the `SPECTRA_SESSION_ID` environment variable.

2. **Project Root Discovery**: The root command must use `SpectraFinder` to locate the project root before dispatching to subcommands. If not found, it must exit with code 1.

3. **Exit Code Consistency**: All subcommands must use the standardized exit codes (0, 1, 2, 3). The root command enforces this by propagating subcommand exit codes unchanged.

4. **Stateless Execution**: The root command must not maintain any state between invocations. Each invocation is independent.

5. **Error Prefix**: All error messages must be prefixed with `"Error: "` when printed to stderr.

6. **No Direct Socket Operations**: The root command must not perform socket operations. All socket communication is delegated to subcommands.

7. **Subcommand Delegation**: The root command's sole responsibility after initialization is to dispatch to the appropriate subcommand handler.

## Edge Cases

- **Condition**: User invokes `spectra-agent` without any arguments.
  **Expected**: Print usage information to stdout and exit with code 0.

- **Condition**: User invokes `spectra-agent --help`.
  **Expected**: Print usage information to stdout and exit with code 0.

- **Condition**: User invokes `spectra-agent foo` (unknown subcommand).
  **Expected**: Exit with code 1, print `"Error: unknown command \"foo\" for \"spectra-agent\""` to stderr.

- **Condition**: User invokes `spectra-agent event emit MyEvent` without `--session-id`.
  **Expected**: Root command detects missing flag, exits with code 1, prints `"Error: --session-id flag is required"` to stderr.

- **Condition**: User invokes `spectra-agent --session-id ""` (empty string).
  **Expected**: Exit with code 1, print `"Error: --session-id flag is required"` to stderr.

- **Condition**: User invokes `spectra-agent --session-id invalid-uuid event emit MyEvent`.
  **Expected**: Root command accepts the value (does not validate UUID format). Subcommand proceeds and fails at socket connection (exit code 2). UUID validation is the Runtime's responsibility.

- **Condition**: User invokes spectra-agent from a directory outside a Spectra project.
  **Expected**: Root command exits with code 1, prints `"Error: .spectra directory not found. Are you in a Spectra project?"` to stderr.

- **Condition**: Subcommand returns exit code 2 (transport error).
  **Expected**: Root command propagates exit code 2 without modification.

- **Condition**: Subcommand returns exit code 3 (Runtime execution error).
  **Expected**: Root command propagates exit code 3 without modification.

## Related

- [event emit Subcommand](./event_emit.md) - Event emission to the workflow runtime
- [error Subcommand](./error.md) - Error reporting to the workflow runtime
- [SocketClient](./client.md) - Shared socket communication logic
- [SpectraFinder](../../storage/spectra_finder.md) - Locates project root directory
- [StorageLayout](../../storage/storage_layout.md) - Provides socket path composition
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - System architecture overview
