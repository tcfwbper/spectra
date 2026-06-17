# SessionWatcher

## Overview

Monitors all `session.json` files under `<projectRoot>/.spectra/sessions/*/` for creation, modification, and deletion. Emits a notification event so that upper-layer controllers can re-scan sessions via SessionScanner. This is a stateful service with a defined lifecycle (create → watch → dispose). It does not read or parse session content itself.

## Boundaries

- Owns: creating and managing two VS Code `FileSystemWatcher` instances (one for `session.json` files, one for session directory creation/deletion), debouncing rapid file-change signals from both watchers into a single notification, firing a `vscode.Event<void>` notification, and releasing resources on dispose.
- Delegates: actual session data reading/parsing to `SessionScanner`.
- Delegates: project root provision to the caller (fixed at construction).
- Must not: read, write, create, or delete any file.
- Must not: invoke SessionScanner or return scanned data.
- Must not: remain active after `dispose()` is called.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | File watcher factory | `createFileSystemWatcher` (called twice: once for file pattern, once for directory pattern) | Must not read/write files |
| `vscode.EventEmitter<void>` | Event infrastructure | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.RelativePattern` | Glob pattern builder | Constructor | — |

Construction constraint: SessionWatcher is instantiated via `new SessionWatcher(projectRoot)`. The parameter is fixed for the instance's lifetime. The instance implements `vscode.Disposable`.

## Behavior

1. On construction, receives `projectRoot` and stores it.
2. On construction, creates a `vscode.EventEmitter<void>` and exposes its `.event` property as the public `onDidChange` event.
3. On construction, creates a `vscode.workspace.FileSystemWatcher` (file watcher) targeting the glob pattern `<projectRoot>/.spectra/sessions/*/session.json`.
4. On construction, creates a second `vscode.workspace.FileSystemWatcher` (directory watcher) targeting the glob pattern `<projectRoot>/.spectra/sessions/*`.
5. Subscribes to the file watcher's `onDidCreate`, `onDidChange`, and `onDidDelete` events.
6. Subscribes to the directory watcher's `onDidCreate` and `onDidDelete` events.
7. When any subscribed event from either watcher fires, starts or resets a shared debounce timer (300ms).
8. When the debounce timer expires without further signals, fires the `EventEmitter` (emits `void` to all subscribers).
9. On `dispose()`, cancels any pending debounce timer, disposes both file watchers, and disposes the event emitter. After disposal, no further events are fired.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| onDidChange | `vscode.Event<void>` | Fires when any `session.json` file is created, modified, or deleted (debounced). |

## Invariants

- Must implement `vscode.Disposable`.
- Must never fire `onDidChange` after `dispose()` has been called.
- Must debounce rapid successive signals from both watchers into a single `onDidChange` firing (300ms quiet period).
- Must watch the glob pattern `<projectRoot>/.spectra/sessions/*/session.json` for file-level mutations.
- Must watch the glob pattern `<projectRoot>/.spectra/sessions/*` for directory-level creation and deletion.
- Must respond to create, change, and delete events from the file watcher.
- Must respond to create and delete events from the directory watcher.
- Must not read file content or validate directory existence at construction time.
- The debounce timer resets on each new signal received from either watcher within the quiet period.

## Edge Cases

- Condition: The `.spectra/sessions/` directory does not exist when the watcher is constructed.
  Expected: No error. The file watcher is still created; it will fire when session directories and files are eventually created.

- Condition: `dispose()` is called while a debounce timer is pending.
  Expected: The pending timer is cancelled; no `onDidChange` event fires.

- Condition: `dispose()` is called multiple times.
  Expected: Subsequent calls are no-ops; no error is thrown.

- Condition: Multiple `session.json` files are created/modified/deleted in rapid succession (within 300ms intervals).
  Expected: Only one `onDidChange` event fires, 300ms after the last signal.

- Condition: A new session directory is created but no `session.json` is written yet.
  Expected: The directory watcher fires `onDidCreate`, triggering a debounced notification. The subsequent scan will find no valid `session.json` and exclude this session from results.

- Condition: A session directory is deleted entirely via `rm -rf` (including its `session.json`).
  Expected: The directory watcher fires `onDidDelete` for the directory, triggering the debounced notification. This is reliable even if the file watcher does not fire `onDidDelete` for the nested `session.json` (platform-dependent behavior).

## Related

- [SessionScanner](./sessionScanner.md) — The controller calls SessionScanner to re-read session data after receiving an `onDidChange` notification.
- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
