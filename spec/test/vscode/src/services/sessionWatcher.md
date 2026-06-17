# Test Specification: `sessionWatcher.test.ts`

## Source File Under Test
`vscode/src/services/sessionWatcher.ts`

## Test File
`vscode/test/suite/sessionWatcher.test.ts`

---

## `SessionWatcher`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should store projectRoot` | `unit` | Construction stores the provided project root. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher with event stubs for `onDidCreate`, `onDidChange`, `onDidDelete`. Stub `vscode.RelativePattern` constructor. | `projectRoot='/project'` | Instance is created without error |
| `should expose onDidChange event` | `unit` | Construction creates and exposes the public event property. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher. Stub `vscode.RelativePattern` constructor. | `projectRoot='/project'` | `instance.onDidChange` is defined and is a function (event accessor) |
| `should create file system watcher with correct glob pattern` | `unit` | Construction creates a watcher targeting all session.json files. | Spy on `vscode.workspace.createFileSystemWatcher`. Stub `vscode.RelativePattern` constructor to capture arguments. | `projectRoot='/my/root'` | `createFileSystemWatcher` called with a pattern matching `/my/root/.spectra/sessions/*/session.json` |

### Happy Path — onDidChange

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should fire onDidChange after debounce when session file is created` | `unit` | A session.json creation results in a debounced event firing. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger mock watcher's `onDidCreate` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should fire onDidChange after debounce when session file is modified` | `unit` | A session.json modification results in a debounced event firing. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger mock watcher's `onDidChange` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should fire onDidChange after debounce when session file is deleted` | `unit` | A session.json deletion results in a debounced event firing. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger mock watcher's `onDidDelete` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should debounce rapid successive signals from mixed event types into single event` | `unit` | Multiple rapid create/change/delete signals within the debounce window produce only one event. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger `onDidCreate`, then `onDidChange`, then `onDidDelete` with 100ms between each | After advancing fake timer to 300ms past the last trigger, the `onDidChange` listener is called exactly once total |
| `should reset debounce timer on each new signal` | `unit` | Each new file-change signal resets the 300ms quiet period. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger `onDidCreate`, advance 200ms, trigger `onDidChange`, advance 200ms, trigger `onDidDelete` | After advancing 300ms from the last trigger, listener called once; at 200ms from last trigger, listener not yet called |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose file watcher and event emitter on dispose` | `unit` | Calling dispose releases all resources. | Stub watcher with `dispose` spy. Track event emitter `dispose` call. Use fake timers. | Call `instance.dispose()` | Both the file watcher's `dispose` and the event emitter's `dispose` are called |
| `should cancel pending debounce timer on dispose` | `unit` | Disposing while a debounce is pending cancels the timer and prevents event firing. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. Trigger a file creation event. | Call `instance.dispose()` before 300ms elapses, then advance timer past 300ms | The `onDidChange` listener is never called |
| `should not fire onDidChange after dispose` | `unit` | No events are emitted after disposal even if the watcher somehow fires. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. Call `instance.dispose()`. | Trigger mock watcher's `onDidCreate` after dispose, advance timer by 300ms | The `onDidChange` listener is never called |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should handle multiple dispose calls without error` | `unit` | Calling dispose multiple times does not throw. | Stub watcher with `dispose` spy. Use fake timers. | Call `instance.dispose()` three times | No error is thrown on any call |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should subscribe to onDidCreate, onDidChange, and onDidDelete` | `unit` | All three file watcher event types are subscribed. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher with spies on `onDidCreate`, `onDidChange`, `onDidDelete`. | `projectRoot='/project'` | All three event subscriptions are registered |
| `should not read or write any files` | `unit` | The watcher never performs filesystem I/O. | Stub watcher. Spy on `fs.readFile`, `fs.writeFile`, `fs.access`. Use fake timers. Trigger a creation event and advance timer. | Trigger file create, advance 300ms | None of the filesystem spies are called |
