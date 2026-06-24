# WorkflowWatcher

## Overview

Monitors all `*.yaml` files under `<projectRoot>/.spectra/workflows/` for creation and deletion (not modification). Emits a notification event so that upper-layer controllers can re-scan workflows via WorkflowScanner. This is a stateful service with a defined lifecycle (create → watch → dispose). It does not read or parse workflow file content itself.

## Boundaries

- Owns: creating and managing a VS Code `FileSystemWatcher` for `<projectRoot>/.spectra/workflows/*.yaml`, debouncing rapid file-change signals, firing a `vscode.Event<void>` notification, and releasing resources on dispose.
- Delegates: actual workflow listing to `WorkflowScanner`.
- Delegates: project root provision to the caller (fixed at construction).
- Must not: read, write, create, or delete any file.
- Must not: invoke WorkflowScanner or return scanned data.
- Must not: fire notifications on file content modifications (only creation and deletion).
- Must not: remain active after `dispose()` is called.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | File watcher factory | `createFileSystemWatcher` | Must not read/write files |
| `vscode.EventEmitter<void>` | Event infrastructure | `new EventEmitter()`, `fire()`, `event`, `dispose()` | — |
| `vscode.RelativePattern` | Glob pattern builder | Constructor | — |

Construction constraint: WorkflowWatcher is instantiated via `new WorkflowWatcher(projectRoot)`. The parameter is fixed for the instance's lifetime. The instance implements `vscode.Disposable`.

## Behavior

1. On construction, receives `projectRoot` and stores it.
2. On construction, creates a `vscode.EventEmitter<void>` and exposes its `.event` property as the public `onDidChange` event.
3. On construction, creates a `vscode.workspace.FileSystemWatcher` targeting the glob pattern `<projectRoot>/.spectra/workflows/*.yaml`.
4. Subscribes to the file watcher's `onDidCreate` and `onDidDelete` events only. Does NOT subscribe to `onDidChange`.
5. When either of the two file watcher events fires, starts or resets a debounce timer (300ms).
6. When the debounce timer expires without further file-change signals, fires the `EventEmitter` (emits `void` to all subscribers).
7. On `dispose()`, cancels any pending debounce timer, disposes the file watcher, and disposes the event emitter. After disposal, no further events are fired.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| onDidChange | `vscode.Event<void>` | Fires when any `*.yaml` file is created or deleted in the workflows directory (debounced). |

## Invariants

- Must implement `vscode.Disposable`.
- Must never fire `onDidChange` after `dispose()` has been called.
- Must debounce rapid successive file-change signals into a single `onDidChange` firing (300ms quiet period).
- Must watch the glob pattern `<projectRoot>/.spectra/workflows/*.yaml` — top-level only, no recursion.
- Must respond only to create and delete events; modifications to existing `.yaml` files must NOT trigger notification.
- Must not read file content or validate directory existence at construction time.
- The debounce timer resets on each new file-change signal received within the quiet period.

## Edge Cases

- Condition: The `.spectra/workflows/` directory does not exist when the watcher is constructed.
  Expected: No error. The file watcher is still created; it will fire when the directory and files are eventually created.

- Condition: `dispose()` is called while a debounce timer is pending.
  Expected: The pending timer is cancelled; no `onDidChange` event fires.

- Condition: `dispose()` is called multiple times.
  Expected: Subsequent calls are no-ops; no error is thrown.

- Condition: Multiple `.yaml` files are created/deleted in rapid succession (within 300ms intervals).
  Expected: Only one `onDidChange` event fires, 300ms after the last file-change signal.

- Condition: An existing `.yaml` file's content is modified (rewritten in place).
  Expected: No `onDidChange` event fires. The watcher ignores modifications.

- Condition: A non-`.yaml` file is created or deleted in the workflows directory.
  Expected: No `onDidChange` event fires. The glob pattern only matches `*.yaml`.

- Condition: A `.yaml` file is created in a subdirectory of workflows (e.g. `.spectra/workflows/sub/foo.yaml`).
  Expected: No `onDidChange` event fires. The glob pattern matches top-level files only.

## Related

- [WorkflowScanner](./workflowScanner.md) — The controller calls WorkflowScanner to re-read the workflow list after receiving an `onDidChange` notification.
- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
