# Extension Entry Point

## Overview

Activation entry point for the Spectra VS Code extension. Assembles all components, wires event subscriptions, registers commands, and manages extension lifecycle. Does not own business logic, state computation, or I/O — acts purely as a composition root and message router.

## Boundaries

- Owns: creating the logger (OutputChannel wrapper), resolving project root, constructing controllers and panel, wiring `onDidUpdate` / `onDidError` / `onDidReceiveMessage` / `onDidDispose` subscriptions, caching last-known `SessionListState`, routing webview messages to the appropriate controller method, registering the `spectra.openPanel` command, and pushing all disposables to `context.subscriptions`.
- Delegates: project root resolution to `ProjectRootResolver.resolve()`.
- Delegates: session list state assembly to `SessionListController`.
- Delegates: session detail state assembly to `SessionDetailController`.
- Delegates: webview panel lifecycle to `SpectraPanel.createOrReveal()`.
- Delegates: error display to `vscode.window.showErrorMessage`.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes.
- Must not: interpret or validate state content — forwards as-is.
- Must not: hold references to watchers, scanners, or launchers directly.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.window.createOutputChannel` | Output channel factory | Create channel with name `'Spectra'` | Must not write to `console.log` |
| `ProjectRootResolver` | Project root computation | `ProjectRootResolver.resolve()` | Must not call any other method |
| `SessionListController` | Session list state owner | Constructor, `onDidUpdate`, `onDidError`, `launch()`, `terminate()`, `dispose()` | Must not call internal scan methods |
| `SessionDetailController` | Session detail state owner | Constructor, `onDidUpdate`, `onDidError`, `open()`, `sendEvent()`, `dispose()` | Must not call internal scan methods |
| `SpectraPanel` | Webview transport | `createOrReveal()`, `showSessionList()`, `showSessionDetail()`, `onDidReceiveMessage`, `onDidDispose` | Must not access `panel.webview` directly |
| `vscode.window.showErrorMessage` | User-facing error display | Call with error message string | — |
| `vscode.commands.registerCommand` | Command registration | Register `spectra.openPanel` | — |
| Logger (constructed internally) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- `activate(context)` is the single exported activation function.
- `deactivate()` is the single exported deactivation function.
- No class instantiation — module-level exported functions.

## Behavior

### activate(context)

1. Creates an `OutputChannel` named `'Spectra'` via `vscode.window.createOutputChannel('Spectra')`.
2. Wraps the OutputChannel in a logger adapter object providing `{ info, warn, error }` methods. Each method prepends a severity tag and delegates to `outputChannel.appendLine`.
3. Logs activation start via `logger.info`.
4. Calls `ProjectRootResolver.resolve()` to obtain `projectRoot`.
5. If `projectRoot` is `undefined`, calls `vscode.window.showErrorMessage` with a descriptive message (e.g., "Spectra: No workspace folder open."), logs the error, and returns immediately without registering any commands or disposables (aside from the OutputChannel itself pushed to subscriptions for cleanup).
6. Creates a `SessionListController(projectRoot, logger)` instance.
7. Creates a `SessionDetailController(projectRoot, logger)` instance.
8. Calls `SpectraPanel.createOrReveal(context, context.extensionUri, logger)` to obtain the panel instance.
9. Initializes an internal `cachedSessionListState` variable to `null`.
10. Subscribes to `sessionListController.onDidUpdate`:
    - Stores the received state in `cachedSessionListState`.
    - Calls `panel.showSessionList(state)`.
11. Subscribes to `sessionDetailController.onDidUpdate`:
    - Calls `panel.showSessionDetail(state)`.
12. Subscribes to `sessionListController.onDidError`:
    - Calls `vscode.window.showErrorMessage(error.message)`.
13. Subscribes to `sessionDetailController.onDidError`:
    - Calls `vscode.window.showErrorMessage(error.message)`.
14. Subscribes to `panel.onDidReceiveMessage` and routes based on `msg.command`:
    - `'navigateToDetail'`: calls `sessionDetailController.open(msg.sessionId, msg.workflowName)`.
    - `'navigateToList'`: if `cachedSessionListState` is not null, calls `panel.showSessionList(cachedSessionListState)`.
    - `'launchSession'`: calls `sessionListController.launch(msg.workflowName)`.
    - `'terminateSession'`: calls `sessionListController.terminate(msg.pid)`.
    - `'sendEvent'`: calls `sessionDetailController.sendEvent(msg.eventType, msg.message)`.
    - Any other command: logs a warning via `logger.warn` with the unrecognized command value.
15. Subscribes to `panel.onDidDispose`:
    - Calls `sessionListController.dispose()`.
    - Calls `sessionDetailController.dispose()`.
16. Registers the `spectra.openPanel` command via `vscode.commands.registerCommand`:
    - Handler calls `SpectraPanel.createOrReveal(context, context.extensionUri, logger)`.
17. Pushes all disposables to `context.subscriptions`: the OutputChannel, sessionListController, sessionDetailController, panel, the command registration, and all subscription disposables.
18. Logs successful activation via `logger.info` including the resolved `projectRoot`.

### deactivate()

19. Empty function body. All cleanup is handled by `context.subscriptions` disposal.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| context | `vscode.ExtensionContext` | Valid extension context provided by VS Code | Yes (activate) |

## Outputs

| Field | Type | Description |
|---|---|---|
| (none) | void | `activate` and `deactivate` return void. |

## Invariants

- Must register all disposables with `context.subscriptions` — no manual cleanup in `deactivate()`.
- Must not proceed past step 5 if `projectRoot` is `undefined`.
- Must cache the latest `SessionListState` for immediate replay on `navigateToList`.
- Must route `terminateSession` to `sessionListController.terminate()` regardless of which page originated the message.
- Must dispose both controllers when the panel is disposed.
- Must log a warning (not throw) for unrecognized webview message commands.
- Must not fire controller methods after the panel's `onDidDispose` has triggered controller disposal.
- Must use the OutputChannel for all diagnostic output — never `console.log`.
- Must show errors via `vscode.window.showErrorMessage` when controllers fire `onDidError`.

## Edge Cases

- Condition: `ProjectRootResolver.resolve()` returns `undefined` (no workspace open).
  Expected: Shows error message, logs error, returns early. No commands registered, no controllers created. Only OutputChannel is pushed to subscriptions.

- Condition: `navigateToList` received before the first `onDidUpdate` from SessionListController (cachedSessionListState is null).
  Expected: No-op — `panel.showSessionList` is not called. The webview remains on whatever page it currently shows until the first scan completes.

- Condition: Panel is disposed by the user closing the tab, then `spectra.openPanel` command is invoked.
  Expected: `SpectraPanel.createOrReveal` creates a new panel. However, controllers were already disposed by the `onDidDispose` handler. The new panel will not receive updates. (Note: this is a known limitation — a full re-activation would require re-creating controllers. Document as future enhancement.)

- Condition: `terminateSession` arrives from the detail page with a `pid` value.
  Expected: Routes identically to a `terminateSession` from the list page — calls `sessionListController.terminate(msg.pid)`.

- Condition: Controller `onDidUpdate` fires after panel has been disposed.
  Expected: `panel.showSessionList`/`panel.showSessionDetail` on a disposed panel is a no-op per VS Code API behavior. No error thrown.

- Condition: `activate()` is called but `SpectraPanel.createOrReveal()` throws (e.g., internal VS Code error).
  Expected: The error propagates to VS Code's extension host, which handles activation failures. No explicit catch in activate().

- Condition: Multiple `onDidError` events fire rapidly from both controllers.
  Expected: Each triggers an independent `showErrorMessage` call. VS Code queues/stacks notifications as needed.

## Related

- [ProjectRootResolver](./services/projectRootResolver.md) — Resolves project root from workspace and configuration.
- [SessionListController](./controllers/sessionListController.md) — Manages session list state and actions.
- [SessionDetailController](./controllers/sessionDetailController.md) — Manages session detail state and actions.
- [SpectraPanel](./views/spectraPanel.md) — Singleton webview panel transport layer.
