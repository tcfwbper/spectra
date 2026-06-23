# Test Specification: `spectraViewProvider.test.ts`

## Source File Under Test
`vscode/src/views/spectraViewProvider.ts`

## Test File
`vscode/test/suite/spectraViewProvider.test.ts`

---

## `SpectraViewProvider`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should store extensionUri and logger` | `unit` | Constructor stores provided arguments. | Create a mock `extensionUri` and a mock `logger` with `info`, `warn`, `error` methods. | `new SpectraViewProvider(mockExtensionUri, mockLogger)` | Instance has internal references to `extensionUri` and `logger` (verified via subsequent method calls) |
| `should expose onDidReceiveMessage event` | `unit` | Instance exposes the message event. | Create a mock `extensionUri` and `logger`. | `new SpectraViewProvider(mockExtensionUri, mockLogger)` | Returned instance has `onDidReceiveMessage` property that is a function |
| `should initialize view as null` | `unit` | No view reference before resolveWebviewView. | Create instance. | `new SpectraViewProvider(mockExtensionUri, mockLogger)` | Calling `showSessionList` with a state stores it as pending (verified by not calling postMessage) |
| `should initialize pendingMessage as null` | `unit` | No pending message at construction. | Create instance. Call `resolveWebviewView` with a mock view. Spy on `view.webview.postMessage`. | `new SpectraViewProvider(mockExtensionUri, mockLogger)` then `resolveWebviewView(mockView, mockContext, mockToken)` | `postMessage` is not called (no pending message to deliver) |

### Happy Path — resolveWebviewView

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should configure webview options with enableScripts and localResourceRoots` | `unit` | Webview options are set correctly. | Create instance with `mockExtensionUri`. Create a mock `webviewView` with a settable `webview.options`. | `instance.resolveWebviewView(mockWebviewView, mockContext, mockToken)` | `mockWebviewView.webview.options` is set to `{ enableScripts: true, localResourceRoots: [mockExtensionUri] }` |
| `should assign HTML from getWebviewContent to webview` | `unit` | Webview html is set from getWebviewContent result. | Stub `getWebviewContent` to return `'<html>test</html>'`. Create mock `webviewView`. | `instance.resolveWebviewView(mockWebviewView, mockContext, mockToken)` | `mockWebviewView.webview.html` is set to `'<html>test</html>'` |
| `should deliver pendingMessage after HTML assignment` | `unit` | Pending message is posted on resolve. | Create instance. Call `instance.showNotInitialized()` (stores as pending). Create mock `webviewView` with spy on `webview.postMessage`. | `instance.resolveWebviewView(mockWebviewView, mockContext, mockToken)` | `webview.postMessage` called with `{ type: 'showNotInitialized' }`; subsequent call to `showSessionList` does not re-deliver the old pending |
| `should clear pendingMessage after delivery` | `unit` | Pending message is cleared once delivered. | Create instance. Call `instance.showSessionList(state)`. Create mock `webviewView`. Call `resolveWebviewView`. Then trigger `onDidDispose` to null the view, and call `resolveWebviewView` again with a new mock. | Second `resolveWebviewView` | `postMessage` is not called on the second resolve (pending was already cleared) |
| `should subscribe to webview onDidReceiveMessage` | `unit` | Incoming messages are forwarded. | Create instance. Create mock `webviewView` with a controllable `webview.onDidReceiveMessage` callback list. Call `resolveWebviewView`. Register a spy on `instance.onDidReceiveMessage`. | Trigger the webview's onDidReceiveMessage with `{ command: 'navigateToList' }` | Spy is called with `{ command: 'navigateToList' }` |
| `should set view to null on webviewView dispose` | `unit` | View reference is cleared on dispose. | Create instance. Call `resolveWebviewView`. Spy on `mockLogger.info`. | Trigger `mockWebviewView.onDidDispose` callback | After dispose, calling `instance.showSessionList(state)` stores message as pending (no postMessage call); `logger.info` called |
| `should log view resolution` | `unit` | Logger is called on resolve. | Create instance. Spy on `mockLogger.info`. | `instance.resolveWebviewView(mockWebviewView, mockContext, mockToken)` | `mockLogger.info` called at least once |

### Happy Path — showSessionList

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should post showSessions message to webview` | `unit` | State is posted with correct type. | Create instance. Call `resolveWebviewView` with mock view. Spy on `view.webview.postMessage`. | `instance.showSessionList({ sessions: [], workflows: ['wf1'] })` | `webview.postMessage` called with `{ type: 'showSessions', state: { sessions: [], workflows: ['wf1'] } }` |
| `should store as pendingMessage when view is null` | `unit` | Message queued when view not resolved. | Create instance (do not call resolveWebviewView). | `instance.showSessionList({ sessions: [], workflows: [] })` | No error thrown; subsequent `resolveWebviewView` delivers the message |

### Happy Path — showSessionDetail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should post showDetail message to webview` | `unit` | State is posted with correct type. | Create instance. Call `resolveWebviewView` with mock view. Spy on `view.webview.postMessage`. | `instance.showSessionDetail({ sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: ['submit'], events: [] })` | `webview.postMessage` called with `{ type: 'showDetail', state: { sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: ['submit'], events: [] } }` |
| `should store as pendingMessage when view is null` | `unit` | Message queued when view not resolved. | Create instance (do not call resolveWebviewView). | `instance.showSessionDetail({ sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: [], events: [] })` | No error thrown; subsequent `resolveWebviewView` delivers the message |

