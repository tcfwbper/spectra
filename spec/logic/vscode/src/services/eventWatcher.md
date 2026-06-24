# EventWatcher

## Overview

Monitors a single session's `events.jsonl` file for modifications and emits a notification event so that upper-layer controllers can re-scan the file via EventScanner. This is a stateful service with a defined lifecycle (create → watch → dispose). It does not read or parse event content itself.

## Boundaries

- Owns: creating and managing a VS Code `FileSystemWatcher` for `<projectRoot>/.spectra/sessions/<sessionId>/events.jsonl`, debouncing rapid file-change signals, firing a `vscode.Event<void>` notification, and releasing resources on dispose.
- Delegates: actual event data reading/parsing to `EventScanner`.
- Delegates: project root and session ID provision to the caller (fixed at construction).
- Must not: read, write, create, or delete any file.
- Must not: invoke EventScanner or return scanned data.
- Must not: remain active after `dispose()` is called.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | File watcher factory | `createFileSystemWatcher` | Must not read/write files |
| `vscode.EventEmitter<void>` | Event infrastructure | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.RelativePattern` | Glob pattern builder | Constructor | — |

Construction constraint: EventWatcher is instantiated via `new EventWatcher(projectRoot, sessionId)`. The parameters are fixed for the instance's lifetime. The instance implements `vscode.Disposable`.

## Behavior

1. On construction, receives `projectRoot` and `sessionId` and stores them.
2. On construction, creates a `vscode.EventEmitter<void>` and exposes its `.event` property as the public `onDidChange` event.
3. On construction, creates a `vscode.workspace.FileSystemWatcher` targeting the glob pattern `<projectRoot>/.spectra/sessions/<sessionId>/events.jsonl`.
4. Subscribes to the file watcher's `onDidChange` event (file modified).
5. When the file watcher fires, starts or resets a debounce timer (300ms).
6. When the debounce timer expires without further file-change signals, fires the `EventEmitter` (emits `void` to all subscribers).
7. On `dispose()`, cancels any pending debounce timer, disposes the file watcher, and disposes the event emitter. After disposal, no further events are fired.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |
| sessionId | string | Non-empty, identifies a session directory | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| onDidChange | `vscode.Event<void>` | Fires when the watched `events.jsonl` file is modified (debounced). |

## Invariants

- Must implement `vscode.Disposable`.
- Must never fire `onDidChange` after `dispose()` has been called.
- Must debounce rapid successive file-change signals into a single `onDidChange` firing (300ms quiet period).
- Must watch exactly one file: `<projectRoot>/.spectra/sessions/<sessionId>/events.jsonl`.
- Must not read file content or validate file existence at construction time.
- The debounce timer resets on each new file-change signal received within the quiet period.

## Edge Cases

- Condition: The `events.jsonl` file does not exist when the watcher is constructed.
  Expected: No error. The file watcher is still created; it will fire when the file is eventually created and modified.

- Condition: `dispose()` is called while a debounce timer is pending.
  Expected: The pending timer is cancelled; no `onDidChange` event fires.

- Condition: `dispose()` is called multiple times.
  Expected: Subsequent calls are no-ops; no error is thrown.

- Condition: The file is modified many times in rapid succession (within 300ms intervals).
  Expected: Only one `onDidChange` event fires, 300ms after the last modification signal.

- Condition: The file is deleted.
  Expected: No `onDidChange` event fires (deletion is not a modification). The watcher remains active for future modifications if the file is recreated.

## Related

- [EventScanner](./eventScanner.md) — The controller calls EventScanner to re-read event data after receiving an `onDidChange` notification.
- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
