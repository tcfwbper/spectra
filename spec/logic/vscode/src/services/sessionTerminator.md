# SessionTerminator

## Overview

Provides a static method that gracefully terminates a Spectra runtime process identified by its PID. Before sending any signal, it verifies that the process is alive and its command name matches the expected Spectra binary — preventing accidental termination of unrelated processes after PID reuse (e.g., machine restart). Uses a SIGTERM → grace period → SIGKILL escalation strategy.

## Boundaries

- Owns: PID liveness check, process command name verification, SIGTERM delivery, grace period wait, SIGKILL escalation, and reporting the termination result.
- Delegates: user-facing error display to the caller (command/controller layer).
- Delegates: session status update to the Spectra runtime itself (runtime writes final status to `session.json` upon receiving SIGTERM).
- Must not: read or write any file (session state persistence is owned by the runtime).
- Must not: display UI elements.
- Must not: hold any instance state — this is a stateless static method.
- Must not: terminate processes whose command name does not match the expected binary.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | Configuration provider | `getConfiguration('spectra').get<string>('binaryPath')` | Must not write configuration |
| Node.js `process` | Signal sender | `process.kill(pid, 0)`, `process.kill(pid, 'SIGTERM')`, `process.kill(pid, 'SIGKILL')` | — |
| Node.js `child_process` | Command verifier | `execFile('ps', ...)` to query process command | Must not spawn spectra processes |
| Logger (`{ info, warn, error }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraint: SessionTerminator is a class with a single static async method. No instantiation required.

## Behavior

1. Reads `spectra.binaryPath` from VS Code configuration. If the value is falsy, defaults to `"spectra"`.
2. Checks whether the process with the given PID is alive by calling `process.kill(pid, 0)`.
3. If the process is not alive (throws ESRCH), returns a result indicating `already_dead`.
4. Verifies the process command name by executing `ps -p <pid> -o comm=` via `execFile`.
5. Extracts the command name from `ps` output (trimmed). Compares it against both the configured binary path's basename and the literal string `"spectra"`. If neither matches, returns a result indicating `not_spectra` — the process is not terminated.
6. Sends SIGTERM to the process via `process.kill(pid, 'SIGTERM')`.
7. Logs an info message via `logger.info` indicating SIGTERM was sent.
8. Waits for a 5-second grace period, polling every 500ms whether the process is still alive (via `process.kill(pid, 0)`).
9. If the process dies during the grace period, returns a result indicating `terminated` with method `sigterm`.
10. If the process is still alive after the grace period, sends SIGKILL via `process.kill(pid, 'SIGKILL')`.
11. Logs a warn message via `logger.warn` indicating SIGKILL escalation.
12. Waits briefly (500ms) and confirms the process is dead.
13. Returns a result indicating `terminated` with method `sigkill`.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| pid | number | Positive integer | Yes |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<TerminationResult>` | Structured result describing the outcome. |

### TerminationResult Type

| Field | Type | Description |
|---|---|---|
| terminated | boolean | Whether the process was successfully terminated (or was already dead). |
| method | `'sigterm' \| 'sigkill' \| 'already_dead' \| 'not_spectra'` | How the termination was achieved, or why it was skipped. |
| error | string \| undefined | Descriptive error message if an unexpected failure occurred (e.g., EPERM on signal). |

## Invariants

- Must read `spectra.binaryPath` on every invocation (never cache).
- Must default to `"spectra"` when the configuration value is falsy.
- Must never send a signal to a process whose command name does not match the expected binary or literal `"spectra"`.
- Must always attempt SIGTERM before SIGKILL (never skip directly to SIGKILL).
- Grace period between SIGTERM and SIGKILL is 5 seconds (hardcoded).
- Must be a static async method (no instance state required).
- Must not throw exceptions to the caller — all outcomes are expressed via the `TerminationResult` structure.
- If `process.kill` fails with EPERM (permission denied), returns a result with `terminated: false` and a descriptive `error` string.

## Edge Cases

- Condition: The PID does not exist (process already dead before method is called).
  Expected: Returns `{ terminated: true, method: 'already_dead' }`.

- Condition: The PID exists but belongs to a non-spectra process (PID reuse after reboot).
  Expected: Returns `{ terminated: false, method: 'not_spectra' }`. No signal is sent.

- Condition: The PID exists, command matches, but the caller lacks permission to signal it (EPERM).
  Expected: Returns `{ terminated: false, method: 'sigterm', error: '<descriptive message>' }`.

- Condition: The process dies between the liveness check and the SIGTERM delivery.
  Expected: `process.kill(pid, 'SIGTERM')` throws ESRCH. Returns `{ terminated: true, method: 'already_dead' }`.

- Condition: The process dies during the grace period (responds to SIGTERM).
  Expected: Polling detects death. Returns `{ terminated: true, method: 'sigterm' }`.

- Condition: The process ignores SIGTERM and survives the full 5-second grace period.
  Expected: SIGKILL is sent. Returns `{ terminated: true, method: 'sigkill' }`.

- Condition: `ps -p <pid> -o comm=` fails (e.g., `ps` not available on the system).
  Expected: Logs an error via `logger.error`. Returns `{ terminated: false, method: 'not_spectra', error: '<descriptive message>' }`. Errs on the side of caution — does not terminate when identity cannot be verified.

- Condition: The configured `spectra.binaryPath` is `/usr/local/bin/spectra` and `ps` reports command as `spectra`.
  Expected: Basename of configured path is `spectra`, which matches `ps` output. Proceeds with termination.

- Condition: The configured `spectra.binaryPath` is a custom name like `spectra-dev` and `ps` reports `spectra-dev`.
  Expected: Basename matches. Proceeds with termination.

## Related

- [SessionLauncher](./sessionLauncher.md) — Counterpart service that spawns sessions.
- [SessionWatcher](./sessionWatcher.md) — Detects when the runtime updates `session.json` status after receiving SIGTERM.
- [SessionScanner](./sessionScanner.md) — Provides the PID value that the caller passes to this method.
