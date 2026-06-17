# SpectraViewProvider

## Overview

Implements `vscode.WebviewViewProvider` to provide a WebviewView in the Activity Bar sidebar. Acts as a thin transport layer that posts state to the webview and forwards incoming messages to subscribers. Does not own business logic, state computation, or UI rendering.

## Boundaries

- Owns: implementing `resolveWebviewView`, configuring webview options, assigning HTML content, posting messages to the webview via `postMessage`, forwarding incoming webview messages to subscribers via `onDidReceiveMessage`, tracking current page state (`'sessions' | 'detail' | 'notInitialized'`), and implementing `vscode.Disposable`.
- Delegates: HTML content generation to `getWebviewContent`.
- Delegates: business logic (state assembly, command handling) to extension.ts via the message events.
- Delegates: controller lifecycle management to extension.ts.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes.
- Must not: instantiate or hold references to controllers.
- Must not: interpret or validate the content of messages — acts as a pass-through.
- Must not: use `vscode.window.createWebviewPanel` — this is a sidebar view, not an editor panel.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.WebviewView` | Underlying view (provided by VS Code in resolveWebviewView) | `webview.postMessage()`, `webview.onDidReceiveMessage`, `webview.html`, `webview.options` | Must not call `show()` or `dispose()` on the view itself (VS Code owns sidebar view lifecycle) |
| `getWebviewContent` | HTML generator | `getWebviewContent(webview, extensionUri)` | Must not call at any time other than resolveWebviewView |
| `vscode.EventEmitter<WebviewMessage>` | Incoming message channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| Logger (`{ info, warn, error }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- Instantiated via `new SpectraViewProvider(extensionUri, logger)`.
- Registered with `vscode.window.registerWebviewViewProvider('spectra.sessionView', provider, { webviewOptions: { retainContextWhenHidden: true } })`.
- Implements `vscode.WebviewViewProvider` and `vscode.Disposable`.

## Behavior

### Construction

1. Stores `extensionUri` and `logger`.
2. Creates a `vscode.EventEmitter<WebviewMessage>` and exposes its `.event` as `onDidReceiveMessage`.
3. Initializes internal `view` reference to `null`.
4. Initializes internal `currentPage` to `'sessions'`.
4a. Initializes internal `pendingMessage` to `null`.

### resolveWebviewView(webviewView, context, token)

5. Stores `webviewView` as internal `view` reference.
6. Configures webview options: `{ enableScripts: true, localResourceRoots: [extensionUri] }`.
7. Calls `getWebviewContent(webviewView.webview, extensionUri)` and assigns the result to `webviewView.webview.html`.
8. Subscribes to `webviewView.webview.onDidReceiveMessage` → fires `onDidReceiveMessage` emitter with the received message object.
9. Subscribes to `webviewView.onDidDispose` → sets internal `view` to `null` and logs via `logger.info`.
10. If `pendingMessage` is not null, calls `view.webview.postMessage(pendingMessage)` and sets `pendingMessage` to `null`.
11. Logs view resolution via `logger.info`.

### showSessionList(state)

12. Sets internal `currentPage` to `'sessions'`.
13. Constructs message `{ type: 'showSessions', state }`.
14. If `view` is null, stores the message as `pendingMessage` and returns.
15. Calls `view.webview.postMessage(message)`.

### showSessionDetail(state)

16. Sets internal `currentPage` to `'detail'`.
17. Constructs message `{ type: 'showDetail', state }`.
18. If `view` is null, stores the message as `pendingMessage` and returns.
19. Calls `view.webview.postMessage(message)`.

### showNotInitialized()

20. Sets internal `currentPage` to `'notInitialized'`.
21. Constructs message `{ type: 'showNotInitialized' }`.
22. If `view` is null, stores the message as `pendingMessage` and returns.
23. Calls `view.webview.postMessage(message)`.

### dispose()

24. Disposes the `EventEmitter` instance.
25. Sets `view` to `null`.
26. Sets `pendingMessage` to `null`.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| extensionUri | `vscode.Uri` | Extension root URI | Yes (constructor) |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes (constructor) |
| webviewView | `vscode.WebviewView` | Provided by VS Code | Yes (resolveWebviewView) |
| context | `vscode.WebviewViewResolveContext` | Provided by VS Code | Yes (resolveWebviewView) |
| token | `vscode.CancellationToken` | Provided by VS Code | Yes (resolveWebviewView) |
| state (showSessionList) | `SessionListState` | Object with `sessions: SessionSummary[]` and `workflows: string[]` | Yes |
| state (showSessionDetail) | `SessionDetailState` | As defined in SessionDetailController | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| onDidReceiveMessage | `vscode.Event<WebviewMessage>` | Fires when the webview posts a message to the extension. |

### WebviewMessage Type (incoming from webview)

| Command | Payload Fields | Description |
|---|---|---|
| `navigateToDetail` | `sessionId: string`, `workflowName: string` | User clicked a session row. |
| `navigateToList` | (none) | User clicked the back button on the detail page. |
| `launchSession` | `workflowName: string` | User clicked the Run button. |
| `terminateSession` | `pid: number` | User clicked the stop button on a session row. |
| `sendEvent` | `eventType: string`, `message: string` | User submitted the event form on the detail page. |

## Invariants

- Must implement `vscode.WebviewViewProvider` and `vscode.Disposable`.
- Must not fire `onDidReceiveMessage` after dispose.
- Must not call `postMessage` when `view` is null — show methods store the message as `pendingMessage` for delivery in `resolveWebviewView`.
- Must deliver `pendingMessage` exactly once in `resolveWebviewView` and then clear it.
- Must not interpret or validate message content — forwards all messages as-is.
- Must set `currentPage` before posting the corresponding message to the webview.
- Must use a Content Security Policy in the generated HTML (responsibility of `getWebviewContent`, but SpectraViewProvider must not override `webview.html` after resolveWebviewView).
- Must set `retainContextWhenHidden: true` in webview options to preserve state when sidebar is hidden.

## Edge Cases

- Condition: `showSessionList`, `showSessionDetail`, or `showNotInitialized` is called before `resolveWebviewView` (view is null).
  Expected: The message is stored as `pendingMessage`. When `resolveWebviewView` fires, the pending message is delivered immediately after HTML assignment. Only the most recent pending message is retained (later calls overwrite earlier ones).

- Condition: The sidebar is hidden by the user (collapsed or switched to another view container).
  Expected: `retainContextWhenHidden: true` keeps the webview alive. Messages can still be posted. No dispose occurs.

- Condition: `resolveWebviewView` is called again after the view was disposed and recreated by VS Code.
  Expected: The new view replaces the old reference. HTML is regenerated and assigned. Subscriptions are set up fresh.

- Condition: The webview sends a message with an unrecognized `command` value.
  Expected: The message is forwarded to `onDidReceiveMessage` subscribers as-is. SpectraViewProvider does not filter.

- Condition: `showNotInitialized` is called, then later `showSessionList` is called (project initialized after extension activation).
  Expected: The webview switches from the not-initialized page to the sessions page via the posted message.

- Condition: `dispose()` is called while the view is still visible.
  Expected: EventEmitter is disposed, view reference set to null. VS Code owns the actual view disposal.

## Related

- [getWebviewContent](./getWebviewContent.md) — Generates the HTML content for the webview.
- [SessionListController](../controllers/sessionListController.md) — Produces `SessionListState` pushed via `showSessionList`.
- [SessionDetailController](../controllers/sessionDetailController.md) — Produces `SessionDetailState` pushed via `showSessionDetail`.
