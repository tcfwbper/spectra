# Extension Entry Point

## Overview

Activation entry point for the Spectra VS Code extension. Assembles all components, wires event subscriptions, registers the WebviewViewProvider, and manages extension lifecycle. Does not own business logic, state computation, or I/O — acts purely as a composition root and message router.

## Boundaries

- Owns: creating the logger (OutputChannel wrapper), resolving project root, checking project initialization, constructing controllers and view provider, wiring `onDidUpdate` / `onDidError` / `onDidReceiveMessage` subscriptions, caching last-known `SessionListState`, routing webview messages to the appropriate controller method, registering `SpectraViewProvider` with VS Code, and pushing all disposables to `context.subscriptions`.
- Delegates: project root resolution to `ProjectRootResolver.resolve()`.
- Delegates: project initialization check to `ProjectRootResolver.isInitialized()`.
- Delegates: session list state assembly to `SessionListController`.
- Delegates: session detail state assembly to `SessionDetailController`.
- Delegates: webview view lifecycle to `SpectraViewProvider`.
- Delegates: error display to `vscode.window.showErrorMessage`.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes.
- Must not: interpret or validate state content — forwards as-is.
- Must not: hold references to watchers, scanners, or launchers directly.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.window.createOutputChannel` | Output channel factory | Create channel with name `'Spectra'` | Must not write to `console.log` |
| `ProjectRootResolver` | Project root computation | `ProjectRootResolver.resolve()`, `ProjectRootResolver.isInitialized(projectRoot)` | Must not call any other method |
| `SessionListController` | Session list state owner | Constructor, `onDidUpdate`, `onDidError`, `launch()`, `terminate()`, `dispose()` | Must not call internal scan methods |
| `SessionDetailController` | Session detail state owner | Constructor, `onDidUpdate`, `onDidError`, `open()`, `sendEvent()`, `dispose()` | Must not call internal scan methods |
| `SpectraViewProvider` | Webview sidebar transport | Constructor, `onDidReceiveMessage`, `showSessionList()`, `showSessionDetail()`, `showNotInitialized()`, `dispose()` | Must not access `view.webview` directly |
| `vscode.window.registerWebviewViewProvider` | View provider registration | Register with viewType `'spectra.chatView'` | — |
| `vscode.window.showErrorMessage` | User-facing error display | Call with error message string | — |
| Logger (constructed internally) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- `activate(context, deps?)` is the single exported activation function. It accepts `vscode.ExtensionContext` as the first parameter (provided by VS Code) and an optional `deps` parameter that defaults to production implementations when omitted. VS Code only passes `context`, so `deps` is `undefined` at runtime and production defaults are used. Tests may pass a mock `deps` object.
- `deactivate()` is the single exported deactivation function.
- No class instantiation for the entry point — module-level exported functions.
- When `deps` is `undefined` or omitted, `activate` must construct all collaborators internally using production implementations — must not throw or crash when called with only `context`.

## Behavior

### activate(context, deps?)

1. Merges `deps` with production defaults: any field not provided in `deps` (or if `deps` is `undefined`) uses the production implementation. This produces a resolved deps object used for all subsequent construction.
1a. Creates an `OutputChannel` named `'Spectra'` via `vscode.window.createOutputChannel('Spectra')` (or uses `deps.outputChannel` if provided).
2. Wraps the OutputChannel in a logger adapter object providing `{ info, warn, error }` methods. Each method prepends a severity tag and delegates to `outputChannel.appendLine`.
3. Logs activation start via `logger.info`.
4. Calls `ProjectRootResolver.resolve()` to obtain `projectRoot`.
5. Creates a `SpectraViewProvider(context.extensionUri, logger)` instance.
6. Registers the view provider via `vscode.window.registerWebviewViewProvider('spectra.chatView', viewProvider, { webviewOptions: { retainContextWhenHidden: true } })`.
7. If `projectRoot` is `undefined`, calls `viewProvider.showNotInitialized()` (ViewProvider stores this as a pending message if the view has not yet resolved), pushes OutputChannel and viewProvider registration to `context.subscriptions`, logs the error, and returns.
8. Calls `ProjectRootResolver.isInitialized(projectRoot)` to check if `.spectra/` directory exists.
9. If not initialized, calls `viewProvider.showNotInitialized()` (ViewProvider stores this as a pending message if the view has not yet resolved), pushes OutputChannel and viewProvider registration to `context.subscriptions`, logs, and returns.
10. Creates a `SessionListController(projectRoot, logger)` instance.
11. Creates a `SessionDetailController(projectRoot, logger)` instance.
12. Initializes an internal `cachedSessionListState` variable to `null`.
13. Subscribes to `sessionListController.onDidUpdate`:
    - Stores the received state in `cachedSessionListState`.
    - Calls `viewProvider.showSessionList(state)`.
14. Subscribes to `sessionDetailController.onDidUpdate`:
    - Calls `viewProvider.showSessionDetail(state)`.
15. Subscribes to `sessionListController.onDidError`:
    - Calls `vscode.window.showErrorMessage(error.message)`.
16. Subscribes to `sessionDetailController.onDidError`:
    - Calls `vscode.window.showErrorMessage(error.message)`.
17. Subscribes to `viewProvider.onDidReceiveMessage` and routes based on `msg.command`:
    - `'navigateToDetail'`: calls `sessionDetailController.open(msg.sessionId, msg.workflowName)`.
    - `'navigateToList'`: if `cachedSessionListState` is not null, calls `viewProvider.showSessionList(cachedSessionListState)`.
    - `'launchSession'`: calls `sessionListController.launch(msg.workflowName)`.
    - `'terminateSession'`: calls `sessionListController.terminate(msg.pid)`.
    - `'sendEvent'`: calls `sessionDetailController.sendEvent(msg.eventType, msg.message)`.
    - Any other command: logs a warning via `logger.warn` with the unrecognized command value.
18. Pushes all disposables to `context.subscriptions`: the OutputChannel, sessionListController, sessionDetailController, viewProvider, the view provider registration, and all subscription disposables.
19. Logs successful activation via `logger.info` including the resolved `projectRoot`.

### deactivate()

20. Empty function body. All cleanup is handled by `context.subscriptions` disposal.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| context | `vscode.ExtensionContext` | Valid extension context provided by VS Code | Yes (activate) |
| deps | `ActivateDeps \| undefined` | Optional object providing collaborator factories/instances for testing. When `undefined`, production defaults are used. | No (activate) |

## Outputs

| Field | Type | Description |
|---|---|---|
| (none) | void | `activate` and `deactivate` return void. |

## Invariants

- `activate` must accept `context: vscode.ExtensionContext` as its first parameter and an optional `deps?: ActivateDeps` as its second parameter. VS Code calls `activate(context)` with no additional arguments, so `deps` is `undefined` at runtime and production defaults are used. Must not crash when `deps` is `undefined`.
- When `deps` is `undefined` or omitted, must construct all collaborators (logger, controllers, view provider) internally using production implementations. When `deps` is provided (test scenarios), must use the supplied implementations.
- Must register `SpectraViewProvider` via `vscode.window.registerWebviewViewProvider('spectra.chatView', ...)` synchronously during activation, before any async work. This ensures the sidebar view is always backed by a provider when the user opens it.
- The extension's `package.json` must declare `"activationEvents": ["onView:spectra.chatView"]` so that VS Code activates the extension when the user opens the sidebar view. Without this, clicking the sidebar icon will not trigger activation and the view will remain unresolved.
- Must register all disposables with `context.subscriptions` — no manual cleanup in `deactivate()`.
- Must always register SpectraViewProvider regardless of projectRoot or initialization state — the sidebar view must always appear.
- Must not proceed past step 9 (creating controllers) if projectRoot is undefined or project is not initialized.
- Must cache the latest `SessionListState` for immediate replay on `navigateToList`.
- Must route `terminateSession` to `sessionListController.terminate()` regardless of which page originated the message.
- Must log a warning (not throw) for unrecognized webview message commands.
- Must not fire controller methods after controllers have been disposed.
- Must use the OutputChannel for all diagnostic output — never `console.log`.
- Must show errors via `vscode.window.showErrorMessage` when controllers fire `onDidError`.

## Edge Cases

- Condition: `ProjectRootResolver.resolve()` returns `undefined` (no workspace open).
  Expected: ViewProvider is registered (sidebar view appears). Shows "not initialized" message in webview. No controllers created. Only OutputChannel and viewProvider registration pushed to subscriptions.

- Condition: `ProjectRootResolver.isInitialized()` returns `false` (.spectra/ directory missing).
  Expected: Same as above — shows "not initialized" message. No controllers created.

- Condition: `navigateToList` received before the first `onDidUpdate` from SessionListController (cachedSessionListState is null).
  Expected: No-op — `viewProvider.showSessionList` is not called. The webview remains on whatever page it currently shows until the first scan completes.

- Condition: `terminateSession` arrives from the detail page with a `pid` value.
  Expected: Routes identically to a `terminateSession` from the list page — calls `sessionListController.terminate(msg.pid)`.

- Condition: Controller `onDidUpdate` fires but the view has not yet been resolved (sidebar not yet visible).
  Expected: `viewProvider.showSessionList`/`showSessionDetail` is a no-op when view is null. The state is still cached in `cachedSessionListState`. When the view resolves and triggers a scan, fresh state will arrive.

- Condition: `showNotInitialized` is called before `resolveWebviewView` is triggered by VS Code.
  Expected: ViewProvider stores the message as `pendingMessage`. When `resolveWebviewView` fires, the pending message is delivered immediately after HTML assignment. No retry or polling needed from extension.ts.

- Condition: Multiple `onDidError` events fire rapidly from both controllers.
  Expected: Each triggers an independent `showErrorMessage` call. VS Code queues/stacks notifications as needed.

## Related

- [ProjectRootResolver](./services/projectRootResolver.md) — Resolves project root from workspace and configuration, checks initialization.
- [SessionListController](./controllers/sessionListController.md) — Manages session list state and actions.
- [SessionDetailController](./controllers/sessionDetailController.md) — Manages session detail state and actions.
- [SpectraViewProvider](./views/spectraViewProvider.md) — WebviewViewProvider for the sidebar.
