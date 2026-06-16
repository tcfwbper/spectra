# SpectraPanel

## Overview

Manages the lifecycle of a single WebviewPanel (singleton pattern) and provides bidirectional message routing between the webview and the extension host. SpectraPanel does not own business logic, state computation, or UI rendering — it is a thin transport layer that posts state to the webview and forwards incoming messages to subscribers.

## Boundaries

- Owns: creating and revealing the singleton WebviewPanel, tracking current page state (`'sessions' | 'detail'`), posting messages to the webview via `postMessage`, forwarding incoming webview messages to subscribers via `onDidReceiveMessage`, notifying subscribers of panel disposal via `onDidDispose`, and implementing `vscode.Disposable`.
- Delegates: HTML content generation to `getWebviewContent`.
- Delegates: business logic (state assembly, command handling) to extension.ts via the message events.
- Delegates: controller lifecycle management to extension.ts via the dispose event.
- Must not: read, write, create, or delete any file.
- Must not: spawn or signal processes.
- Must not: instantiate or hold references to controllers.
- Must not: interpret or validate the content of messages — acts as a pass-through.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.window.createWebviewPanel` | Panel factory | Create panel with viewType, title, column, options | — |
| `vscode.WebviewPanel` | Underlying panel | `reveal()`, `webview.postMessage()`, `webview.onDidReceiveMessage`, `onDidDispose`, `webview.html` | Must not access `webview.asWebviewUri` directly (delegated to `getWebviewContent`) |
| `getWebviewContent` | HTML generator | `getWebviewContent(webview, extensionUri)` | Must not call at any time other than panel creation |
| `vscode.EventEmitter<WebviewMessage>` | Incoming message channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.EventEmitter<void>` | Dispose notification channel | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| Logger (`{ info, warn, error }`) | Diagnostic output | `info()`, `warn()`, `error()` | — |

Construction constraints:
- Must not be instantiated directly — uses a static factory method `createOrReveal`.
- Private constructor enforces singleton via a static `instance` field.
- Implements `vscode.Disposable`.

## Behavior

### createOrReveal(context, extensionUri, logger)

1. If a live instance already exists (static `instance` is not null and underlying panel is not disposed), calls `panel.reveal()` on the existing instance and returns the existing instance.
2. Otherwise, creates a new `vscode.WebviewPanel` with:
   - viewType: `'spectra'`
   - title: `'Spectra'`
   - showOptions: `vscode.ViewColumn.One`
   - options: `{ enableScripts: true, retainContextWhenHidden: true, localResourceRoots: [extensionUri] }`
3. Calls `getWebviewContent(panel.webview, extensionUri)` and assigns the result to `panel.webview.html`.
4. Creates a `vscode.EventEmitter<WebviewMessage>` and exposes its `.event` as `onDidReceiveMessage`.
5. Creates a `vscode.EventEmitter<void>` and exposes its `.event` as `onDidDispose`.
6. Subscribes to `panel.webview.onDidReceiveMessage` → fires `onDidReceiveMessage` emitter with the received message object.
7. Subscribes to `panel.onDidDispose` → fires `onDidDispose` emitter, disposes both EventEmitter instances, sets static `instance` to null, and logs via `logger.info`.
8. Stores the new instance as the static `instance`.
9. Pushes the instance into `context.subscriptions` for automatic cleanup.
10. Logs panel creation via `logger.info`.
11. Initializes internal `currentPage` state to `'sessions'`.
12. Returns the new instance.

### showSessionList(state)

13. Sets internal `currentPage` to `'sessions'`.
14. Calls `panel.webview.postMessage({ type: 'showSessions', state })`.

### showSessionDetail(state)

15. Sets internal `currentPage` to `'detail'`.
16. Calls `panel.webview.postMessage({ type: 'showDetail', state })`.

### dispose()

