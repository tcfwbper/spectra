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
| `should default fallbackScanDelayMs to 800 when not provided` | `unit` | Optional parameter defaults correctly. | Stub `vscode.EventEmitter` constructor. Use fake timers. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. | `new SessionDetailController('/project', mockLogger)` then `sendEvent('go','msg')` | Fallback timer is scheduled with 800ms delay |
| `should accept custom fallbackScanDelayMs` | `unit` | Custom delay is stored and used for timer scheduling. | Stub `vscode.EventEmitter` constructor. Use fake timers. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. | `new SessionDetailController('/project', mockLogger, 0)` then `sendEvent('go','msg')` | Fallback timer is scheduled with 0ms delay |

### Happy Path — open

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should create EventWatcher and fire onDidUpdate with assembled state` | `unit` | Opening a session creates watcher and scans data. | Stub `EventWatcher` constructor to return mock with `onDidChange` stub and `dispose` spy. Stub `WorkflowDefinitionParser.parse` to resolve with `{ entryNode: 'start', eventTypes: ['submit'] }`. Stub `EventScanner.scan` to resolve with `[{type:'submit', ts:100}]`. Stub `SessionScanner.scan` to resolve with `[{id:'s1', currentState:'running', status:'running', pid:42}]`. Register spy on `onDidUpdate`. | `instance.open('s1', 'wf1')` | `onDidUpdate` fires with `{ sessionId:'s1', workflowName:'wf1', entryNode:'start', currentState:'running', status:'running', pid:42, eventTypes:['submit'], events:[{type:'submit', ts:100}] }` |
| `should pass correct arguments to EventWatcher constructor` | `unit` | EventWatcher is created with projectRoot and sessionId. | Spy on `EventWatcher` constructor. Stub scanners to resolve. | `instance.open('sess-abc', 'wf1')` | `EventWatcher` constructor called with `('/project', 'sess-abc')` |
| `should pass correct arguments to WorkflowDefinitionParser.parse` | `unit` | Parser is called with projectRoot, workflowName, and logger. | Stub `WorkflowDefinitionParser.parse` to resolve. Stub other scanners. | `instance.open('s1', 'my-workflow')` | `WorkflowDefinitionParser.parse` called with `('/project', 'my-workflow', mockLogger)` |
| `should subscribe to EventWatcher.onDidChange` | `unit` | After open, watcher changes trigger re-scan. | Stub `EventWatcher` constructor returning mock. Stub scanners to resolve. | `instance.open('s1', 'wf1')` | Subscription registered on mock's `onDidChange` |
| `should cancel pending fallback timer on open` | `unit` | Opening a new session cancels any existing fallback timer. | Use fake timers. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Call `sendEvent('go','msg')` to schedule timer. Stub scanners for second open. | Call `instance.open('s2','wf2')`, then advance fake timers past `fallbackScanDelayMs` | No fallback scan runs for s1; no extra `onDidUpdate` fires from the cancelled timer |

### Happy Path — internal scan routine

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should re-scan and fire onDidUpdate when onDidChange fires` | `unit` | File change triggers fresh scan. | Stub `EventWatcher` with controllable `onDidChange`. Let `open('s1','wf1')` complete with initial state. Stub `EventScanner.scan` to resolve with new events `[{type:'ack', ts:200}]`. Stub `SessionScanner.scan` to resolve with updated state. Register spy on `onDidUpdate`. | Trigger `onDidChange` on mock watcher | `onDidUpdate` fires with updated events and state; `entryNode` preserved from initial open |
| `should include previously stored entryNode and eventTypes in re-scan state` | `unit` | Re-scan uses cached workflow parse results. | Let `open('s1','wf1')` complete with `entryNode:'start'`, `eventTypes:['go']`. Stub `EventScanner.scan` to resolve with new events. | Trigger `onDidChange` | `onDidUpdate` state includes `entryNode:'start'` and `eventTypes:['go']` |

