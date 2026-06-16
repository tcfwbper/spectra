# SessionListController

## Overview

Manages all state and actions for the Sessions list view. Orchestrates SessionWatcher, WorkflowWatcher, SessionScanner, WorkflowScanner, SessionLauncher, and SessionTerminator to maintain a unified `SessionListState` and push updates to subscribers. Does not read/write files or spawn processes directly — all I/O is delegated to collaborators.

## Boundaries

- Owns: constructing and disposing SessionWatcher and WorkflowWatcher, subscribing to their `onDidChange` events, triggering scans and assembling state, coalescing overlapping scan requests (dirty-flag mechanism), firing `onDidUpdate` with current state, firing `onDidError` for actionable failures, classifying termination results, and guarding callbacks after dispose.
- Delegates: filesystem watching to SessionWatcher and WorkflowWatcher.
- Delegates: session data reading to SessionScanner (static method).
- Delegates: workflow listing to WorkflowScanner (static method).
- Delegates: process spawning to SessionLauncher (static method).
- Delegates: process termination to SessionTerminator (static method).
- Delegates: diagnostic logging to the injected logger.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes directly.
- Must not: display UI elements (error display is the subscriber's responsibility).
- Must not: manage watcher debounce logic (owned by the watchers themselves).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `SessionWatcher` | Session file change notification | Constructor, `onDidChange`, `dispose()` | Must not call scan or read files through it |
| `WorkflowWatcher` | Workflow file change notification | Constructor, `onDidChange`, `dispose()` | Must not call scan or read files through it |
| `SessionScanner` | Session data reader | `SessionScanner.scan(projectRoot, logger)` | Must not instantiate |
| `WorkflowScanner` | Workflow data reader | `WorkflowScanner.scan(projectRoot, logger)` | Must not instantiate |
| `SessionLauncher` | Process spawner | `SessionLauncher.launch(workflowName, logger)` | Must not instantiate or retain references |
| `SessionTerminator` | Process terminator | `SessionTerminator.terminate(pid, logger)` | Must not instantiate or send signals directly |
| `vscode.EventEmitter<SessionListState>` | State push channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.EventEmitter<Error>` | Error push channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| Logger (`{ info, warn, error }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- Instantiated via `new SessionListController(projectRoot, logger)`.
- Internally constructs `SessionWatcher` and `WorkflowWatcher` during construction.
- Internally constructs two `vscode.EventEmitter` instances (for state and error events).
- Implements `vscode.Disposable`.

## Behavior

### Construction

1. Stores `projectRoot` and `logger`.
2. Creates a `SessionWatcher(projectRoot)` instance.
3. Creates a `WorkflowWatcher(projectRoot)` instance.
4. Creates a `vscode.EventEmitter<SessionListState>` and exposes its `.event` as `onDidUpdate`.
5. Creates a `vscode.EventEmitter<Error>` and exposes its `.event` as `onDidError`.
6. Subscribes to `SessionWatcher.onDidChange` → calls the internal scan-sessions routine.
7. Subscribes to `WorkflowWatcher.onDidChange` → calls the internal scan-workflows routine.
8. Initializes internal state: `sessions` as empty array, `workflows` as empty array.
9. Initializes the dirty flag to `false` and scanning flag to `false`.
10. Kicks off an asynchronous initial scan of both sessions and workflows (does not block construction).

### Internal Scan Routine (sessions)

11. If a scan is already in-flight, sets the dirty flag to `true` and returns immediately.
12. Sets scanning flag to `true`.
13. Calls `SessionScanner.scan(projectRoot, logger)`.
14. Stores the result in internal `sessions` state.
15. Fires `onDidUpdate` with the current composite state.
16. Sets scanning flag to `false`.
17. If the dirty flag is `true`, resets it to `false` and re-invokes this routine (loop until clean).

### Internal Scan Routine (workflows)

18. Follows the same coalescing pattern as the sessions scan routine (steps 11–17) but with its own independent dirty/scanning flags.
19. Calls `WorkflowScanner.scan(projectRoot, logger)`.
20. Stores the result in internal `workflows` state.
21. Fires `onDidUpdate` with the current composite state.

### launch(workflowName)

22. Calls `SessionLauncher.launch(workflowName, logger)`.
23. If the call throws, logs the error via `logger.error` and fires `onDidError` with the caught error.
24. If the call succeeds, no immediate action is needed (the watcher will detect the new session.json).

### terminate(pid)

25. Calls `SessionTerminator.terminate(pid, logger)`.
26. Inspects the returned `TerminationResult`:
    - If `method` is `'already_dead'` or (`terminated` is `true` and method is `'sigterm'` or `'sigkill'`): treats as success, no error event.
    - If `method` is `'not_spectra'`: fires `onDidError` with an Error describing that the process no longer belongs to Spectra (PID reuse suspected). Logs via `logger.error`.
    - If `terminated` is `false` and `error` is present (EPERM): fires `onDidError` with an Error describing permission failure. Logs via `logger.error`.

### Dispose

27. Sets the disposed flag to `true`.
28. Disposes `SessionWatcher`.
29. Disposes `WorkflowWatcher`.
30. Disposes both `EventEmitter` instances.
31. All in-flight async operations that complete after dispose check the disposed flag and suppress any callback invocations (`fire()` calls).

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes (constructor) |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes (constructor) |
| workflowName | string | Non-empty | Yes (launch method) |
| pid | number | Positive integer | Yes (terminate method) |

## Outputs

| Field | Type | Description |
|---|---|---|
| onDidUpdate | `vscode.Event<SessionListState>` | Fires when session or workflow state changes. |
| onDidError | `vscode.Event<Error>` | Fires for actionable errors (launch failure, terminate permission/identity failure). |

### SessionListState Type

| Field | Type | Description |
|---|---|---|
| sessions | `SessionSummary[]` | Current known sessions, sorted by `createdAt` descending. |
| workflows | `string[]` | Current known workflow names. |

## Invariants

- Must implement `vscode.Disposable`.
- Must never fire `onDidUpdate` or `onDidError` after `dispose()` has been called.
- Must coalesce overlapping scan requests: at most one scan in-flight per scan type (sessions, workflows), with at most one pending re-scan queued via dirty flag.
- Must not await the initial scan during construction — construction is synchronous; the scan runs asynchronously.
- Must always push the full composite state (both `sessions` and `workflows`) on every `onDidUpdate` firing, even if only one changed.
- The sessions scan and workflows scan are independent — one does not block the other.
- Must fire `onDidError` for launch failures (ENOENT, EACCES) and terminate failures (not_spectra, EPERM).
- Must not fire `onDidError` for `already_dead` termination results.
- Must log via `logger.error` every time `onDidError` is fired.

## Edge Cases

- Condition: Watcher fires `onDidChange` while a scan of the same type is already in-flight.
  Expected: Sets dirty flag; after the current scan completes, a fresh scan runs automatically. Only one re-scan is queued regardless of how many events fired.

- Condition: Both SessionWatcher and WorkflowWatcher fire simultaneously.
  Expected: Both scan routines run concurrently and independently. Each fires `onDidUpdate` upon completion with the latest composite state.

- Condition: `launch()` is called after `dispose()`.
  Expected: The call to SessionLauncher still executes (no cancel mechanism), but any resulting `onDidError` is suppressed (not fired).

- Condition: `terminate()` is called after `dispose()`.
  Expected: The call to SessionTerminator still executes, but any resulting `onDidError` is suppressed.

- Condition: `dispose()` is called while the initial scan is in-flight.
  Expected: The scan completes internally but the `onDidUpdate` callback is suppressed by the disposed guard.

- Condition: `launch()` is called with a workflow name that does not exist as a `.yaml` file.
  Expected: SessionLauncher spawns the process regardless (the CLI validates workflow existence). If the CLI fails, the runtime never writes session.json — no session appears in subsequent scans.

- Condition: `terminate()` returns `already_dead`.
  Expected: Treated as success. No `onDidError` fired. The watcher will pick up the session.json status change if the runtime wrote a final state before dying.

- Condition: Two concurrent `terminate()` calls for the same PID.
  Expected: Both proceed independently through SessionTerminator. One may see `already_dead` if the other's SIGTERM took effect first. No coordination is required by this controller.

## Related

- [SessionWatcher](../services/sessionWatcher.md) — Provides `onDidChange` events for session file mutations.
- [WorkflowWatcher](../services/workflowWatcher.md) — Provides `onDidChange` events for workflow file mutations.
- [SessionScanner](../services/sessionScanner.md) — Reads and returns session summaries.
- [WorkflowScanner](../services/workflowScanner.md) — Reads and returns workflow names.
- [SessionLauncher](../services/sessionLauncher.md) — Spawns new workflow sessions.
- [SessionTerminator](../services/sessionTerminator.md) — Terminates running sessions.
