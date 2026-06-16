# SessionLauncher

## Overview

Provides a static method that launches a new Spectra workflow session by spawning a detached `spectra run` CLI process. The spawned process is fully decoupled from the VS Code extension host — it survives extension reload, window close, and VS Code restart. The method's only job is to spawn successfully; session lifecycle tracking is owned by SessionWatcher and SessionScanner.

## Boundaries

- Owns: resolving `spectra.binaryPath` configuration, generating a UUIDv4 session identifier, spawning a detached child process with `unref()`, detecting spawn-level failures.
- Delegates: session lifecycle tracking (status, state, pid) to SessionWatcher and SessionScanner via the runtime-written `session.json`.
- Delegates: user-facing error display to the caller (command/controller layer).
- Must not: wait for the child process to exit or monitor its health after spawn.
- Must not: read or write any file (the runtime writes `session.json`).
- Must not: display UI elements.
- Must not: hold any instance state — this is a stateless static method.
- Must not: return pid or session metadata — the watcher/scanner pipeline provides this.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | Configuration provider | `getConfiguration('spectra').get<string>('binaryPath')` | Must not write configuration |
| Node.js `child_process` | Process spawner | `spawn` with `detached: true`, `stdio: 'ignore'` | Must not use `exec` or `execFile` (need detach semantics) |
| Node.js `crypto` | UUID generator | `crypto.randomUUID()` | — |
| Logger (`{ info, warn, error }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraint: SessionLauncher is a class with a single static async method. No instantiation required.

## Behavior

1. Reads `spectra.binaryPath` from VS Code configuration. If the value is falsy, defaults to `"spectra"`.
2. Generates a new UUIDv4 via `crypto.randomUUID()`.
3. Constructs the argument list: `["run", "--workflow", workflowName, "--session-id", generatedUUID]`.
4. Spawns the binary using `spawn` with options: `{ detached: true, stdio: 'ignore' }`.
5. Calls `unref()` on the child process to ensure the VS Code extension host does not wait for it.
6. Registers an `error` event handler on the child process. If the spawn fails (e.g., `ENOENT`), throws an error with a descriptive message including the binary path.
7. Logs an info message via `logger.info` recording the workflow name and generated session ID.
8. Resolves with `void` once spawn succeeds.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| workflowName | string | Non-empty | Yes |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<void>` | Resolves when spawn succeeds. Rejects if spawn itself fails. |

## Invariants

- Must read `spectra.binaryPath` on every invocation (never cache).
- Must default to `"spectra"` when the configuration value is falsy.
- Must spawn with `detached: true` and call `unref()` — the child process must outlive the extension host.
- Must use `stdio: 'ignore'` — no pipe connection to the extension host process.
- Must throw on spawn failure (ENOENT, EACCES, etc.) — these indicate configuration errors.
- Must be a static async method (no instance state required).
- Must generate a fresh UUIDv4 for every invocation — never reuse identifiers.
- Must not retain a reference to the child process after spawn and unref.

## Edge Cases

- Condition: `spectra.binaryPath` is configured to a non-existent path.
  Expected: Spawn fails with ENOENT; method throws an error including the configured path.

- Condition: `spectra.binaryPath` is configured to a path without execute permission.
  Expected: Spawn fails with EACCES; method throws an error.

- Condition: `spectra.binaryPath` configuration is set to an empty string.
  Expected: Treated as falsy; defaults to `"spectra"`.

- Condition: VS Code extension host is terminated immediately after spawn.
  Expected: The detached child process continues running independently.

- Condition: The spawned process crashes immediately after starting.
  Expected: SessionLauncher has already resolved. The runtime either writes a `failed` status to `session.json` (detected by SessionWatcher), or never writes `session.json` at all (session never appears in SessionScanner results).

- Condition: `workflowName` contains spaces or special characters.
  Expected: Passed as a single argv element (no shell); the runtime receives it verbatim.

## Related

- [SessionWatcher](./sessionWatcher.md) — Detects when the runtime writes `session.json` after successful startup.
- [SessionScanner](./sessionScanner.md) — Reads session metadata (id, pid, status, etc.) from `session.json`.
- [SessionTerminator](./sessionTerminator.md) — Counterpart service for stopping sessions.
- Architecture reference: `spectra run --workflow <WorkflowName> --session-id <UUID>` in ARCHITECTURE.md.