### Happy Path — showNotInitialized

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should post showNotInitialized message to webview` | `unit` | Message is posted with correct type. | Create instance. Call `resolveWebviewView` with mock view. Spy on `view.webview.postMessage`. | `instance.showNotInitialized()` | `webview.postMessage` called with `{ type: 'showNotInitialized' }` |
| `should store as pendingMessage when view is null` | `unit` | Message queued when view not resolved. | Create instance (do not call resolveWebviewView). | `instance.showNotInitialized()` | No error thrown; subsequent `resolveWebviewView` delivers the message |

### Happy Path — postSendResult

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should post sendResult message with success true to webview` | `unit` | Posts sendResult true when view is resolved. | Create instance. Call `resolveWebviewView` with mock view. Spy on `view.webview.postMessage`. | `instance.postSendResult(true)` | `webview.postMessage` called with `{ type: 'sendResult', success: true }` |
| `should post sendResult message with success false to webview` | `unit` | Posts sendResult false when view is resolved. | Create instance. Call `resolveWebviewView` with mock view. Spy on `view.webview.postMessage`. | `instance.postSendResult(false)` | `webview.postMessage` called with `{ type: 'sendResult', success: false }` |
| `should do nothing when view is null` | `unit` | Ephemeral result is dropped when view is not available. | Create instance (do not call resolveWebviewView). | `instance.postSendResult(true)` | No error thrown; no `pendingMessage` stored (subsequent `resolveWebviewView` does not deliver a sendResult message) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should overwrite previous pendingMessage with latest call` | `unit` | Only the most recent pending message is retained. | Create instance (no resolveWebviewView). Call `instance.showNotInitialized()`. Then call `instance.showSessionList(state)`. Then call `resolveWebviewView` with mock view. Spy on `postMessage`. | `resolveWebviewView` after multiple show calls | `postMessage` called exactly once with `{ type: 'showSessions', state }` (the latest message) |
| `should handle resolveWebviewView called again after view disposal` | `unit` | New view replaces old reference, HTML regenerated. | Create instance. Call `resolveWebviewView` with first mock view. Trigger `onDidDispose` on first view. Stub `getWebviewContent` to return `'<html>new</html>'`. Create a second mock view. | `instance.resolveWebviewView(secondMockView, mockContext, mockToken)` | Second mock view's `webview.html` set to `'<html>new</html>'`; `webview.onDidReceiveMessage` subscribed on new view |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should forward webview messages to onDidReceiveMessage subscribers` | `unit` | Incoming messages are relayed. | Create instance. Call `resolveWebviewView`. Register a sinon spy as a listener on `instance.onDidReceiveMessage`. | Simulate mock view's `webview.onDidReceiveMessage` callback firing with `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }` | Spy is called with `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }` |
| `should forward unrecognized commands without filtering` | `unit` | Unknown commands pass through. | Create instance. Call `resolveWebviewView`. Register a sinon spy as a listener on `instance.onDidReceiveMessage`. | Simulate mock view's `webview.onDidReceiveMessage` callback firing with `{ command: 'unknownCommand', data: 123 }` | Spy is called with `{ command: 'unknownCommand', data: 123 }` |
| `should call getWebviewContent with webview and extensionUri` | `unit` | HTML generator receives correct args. | Stub `getWebviewContent` as a sinon stub. Create instance with `mockExtensionUri`. | `instance.resolveWebviewView(mockWebviewView, mockContext, mockToken)` | `getWebviewContent` called with `(mockWebviewView.webview, mockExtensionUri)` |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose EventEmitter on dispose` | `unit` | EventEmitter is cleaned up. | Create instance. Spy on the internal EventEmitter's `dispose` method (accessible via testing seam or by verifying behavior). | `instance.dispose()` | After dispose, firing a message on the underlying webview does not trigger `onDidReceiveMessage` subscribers |
| `should set view to null on dispose` | `unit` | View reference cleared. | Create instance. Call `resolveWebviewView`. | `instance.dispose()` | Subsequent `showSessionList` call stores as pending (no postMessage on disposed view) |
| `should set pendingMessage to null on dispose` | `unit` | Pending message cleared. | Create instance. Call `instance.showNotInitialized()` (stores pending). | `instance.dispose()` then `resolveWebviewView(newMockView, ...)` | `postMessage` is not called on the new view (pending was cleared by dispose) |
| `should not fire onDidReceiveMessage after dispose` | `unit` | Events stop after dispose. | Create instance. Call `resolveWebviewView`. Register spy on `instance.onDidReceiveMessage`. Call `instance.dispose()`. | Simulate `webview.onDidReceiveMessage` callback after disposal | Spy is not called after disposal |

### Asynchronous Flow

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should transition from notInitialized to sessions when showSessionList called later` | `unit` | Webview switches pages via posted messages. | Create instance. Call `resolveWebviewView`. Spy on `postMessage`. Call `instance.showNotInitialized()`. Then call `instance.showSessionList(state)`. | Sequential calls | `postMessage` called first with `{ type: 'showNotInitialized' }`, then with `{ type: 'showSessions', state }` |
