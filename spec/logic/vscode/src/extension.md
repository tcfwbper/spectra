# Extension Entry Point

## Overview

Activation entry point for the Spectra VS Code extension. Assembles all components, wires event subscriptions, registers the WebviewViewProvider, and manages extension lifecycle. Does not own business logic, state computation, or I/O — acts purely as a composition root and message router.

## Boundaries

- Owns: creating the logger (OutputChannel wrapper), resolving project root, constructing controllers and view provider, wiring production dependencies for controllers (watchers, scanners, launchers, terminators, dispatchers), wiring `onDidUpdate` / `onDidError` / `onDidReceiveMessage` subscriptions, caching last-known `SessionListState`, routing webview messages to the appropriate controller method, registering `SpectraViewProvider` with VS Code, registering the `spectra.openPanel` command, and pushing all disposables to `context.subscriptions`.
- Delegates: project root resolution to `ProjectRootResolver.resolve()`.
- Delegates: session list state assembly to `SessionListController`.
- Delegates: session detail state assembly to `SessionDetailController`.
- Delegates: webview view lifecycle to `SpectraViewProvider`.
- Delegates: error display to `vscode.window.showErrorMessage`.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes directly (delegates to `SessionLauncher` / `SessionTerminator` via controller deps).
- Must not: interpret or validate state content — forwards as-is.
- Must not: hold references to watchers, scanners, or launchers directly — passes them as controller constructor deps.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.window.createOutputChannel` | Output channel factory | Create channel with name `'Spectra'` | Must not write to `console.log` |
| `vscode.workspace` | Workspace state | Pass to `ProjectRootResolver.resolve()`, `createFileSystemWatcher()`, `getConfiguration("spectra")` | Must not read files directly |
| `vscode.RelativePattern` | Glob pattern construction | Construct via `new vscode.RelativePattern(base, pattern)` for watcher deps | — |
| `vscode.EventEmitter` | Event emitter factory | Construct via `new vscode.EventEmitter()` for controller state/error emitters | — |
| `vscode.commands.registerCommand` | Command registration | Register `'spectra.openPanel'` command | — |
| `ProjectRootResolver` | Project root computation | `ProjectRootResolver.resolve(vscode.workspace)` | Must not call `isInitialized` |
| `SessionListController` | Session list state owner | Constructor (with full deps object), `onDidUpdate`, `onDidError`, `launch()`, `terminate()`, `dispose()` | Must not call internal scan methods |
| `SessionDetailController` | Session detail state owner | Constructor (with full deps object), `onDidUpdate`, `onDidError`, `open()`, `sendEvent()`, `dispose()` | Must not call internal scan methods |
| `SessionWatcher` | File system watcher for sessions | Construct and pass as controller dep | Must not subscribe to events directly |
| `WorkflowWatcher` | File system watcher for workflows | Construct and pass as controller dep | Must not subscribe to events directly |
| `EventWatcher` | File system watcher for events | Construct and pass as controller dep (via factory) | Must not subscribe to events directly |
| `SessionScanner` | Session directory scanner | Pass `scanSessions` as controller dep | Must not call directly |
| `WorkflowScanner` | Workflow directory scanner | Pass `scanWorkflows` as controller dep | Must not call directly |
| `SessionLauncher` | Process spawning for sessions | Pass `launch` as controller dep | Must not call directly |
| `SessionTerminator` | Process termination | Pass `terminate` as controller dep | Must not call directly |
| `EventScanner` | Event file scanner | Pass `scanEvents` as controller dep | Must not call directly |
| `EventDispatcher` | Event CLI dispatch | Pass `dispatch` as controller dep | Must not call directly |
| `WorkflowDefinitionParser` | Workflow YAML parser | Pass `parseWorkflowDefinition` as controller dep | Must not call directly |
| `SpectraViewProvider` | Webview sidebar transport | Constructor, `onDidReceiveMessage`, `showSessionList()`, `showSessionDetail()`, `dispose()` | Must not access `view.webview` directly |
| `vscode.window.registerWebviewViewProvider` | View provider registration | Register with viewType `'spectra.chatView'` | — |
| `vscode.window.showErrorMessage` | User-facing error display | Call with error message string | — |
| Logger (constructed internally) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- `activate(context, deps?)` is the single exported activation function. It accepts `IExtensionContext` as the first parameter (provided by VS Code) and an optional `deps` parameter that defaults to `{}`. VS Code only passes `context`, so `deps` is `{}` at runtime and production defaults are used. Tests may pass a mock `deps` object with any subset of fields.
- `deactivate()` is the single exported deactivation function.
- No class instantiation for the entry point — module-level exported functions.
- When `deps` does not supply output channel factories (neither `outputChannel` nor `createOutputChannel` is present), `activate` must `require("vscode")` and construct all collaborators internally using production implementations — must not throw or crash.
- When `deps` supplies output channel factories, `vscode` is not required — enables testing without the VS Code runtime.

## Behavior

### activate(context, deps?)

1. Determines whether production defaults are needed by checking if `deps` provides an `outputChannel` or `createOutputChannel` field. If neither is present, lazily requires the `vscode` module for production wiring.
2. Creates an `OutputChannel` named `'Spectra'` — uses `deps.outputChannel` if provided, else `deps.createOutputChannel('Spectra')` if provided, else `vscode.window.createOutputChannel('Spectra')`.
3. Wraps the OutputChannel in a logger adapter object providing `{ info, warn, error }` methods. Each method prepends a severity tag (`[INFO]`, `[WARN]`, `[ERROR]`) and delegates to `outputChannel.appendLine`.
4. Logs activation start via `logger.info("Spectra extension activating...")`.
5. Resolves project root — uses `deps.resolveProjectRoot()` if provided, else calls `ProjectRootResolver.resolve(vscode.workspace)`.
6. Resolves `showErrorMessage` function — uses `deps.showErrorMessage` if provided, else `vscode.window.showErrorMessage`.
7. If `projectRoot` is `undefined`: calls `showErrorMessage("Spectra: No workspace folder open.")`, logs the error, pushes `outputChannel` to `context.subscriptions`, and returns early. No controllers or ViewProvider are created.
8. Resolves `registerCommand` function — uses `deps.registerCommand` if provided, else `vscode.commands.registerCommand`.
9. Creates a `SessionListController(projectRoot, logger, controllerDeps)` instance. In production, `controllerDeps` wires:
    - `createSessionWatcher`: constructs a `SessionWatcher` with vscode watcher deps.
    - `createWorkflowWatcher`: constructs a `WorkflowWatcher` with vscode watcher deps.
    - `scanSessions`: delegates to `SessionScanner.scanSessions`.
    - `scanWorkflows`: delegates to `WorkflowScanner.scanWorkflows`.
    - `launch`: delegates to `SessionLauncher.launch` with `child_process.spawn`, `crypto.randomUUID`, and `spectra` config.
    - `terminate`: delegates to `SessionTerminator.terminate` with `process.kill`, `child_process.execFile`, and `spectra` config.
    - `createStateEmitter` / `createErrorEmitter`: construct `vscode.EventEmitter` instances.
10. Creates a `SessionDetailController(projectRoot, logger, controllerDeps)` instance. In production, `controllerDeps` wires:
    - `createEventWatcher`: constructs an `EventWatcher` with vscode watcher deps.
    - `scanEvents`: delegates to `EventScanner.scanEvents`.
    - `scanSessions`: delegates to `SessionScanner.scanSessions`.
    - `parseWorkflowDefinition`: delegates to `WorkflowDefinitionParser.parseWorkflowDefinition`.
    - `dispatchEvent`: delegates to `EventDispatcher.dispatch` with `child_process.execFile` and `spectra` config.
    - `createStateEmitter` / `createErrorEmitter`: construct `vscode.EventEmitter` instances.
11. Creates a `SpectraViewProvider(context.extensionUri, logger)` instance — uses `deps.createViewProvider` if provided.
12. Registers the view provider via `vscode.window.registerWebviewViewProvider('spectra.chatView', viewProvider, { webviewOptions: { retainContextWhenHidden: true } })` — uses `deps.registerWebviewViewProvider` if provided.
13. Initializes an internal `cachedSessionListState` variable to `null`.
14. Subscribes to `sessionListController.onDidUpdate`:
    - Stores the received state in `cachedSessionListState`.
    - Calls `viewProvider.showSessionList(state)`.
15. Subscribes to `sessionDetailController.onDidUpdate`:
    - Calls `viewProvider.showSessionDetail(state)`.
16. Subscribes to `sessionListController.onDidError`:
    - Calls `showErrorMessage(error.message)`.
17. Subscribes to `sessionDetailController.onDidError`:
    - Calls `showErrorMessage(error.message)`.
18. Subscribes to `viewProvider.onDidReceiveMessage` and routes based on `msg.command`:
    - `'navigateToDetail'`: calls `sessionDetailController.open(msg.sessionId, msg.workflowName)`.
    - `'navigateToList'`: if `cachedSessionListState` is not null, calls `viewProvider.showSessionList(cachedSessionListState)`.
    - `'launchSession'`: calls `sessionListController.launch(msg.workflowName)`.
    - `'terminateSession'`: calls `sessionListController.terminate(msg.pid)`.
    - `'sendEvent'`: calls `sessionDetailController.sendEvent(msg.eventType, msg.message)`.
    - Any other command: logs a warning via `logger.warn` with the unrecognized command value.
19. Registers the `spectra.openPanel` command via `registerCommand` — the handler is a no-op (view is managed by VS Code sidebar).
20. Pushes all disposables to `context.subscriptions`: the OutputChannel, sessionListController, sessionDetailController, viewProvider, the view provider registration, the command disposable, and all subscription disposables.
21. Logs successful activation via `logger.info` including the resolved `projectRoot`.

### deactivate()

22. Empty function body. All cleanup is handled by `context.subscriptions` disposal.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| context | `IExtensionContext` | Object with `subscriptions: IDisposable[]` and `extensionUri` | Yes (activate) |
| deps | `ActivateDeps \| ExtensionDeps \| {}` | Optional object providing collaborator factories/instances for testing. When `{}` or fields are missing, production defaults are used for missing fields. | No (activate) |

## Outputs

| Field | Type | Description |
|---|---|---|
| (none) | void | `activate` and `deactivate` return void. |

## Invariants

- `activate` must accept `context: IExtensionContext` as its first parameter and an optional `deps` as its second parameter (defaulting to `{}`). VS Code calls `activate(context)` with no additional arguments. Must not crash when `deps` is `{}` or `undefined`.
- When `deps` does not supply output channel factories, must `require("vscode")` and construct all collaborators (logger, controllers, view provider, watchers, scanners, launchers, terminators) internally using production implementations.
- When `deps` supplies collaborator factories (test scenarios), must use the supplied implementations and must not `require("vscode")`.
- Must not proceed past step 8 (creating controllers) if `projectRoot` is `undefined`.
- Must register all disposables with `context.subscriptions` — no manual cleanup in `deactivate()`.
- Must cache the latest `SessionListState` for immediate replay on `navigateToList`.
- Must route `terminateSession` to `sessionListController.terminate()` regardless of which page originated the message.
- Must log a warning (not throw) for unrecognized webview message commands.
- Must use the OutputChannel for all diagnostic output — never `console.log`.
- Must show errors via `showErrorMessage` when controllers fire `onDidError`.
- Production wiring for controller deps must use `vscode.workspace.getConfiguration("spectra")` for configuration access — reads `binaryPath` and `agentBinaryPath` settings.
- Production wiring must use `child_process.spawn` (detached) for session launching and `child_process.execFile` for termination verification and event dispatch.
- The `spectra.openPanel` command handler must be a no-op — the sidebar view is managed by VS Code via the registered WebviewViewProvider.

## Edge Cases

- Condition: `ProjectRootResolver.resolve()` returns `undefined` (no workspace open).
  Expected: Shows error message "Spectra: No workspace folder open." via `showErrorMessage`. Logs error. Only OutputChannel pushed to subscriptions. No controllers, no ViewProvider created. Returns early.

- Condition: `navigateToList` received before the first `onDidUpdate` from SessionListController (cachedSessionListState is null).
  Expected: No-op — `viewProvider.showSessionList` is not called. The webview remains on whatever page it currently shows until the first scan completes.

- Condition: `terminateSession` arrives from the detail page with a `pid` value.
  Expected: Routes identically to a `terminateSession` from the list page — calls `sessionListController.terminate(msg.pid)`.

- Condition: Controller `onDidUpdate` fires but the view has not yet been resolved (sidebar not yet visible).
  Expected: `viewProvider.showSessionList`/`showSessionDetail` stores the message as pending internally. The state is still cached in `cachedSessionListState`. When the view resolves, the pending message is delivered.

- Condition: Multiple `onDidError` events fire rapidly from both controllers.
  Expected: Each triggers an independent `showErrorMessage` call. VS Code queues/stacks notifications as needed.

- Condition: `deps` provides some fields but not others (partial deps).
  Expected: Provided fields are used; missing fields fall back to production defaults. The `??` chain resolves each field independently.

## Related

- [ProjectRootResolver](./services/projectRootResolver.md) — Resolves project root from workspace and configuration.
- [SessionListController](./controllers/sessionListController.md) — Manages session list state and actions.
- [SessionDetailController](./controllers/sessionDetailController.md) — Manages session detail state and actions.
- [SpectraViewProvider](./views/spectraViewProvider.md) — WebviewViewProvider for the sidebar.
- [SessionWatcher](./services/sessionWatcher.md) — Monitors session.json files for changes.
- [WorkflowWatcher](./services/workflowWatcher.md) — Monitors workflow YAML files for changes.
- [EventWatcher](./services/eventWatcher.md) — Monitors events.jsonl for changes.
- [SessionScanner](./services/sessionScanner.md) — Scans session directories.
- [WorkflowScanner](./services/workflowScanner.md) — Scans workflow directory.
- [SessionLauncher](./services/sessionLauncher.md) — Spawns workflow session processes.
- [SessionTerminator](./services/sessionTerminator.md) — Terminates session processes.
- [EventScanner](./services/eventScanner.md) — Reads events.jsonl files.
- [EventDispatcher](./services/eventDispatcher.md) — Dispatches events via CLI.
- [WorkflowDefinitionParser](./services/workflowDefinitionParser.md) — Parses workflow YAML definitions.
