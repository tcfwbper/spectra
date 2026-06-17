# EventDispatcher

## Overview

Provides a static method that dispatches an event to the Spectra runtime by spawning an external `spectra-agent event emit` CLI process. This is a fire-and-forget operation for the happy path — the method does not wait for the child process to exit. However, if the spawn itself fails (e.g., binary not found), the method throws immediately so the caller can surface the error.

## Boundaries

- Owns: resolving `spectra.agentBinaryPath` configuration, constructing the CLI argument list, spawning the child process with `cwd: projectRoot`, detecting spawn-level failures.
- Delegates: session lifecycle management to the Spectra runtime.
- Delegates: user-facing error display to the caller (command/controller layer).
- Must not: wait for the child process to exit or capture its output on the happy path.
- Must not: read or write any file.
- Must not: display UI elements (showErrorMessage, showInformationMessage).
- Must not: hold any instance state — this is a stateless static method.
- Must not: validate the semantic correctness of eventType, sessionId, or message values.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | Configuration provider | `getConfiguration('spectra').get<string>('agentBinaryPath')` | Must not write configuration |
| Node.js `child_process` | Process spawner | `execFile` (no shell) | Must not use `exec` (shell injection risk) |
| Logger (`{ info(msg: string): void; warn(msg: string): void; error(msg: string): void }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraint: EventDispatcher is a class with a single static async method. No instantiation required.

## Behavior

1. Reads `spectra.agentBinaryPath` from VS Code configuration. If the value is falsy, defaults to `"spectra-agent"`.
2. Constructs the argument list: `["event", "emit", eventType, "--session-id", sessionId, "--message", message]`.
3. Spawns the binary using `execFile` (no shell) with the constructed arguments and option `{ cwd: projectRoot }`.
4. Logs an info message via `logger.info` recording the dispatched event type and session ID.
5. Registers an `error` event handler on the child process. If the spawn fails (e.g., `ENOENT` — binary not found), throws an error with a descriptive message including the binary path.
6. Does not await the child process exit. On the happy path, the method resolves as soon as spawn succeeds.
7. Registers an `exit` event handler on the child process for diagnostic purposes: if the process exits with a non-zero code, calls `logger.warn` with the exit code.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| eventType | string | Non-empty | Yes |
| sessionId | string | Non-empty | Yes |
| message | string | Non-empty | Yes |
| projectRoot | string | Non-empty, absolute path | Yes |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<void>` | Resolves when spawn succeeds. Rejects if spawn itself fails. |

## Invariants

- Must read `spectra.agentBinaryPath` on every invocation (never cache).
- Must default to `"spectra-agent"` when the configuration value is falsy.
- Must not use a shell to spawn the process (prevents shell injection).
- Must set `cwd: projectRoot` in execFile options — the spectra-agent CLI resolves `.spectra/` relative to its working directory.
- Must throw on spawn failure (ENOENT, EACCES, etc.) — these indicate configuration errors.
- Must not throw on non-zero exit code from the child process — only log a warning.
- Must be a static async method (no instance state required).
- Must not block on child process completion for the happy path.

## Edge Cases

- Condition: `spectra.agentBinaryPath` is configured to a non-existent path.
  Expected: Spawn fails with ENOENT; method throws an error including the configured path.

- Condition: `spectra.agentBinaryPath` is configured to a path without execute permission.
  Expected: Spawn fails with EACCES; method throws an error.

- Condition: The child process starts successfully but exits with a non-zero code.
  Expected: Method has already resolved. Logger receives a warning with the exit code.

- Condition: `spectra.agentBinaryPath` configuration is set to an empty string.
  Expected: Treated as falsy; defaults to `"spectra-agent"`.

- Condition: The `message` argument contains special characters (quotes, newlines, etc.).
  Expected: No shell interpretation occurs because `execFile` is used without a shell. Characters are passed as-is to the child process argv.

## Related

- [ProjectRootResolver](./projectRootResolver.md) — Caller may use this for context but EventDispatcher does not depend on it.
- [SessionWatcher](./sessionWatcher.md) — Events emitted via this dispatcher may cause state changes detected by SessionWatcher.
- Architecture reference: `spectra-agent event emit <type> --session-id <UUID> --message <message>` in ARCHITECTURE.md.
