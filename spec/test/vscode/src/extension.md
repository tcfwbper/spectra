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
| `test_activate_createsOutputChannel` | `unit` | Creates an OutputChannel named 'Spectra' on activation. | Stub `vscode.window.createOutputChannel` to return a mock channel. Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `SessionListController`, `SessionDetailController`, and `SpectraPanel.createOrReveal` constructors/methods. Provide a fake `ExtensionContext` with a `subscriptions` array. | `context` (fake ExtensionContext) | `vscode.window.createOutputChannel` called with `'Spectra'` |
| `test_activate_logsActivationStart` | `unit` | Logs activation start before resolving project root. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. Capture logger `info` calls. | `context` | Logger `info` called with a message indicating activation start |
| `test_activate_resolvesProjectRoot` | `unit` | Calls ProjectRootResolver.resolve() to obtain the project root. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub remaining dependencies. | `context` | `ProjectRootResolver.resolve()` called exactly once |
| `test_activate_createsSessionListController` | `unit` | Constructs SessionListController with projectRoot and logger. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Spy on `SessionListController` constructor. | `context` | `SessionListController` constructed with `'/workspace'` and logger |
| `test_activate_createsSessionDetailController` | `unit` | Constructs SessionDetailController with projectRoot and logger. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Spy on `SessionDetailController` constructor. | `context` | `SessionDetailController` constructed with `'/workspace'` and logger |
| `test_activate_callsCreateOrReveal` | `unit` | Calls SpectraPanel.createOrReveal with context, extensionUri, and logger. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. Spy on `SpectraPanel.createOrReveal`. | `context` | `SpectraPanel.createOrReveal` called with `context`, `context.extensionUri`, logger |
| `test_activate_registersOpenPanelCommand` | `unit` | Registers the spectra.openPanel command. | Stub all dependencies. Spy on `vscode.commands.registerCommand`. | `context` | `vscode.commands.registerCommand` called with `'spectra.openPanel'` and a handler function |
| `test_activate_pushesAllDisposablesToSubscriptions` | `unit` | Pushes OutputChannel, controllers, panel, command registration, and subscriptions to context.subscriptions. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. Provide a fake `context` with an empty `subscriptions` array. | `context` | `context.subscriptions` contains at least: OutputChannel, sessionListController, sessionDetailController, panel, command registration disposable |
| `test_activate_logsSuccessWithProjectRoot` | `unit` | Logs successful activation including the resolved projectRoot. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/my/project'`. Capture logger `info` calls. | `context` | Logger `info` called with a message containing `'/my/project'` |

