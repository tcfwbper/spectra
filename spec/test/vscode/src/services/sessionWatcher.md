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
| `should store projectRoot` | `unit` | Construction stores the provided project root. | Stub `vscode.workspace.createFileSystemWatcher` to return mock watchers with event stubs for `onDidCreate`, `onDidChange`, `onDidDelete`. Stub `vscode.RelativePattern` constructor. | `projectRoot='/project'` | Instance is created without error |
| `should expose onDidChange event` | `unit` | Construction creates and exposes the public event property. | Stub `vscode.workspace.createFileSystemWatcher` to return mock watchers. Stub `vscode.RelativePattern` constructor. | `projectRoot='/project'` | `instance.onDidChange` is defined and is a function (event accessor) |
| `should create file watcher with correct glob pattern for session.json files` | `unit` | Construction creates a file watcher targeting all session.json files. | Spy on `vscode.workspace.createFileSystemWatcher`. Stub `vscode.RelativePattern` constructor to capture arguments. | `projectRoot='/my/root'` | `createFileSystemWatcher` called with a pattern matching `/my/root/.spectra/sessions/*/session.json` |
| `should create directory watcher with correct glob pattern for session directories` | `unit` | Construction creates a second watcher targeting session directories. | Spy on `vscode.workspace.createFileSystemWatcher`. Stub `vscode.RelativePattern` constructor to capture arguments. | `projectRoot='/my/root'` | `createFileSystemWatcher` called a second time with a pattern matching `/my/root/.spectra/sessions/*` |

### Happy Path — onDidChange

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should fire onDidChange after debounce when session file is created` | `unit` | A session.json creation results in a debounced event firing. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger file watcher's `onDidCreate` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should fire onDidChange after debounce when session file is modified` | `unit` | A session.json modification results in a debounced event firing. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger file watcher's `onDidChange` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should fire onDidChange after debounce when session file is deleted` | `unit` | A session.json deletion results in a debounced event firing. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger file watcher's `onDidDelete` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should fire onDidChange after debounce when session directory is created` | `unit` | A session directory creation via directory watcher results in a debounced event firing. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger directory watcher's `onDidCreate` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should fire onDidChange after debounce when session directory is deleted` | `unit` | A session directory deletion via directory watcher results in a debounced event firing. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger directory watcher's `onDidDelete` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should debounce rapid successive signals from both watchers into single event` | `unit` | Multiple rapid signals from file and directory watchers within the debounce window produce only one event. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger file watcher's `onDidCreate`, then directory watcher's `onDidCreate`, then file watcher's `onDidChange` with 100ms between each | After advancing fake timer to 300ms past the last trigger, the `onDidChange` listener is called exactly once total |
| `should reset debounce timer on each new signal from either watcher` | `unit` | Each new signal from either watcher resets the 300ms quiet period. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger file watcher's `onDidCreate`, advance 200ms, trigger directory watcher's `onDidCreate`, advance 200ms, trigger file watcher's `onDidDelete` | After advancing 300ms from the last trigger, listener called once; at 200ms from last trigger, listener not yet called |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose both watchers and event emitter on dispose` | `unit` | Calling dispose releases all resources. | Stub both watchers with `dispose` spies. Track event emitter `dispose` call. Use fake timers. | Call `instance.dispose()` | Both the file watcher's `dispose`, the directory watcher's `dispose`, and the event emitter's `dispose` are called |
| `should cancel pending debounce timer on dispose` | `unit` | Disposing while a debounce is pending cancels the timer and prevents event firing. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. Trigger a file creation event. | Call `instance.dispose()` before 300ms elapses, then advance timer past 300ms | The `onDidChange` listener is never called |
| `should not fire onDidChange after dispose` | `unit` | No events are emitted after disposal even if a watcher somehow fires. | Stub both watchers. Use fake timers. Register a spy on `instance.onDidChange`. Call `instance.dispose()`. | Trigger file watcher's `onDidCreate` after dispose, advance timer by 300ms | The `onDidChange` listener is never called |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should handle multiple dispose calls without error` | `unit` | Calling dispose multiple times does not throw. | Stub both watchers with `dispose` spies. Use fake timers. | Call `instance.dispose()` three times | No error is thrown on any call |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should subscribe to onDidCreate, onDidChange, and onDidDelete on file watcher` | `unit` | All three file watcher event types are subscribed. | Stub `vscode.workspace.createFileSystemWatcher` to return mock watchers with spies on `onDidCreate`, `onDidChange`, `onDidDelete`. | `projectRoot='/project'` | File watcher's `onDidCreate`, `onDidChange`, and `onDidDelete` subscriptions are registered |
| `should subscribe to onDidCreate and onDidDelete on directory watcher` | `unit` | Directory watcher's create and delete event types are subscribed. | Stub `vscode.workspace.createFileSystemWatcher` to return mock watchers with spies on `onDidCreate` and `onDidDelete`. | `projectRoot='/project'` | Directory watcher's `onDidCreate` and `onDidDelete` subscriptions are registered |
| `should not read or write any files` | `unit` | The watcher never performs filesystem I/O. | Stub both watchers. Spy on `fs.readFile`, `fs.writeFile`, `fs.access`. Use fake timers. Trigger a creation event and advance timer. | Trigger file create, advance 300ms | None of the filesystem spies are called |
