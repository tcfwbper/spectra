# Test Specification: `eventWatcher.test.ts`

## Source File Under Test
`vscode/src/services/eventWatcher.ts`

## Test File
`vscode/test/suite/eventWatcher.test.ts`

---

## `EventWatcher`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should store projectRoot and sessionId` | `unit` | Construction stores the provided parameters for later use. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher with `onDidChange` event stub. Stub `vscode.RelativePattern` constructor. | `projectRoot='/project'`, `sessionId='sess-1'` | Instance is created without error |
| `should expose onDidChange event` | `unit` | Construction creates and exposes the public event property. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher. Stub `vscode.RelativePattern` constructor. | `projectRoot='/project'`, `sessionId='sess-1'` | `instance.onDidChange` is defined and is a function (event accessor) |
| `should create file system watcher with correct glob pattern` | `unit` | Construction creates a watcher targeting the session's events.jsonl file. | Spy on `vscode.workspace.createFileSystemWatcher`. Stub `vscode.RelativePattern` constructor to capture arguments. | `projectRoot='/my/root'`, `sessionId='abc-123'` | `createFileSystemWatcher` called with a pattern matching `/my/root/.spectra/sessions/abc-123/events.jsonl` |

### Happy Path — onDidChange

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should fire onDidChange after debounce when file is modified` | `unit` | A single file modification results in a debounced event firing. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher. Use fake timers (sinon). Register a spy on `instance.onDidChange`. | Trigger mock watcher's `onDidChange` once | After advancing fake timer by 300ms, the `onDidChange` listener is called exactly once |
| `should debounce rapid successive modifications into single event` | `unit` | Multiple rapid modifications within the debounce window produce only one event. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger mock watcher's `onDidChange` three times with 100ms between each | After advancing fake timer to 300ms past the last trigger, the `onDidChange` listener is called exactly once total |
| `should reset debounce timer on each new modification` | `unit` | Each new modification resets the 300ms quiet period. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher. Use fake timers. Register a spy on `instance.onDidChange`. | Trigger modification, advance 200ms, trigger again, advance 200ms, trigger again | After advancing 300ms from the last trigger, listener called once; at 200ms from last trigger, listener not yet called |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should dispose file watcher and event emitter on dispose` | `unit` | Calling dispose releases all resources. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher with `dispose` spy. Track event emitter `dispose` call. Use fake timers. | Call `instance.dispose()` | Both the file watcher's `dispose` and the event emitter's `dispose` are called |
| `should cancel pending debounce timer on dispose` | `unit` | Disposing while a debounce is pending cancels the timer and prevents event firing. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. Trigger a file modification. | Call `instance.dispose()` before 300ms elapses, then advance timer past 300ms | The `onDidChange` listener is never called |
| `should not fire onDidChange after dispose` | `unit` | No events are emitted after disposal even if the watcher somehow fires. | Stub watcher. Use fake timers. Register a spy on `instance.onDidChange`. Call `instance.dispose()`. | Trigger mock watcher's `onDidChange` after dispose, advance timer by 300ms | The `onDidChange` listener is never called |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should handle multiple dispose calls without error` | `unit` | Calling dispose multiple times does not throw. | Stub watcher with `dispose` spy. Use fake timers. | Call `instance.dispose()` three times | No error is thrown on any call |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should subscribe only to onDidChange of file watcher` | `unit` | Only file modification events are subscribed, not create or delete. | Stub `vscode.workspace.createFileSystemWatcher` to return a mock watcher with spies on `onDidChange`, `onDidCreate`, `onDidDelete`. | `projectRoot='/project'`, `sessionId='s1'` | `onDidChange` subscription is registered; `onDidCreate` and `onDidDelete` are not subscribed |
| `should not read or write any files` | `unit` | The watcher never performs filesystem I/O. | Stub watcher. Spy on `fs.readFile`, `fs.writeFile`, `fs.access`. Use fake timers. Trigger modification and advance timer. | Trigger file change, advance 300ms | None of the filesystem spies are called |
