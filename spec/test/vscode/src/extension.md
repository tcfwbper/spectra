# Test Specification: `extension.test.ts`

## Source File Under Test
`vscode/src/extension.ts`

## Test File
`vscode/test/suite/extension.test.ts`

---

## `activate`

### Happy Path — activate

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_createsOutputChannel` | `unit` | Creates an OutputChannel named 'Spectra' on activation. | Stub `vscode.window.createOutputChannel` to return a mock channel. Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `ProjectRootResolver.isInitialized()` to return `true`. Stub `SessionListController`, `SessionDetailController`, and `SpectraViewProvider` constructors/methods. Stub `vscode.window.registerWebviewViewProvider`. Provide a fake `ExtensionContext` with a `subscriptions` array and `extensionUri`. | `context` (fake ExtensionContext) | `vscode.window.createOutputChannel` called with `'Spectra'` |
| `test_activate_logsActivationStart` | `unit` | Logs activation start before resolving project root. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. `ProjectRootResolver.isInitialized()` returns `true`. Capture logger `info` calls. | `context` | Logger `info` called with a message indicating activation start |
| `test_activate_resolvesProjectRoot` | `unit` | Calls ProjectRootResolver.resolve() to obtain the project root. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `ProjectRootResolver.isInitialized()` to return `true`. Stub remaining dependencies. | `context` | `ProjectRootResolver.resolve()` called exactly once |
| `test_activate_createsViewProvider` | `unit` | Constructs SpectraViewProvider with extensionUri and logger. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `ProjectRootResolver.isInitialized()` to return `true`. Spy on `SpectraViewProvider` constructor. | `context` | `SpectraViewProvider` constructed with `context.extensionUri` and logger |
| `test_activate_registersViewProvider` | `unit` | Registers the view provider with VS Code using the correct viewType and options. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `vscode.window.registerWebviewViewProvider`. | `context` | `vscode.window.registerWebviewViewProvider` called with `'spectra.sessionView'`, the viewProvider instance, and options containing `{ webviewOptions: { retainContextWhenHidden: true } }` |
| `test_activate_createsSessionListController` | `unit` | Constructs SessionListController with projectRoot and logger. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `ProjectRootResolver.isInitialized()` to return `true`. Spy on `SessionListController` constructor. | `context` | `SessionListController` constructed with `'/workspace'` and logger |
| `test_activate_createsSessionDetailController` | `unit` | Constructs SessionDetailController with projectRoot and logger. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `ProjectRootResolver.isInitialized()` to return `true`. Spy on `SessionDetailController` constructor. | `context` | `SessionDetailController` constructed with `'/workspace'` and logger |
| `test_activate_pushesAllDisposablesToSubscriptions` | `unit` | Pushes OutputChannel, controllers, viewProvider, registration, and subscriptions to context.subscriptions. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. `ProjectRootResolver.isInitialized()` returns `true`. Provide a fake `context` with an empty `subscriptions` array. | `context` | `context.subscriptions` contains at least: OutputChannel, sessionListController, sessionDetailController, viewProvider, view provider registration disposable |
| `test_activate_logsSuccessWithProjectRoot` | `unit` | Logs successful activation including the resolved projectRoot. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/my/project'`. `ProjectRootResolver.isInitialized()` returns `true`. Capture logger `info` calls. | `context` | Logger `info` called with a message containing `'/my/project'` |
| `test_activate_checksProjectInitialization` | `unit` | Calls ProjectRootResolver.isInitialized with projectRoot after resolving. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Spy on `ProjectRootResolver.isInitialized`. Stub it to return `true`. Stub remaining dependencies. | `context` | `ProjectRootResolver.isInitialized` called with `'/workspace'` |

