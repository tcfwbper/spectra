# Test Specification: `sessionListController.test.ts`

## Source File Under Test
`vscode/src/controllers/sessionListController.ts`

## Test File
`vscode/test/suite/sessionListController.test.ts`

---

## `SessionListController`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should create SessionWatcher and WorkflowWatcher during construction` | `unit` | Construction instantiates both watchers. | Stub `SessionWatcher` and `WorkflowWatcher` constructors to return mock instances with `onDidChange` event stubs and `dispose` spies. Stub `SessionScanner.scan` and `WorkflowScanner.scan` to resolve with empty arrays. | `projectRoot='/project'`, `logger=mockLogger` | Both `SessionWatcher` and `WorkflowWatcher` constructors are called with `'/project'` |
| `should expose onDidUpdate and onDidError events` | `unit` | Construction creates and exposes the public event accessors. | Stub `SessionWatcher` and `WorkflowWatcher` constructors. Stub scanners to resolve with empty arrays. | `projectRoot='/project'`, `logger=mockLogger` | `instance.onDidUpdate` and `instance.onDidError` are defined and are functions (event accessors) |
| `should kick off initial scan asynchronously without blocking construction` | `unit` | Construction returns immediately while scan is still pending. | Stub `SessionWatcher` and `WorkflowWatcher` constructors. Stub `SessionScanner.scan` to return an unresolved promise. Stub `WorkflowScanner.scan` to return an unresolved promise. | `projectRoot='/project'`, `logger=mockLogger` | Constructor returns synchronously; `SessionScanner.scan` and `WorkflowScanner.scan` are called |

### Happy Path — onDidUpdate

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should fire onDidUpdate with sessions and workflows after initial scan completes` | `unit` | Initial scan produces composite state. | Stub watchers. Stub `SessionScanner.scan` to resolve with `[{id:'s1', createdAt:100}]`. Stub `WorkflowScanner.scan` to resolve with `['wf1']`. Register spy on `onDidUpdate`. | Construction completes, await pending promises | `onDidUpdate` fires with `{ sessions: [{id:'s1', createdAt:100}], workflows: ['wf1'] }` |
| `should fire onDidUpdate when SessionWatcher triggers onDidChange` | `unit` | Session file change triggers re-scan and state push. | Stub watchers. Stub `SessionScanner.scan` to resolve with `[{id:'s2', createdAt:200}]`. Stub `WorkflowScanner.scan` to resolve with `['wf1']`. Let initial scan complete. Register spy on `onDidUpdate`. | Trigger `SessionWatcher.onDidChange` | `onDidUpdate` fires with updated sessions array |
| `should fire onDidUpdate when WorkflowWatcher triggers onDidChange` | `unit` | Workflow file change triggers re-scan and state push. | Stub watchers. Stub `WorkflowScanner.scan` to resolve with `['wf1','wf2']`. Let initial scan complete. Register spy on `onDidUpdate`. | Trigger `WorkflowWatcher.onDidChange` | `onDidUpdate` fires with updated workflows array |
| `should push full composite state even when only sessions changed` | `unit` | Partial change still results in full state. | Stub watchers. Let initial scan complete with sessions `[{id:'s1'}]` and workflows `['wf1']`. Register spy on `onDidUpdate`. Stub `SessionScanner.scan` to resolve with `[{id:'s1'},{id:'s2'}]`. | Trigger `SessionWatcher.onDidChange` | `onDidUpdate` fires with both updated `sessions` and existing `workflows` |

### Happy Path — launch

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should call SessionLauncher.launch with workflowName, projectRoot, and logger` | `unit` | Successful launch delegates to SessionLauncher. | Stub `SessionLauncher.launch` to resolve successfully. Stub watchers and scanners. | `instance.launch('my-workflow')` | `SessionLauncher.launch` called with `('my-workflow', '/project', mockLogger)` |

