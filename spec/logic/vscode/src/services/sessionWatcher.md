# SessionWatcher

## Overview

Monitors all `session.json` files under `<projectRoot>/.spectra/sessions/*/` for creation, modification, and deletion. Emits a notification event so that upper-layer controllers can re-scan sessions via SessionScanner. This is a stateful service with a defined lifecycle (create → watch → dispose). It does not read or parse session content itself.

## Boundaries

- Owns: creating and managing a VS Code `FileSystemWatcher` for `<projectRoot>/.spectra/sessions/*/session.json`, debouncing rapid file-change signals, firing a `vscode.Event<void>` notification, and releasing resources on dispose.
- Delegates: actual session data reading/parsing to `SessionScanner`.
- Delegates: project root provision to the caller (fixed at construction).
- Must not: read, write, create, or delete any file.
- Must not: invoke SessionScanner or return scanned data.
- Must not: remain active after `dispose()` is called.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | File watcher factory | `createFileSystemWatcher` | Must not read/write files |
| `vscode.EventEmitter<void>` | Event infrastructure | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.RelativePattern` | Glob pattern builder | Constructor | — |

Construction constraint: SessionWatcher is instantiated via `new SessionWatcher(projectRoot)`. The parameter is fixed for the instance's lifetime. The instance implements `vscode.Disposable`.

## Behavior

1. On construction, receives `projectRoot` and stores it.
2. On construction, creates a `vscode.EventEmitter<void>` and exposes its `.event` property as the public `onDidChange` event.
3. On construction, creates a `vscode.workspace.FileSystemWatcher` targeting the glob pattern `<projectRoot>/.spectra/sessions/*/session.json`.
4. Subscribes to the file watcher's `onDidCreate`, `onDidChange`, and `onDidDelete` events.
5. When any of the three file watcher events fire, starts or resets a debounce timer (300ms).
6. When the debounce timer expires without further file-change signals, fires the `EventEmitter` (emits `void` to all subscribers).
7. On `dispose()`, cancels any pending debounce timer, disposes the file watcher, and disposes the event emitter. After disposal, no further events are fired.

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
- Must debounce rapid successive file-change signals into a single `onDidChange` firing (300ms quiet period).
- Must watch the glob pattern `<projectRoot>/.spectra/sessions/*/session.json` — covering all session directories.
- Must respond to all three event types: create, change, and delete.
- Must not read file content or validate directory existence at construction time.
- The debounce timer resets on each new file-change signal received within the quiet period.

## Edge Cases

- Condition: The `.spectra/sessions/` directory does not exist when the watcher is constructed.
  Expected: No error. The file watcher is still created; it will fire when session directories and files are eventually created.

- Condition: `dispose()` is called while a debounce timer is pending.
  Expected: The pending timer is cancelled; no `onDidChange` event fires.

- Condition: `dispose()` is called multiple times.
  Expected: Subsequent calls are no-ops; no error is thrown.

- Condition: Multiple `session.json` files are created/modified/deleted in rapid succession (within 300ms intervals).
  Expected: Only one `onDidChange` event fires, 300ms after the last file-change signal.

- Condition: A new session directory is created but no `session.json` is written yet.
  Expected: No `onDidChange` event fires until a `session.json` file appears in that directory.

- Condition: A session directory is deleted entirely (including its `session.json`).
  Expected: The file watcher fires `onDidDelete` for the `session.json`, triggering the debounced notification.

## Related

- [SessionScanner](./sessionScanner.md) — The controller calls SessionScanner to re-read session data after receiving an `onDidChange` notification.
- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