### Happy Path — onDidUpdate subscriptions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_sessionListOnDidUpdate_cachesStateAndShowsList` | `unit` | Caches received state and calls panel.showSessionList on sessionListController update. | Stub all dependencies. `ProjectRootResolver.resolve()` returns `'/workspace'`. Configure mock `sessionListController.onDidUpdate` to accept a callback. After activation, trigger the callback with a fake state object. | Callback triggered with `state` | `panel.showSessionList` called with the same `state`; state is cached internally |
| `test_activate_sessionDetailOnDidUpdate_showsDetail` | `unit` | Calls panel.showSessionDetail on sessionDetailController update. | Stub all dependencies. Configure mock `sessionDetailController.onDidUpdate` to accept a callback. After activation, trigger the callback with a fake detail state. | Callback triggered with `detailState` | `panel.showSessionDetail` called with `detailState` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_sessionListOnDidError_showsErrorMessage` | `unit` | Shows error message when sessionListController fires onDidError. | Stub all dependencies. Configure mock `sessionListController.onDidError` to accept a callback. Spy on `vscode.window.showErrorMessage`. After activation, trigger the error callback with `{ message: 'scan failed' }`. | Error event with `message: 'scan failed'` | `vscode.window.showErrorMessage` called with `'scan failed'` |
| `test_activate_sessionDetailOnDidError_showsErrorMessage` | `unit` | Shows error message when sessionDetailController fires onDidError. | Stub all dependencies. Configure mock `sessionDetailController.onDidError` to accept a callback. Spy on `vscode.window.showErrorMessage`. After activation, trigger the error callback with `{ message: 'detail error' }`. | Error event with `message: 'detail error'` | `vscode.window.showErrorMessage` called with `'detail error'` |
| `test_activate_createOrRevealThrows_propagatesError` | `unit` | Error propagates when SpectraPanel.createOrReveal throws. | Stub `ProjectRootResolver.resolve()` to return `'/workspace'`. Stub `SpectraPanel.createOrReveal` to throw `new Error('internal error')`. | `context` | `activate` throws/rejects with `'internal error'` |

### Happy Path — onDidReceiveMessage routing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_messageRouting_navigateToDetail` | `unit` | Routes navigateToDetail to sessionDetailController.open. | Stub all dependencies. Configure mock `panel.onDidReceiveMessage` to accept a callback. Spy on `sessionDetailController.open`. Trigger callback with `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }`. | Message `{ command: 'navigateToDetail', sessionId: 's1', workflowName: 'wf1' }` | `sessionDetailController.open` called with `'s1'`, `'wf1'` |
| `test_activate_messageRouting_navigateToList_withCache` | `unit` | Routes navigateToList to panel.showSessionList with cached state. | Stub all dependencies. First trigger `sessionListController.onDidUpdate` with `cachedState` to populate cache. Then trigger `panel.onDidReceiveMessage` with `{ command: 'navigateToList' }`. | Message `{ command: 'navigateToList' }` | `panel.showSessionList` called with `cachedState` |
| `test_activate_messageRouting_navigateToList_noCacheNoOp` | `unit` | No-op when navigateToList received before first onDidUpdate. | Stub all dependencies. Do NOT trigger `sessionListController.onDidUpdate`. Trigger `panel.onDidReceiveMessage` with `{ command: 'navigateToList' }`. | Message `{ command: 'navigateToList' }` | `panel.showSessionList` is not called |
| `test_activate_messageRouting_launchSession` | `unit` | Routes launchSession to sessionListController.launch. | Stub all dependencies. Spy on `sessionListController.launch`. Trigger message `{ command: 'launchSession', workflowName: 'deploy' }`. | Message `{ command: 'launchSession', workflowName: 'deploy' }` | `sessionListController.launch` called with `'deploy'` |
| `test_activate_messageRouting_terminateSession` | `unit` | Routes terminateSession to sessionListController.terminate. | Stub all dependencies. Spy on `sessionListController.terminate`. Trigger message `{ command: 'terminateSession', pid: 1234 }`. | Message `{ command: 'terminateSession', pid: 1234 }` | `sessionListController.terminate` called with `1234` |
| `test_activate_messageRouting_sendEvent` | `unit` | Routes sendEvent to sessionDetailController.sendEvent. | Stub all dependencies. Spy on `sessionDetailController.sendEvent`. Trigger message `{ command: 'sendEvent', eventType: 'input', message: 'hello' }`. | Message `{ command: 'sendEvent', eventType: 'input', message: 'hello' }` | `sessionDetailController.sendEvent` called with `'input'`, `'hello'` |
| `test_activate_messageRouting_unknownCommand_logsWarning` | `unit` | Logs a warning for unrecognized webview message commands. | Stub all dependencies. Capture logger `warn` calls. Trigger message `{ command: 'unknownCmd' }`. | Message `{ command: 'unknownCmd' }` | Logger `warn` called with a message containing `'unknownCmd'` |

### Happy Path — onDidDispose

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_panelOnDidDispose_disposesBothControllers` | `unit` | Disposes both controllers when panel is disposed. | Stub all dependencies. Configure mock `panel.onDidDispose` to accept a callback. Spy on `sessionListController.dispose` and `sessionDetailController.dispose`. Trigger the dispose callback. | Panel dispose event | `sessionListController.dispose()` and `sessionDetailController.dispose()` each called once |

### Happy Path — spectra.openPanel command

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_openPanelCommand_callsCreateOrReveal` | `unit` | The spectra.openPanel command handler calls SpectraPanel.createOrReveal. | Stub all dependencies. Capture the handler registered with `vscode.commands.registerCommand` for `'spectra.openPanel'`. Spy on `SpectraPanel.createOrReveal`. Invoke the captured handler. | Command invoked | `SpectraPanel.createOrReveal` called with `context`, `context.extensionUri`, logger |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_projectRootUndefined_showsErrorAndReturnsEarly` | `unit` | Shows error and returns early when projectRoot is undefined. | Stub `ProjectRootResolver.resolve()` to return `undefined`. Spy on `vscode.window.showErrorMessage`. Provide a fake `context` with a `subscriptions` array. | `context` | `vscode.window.showErrorMessage` called with a descriptive message; no commands registered; no controllers created; only OutputChannel pushed to `context.subscriptions` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_activate_loggerWrapsOutputChannel` | `unit` | Logger adapter delegates info/warn/error to outputChannel.appendLine with severity tags. | Stub `vscode.window.createOutputChannel` to return a mock channel with a spy on `appendLine`. Stub `ProjectRootResolver.resolve()` to return `undefined` (so activate returns early after logging). | `context` | `outputChannel.appendLine` called with strings containing severity prefix (e.g., `[INFO]`) |
| `test_activate_terminateFromDetailPage_routesToSessionListController` | `unit` | terminateSession from detail page routes to sessionListController.terminate identically. | Stub all dependencies. Spy on `sessionListController.terminate`. Trigger message `{ command: 'terminateSession', pid: 5678 }` (simulating origin from detail page). | Message `{ command: 'terminateSession', pid: 5678 }` | `sessionListController.terminate` called with `5678` |

---

## `deactivate`

### Happy Path — deactivate

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `test_deactivate_isEmptyFunction` | `unit` | deactivate does nothing — cleanup is handled by context.subscriptions. | None | (no arguments) | Function returns `undefined`; no errors thrown |
