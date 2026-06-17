# Test Specification: `spectraPanel.test.ts`

## Source File Under Test
`vscode/src/views/spectraPanel.ts`

## Test File
`vscode/test/suite/spectraPanel.test.ts`

---

## `SpectraPanel`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should create a new WebviewPanel with correct options` | `unit` | First call creates a panel. | Stub `vscode.window.createWebviewPanel` to return a mock panel with `webview` (stub `postMessage`, `onDidReceiveMessage`, `html` setter), `onDidDispose`, and `reveal`. Stub `getWebviewContent` to return `'<html></html>'`. Create mock `context` with `subscriptions` array. Create mock `logger`. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | `createWebviewPanel` called with viewType `'spectra'`, title `'Spectra'`, `vscode.ViewColumn.One`, options containing `enableScripts: true`, `retainContextWhenHidden: true`, `localResourceRoots: [mockExtensionUri]` |
| `should assign HTML from getWebviewContent to webview` | `unit` | Panel's webview.html is set. | Stub `vscode.window.createWebviewPanel` to return mock panel. Stub `getWebviewContent` to return `'<html>test</html>'`. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | Mock panel's `webview.html` is set to `'<html>test</html>'` |
| `should push instance into context.subscriptions` | `unit` | Instance is registered for automatic disposal. | Stub `vscode.window.createWebviewPanel` to return mock panel. Create mock `context` with `subscriptions` as an empty array. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | `mockContext.subscriptions` array length increases by 1 |
| `should log panel creation` | `unit` | Logger is called on creation. | Stub `vscode.window.createWebviewPanel` to return mock panel. Spy on `mockLogger.info`. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | `mockLogger.info` called at least once |
| `should expose onDidReceiveMessage event` | `unit` | Instance exposes the message event. | Stub `vscode.window.createWebviewPanel` to return mock panel. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | Returned instance has `onDidReceiveMessage` property that is a function |
| `should expose onDidDispose event` | `unit` | Instance exposes the dispose event. | Stub `vscode.window.createWebviewPanel` to return mock panel. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | Returned instance has `onDidDispose` property that is a function |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should reveal existing panel when called again` | `unit` | Second call reveals instead of creating. | Stub `vscode.window.createWebviewPanel` to return mock panel with `reveal` spy. Call `createOrReveal` once to establish singleton. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` (second call) | `createWebviewPanel` called only once total; mock panel's `reveal` called once; same instance returned |
| `should create new panel after previous was disposed` | `unit` | Disposal clears singleton. | Stub `vscode.window.createWebviewPanel` to return mock panel. Call `createOrReveal` once. Trigger `onDidDispose` callback on the mock panel. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` (after dispose) | `createWebviewPanel` called a second time; a new instance is returned |

### Happy Path — showSessionList

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should post showSessions message to webview` | `unit` | State is posted with correct type. | Call `createOrReveal` to get instance. Spy on mock panel's `webview.postMessage`. | `instance.showSessionList({ sessions: [], workflows: ['wf1'] })` | `webview.postMessage` called with `{ type: 'showSessions', state: { sessions: [], workflows: ['wf1'] } }` |
| `should update currentPage to sessions` | `unit` | Internal page tracking updates. | Call `createOrReveal` to get instance. First call `showSessionDetail` to set page to detail. Spy on mock panel's `webview.postMessage`. | `instance.showSessionList({ sessions: [], workflows: [] })` | No error thrown; `postMessage` called with type `'showSessions'` |