### Happy Path — sendEvent

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should call EventDispatcher.dispatch with correct arguments` | `unit` | Successful event dispatch. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. | `instance.sendEvent('submit', 'hello')` | `EventDispatcher.dispatch` called with `('submit', 's1', 'hello', '/project', mockLogger)` |
| `should return true when dispatch succeeds` | `unit` | Successful dispatch returns true. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve successfully. | `instance.sendEvent('submit', 'hello')` | Returned promise resolves to `true` |
| `should return false when disposed` | `unit` | Disposed controller returns false without dispatching. | Construct instance, call `instance.dispose()`. Spy on `EventDispatcher.dispatch`. | `instance.sendEvent('submit', 'msg')` | Returned promise resolves to `false`; `EventDispatcher.dispatch` is not called |
| `should schedule fallback timer after successful dispatch when session is open` | `unit` | Fallback timer is scheduled to trigger a scan. | Use fake timers. Construct with `fallbackScanDelayMs=50`. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Stub `EventScanner.scan` and `SessionScanner.scan` to resolve with updated state. Register spy on `onDidUpdate`. | Call `instance.sendEvent('go','msg')`, then advance fake timer by 50ms | `onDidUpdate` fires with refreshed state from the fallback scan |
| `should log info when fallback timer fires` | `unit` | Fallback scan logs a diagnostic message. | Use fake timers. Construct with `fallbackScanDelayMs=10`. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Stub scanners to resolve. Spy on `mockLogger.info`. | Call `instance.sendEvent('go','msg')`, then advance fake timer by 10ms | `mockLogger.info` is called with a message indicating fallback scan triggered |
| `should not schedule fallback timer when currentWatcher is null` | `unit` | No timer if no session is open. | Use fake timers. Construct instance (do not call `open()`). Stub `EventDispatcher.dispatch` to resolve. Register spy on `onDidUpdate`. | Call `instance.sendEvent('go','msg')`, then advance fake timer past `fallbackScanDelayMs` | Returned promise resolves to `true`; no `onDidUpdate` fires; no timer scheduled |
| `should debounce fallback timer on rapid sendEvent calls` | `unit` | Multiple rapid calls reset the timer; only one fires. | Use fake timers. Construct with `fallbackScanDelayMs=100`. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Stub scanners to resolve. Register spy on `onDidUpdate`. | Call `sendEvent` three times at 0ms, 30ms, 60ms (advancing fake timer between); then advance to 160ms | Only one fallback scan fires (at 60ms + 100ms = 160ms); `EventScanner.scan` called once for the fallback |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should propagate EventWatcher construction error to caller` | `unit` | EventWatcher constructor throw propagates. | Stub `EventWatcher` constructor to throw `new Error('ENOENT')`. | `instance.open('s1', 'wf1')` | The call throws/rejects with the `ENOENT` error |
| `should fire onDidError and log when sendEvent dispatch fails with ENOENT` | `unit` | Spawn failure fires error event and returns false. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to throw `new Error('ENOENT')`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.sendEvent('submit', 'msg')` | `onDidError` fires with the error; `mockLogger.error` is called; returned promise resolves to `false` |
| `should fire onDidError and log when sendEvent dispatch fails with EACCES` | `unit` | Permission failure fires error event and returns false. | Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to throw `new Error('EACCES')`. Register spy on `onDidError`. Spy on `mockLogger.error`. | `instance.sendEvent('submit', 'msg')` | `onDidError` fires with the error; `mockLogger.error` is called; returned promise resolves to `false` |
| `should not fire onDidError when fallback scan throws` | `unit` | Fallback scan error is logged but not surfaced as onDidError. | Use fake timers. Construct with `fallbackScanDelayMs=10`. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Stub `EventScanner.scan` to reject with `new Error('ENOENT')` on next call. Spy on `mockLogger.error`. Register spy on `onDidError`. | Call `instance.sendEvent('go','msg')`, then advance fake timer by 10ms | `mockLogger.error` is called with the error; `onDidError` is NOT fired; controller continues operating |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should discard stale scan results when open is called again` | `unit` | Generation counter prevents stale results from firing. | Stub `EventWatcher` constructor. Stub `EventScanner.scan` for first open to return a deferred promise. Register spy on `onDidUpdate`. | Call `instance.open('s1','wf1')`, then immediately call `instance.open('s2','wf2')` before first resolves, then resolve first open's scan | `onDidUpdate` is not fired for s1 results; only s2 results fire |
| `should dispose previous watcher when open is called again` | `unit` | Previous watcher is cleaned up on re-open. | Stub `EventWatcher` constructor returning mock with `dispose` spy. Let first `open('s1','wf1')` complete. | Call `instance.open('s2','wf2')` | First watcher's `dispose` is called before second watcher is created |
| `should coalesce overlapping scans via dirty flag` | `unit` | Multiple onDidChange during in-flight scan results in one re-scan. | Let `open('s1','wf1')` complete. Stub `EventScanner.scan` to return a deferred promise. Register spy on `onDidUpdate`. | Trigger `onDidChange` three times while scan is in-flight, then resolve | `EventScanner.scan` called twice total (in-flight + one re-scan after dirty flag) |
| `should discard scan result when generation changes mid-scan` | `unit` | Stale onDidChange callback is discarded. | Let `open('s1','wf1')` complete. Stub `EventScanner.scan` with deferred promise. Register spy on `onDidUpdate`. | Trigger `onDidChange`, then call `open('s2','wf2')` before scan resolves, then resolve original scan | `onDidUpdate` is not fired for the stale scan |
| `should coalesce fallback scan with in-flight watcher scan via dirty flag` | `unit` | Fallback timer fires during an in-flight scan and is coalesced. | Use fake timers. Construct with `fallbackScanDelayMs=50`. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Call `sendEvent('go','msg')`. Stub `EventScanner.scan` to return a deferred promise. Trigger `onDidChange` (starts an in-flight scan). | Advance fake timer by 50ms (fallback fires while scan in-flight), then resolve the in-flight scan | Dirty flag is set by fallback; after in-flight scan completes, one re-scan fires; total `EventScanner.scan` calls = 2 (in-flight + re-scan) |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose watcher and emitters on dispose` | `unit` | Dispose releases all resources. | Let `open('s1','wf1')` complete. Stub watcher with `dispose` spy. | Call `instance.dispose()` | Watcher's `dispose()` and both event emitter `dispose()` methods are called |
| `should suppress onDidUpdate after dispose` | `unit` | No state events after disposal. | Let `open('s1','wf1')` complete. Stub `EventScanner.scan` with deferred promise. Trigger `onDidChange`. Register spy on `onDidUpdate`. | Call `instance.dispose()`, then resolve the pending scan | `onDidUpdate` is not fired |
| `should no-op on open after dispose` | `unit` | Open after disposal does nothing. | Construct instance, call `instance.dispose()`. Spy on `EventWatcher` constructor. | Call `instance.open('s1','wf1')` | `EventWatcher` constructor is not called; no error thrown |
| `should return false on sendEvent after dispose` | `unit` | sendEvent after disposal returns false without dispatching. | Construct instance, call `instance.dispose()`. Spy on `EventDispatcher.dispatch`. | Call `instance.sendEvent('submit','msg')` | `EventDispatcher.dispatch` is not called; returned promise resolves to `false` |
| `should set watcher to null after dispose` | `unit` | Watcher reference cleared. | Let `open('s1','wf1')` complete. | Call `instance.dispose()` | Subsequent `open` after un-disposed re-construction would not double-dispose |
| `should cancel pending fallback timer on dispose` | `unit` | Dispose cancels any scheduled fallback. | Use fake timers. Construct with `fallbackScanDelayMs=100`. Let `open('s1','wf1')` complete. Stub `EventDispatcher.dispatch` to resolve. Call `sendEvent('go','msg')` to schedule timer. Register spy on `onDidUpdate`. | Call `instance.dispose()`, then advance fake timer past 100ms | No fallback scan runs; no `onDidUpdate` fires after dispose |

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
