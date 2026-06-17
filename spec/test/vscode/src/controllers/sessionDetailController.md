# Test Specification: `sessionDetailController.test.ts`

## Source File Under Test
`vscode/src/controllers/sessionDetailController.ts`

## Test File
`vscode/test/suite/sessionDetailController.test.ts`

---

## `SessionDetailController`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should store projectRoot and logger` | `unit` | Construction stores provided parameters. | Stub `vscode.EventEmitter` constructor. | `projectRoot='/project'`, `logger=mockLogger` | Instance is created without error |
| `should expose onDidUpdate and onDidError events` | `unit` | Construction creates and exposes public event accessors. | Stub `vscode.EventEmitter` constructor. | `projectRoot='/project'`, `logger=mockLogger` | `instance.onDidUpdate` and `instance.onDidError` are defined and are functions (event accessors) |
| `should not create EventWatcher during construction` | `unit` | EventWatcher is created lazily in open(). | Spy on `EventWatcher` constructor. | `projectRoot='/project'`, `logger=mockLogger` | `EventWatcher` constructor is not called |
| `should initialize with null currentSessionId and zero generation` | `unit` | Internal state starts at defaults. | Stub `vscode.EventEmitter` constructor. | `projectRoot='/project'`, `logger=mockLogger` | Instance created; no `onDidUpdate` fired during construction |

### Happy Path — open

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should create EventWatcher and fire onDidUpdate with assembled state` | `unit` | Opening a session creates watcher and scans data. | Stub `EventWatcher` constructor to return mock with `onDidChange` stub and `dispose` spy. Stub `WorkflowDefinitionParser.parse` to resolve with `{ entryNode: 'start', eventTypes: ['submit'] }`. Stub `EventScanner.scan` to resolve with `[{type:'submit', ts:100}]`. Stub `SessionScanner.scan` to resolve with `[{id:'s1', currentState:'running', status:'running', pid:42}]`. Register spy on `onDidUpdate`. | `instance.open('s1', 'wf1')` | `onDidUpdate` fires with `{ sessionId:'s1', workflowName:'wf1', entryNode:'start', currentState:'running', status:'running', pid:42, eventTypes:['submit'], events:[{type:'submit', ts:100}] }` |
| `should pass correct arguments to EventWatcher constructor` | `unit` | EventWatcher is created with projectRoot and sessionId. | Spy on `EventWatcher` constructor. Stub scanners to resolve. | `instance.open('sess-abc', 'wf1')` | `EventWatcher` constructor called with `('/project', 'sess-abc')` |
| `should pass correct arguments to WorkflowDefinitionParser.parse` | `unit` | Parser is called with projectRoot, workflowName, and logger. | Stub `WorkflowDefinitionParser.parse` to resolve. Stub other scanners. | `instance.open('s1', 'my-workflow')` | `WorkflowDefinitionParser.parse` called with `('/project', 'my-workflow', mockLogger)` |
| `should subscribe to EventWatcher.onDidChange` | `unit` | After open, watcher changes trigger re-scan. | Stub `EventWatcher` constructor returning mock. Stub scanners to resolve. | `instance.open('s1', 'wf1')` | Subscription registered on mock's `onDidChange` |

### Happy Path — internal scan routine

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should re-scan and fire onDidUpdate when onDidChange fires` | `unit` | File change triggers fresh scan. | Stub `EventWatcher` with controllable `onDidChange`. Let `open('s1','wf1')` complete with initial state. Stub `EventScanner.scan` to resolve with new events `[{type:'ack', ts:200}]`. Stub `SessionScanner.scan` to resolve with updated state. Register spy on `onDidUpdate`. | Trigger `onDidChange` on mock watcher | `onDidUpdate` fires with updated events and state; `entryNode` preserved from initial open |
| `should include previously stored entryNode and eventTypes in re-scan state` | `unit` | Re-scan uses cached workflow parse results. | Let `open('s1','wf1')` complete with `entryNode:'start'`, `eventTypes:['go']`. Stub `EventScanner.scan` to resolve with new events. | Trigger `onDidChange` | `onDidUpdate` state includes `entryNode:'start'` and `eventTypes:['go']` |

