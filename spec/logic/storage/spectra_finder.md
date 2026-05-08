# SpectraFinder

## Overview

SpectraFinder locates the `.spectra` directory by searching upward from a starting directory toward the filesystem root. Upon finding a `.spectra` directory, it returns the absolute path to the project root (the parent directory containing `.spectra`). If `.spectra` is not found, it returns a sentinel error. SpectraFinder does not validate the internal structure of `.spectra/`.

## Boundaries

- Owns: upward directory traversal logic, `.spectra` directory detection, and CWD resolution when no StartDir is provided.
- Owns: symlink resolution during traversal via `filepath.EvalSymlinks`.
- Delegates: path composition within `.spectra/` to StorageLayout.
- Delegates: `.spectra/` internal structure validation to consuming modules (e.g., init checks).
- Must not: create, modify, or delete any files or directories.
- Must not: validate the contents or structure of `.spectra/` beyond confirming it is a directory.
- Must not: hold any state or cache results between invocations.

## Dependencies

None. Depends only on Go standard library (`os`, `path/filepath`).

## Behavior

1. Accepts an optional `startDir` parameter. If empty, internally calls `os.Getwd()` to obtain the current working directory.
2. If `startDir` is provided as a relative path, resolves it to an absolute path via `filepath.Abs`.
3. Validates that the resolved start directory exists and is a directory. If not, returns error `"invalid start directory: <path>"`.
4. Beginning from the resolved start directory, checks if a `.spectra` entry exists in the current directory.
5. If `.spectra` exists and is a directory (determined via `os.Stat`), returns the absolute path of the current directory as the project root.
6. If `.spectra` exists but is not a directory (e.g., a file), treats it as "not found" and continues upward.
7. If `.spectra` does not exist, moves to the parent directory and repeats the check.
8. If the parent directory is the same as the current directory (filesystem root reached), returns `ErrNotInitialized`.
9. If accessing a parent directory fails due to permission error, returns `ErrNotInitialized`.
10. Resolves symbolic links during traversal using `filepath.EvalSymlinks` to detect loops.
11. If a symlink loop is detected (resolved path equals a previously visited path), returns `ErrNotInitialized`.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| startDir | string | Absolute or relative path to a valid directory; empty string means use CWD | No (defaults to CWD via `os.Getwd()`) |

## Outputs

### Success Case

| Field | Type | Description |
|-------|------|-------------|
| projectRoot | string | Absolute path to the directory containing `.spectra` |

### Error Cases

| Error | Description |
|-------|-------------|
| `ErrNotInitialized` | Sentinel error: `.spectra` not found after exhausting traversal (root reached, permission denied, or symlink loop) |
| `"invalid start directory: <path>"` | `startDir` does not exist or is not a directory |
| `"failed to get working directory: <error>"` | `os.Getwd()` failed when `startDir` is empty |

## Invariants

1. **Upward-Only Traversal**: The finder must only traverse toward parent directories, never downward or sideways.
2. **Absolute Path Return**: `projectRoot` is always an absolute path on success.
3. **Directory Validation**: A `.spectra` entry is only accepted if it is a directory. Files named `.spectra` are ignored.
4. **Filesystem Root Boundary**: Traversal terminates at the filesystem root without error propagation — returns `ErrNotInitialized`.
5. **Permission Error Handling**: Permission errors during parent access terminate traversal immediately with `ErrNotInitialized`.
6. **Symlink Resolution**: Symbolic links are resolved to detect loops and determine real paths.
7. **No Caching**: Each invocation performs a fresh filesystem traversal.
8. **No State**: SpectraFinder is a stateless exported function, not a struct.
9. **No Mutation**: Must not create, modify, or delete any filesystem entries.

## Edge Cases

- Condition: `.spectra` exists in the start directory and is a directory.
  Expected: Returns the start directory immediately as project root.

- Condition: `.spectra` exists as a regular file (not a directory) in the start directory.
  Expected: Treats as "not found" and continues searching upward.

- Condition: Multiple `.spectra` directories exist in ancestor chain (e.g., `/home/user/.spectra` and `/home/user/project/.spectra`).
  Expected: Returns the nearest (deepest) directory containing `.spectra`.

- Condition: Start directory is the filesystem root and `.spectra` does not exist there.
  Expected: Returns `ErrNotInitialized`.

- Condition: Symbolic link loop is encountered during traversal.
  Expected: Returns `ErrNotInitialized`.

- Condition: Permission denied when accessing a parent directory.
  Expected: Returns `ErrNotInitialized`.

- Condition: `startDir` is a non-existent path.
  Expected: Returns `"invalid start directory: <path>"`.

- Condition: `startDir` is a file (not a directory).
  Expected: Returns `"invalid start directory: <path>"`.

- Condition: `startDir` is empty and `os.Getwd()` fails.
  Expected: Returns `"failed to get working directory: <error>"`.

- Condition: `startDir` is a relative path (e.g., `../project`).
  Expected: Resolves to absolute path via `filepath.Abs` before starting traversal.

## Related

- [StorageLayout](./storage_layout.md) — Defines paths within `.spectra/`; consumes projectRoot from SpectraFinder
- [SessionDirectoryManager](./session_directory_manager.md) — Requires projectRoot from SpectraFinder