### Happy Path — terminate

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should call SessionTerminator.terminate with pid and logger` | `unit` | Terminate delegates to SessionTerminator. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'sigterm', terminated: true }`. Stub watchers and scanners. | `instance.terminate(1234)` | `SessionTerminator.terminate` called with `(1234, mockLogger)` |
| `should treat already_dead as success without firing onDidError` | `unit` | Process already dead is not an error. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'already_dead', terminated: true }`. Register spy on `onDidError`. | `instance.terminate(5678)` | `onDidError` is not fired |
| `should treat sigterm terminated as success without firing onDidError` | `unit` | Successful SIGTERM is not an error. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'sigterm', terminated: true }`. Register spy on `onDidError`. | `instance.terminate(5678)` | `onDidError` is not fired |
| `should treat sigkill terminated as success without firing onDidError` | `unit` | Successful SIGKILL is not an error. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'sigkill', terminated: true }`. Register spy on `onDidError`. | `instance.terminate(5678)` | `onDidError` is not fired |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should fire onDidError and log when launch throws` | `unit` | Launch failure fires error event. | Stub `SessionLauncher.launch` to reject with `new Error('ENOENT')`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.launch('bad-workflow')` | `onDidError` fires with the error; `mockLogger.error` is called |
| `should fire onDidError and log when terminate returns not_spectra` | `unit` | PID reuse detection fires error. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'not_spectra', terminated: false }`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.terminate(9999)` | `onDidError` fires with Error describing PID reuse; `mockLogger.error` is called |
| `should fire onDidError and log when terminate returns EPERM` | `unit` | Permission failure fires error. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'sigterm', terminated: false, error: new Error('EPERM') }`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.terminate(9999)` | `onDidError` fires with Error describing permission failure; `mockLogger.error` is called |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should coalesce overlapping session scans via dirty flag` | `unit` | Multiple onDidChange during in-flight scan results in one re-scan. | Stub watchers. Stub `SessionScanner.scan` to return a deferred promise (manually resolved). Register spy on `onDidUpdate`. Let initial scan complete. | Trigger `SessionWatcher.onDidChange` three times while first scan is in-flight, then resolve scan | `SessionScanner.scan` is called exactly twice total (initial in-flight + one re-scan); `onDidUpdate` fires twice |
| `should coalesce overlapping workflow scans independently` | `unit` | Workflow scan coalescing is independent from session scan. | Stub watchers. Stub `WorkflowScanner.scan` to return a deferred promise. Let initial scan complete. | Trigger `WorkflowWatcher.onDidChange` twice while scan is in-flight, then resolve | `WorkflowScanner.scan` is called exactly twice (in-flight + one re-scan) |
| `should run session and workflow scans concurrently` | `unit` | Both scan types run in parallel without blocking each other. | Stub watchers. Stub both scanners with deferred promises. | Trigger both `SessionWatcher.onDidChange` and `WorkflowWatcher.onDidChange` simultaneously | Both `SessionScanner.scan` and `WorkflowScanner.scan` are called without waiting for the other |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose watchers and emitters on dispose` | `unit` | Dispose releases all resources. | Stub watchers with `dispose` spies. Let initial scan complete. | Call `instance.dispose()` | `SessionWatcher.dispose()`, `WorkflowWatcher.dispose()`, and both event emitter `dispose()` methods are called |
| `should suppress onDidUpdate after dispose` | `unit` | No state events after disposal. | Stub watchers. Stub `SessionScanner.scan` with a deferred promise. Trigger `SessionWatcher.onDidChange`. Register spy on `onDidUpdate`. | Call `instance.dispose()`, then resolve the pending scan | `onDidUpdate` is not fired |
| `should suppress onDidError from launch after dispose` | `unit` | No error events from launch after disposal. | Stub `SessionLauncher.launch` to reject. Register spy on `onDidError`. | Call `instance.dispose()`, then call `instance.launch('wf')` and await | `onDidError` is not fired |
| `should suppress onDidError from terminate after dispose` | `unit` | No error events from terminate after disposal. | Stub `SessionTerminator.terminate` to resolve with `{ method: 'not_spectra', terminated: false }`. Register spy on `onDidError`. | Call `instance.dispose()`, then call `instance.terminate(123)` and await | `onDidError` is not fired |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should handle multiple dispose calls without error` | `unit` | Repeated disposal does not throw. | Stub watchers with `dispose` spies. | Call `instance.dispose()` three times | No error is thrown on any call |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should not read or write any files directly` | `unit` | Controller never performs filesystem I/O. | Stub watchers and scanners. Spy on `fs.readFile`, `fs.writeFile`. | Construct instance, let initial scan complete, trigger onDidChange, call launch and terminate | None of the filesystem spies are called |
| `should not spawn processes directly` | `unit` | Controller delegates spawning to SessionLauncher. | Stub watchers and scanners. Spy on `child_process.spawn`, `child_process.exec`. | Call `instance.launch('wf')` | None of the process spies are called directly; only `SessionLauncher.launch` is called |
| `should not send signals directly` | `unit` | Controller delegates signaling to SessionTerminator. | Stub watchers and scanners. Spy on `process.kill`. | Call `instance.terminate(123)` | `process.kill` is not called directly; only `SessionTerminator.terminate` is called |