### Happy Path — showSessionDetail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should post showDetail message to webview` | `unit` | State is posted with correct type. | Call `createOrReveal` to get instance. Spy on mock panel's `webview.postMessage`. | `instance.showSessionDetail({ sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: ['submit'], events: [] })` | `webview.postMessage` called with `{ type: 'showDetail', state: { sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: ['submit'], events: [] } }` |
| `should update currentPage to detail` | `unit` | Internal page tracking updates. | Call `createOrReveal` to get instance. Spy on mock panel's `webview.postMessage`. | `instance.showSessionDetail({ sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: [], events: [] })` | No error thrown; `postMessage` called with type `'showDetail'` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should forward webview messages to onDidReceiveMessage subscribers` | `unit` | Incoming messages are relayed. | Call `createOrReveal` to get instance. Register a sinon spy as a listener on `instance.onDidReceiveMessage`. | Simulate mock panel's `webview.onDidReceiveMessage` callback firing with `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }` | Spy is called with `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }` |
| `should forward unrecognized commands without filtering` | `unit` | Unknown commands pass through. | Call `createOrReveal` to get instance. Register a sinon spy as a listener on `instance.onDidReceiveMessage`. | Simulate mock panel's `webview.onDidReceiveMessage` callback firing with `{ command: 'unknownCommand', data: 123 }` | Spy is called with `{ command: 'unknownCommand', data: 123 }` |
| `should fire onDidDispose when panel is closed` | `unit` | Panel closure notifies subscribers. | Call `createOrReveal` to get instance. Register a sinon spy as a listener on `instance.onDidDispose`. | Trigger `onDidDispose` callback on mock panel | Spy is called once |
| `should call getWebviewContent with webview and extensionUri` | `unit` | HTML generator receives correct args. | Stub `getWebviewContent` as a sinon stub. Stub `vscode.window.createWebviewPanel` to return mock panel. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` | `getWebviewContent` called with `(mockPanel.webview, mockExtensionUri)` |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose underlying panel when dispose is called` | `unit` | Programmatic disposal triggers panel dispose. | Call `createOrReveal` to get instance. Spy on mock panel's `dispose`. | `instance.dispose()` | Mock panel's `dispose` is called |
| `should set static instance to null on disposal` | `unit` | Singleton reference is cleared. | Call `createOrReveal` to get instance. Trigger `onDidDispose` on mock panel. | `SpectraPanel.createOrReveal(mockContext, mockExtensionUri, mockLogger)` (after dispose) | A new panel is created (proving instance was null) |
| `should log on panel disposal` | `unit` | Logger is called on disposal. | Call `createOrReveal` to get instance. Spy on `mockLogger.info`. | Trigger `onDidDispose` on mock panel | `mockLogger.info` called |
| `should not fire onDidReceiveMessage after disposal` | `unit` | Events stop after dispose. | Call `createOrReveal` to get instance. Register spy on `instance.onDidReceiveMessage`. Trigger `onDidDispose` on mock panel. | Simulate `webview.onDidReceiveMessage` callback after disposal | Spy is not called again after disposal |
| `should handle multiple dispose calls gracefully` | `unit` | Second dispose is a no-op. | Call `createOrReveal` to get instance. Call `instance.dispose()`. | `instance.dispose()` (second call) | No error is thrown |
| `should not throw when showSessionList is called after disposal` | `unit` | Post-disposal postMessage is a no-op. | Call `createOrReveal` to get instance. Trigger `onDidDispose` on mock panel. Stub `webview.postMessage` to be a no-op. | `instance.showSessionList({ sessions: [], workflows: [] })` | No error is thrown |
| `should not throw when showSessionDetail is called after disposal` | `unit` | Post-disposal postMessage is a no-op. | Call `createOrReveal` to get instance. Trigger `onDidDispose` on mock panel. Stub `webview.postMessage` to be a no-op. | `instance.showSessionDetail({ sessionId: 's1', workflowName: 'wf1', entryNode: 'start', currentState: 'start', status: 'running', pid: 42, eventTypes: [], events: [] })` | No error is thrown |

### Asynchronous Flow

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should not lose messages posted before webview JS initializes` | `unit` | postMessage queues message delivery. | Call `createOrReveal` to get instance. Spy on mock panel's `webview.postMessage` (resolves successfully). | `instance.showSessionDetail(state)` called immediately after creation | `webview.postMessage` is called (VS Code handles queuing internally); no error thrown |
