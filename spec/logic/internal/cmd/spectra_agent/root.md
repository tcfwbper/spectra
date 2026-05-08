# spectra-agent Root Command

## Overview

The root command for the `spectra-agent` CLI binary. It provides the Cobra root command that handles global flag parsing (`--session-id`), project root discovery via SpectraFinder, and dispatches to subcommands (`event emit`, `error`). It does not perform any socket operations or business logic directly.

## Boundaries

- Owns: Cobra root command definition and global flag registration.
- Owns: `--session-id` flag presence validation (must be non-empty).
- Owns: project root discovery invocation (SpectraFinder).
- Owns: propagation of sessionID and projectRoot to subcommands via Cobra's context or persistent pre-run.
- Owns: exit code propagation from subcommands to the process.
- Owns: usage/help text rendering.
- Delegates: socket communication to subcommands (which use `cmdutil.SendAndHandle`).
- Delegates: project root directory discovery to `storage.SpectraFinder`.
- Delegates: subcommand-specific argument parsing to each subcommand.
- Must not: perform socket operations.
- Must not: validate session ID format (UUID validation is the Runtime's responsibility).
- Must not: read environment variables (`SPECTRA_SESSION_ID`, etc.). Flags are the only input mechanism.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `storage.SpectraFinder` | Project root discovery | `FindProjectRoot(startDir)` | Must not use any other storage function |
| `cobra` (spf13/cobra) | CLI framework | Define commands, register flags, execute | — |

Construction constraint: Exposes a function (e.g., `Execute() int`) that builds and runs the Cobra command tree. Returns the process exit code. The caller (main.go) calls `os.Exit` with this value.

## Behavior

1. Defines a Cobra root command `spectra-agent` with usage `"spectra-agent [command]"`.
2. Registers a persistent string flag `--session-id` on the root command (available to all subcommands).
3. In `PersistentPreRunE`, validates that `--session-id` is non-empty. If empty or not provided, returns error `"--session-id flag is required"` (Cobra prints to stderr with exit code 1).
4. In `PersistentPreRunE`, calls `storage.SpectraFinder.FindProjectRoot("")` (empty string = use CWD). If it returns `ErrNotInitialized`, returns error `".spectra directory not found. Are you in a Spectra project?"`.
5. Stores the resolved `sessionID` and `projectRoot` in a shared context accessible by subcommands.
6. Registers subcommands: `event` (with nested `emit`) and `error`.
7. If invoked without a subcommand, Cobra prints usage information and exits with code 0.
8. If invoked with an unknown subcommand, Cobra prints error and exits with code 1.
9. The `Execute()` function captures Cobra's exit behavior and returns the appropriate exit code to the caller.

## Inputs

| Input | Type | Source | Required |
|-------|------|--------|----------|
| `--session-id` | string | CLI flag (persistent) | Yes |
| Current Working Directory | string | Process CWD (implicit, passed to SpectraFinder) | Yes (implicit) |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| Exit code | int | 0 (success/help), 1 (invocation error), 2 (transport error), 3 (runtime error) |
| stdout | string | Usage/help text when no subcommand; success messages from subcommands |
| stderr | string | Error messages prefixed with `"Error: "` |

## Invariants

1. **Mandatory Session ID**: `--session-id` must be validated as non-empty before any subcommand runs.
2. **Project Root Discovery**: SpectraFinder must succeed before any subcommand runs.
3. **Exit Code Consistency**: All subcommands use the standardized exit codes (0, 1, 2, 3). The root command propagates them unchanged.
4. **No Environment Variable Reads**: The root command does not read `SPECTRA_SESSION_ID` or any other environment variable.
5. **Stateless Execution**: Each invocation is independent.
6. **Error Prefix**: All error messages printed to stderr are prefixed with `"Error: "` (Cobra's default behavior with `SilenceErrors` disabled or custom error handling).

## Edge Cases

- Condition: `--session-id` is not provided.
  Expected: Exit code 1, stderr `"Error: --session-id flag is required"`.

- Condition: `--session-id ""` (empty string).
  Expected: Exit code 1, stderr `"Error: --session-id flag is required"`.

- Condition: `--session-id` is an invalid UUID (e.g., `"not-a-uuid"`).
  Expected: Accepted. Root command does not validate UUID format.

- Condition: CWD is outside a Spectra project (no `.spectra/` in ancestors).
  Expected: Exit code 1, stderr `"Error: .spectra directory not found. Are you in a Spectra project?"`.

- Condition: Invoked without subcommand (`spectra-agent --session-id <UUID>`).
  Expected: Prints usage, exits with code 0.

- Condition: Invoked with `--help`.
  Expected: Prints usage, exits with code 0.

- Condition: Invoked with unknown subcommand (`spectra-agent foo`).
  Expected: Exit code 1, stderr with unknown command error.

- Condition: Subcommand returns exit code 2 or 3.
  Expected: Root command propagates exit code unchanged.

## Related

- [EventEmit](./event_emit.md) — event emit subcommand
- [ErrorCmd](./error_cmd.md) — error subcommand
- [SpectraFinder](../../../storage/spectra_finder.md) — locates project root
- [ExitCodes](../../cmdutil/exit_codes.md) — exit code constants
