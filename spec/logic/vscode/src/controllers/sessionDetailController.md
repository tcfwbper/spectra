# SessionDetailController

## Overview

Manages state and interactions for a single Session detail view. Orchestrates EventWatcher, EventScanner, SessionScanner, WorkflowDefinitionParser, and EventDispatcher to maintain a `SessionDetailState` and push updates to subscribers. Does not read/write files or spawn processes directly — all I/O is delegated to collaborators.

## Boundaries

- Owns: constructing and disposing EventWatcher instances (one per open session), subscribing to `onDidChange`, triggering scans via EventScanner and SessionScanner, parsing event types and entry node via WorkflowDefinitionParser, assembling and pushing `SessionDetailState` (including `entryNode`, `currentState`, `status`), coalescing overlapping scan requests (dirty-flag mechanism), maintaining a generation counter to discard stale scan results, dispatching events via EventDispatcher, and suppressing callbacks after dispose.
- Delegates: filesystem watching and debounce to EventWatcher.
- Delegates: event file reading/parsing to EventScanner (static method).
- Delegates: workflow definition parsing to WorkflowDefinitionParser (static method).
- Delegates: CLI process spawning to EventDispatcher (static method).
- Delegates: diagnostic logging to the injected logger.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes directly.
- Must not: display UI elements (error display is the subscriber's responsibility).
- Must not: manage file-watcher debounce logic (owned by EventWatcher).
- Must not: catch errors thrown during EventWatcher construction — lets them propagate to the caller.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `EventWatcher` | File change notification for a single session | Constructor (`new EventWatcher(projectRoot, sessionId)`), `onDidChange`, `dispose()` | Must not call scan or read files through it |
| `EventScanner` | Event data reader | `EventScanner.scan(projectRoot, sessionId, logger)` | Must not instantiate |
| `SessionScanner` | Session metadata reader | `SessionScanner.scan(projectRoot, logger)` | Must not instantiate |
| `WorkflowDefinitionParser` | Workflow definition reader | `WorkflowDefinitionParser.parse(projectRoot, workflowName, logger)` | Must not instantiate |
| `EventDispatcher` | Event CLI dispatcher | `EventDispatcher.dispatch(eventType, sessionId, message, projectRoot, logger)` | Must not instantiate or retain references |
| `vscode.EventEmitter<SessionDetailState>` | State push channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.EventEmitter<Error>` | Error push channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| Logger (`{ info, warn, error }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- Instantiated via `new SessionDetailController(projectRoot, logger)`.
- Internally constructs two `vscode.EventEmitter` instances (for state and error events) during construction.
- Does NOT construct an EventWatcher during construction — EventWatcher is created lazily in `open()`.
- Implements `vscode.Disposable`.

## Behavior

### Construction

1. Stores `projectRoot` and `logger`.
2. Creates a `vscode.EventEmitter<SessionDetailState>` and exposes its `.event` as `onDidUpdate`.
3. Creates a `vscode.EventEmitter<Error>` and exposes its `.event` as `onDidError`.
4. Initializes internal state: `currentWatcher` as `null`, `currentSessionId` as `null`, `currentWorkflowName` as `null`, `generation` as `0`.
5. Initializes the dirty flag to `false` and scanning flag to `false`.
6. Sets the disposed flag to `false`.

### open(sessionId, workflowName)

7. If disposed, returns immediately (no-op).
8. If `currentWatcher` is not null, calls `currentWatcher.dispose()`.
9. Increments `generation` by 1. Captures the new value as `openGeneration`.
10. Stores `sessionId` as `currentSessionId` and `workflowName` as `currentWorkflowName`.
11. Resets dirty flag to `false` and scanning flag to `false`.
12. Constructs a new `EventWatcher(projectRoot, sessionId)`. If construction throws, the error propagates to the caller (no catch).
13. Stores the new instance as `currentWatcher`.
14. Subscribes to `currentWatcher.onDidChange` → calls the internal scan routine.
15. Calls `WorkflowDefinitionParser.parse(projectRoot, workflowName, logger)` to obtain the `WorkflowParseResult` (`entryNode` and `eventTypes`).
16. Calls `EventScanner.scan(projectRoot, sessionId, logger)` to load historical events.
17. Calls `SessionScanner.scan(projectRoot, logger)` to load all session summaries, then finds the matching session by `sessionId` to extract `currentState`, `status`, and `pid`.
18. After all calls resolve: checks if `generation` still equals `openGeneration`. If not, discards results and returns (a newer open() has taken over).
19. Stores `entryNode` from the parse result for use in subsequent scan routines.
20. Assembles `SessionDetailState` from the results (including `entryNode`, `currentState`, `status`, `pid`, `eventTypes`, `events`) and fires `onDidUpdate`.

### Internal Scan Routine (triggered by onDidChange)

21. If disposed, returns immediately.
22. Checks if `generation` matches the generation captured at subscription time. If not, returns (watcher is stale).
23. If a scan is already in-flight, sets the dirty flag to `true` and returns immediately.
24. Sets scanning flag to `true`.
25. Captures the current `generation` as `scanGeneration`.
26. Calls `EventScanner.scan(projectRoot, currentSessionId, logger)`.
27. Calls `SessionScanner.scan(projectRoot, logger)` and finds the matching session by `currentSessionId` to extract `currentState`, `status`, and `pid`.
28. After both calls resolve: checks if `generation` still equals `scanGeneration`. If not, sets scanning flag to `false` and returns.
29. Assembles `SessionDetailState` using the new events, previously stored `entryNode` and `eventTypes`, and freshly read `currentState`, `status`, and `pid`, then fires `onDidUpdate`.
30. Sets scanning flag to `false`.
31. If the dirty flag is `true`, resets it to `false` and re-invokes this routine (loop until clean).

### sendEvent(eventType, message) → Promise\<boolean\>

32. If disposed, returns `false` immediately.
33. Calls `EventDispatcher.dispatch(eventType, currentSessionId, message, projectRoot, logger)`.
34. If the call throws (spawn failure — ENOENT, EACCES), catches the error, logs via `logger.error`, fires `onDidError` with the caught error, and returns `false`.
35. If the call succeeds, returns `true`. No other immediate action (the EventWatcher will detect the resulting file change and trigger a re-scan).

### Dispose

36. Sets the disposed flag to `true`.
37. If `currentWatcher` is not null, calls `currentWatcher.dispose()` and sets `currentWatcher` to null.
38. Disposes both `EventEmitter` instances.
39. All in-flight async operations that complete after dispose check the disposed flag and suppress any callback invocations (`fire()` calls).

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes (constructor) |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes (constructor) |
| sessionId | string | Non-empty | Yes (open method) |
| workflowName | string | Non-empty | Yes (open method) |
| eventType | string | Non-empty | Yes (sendEvent method) |
| message | string | Non-empty | Yes (sendEvent method) |

## Outputs

### sendEvent Return Value

| Value | Meaning |
|---|---|
| `true` | Dispatch succeeded. Caller may clear input state. |
| `false` | Dispatch failed or controller is disposed. Caller should preserve input state. |

| Field | Type | Description |
|---|---|---|
| onDidUpdate | `vscode.Event<SessionDetailState>` | Fires when session detail state changes (after open or onDidChange re-scan). |
| onDidError | `vscode.Event<Error>` | Fires for actionable errors (EventDispatcher spawn failure). |

### SessionDetailState Type

| Field | Type | Description |
|---|---|---|
| sessionId | string | The currently open session's ID. |
| workflowName | string | The workflow name associated with this session. |
| entryNode | string | The workflow's entry node name (from WorkflowDefinitionParser). |
| currentState | string | The session's current active node (from SessionScanner). |
| status | string | The session's current status: `'initializing' \| 'running' \| 'completed' \| 'failed'` (from SessionScanner). |
| pid | number | The OS process ID of the session (from SessionScanner). Required for the Stop button on the detail page. |
| eventTypes | string[] | Available event types for this session (from workflow definition, filtered by entryNode). |
| events | EventSummary[] | Historical events in file order. |

## Invariants

- Must implement `vscode.Disposable`.
- Must never fire `onDidUpdate` or `onDidError` after `dispose()` has been called.
- Must dispose the previous EventWatcher before creating a new one in `open()`.
- Must not catch errors thrown during EventWatcher construction — propagates to caller.
- Must use the generation counter to discard scan results from a superseded `open()` call.
- Must coalesce overlapping scan requests triggered by `onDidChange`: at most one scan in-flight, with at most one pending re-scan queued via dirty flag.
- Must fire `onDidError` for EventDispatcher spawn failures (ENOENT, EACCES).
- Must log via `logger.error` every time `onDidError` is fired.
- Must not fire events or invoke scans after disposed flag is set.
- At most one EventWatcher exists at any time (the current one).

## Edge Cases

- Condition: `open()` is called while a previous `open()`'s scan is still in-flight.
  Expected: The previous watcher is disposed, generation increments. When the old scan resolves, its generation check fails and results are discarded. The new open proceeds independently.

- Condition: `onDidChange` fires while a scan is already in-flight.
  Expected: Sets dirty flag. After the current scan completes and passes generation check, a fresh scan runs. Only one re-scan is queued regardless of how many onDidChange events fired.

- Condition: `onDidChange` fires after `open()` replaced the watcher but before the old subscription callback runs.
  Expected: The generation check in the scan routine detects mismatch and discards the stale invocation.

- Condition: `sendEvent()` is called when `currentSessionId` is null (before any `open()` call).
  Expected: The call proceeds with null sessionId — EventDispatcher will spawn the CLI with invalid arguments. The CLI may fail with non-zero exit (logged as warning by EventDispatcher). No onDidError is fired because spawn itself succeeds.

- Condition: `sendEvent()` is called after `dispose()`.
  Expected: Returns immediately (no-op). No dispatch, no error event.

- Condition: `open()` is called after `dispose()`.
  Expected: Returns immediately (no-op). No watcher created, no scan performed.

- Condition: `dispose()` is called while a scan from `onDidChange` is in-flight.
  Expected: The scan completes internally but the disposed guard suppresses the `onDidUpdate` fire.

- Condition: `WorkflowDefinitionParser.parse()` returns an empty array (no matching transitions).
  Expected: `eventTypes` in the pushed state is an empty array. No error fired.

- Condition: `EventScanner.scan()` returns an empty array (file missing or no valid lines).
  Expected: `events` in the pushed state is an empty array. No error fired.

- Condition: `dispose()` is called multiple times.
  Expected: First call performs cleanup. Subsequent calls are no-ops (watcher is already null, emitters already disposed).

- Condition: `SessionScanner.scan()` returns an array that does not contain a session matching `currentSessionId`.
  Expected: `currentState` defaults to empty string, `status` defaults to `'initializing'`, and `pid` defaults to `0` in the pushed state. No error fired.

- Condition: `WorkflowDefinitionParser.parse()` returns failure result (empty `entryNode` and empty `eventTypes`).
  Expected: `entryNode` in the pushed state is empty string, `eventTypes` is empty array. No error fired. The send-button in the webview will remain disabled because the guard condition (`currentState === entryNode`) cannot meaningfully match.

## Related

- [SessionListController](./sessionListController.md) — Sibling controller managing the session list view; follows the same EventEmitter + dirty-flag patterns.
- [EventWatcher](../services/eventWatcher.md) — Provides `onDidChange` events for a single session's event file.
- [EventScanner](../services/eventScanner.md) — Reads and returns event summaries from a session's events.jsonl.
- [SessionScanner](../services/sessionScanner.md) — Reads session metadata to provide `currentState` and `status`.
- [WorkflowDefinitionParser](../services/workflowDefinitionParser.md) — Parses workflow YAML to extract `entryNode` and available event types.
- [EventDispatcher](../services/eventDispatcher.md) — Spawns the spectra-agent CLI to emit events.