### Happy Path — sendEvent

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should call EventDispatcher.dispatch with correct arguments` | `unit` | Successful event dispatch. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. | `instance.sendEvent('submit', 'hello')` | `EventDispatcher.dispatch` called with `('submit', 's1', 'hello', '/project', mockLogger)` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should propagate EventWatcher construction error to caller` | `unit` | EventWatcher constructor throw propagates. | Stub `EventWatcher` constructor to throw `new Error('ENOENT')`. | `instance.open('s1', 'wf1')` | The call throws/rejects with the `ENOENT` error |
| `should fire onDidError and log when sendEvent dispatch fails with ENOENT` | `unit` | Spawn failure fires error event. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to throw `new Error('ENOENT')`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.sendEvent('submit', 'msg')` | `onDidError` fires with the error; `mockLogger.error` is called |
| `should fire onDidError and log when sendEvent dispatch fails with EACCES` | `unit` | Permission failure fires error event. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to throw `new Error('EACCES')`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.sendEvent('submit', 'msg')` | `onDidError` fires with the error; `mockLogger.error` is called |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should discard stale scan results when open is called again` | `unit` | Generation counter prevents stale results from firing. | Stub `EventWatcher` constructor. Stub `EventScanner.scan` for first open to return a deferred promise. Register spy on `onDidUpdate`. | Call `instance.open('s1','wf1')`, then immediately call `instance.open('s2','wf2')` before first resolves, then resolve first open's scan | `onDidUpdate` is not fired for s1 results; only s2 results fire |
| `should dispose previous watcher when open is called again` | `unit` | Previous watcher is cleaned up on re-open. | Stub `EventWatcher` constructor returning mock with `dispose` spy. Let first `open('s1','wf1')` complete. | Call `instance.open('s2','wf2')` | First watcher's `dispose` is called before second watcher is created |
| `should coalesce overlapping scans via dirty flag` | `unit` | Multiple onDidChange during in-flight scan results in one re-scan. | Let `open('s1','wf1')` complete. Stub `EventScanner.scan` to return a deferred promise. Register spy on `onDidUpdate`. | Trigger `onDidChange` three times while scan is in-flight, then resolve | `EventScanner.scan` called twice total (in-flight + one re-scan after dirty flag) |
| `should discard scan result when generation changes mid-scan` | `unit` | Stale onDidChange callback is discarded. | Let `open('s1','wf1')` complete. Stub `EventScanner.scan` with deferred promise. Register spy on `onDidUpdate`. | Trigger `onDidChange`, then call `open('s2','wf2')` before scan resolves, then resolve original scan | `onDidUpdate` is not fired for the stale scan |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose watcher and emitters on dispose` | `unit` | Dispose releases all resources. | Let `open('s1','wf1')` complete. Stub watcher with `dispose` spy. | Call `instance.dispose()` | Watcher's `dispose()` and both event emitter `dispose()` methods are called |
| `should suppress onDidUpdate after dispose` | `unit` | No state events after disposal. | Let `open('s1','wf1')` complete. Stub `EventScanner.scan` with deferred promise. Trigger `onDidChange`. Register spy on `onDidUpdate`. | Call `instance.dispose()`, then resolve the pending scan | `onDidUpdate` is not fired |
| `should no-op on open after dispose` | `unit` | Open after disposal does nothing. | Construct instance, call `instance.dispose()`. Spy on `EventWatcher` constructor. | Call `instance.open('s1','wf1')` | `EventWatcher` constructor is not called; no error thrown |
| `should no-op on sendEvent after dispose` | `unit` | sendEvent after disposal does nothing. | Construct instance, call `instance.dispose()`. Spy on `EventDispatcher.dispatch`. | Call `instance.sendEvent('submit','msg')` | `EventDispatcher.dispatch` is not called; no error thrown |
| `should set watcher to null after dispose` | `unit` | Watcher reference cleared. | Let `open('s1','wf1')` complete. | Call `instance.dispose()` | Subsequent `open` after un-disposed re-construction would not double-dispose |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should handle multiple dispose calls without error` | `unit` | Repeated disposal does not throw. | Let `open('s1','wf1')` complete. Stub watcher with `dispose` spy. | Call `instance.dispose()` three times | No error is thrown on any call |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should push empty events array when EventScanner returns empty` | `unit` | No events file yields empty array in state. | Stub `EventWatcher` constructor. Stub `EventScanner.scan` to resolve with `[]`. Stub `WorkflowDefinitionParser.parse` to resolve with `{ entryNode:'start', eventTypes:['go'] }`. Stub `SessionScanner.scan` to resolve with `[{id:'s1', currentState:'start', status:'running', pid:1}]`. Register spy on `onDidUpdate`. | `instance.open('s1','wf1')` | `onDidUpdate` fires with `events: []` |
| `should push empty eventTypes when WorkflowDefinitionParser returns empty` | `unit` | No transitions yields empty eventTypes. | Stub `EventWatcher` constructor. Stub `WorkflowDefinitionParser.parse` to resolve with `{ entryNode:'', eventTypes:[] }`. Stub `EventScanner.scan` to resolve with `[]`. Stub `SessionScanner.scan` to resolve with `[{id:'s1', currentState:'', status:'initializing', pid:0}]`. Register spy on `onDidUpdate`. | `instance.open('s1','wf1')` | `onDidUpdate` fires with `entryNode: ''` and `eventTypes: []` |
| `should default session fields when SessionScanner has no matching session` | `unit` | Missing session yields defaults. | Stub `EventWatcher` constructor. Stub `SessionScanner.scan` to resolve with `[{id:'other', currentState:'done', status:'completed', pid:99}]`. Stub other scanners. Register spy on `onDidUpdate`. | `instance.open('s1','wf1')` | `onDidUpdate` fires with `currentState: ''`, `status: 'initializing'`, `pid: 0` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should not read or write any files directly` | `unit` | Controller never performs filesystem I/O. | Stub watchers and scanners. Spy on `fs.readFile`, `fs.writeFile`. | Construct, open, trigger onDidChange, call sendEvent | None of the filesystem spies are called |
| `should not spawn processes directly` | `unit` | Controller delegates spawning to EventDispatcher. | Stub watchers and scanners. Spy on `child_process.spawn`, `child_process.exec`. | Call `instance.sendEvent('submit','msg')` | None of the process spies are called directly; only `EventDispatcher.dispatch` is called |
