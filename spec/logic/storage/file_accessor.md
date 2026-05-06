# FileAccessor

## Overview

FileAccessor provides a common pattern for accessing files in `.spectra/` storage with automatic preparation fallback. It checks file existence via `os.Stat` and, if the file does not exist, invokes a caller-provided callback to prepare (create) the file. FileAccessor is a stateless exported function — it does not cache results, read file contents, or manage directories.

## Boundaries

- Owns: file existence check (stat), callback invocation on absence, and post-callback verification.
- Delegates: file creation logic (including parent directory creation) to the caller-provided callback.
- Delegates: path composition to StorageLayout (caller provides the composed path).
- Must not: read, write, or open file contents.
- Must not: create directories.
- Must not: cache file handles, paths, or existence state.
- Must not: recover from panics in the callback.

## Dependencies

None. Depends only on Go standard library (`os`, `fmt`, `errors`).

## Behavior

1. Accepts an absolute file path and a preparation callback function.
2. Calls `os.Stat` on the file path to determine existence.
3. If the file exists (stat succeeds), returns the file path immediately without invoking the callback.
4. If stat returns `os.ErrNotExist`, invokes the preparation callback.
5. If the callback returns an error, returns a wrapped error: `"failed to prepare file <path>: <callback error>"`.
6. If the callback returns nil, calls `os.Stat` again to verify the file now exists.
7. If the post-callback stat succeeds, returns the file path.
8. If the post-callback stat returns `os.ErrNotExist`, returns error: `"file preparation succeeded but file was not created: <path>"`.
9. If the post-callback stat returns another error (e.g., permission denied), returns that error wrapped with path context.
10. If the initial stat returns an error other than `os.ErrNotExist` (e.g., permission denied), returns the error immediately without invoking the callback.
11. Each invocation is independent — no state is carried between calls.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| filePath | string | Absolute path to the target file | Yes |
| prepare | `func() error` | Callback responsible for creating the file; must be non-nil | Yes |

## Outputs

### Success Case

| Field | Type | Description |
|-------|------|-------------|
| filePath | string | The same path provided as input, confirmed to exist |
| err | error | nil |

### Error Cases

| Error Format | Condition |
|--------------|-----------|
| `"failed to prepare file <path>: <callback error>"` | Callback returned an error |
| `"file preparation succeeded but file was not created: <path>"` | Callback returned nil but file still does not exist |
| Wrapped stat error with path context | Initial or post-callback stat failed with non-ErrNotExist error |

## Invariants

1. **Callback Invocation Condition**: The callback is invoked if and only if the initial stat returns `os.ErrNotExist`.
2. **Post-Callback Verification**: If the callback returns nil, file existence must be immediately re-verified via stat.
3. **No Content Access**: FileAccessor must never open, read, or write file contents. It only checks existence.
4. **No Directory Creation**: FileAccessor must not create parent directories. This is the callback's responsibility if needed.
5. **Error Wrapping**: All errors include the file path as context for diagnostics.
6. **Absolute Path Requirement**: `filePath` must be an absolute path. FileAccessor does not resolve relative paths.
7. **No State**: FileAccessor is a stateless function. Each call performs fresh filesystem checks.
8. **Nil Callback Panics**: If `prepare` is nil and the file does not exist, the nil function call panics. FileAccessor does not guard against nil callbacks.
9. **No Panic Recovery**: If the callback panics, the panic propagates to the caller.

## Edge Cases

- Condition: File exists at the time of the initial stat.
  Expected: Returns the path immediately without invoking the callback.

- Condition: File does not exist; callback creates the file successfully.
  Expected: Post-callback stat succeeds; returns the file path.

- Condition: File does not exist; callback returns an error.
  Expected: Returns wrapped error `"failed to prepare file <path>: <callback error>"`.

- Condition: File does not exist; callback returns nil but file still does not exist.
  Expected: Returns error `"file preparation succeeded but file was not created: <path>"`.

- Condition: Initial stat fails with permission denied.
  Expected: Returns the permission error immediately without invoking the callback.

- Condition: File does not exist; callback creates it; post-callback stat fails with permission denied.
  Expected: Returns the permission error from the post-callback stat.

- Condition: `filePath` is an empty string.
  Expected: Stat fails (likely with "no such file or directory"); returns stat error.

- Condition: `prepare` is nil and the file does not exist.
  Expected: Panics when attempting to call nil function.

- Condition: Callback panics during execution.
  Expected: Panic propagates to the caller. No recovery.

- Condition: `filePath` points to a directory.
  Expected: Stat succeeds (directories are valid stat targets); returns the path. Caller encounters errors when opening as a file.

- Condition: Two goroutines call FileAccessor for the same non-existent file simultaneously.
  Expected: Both may invoke their callbacks. Race conditions are possible. Callbacks should handle concurrent creation gracefully.

- Condition: Another process deletes the file between callback completion and post-callback stat.
  Expected: Returns error `"file preparation succeeded but file was not created: <path>"`.

## Related

- [StorageLayout](./storage_layout.md) — Provides file paths consumed by FileAccessor callers
- [SessionDirectoryManager](./session_directory_manager.md) — Sibling I/O utility; manages directories while FileAccessor manages file access