### Happy Path — onDidUpdate subscriptions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_sessionListOnDidUpdate_cachesStateAndShowsList` | `unit` | Caches received state and calls viewProvider.showSessionList on sessionListController update. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. `ProjectRootResolver.isInitialized()` returns `true`. Configure mock `sessionListController.onDidUpdate` to accept a callback. After activation, trigger the callback with a fake state object. | Callback triggered with `state` | `viewProvider.showSessionList` called with the same `state`; state is cached internally |
| `test_activate_sessionDetailOnDidUpdate_showsDetail` | `unit` | Calls viewProvider.showSessionDetail on sessionDetailController update. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Configure mock `sessionDetailController.onDidUpdate` to accept a callback. After activation, trigger the callback with a fake detail state. | Callback triggered with `detailState` | `viewProvider.showSessionDetail` called with `detailState` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_sessionListOnDidError_showsErrorMessage` | `unit` | Shows error message when sessionListController fires onDidError. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Configure mock `sessionListController.onDidError` to accept a callback. Spy on `vscode.window.showErrorMessage`. After activation, trigger the error callback with `{ message: 'scan failed' }`. | Error event with `message: 'scan failed'` | `vscode.window.showErrorMessage` called with `'scan failed'` |
| `test_activate_sessionDetailOnDidError_showsErrorMessage` | `unit` | Shows error message when sessionDetailController fires onDidError. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Configure mock `sessionDetailController.onDidError` to accept a callback. Spy on `vscode.window.showErrorMessage`. After activation, trigger the error callback with `{ message: 'detail error' }`. | Error event with `message: 'detail error'` | `vscode.window.showErrorMessage` called with `'detail error'` |

### Happy Path — onDidReceiveMessage routing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_messageRouting_navigateToDetail` | `unit` | Routes navigateToDetail to sessionDetailController.open. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Configure mock `viewProvider.onDidReceiveMessage` to accept a callback. Spy on `sessionDetailController.open`. Trigger callback with `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }`. | Message `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }` | `sessionDetailController.open` called with `'s1'`, `'wf1'` |
| `test_activate_messageRouting_navigateToList_withCache` | `unit` | Routes navigateToList to viewProvider.showSessionList with cached state. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. First trigger `sessionListController.onDidUpdate` with `cachedState` to populate cache. Then trigger `viewProvider.onDidReceiveMessage` with `{ command: 'navigateToList' }`. | Message `{ command: 'navigateToList' }` | `viewProvider.showSessionList` called with `cachedState` |
| `test_activate_messageRouting_navigateToList_noCacheNoOp` | `unit` | No-op when navigateToList received before first onDidUpdate. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Do NOT trigger `sessionListController.onDidUpdate`. Trigger `viewProvider.onDidReceiveMessage` with `{ command: 'navigateToList' }`. | Message `{ command: 'navigateToList' }` | `viewProvider.showSessionList` is not called |
| `test_activate_messageRouting_launchSession` | `unit` | Routes launchSession to sessionListController.launch. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `sessionListController.launch`. Trigger message `{ command: 'launchSession', workflowName: 'deploy' }`. | Message `{ command: 'launchSession', workflowName: 'deploy' }` | `sessionListController.launch` called with `'deploy'` |
| `test_activate_messageRouting_terminateSession` | `unit` | Routes terminateSession to sessionListController.terminate. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `sessionListController.terminate`. Trigger message `{ command: 'terminateSession', pid: 1234 }`. | Message `{ command: 'terminateSession', pid: 1234 }` | `sessionListController.terminate` called with `1234` |
| `test_activate_messageRouting_sendEvent` | `unit` | Routes sendEvent to sessionDetailController.sendEvent. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `sessionDetailController.sendEvent`. Trigger message `{ command: 'sendEvent', eventType: 'input', message: 'hello' }`. | Message `{ command: 'sendEvent', eventType: 'input', message: 'hello' }` | `sessionDetailController.sendEvent` called with `'input'`, `'hello'` |
| `test_activate_messageRouting_unknownCommand_logsWarning` | `unit` | Logs a warning for unrecognized webview message commands. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Capture logger `warn` calls. Trigger message `{ command: 'unknownCmd' }`. | Message `{ command: 'unknownCmd' }` | Logger `warn` called with a message containing `'unknownCmd'` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_projectRootUndefined_showsNotInitializedAndReturnsEarly` | `unit` | Shows not-initialized message and returns early when projectRoot is undefined. | Stub `ProjectRootResolver.resolve()` to return `undefined`. Spy on `viewProvider.showNotInitialized`. Stub `vscode.window.registerWebviewViewProvider`. Provide a fake `context` with a `subscriptions` array. | `context` | `viewProvider.showNotInitialized()` called; no controllers created; ViewProvider registration and OutputChannel pushed to `context.subscriptions` |
| `test_activate_projectNotInitialized_showsNotInitializedAndReturnsEarly` | `unit` | Shows not-initialized message and returns early when .spectra/ directory is missing. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `ProjectRootResolver.isInitialized()` to return `false`. Spy on `viewProvider.showNotInitialized`. | `context` | `viewProvider.showNotInitialized()` called; no controllers created; ViewProvider registration and OutputChannel pushed to `context.subscriptions` |
| `test_activate_projectRootUndefined_viewProviderStillRegistered` | `unit` | ViewProvider is registered even when projectRoot is undefined. | Stub `ProjectRootResolver.resolve()` to return `undefined`. Spy on `vscode.window.registerWebviewViewProvider`. | `context` | `vscode.window.registerWebviewViewProvider` called with `'spectra.sessionView'` before early return |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_loggerWrapsOutputChannel` | `unit` | Logger adapter delegates info/warn/error to outputChannel.appendLine with severity tags. | Stub `vscode.window.createOutputChannel` to return a mock channel with a spy on `appendLine`. Stub `ProjectRootResolver.resolve()` to return `undefined` (so activate returns early after logging). Stub `vscode.window.registerWebviewViewProvider`. | `context` | `outputChannel.appendLine` called with strings containing severity prefix (e.g., `[INFO]`) |
| `test_activate_terminateFromDetailPage_routesToSessionListController` | `unit` | terminateSession from detail page routes to sessionListController.terminate identically. | Stub all dependencies. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `sessionListController.terminate`. Trigger message `{ command: 'terminateSession', pid: 5678 }` (simulating origin from detail page). | Message `{ command: 'terminateSession', pid: 5678 }` | `sessionListController.terminate` called with `5678` |
| `test_activate_acceptsOnlyOneParameter` | `unit` | activate function signature accepts exactly one parameter. | Import the `activate` function from the extension module. | Inspect `activate.length` | `activate.length` equals `1` |
| `test_activate_constructsAllCollaboratorsInternally` | `unit` | All collaborators are constructed inside activate without external injection. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `SessionListController`, `SessionDetailController`, and `SpectraViewProvider` constructors. | `context` (fake ExtensionContext with no extra arguments) | All three constructors are called; `activate` was called with only `context` (no second argument) |
| `test_activate_registersViewProviderSynchronouslyDuringActivation` | `unit` | ViewProvider registration occurs synchronously during activation, before any async work. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. `ProjectRootResolver.isInitialized()` returns `true`. Spy on `vscode.window.registerWebviewViewProvider`. Record call order of `registerWebviewViewProvider` relative to any async operations. | `context` | `registerWebviewViewProvider` is called with `'spectra.sessionView'` synchronously within the `activate` call before any `await` or microtask completes |

---

## `deactivate`

### Happy Path — deactivate

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_deactivate_isEmptyFunction` | `unit` | deactivate does nothing — cleanup is handled by context.subscriptions. | None | (no arguments) | Function returns `undefined`; no errors thrown |
