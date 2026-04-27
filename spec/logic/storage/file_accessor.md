# FileAccessor

## Overview

FileAccessor provides a common pattern for accessing files in the `.spectra/` storage with automatic preparation fallback. When a DAO or DTO attempts to access a file, FileAccessor checks for file existence and invokes a caller-provided callback to prepare the file if it does not exist. FileAccessor distinguishes between "file not found but can be created" and "file not found and should not exist" scenarios based on callback behavior. It does not cache file handles or contents.

## Behavior

1. FileAccessor accepts a file path and a preparation callback function.
2. FileAccessor attempts to stat the file to determine if it exists.
3. If the file exists, FileAccessor returns the file path immediately without invoking the callback.
4. If the file does not exist (os.IsNotExist), FileAccessor invokes the preparation callback.
5. The preparation callback is responsible for creating the file, creating parent directories if needed, or returning an error if the file should not be created.
6. If the callback succeeds (returns nil error), FileAccessor verifies that the file now exists by calling stat again.
7. If the file exists after callback execution, FileAccessor returns the file path.
8. If the file still does not exist after callback execution, FileAccessor returns an error: "file preparation succeeded but file was not created: <path>".
9. If the callback returns an error, FileAccessor wraps the error with context: "failed to prepare file <path>: <callback error>".
10. If stat fails with an error other than "not exist" (e.g., permission denied), FileAccessor returns the stat error immediately without invoking the callback.
11. FileAccessor is stateless and does not cache results. Each invocation performs fresh filesystem checks.
12. FileAccessor does not read or write file contents. It only verifies existence and delegates preparation.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| FilePath | string | Absolute path to the target file | Yes |
| PrepareCallback | function | Signature: `func() error`. Must be safe to call multiple times if caller retries. | Yes |

## Outputs

### Success Case

| Field | Type | Description |
|-------|------|-------------|
| FilePath | string | The same path provided as input, confirmed to exist |

### Error Cases

| Error Message Format | Description |
|---------------------|-------------|
| `"file preparation succeeded but file was not created: <path>"` | Callback returned nil, but the file does not exist after callback execution |
| `"failed to prepare file <path>: <callback error>"` | Callback returned an error |
| `"permission denied: <path>"` | Stat operation failed due to insufficient permissions (before or after callback) |
| `"invalid file path: <path>"` | Stat operation failed for other reasons (e.g., path is a directory, not a file) |

## Invariants

1. **Callback Invocation Condition**: The callback must be invoked if and only if the initial stat call returns `os.IsNotExist`.

2. **Post-Callback Verification**: If the callback returns nil, FileAccessor must immediately verify file existence by calling stat again.

3. **No Content Access**: FileAccessor must never open, read, or write the file. It only checks existence via stat.

4. **Error Wrapping**: All errors returned by the callback must be wrapped with context indicating the file path.

5. **Idempotent Callback**: The callback may be invoked multiple times (e.g., if caller retries) and must be idempotent or handle existing files gracefully.

6. **No Directory Creation**: FileAccessor itself must not create parent directories. Directory creation is the responsibility of the callback if needed.

7. **Absolute Path Requirement**: `FilePath` must be an absolute path. FileAccessor does not resolve relative paths.

8. **Thread Safety**: FileAccessor must be safe to call concurrently from multiple goroutines, but race conditions between stat and callback execution are not prevented (e.g., file created by another process between stat checks).

## Edge Cases

- **Condition**: File exists at the time of the first stat call.
  **Expected**: FileAccessor returns the path immediately without invoking the callback.

- **Condition**: File does not exist, callback creates the file successfully.
  **Expected**: FileAccessor verifies the file exists and returns the path.

- **Condition**: File does not exist, callback returns an error indicating the file should not be created (e.g., "parent directory does not exist").
  **Expected**: FileAccessor returns the wrapped error: `"failed to prepare file <path>: parent directory does not exist"`.

- **Condition**: File does not exist, callback returns nil, but the file still does not exist after callback execution (callback bug).
  **Expected**: FileAccessor returns an error: `"file preparation succeeded but file was not created: <path>"`.

- **Condition**: File path points to a directory, not a file.
  **Expected**: Stat succeeds (directories are valid stat targets), FileAccessor returns the path. Caller will encounter errors when attempting to open the path as a file.

- **Condition**: File path does not exist, callback creates the file, but another process deletes the file before the post-callback stat.
  **Expected**: FileAccessor returns an error: `"file preparation succeeded but file was not created: <path>"`.

- **Condition**: Stat fails with "permission denied" before callback invocation.
  **Expected**: FileAccessor returns the permission error immediately without invoking the callback.

- **Condition**: File does not exist, callback creates the file, but post-callback stat fails with "permission denied" (permissions changed during callback).
  **Expected**: FileAccessor returns the permission error from the post-callback stat.

- **Condition**: `FilePath` is an empty string.
  **Expected**: Stat fails with an error (likely "no such file or directory"), FileAccessor returns the stat error.

- **Condition**: `PrepareCallback` is nil.
  **Expected**: If the file does not exist and FileAccessor attempts to invoke a nil callback, it panics. Caller must provide a valid callback.

- **Condition**: Callback panics during execution.
  **Expected**: The panic propagates to the caller. FileAccessor does not recover from panics.

- **Condition**: Two goroutines call FileAccessor for the same non-existent file simultaneously.
  **Expected**: Both goroutines may invoke their respective callbacks. Race conditions are possible. Callbacks should handle concurrent execution gracefully (e.g., using file-level locking or checking existence before creation).

## Related

- [StorageLayout](./storage_layout.md) - Provides file paths to FileAccessor
- [EventStore](./event_store.md) - Uses FileAccessor to access `events.jsonl`