17. Calls the underlying `panel.dispose()`, which triggers the `onDidDispose` handler in step 7.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| context | `vscode.ExtensionContext` | Valid extension context | Yes (createOrReveal) |
| extensionUri | `vscode.Uri` | Extension root URI | Yes (createOrReveal) |
| logger | `{ info, warn, error }` | Must provide info, warn, error methods | Yes (createOrReveal) |
| state (showSessionList) | `SessionListState` | Object with `sessions: SessionSummary[]` and `workflows: string[]` | Yes |
| state (showSessionDetail) | `SessionDetailState` | As defined in [SessionDetailController](../controllers/sessionDetailController.md) — includes `sessionId`, `workflowName`, `entryNode`, `currentState`, `status`, `eventTypes`, `events` | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| onDidReceiveMessage | `vscode.Event<WebviewMessage>` | Fires when the webview posts a message to the extension. |
| onDidDispose | `vscode.Event<void>` | Fires when the panel is closed by the user or disposed programmatically. |

### WebviewMessage Type (incoming from webview)

| Command | Payload Fields | Description |
|---|---|---|
| `navigateToDetail` | `sessionId: string`, `workflowName: string` | User clicked a session row. |
| `navigateToList` | (none) | User clicked the back button on the detail page. |
| `launchSession` | `workflowName: string` | User clicked the Run button. |
| `terminateSession` | `pid: number` | User clicked the stop button on a session row. |
| `sendEvent` | `eventType: string`, `message: string` | User submitted the event form on the detail page. |

### SessionDetailState Type

Defined authoritatively in [SessionDetailController](../controllers/sessionDetailController.md). Includes `sessionId`, `workflowName`, `entryNode`, `currentState`, `status`, `pid`, `eventTypes`, and `events`. The `entryNode`, `currentState`, and `status` fields are required by the webview's send-button guard logic. The `pid` field is required by the Stop button on the detail page (same behavior as the list page's terminate button).

## Invariants

- Must implement `vscode.Disposable`.
- Must enforce singleton: at most one live WebviewPanel exists at any time.
- Must never fire `onDidReceiveMessage` or `onDidDispose` after the panel has been disposed.
- Must set static `instance` to null upon panel disposal.
- Must not interpret or validate message content — forwards all messages as-is.
- Must set `currentPage` before posting the corresponding message to the webview.
- Must use a Content Security Policy in the generated HTML (responsibility of `getWebviewContent`, but SpectraPanel must not override `webview.html` after creation).
- Must register itself with `context.subscriptions` for automatic disposal.

## Edge Cases

- Condition: `createOrReveal` is called when a panel already exists.
  Expected: The existing panel is revealed (brought to front). No new panel is created. The same instance is returned.

- Condition: `createOrReveal` is called after the user manually closed the panel.
  Expected: Static `instance` is null (cleared by `onDidDispose` handler). A new panel is created.

- Condition: `showSessionList` or `showSessionDetail` is called after the panel is disposed.
  Expected: `postMessage` on a disposed panel is a no-op (VS Code API behavior). No error is thrown.

- Condition: The webview sends a message with an unrecognized `command` value.
  Expected: The message is forwarded to `onDidReceiveMessage` subscribers as-is. SpectraPanel does not filter.

- Condition: `dispose()` is called multiple times.
  Expected: The first call disposes the underlying panel and triggers cleanup. Subsequent calls are no-ops (panel is already disposed).

- Condition: `showSessionDetail` is called immediately after `createOrReveal` before the webview has initialized its JS.
  Expected: `postMessage` queues the message; VS Code delivers it once the webview is ready. No message is lost.

## Related

- [getWebviewContent](./getWebviewContent.md) — Generates the HTML content for the webview.
- [SessionListController](../controllers/sessionListController.md) — Produces `SessionListState` pushed via `showSessionList`.
- [SessionDetailController](../controllers/sessionDetailController.md) — Produces `SessionDetailState` pushed via `showSessionDetail`.
