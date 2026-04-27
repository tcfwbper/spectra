# SpectraFinder

## Overview

SpectraFinder is responsible for locating the `.spectra` directory by searching upward from the current working directory. It traverses the directory hierarchy toward the filesystem root until it finds a `.spectra` directory or encounters a boundary condition (reached root or permission denied). Upon success, it returns the absolute path to the project root directory containing `.spectra`. If `.spectra` is not found, it reports an initialization error.

## Behavior

1. The finder starts searching from the current working directory.
2. It checks if `.spectra` exists in the current directory.
3. If `.spectra` exists and is a directory, the finder returns the absolute path of the current directory.
4. If `.spectra` does not exist, or exists but is not a directory, the finder attempts to access the parent directory (`..`).
5. If accessing the parent directory fails due to insufficient read permissions, the finder reports an error.
6. If the parent directory is accessible, the finder repeats steps 2-5 in the parent directory.
7. The search terminates when:
   - A `.spectra` directory is found (success), OR
   - The filesystem root is reached without finding `.spectra` (failure), OR
   - Attempting to access the parent directory is denied due to permissions (failure).
8. Symbolic links are followed during upward traversal.
9. The finder validates that `.spectra` is a directory (not a file). It does NOT validate the internal structure (e.g., presence of `sessions/`, `workflows/`, `agents/`).

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| StartDir | string | Absolute path to a valid directory | No (defaults to current working directory) |

## Outputs

### Success Case

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| ProjectRoot | string | Absolute path | The absolute path to the directory containing `.spectra` |

### Error Case

| Field | Type | Description |
|-------|------|-------------|
| ErrorMessage | string | Human-readable error message |

**Error message formats**:
- `"spectra not initialized"` — `.spectra` was not found after searching up to the filesystem root, or a permission error was encountered during traversal, or a symlink loop was detected.
- `"invalid start directory: <path>"` — `StartDir` does not exist or is not a directory.

## Invariants

1. **Upward-Only Traversal**: The finder must only traverse upward (toward parent directories), never downward or sideways.

2. **Absolute Path Return**: `ProjectRoot` must always be an absolute path (starting with `/` on Unix or a drive letter on Windows).

3. **Directory Validation**: If a `.spectra` path is found, it must be a directory. If it is a file, the finder must continue searching upward.

4. **Filesystem Root Boundary**: The search must terminate at or before reaching the filesystem root (`/` on Unix, drive root like `C:\` on Windows). If `.spectra` is not found by that point, the finder must return an error.

5. **Symlink Resolution**: The finder must resolve symbolic links during traversal to determine the real path of each directory.

6. **Permission Error Handling**: If the finder encounters a permission error when attempting to access a parent directory, it must immediately terminate and return an error.

7. **No Caching**: The finder must perform a fresh search on every invocation. It must not cache the result.

## Edge Cases

- **Condition**: `.spectra` exists in the current directory and is a directory.
  **Expected**: The finder immediately returns the absolute path of the current directory.

- **Condition**: `.spectra` is a file, not a directory.
  **Expected**: The finder treats this as "not found" and continues searching upward.

- **Condition**: Multiple `.spectra` directories exist in the path hierarchy (e.g., `/home/user/project/.spectra` and `/home/user/.spectra`).
  **Expected**: The finder returns the first (nearest) `.spectra` directory found, starting from the current directory.

- **Condition**: `StartDir` is a non-root directory, and the finder traverses upward through multiple levels without finding `.spectra`, eventually reaching the filesystem root.
  **Expected**: The finder returns an error: `"spectra not initialized"`.

- **Condition**: The finder encounters a symbolic link pointing to a parent directory.
  **Expected**: The finder follows the symbolic link and continues searching upward from the resolved path.

- **Condition**: The finder encounters a symbolic link loop (e.g., `a -> b -> a`).
  **Expected**: The finder detects the loop and returns an error: `"spectra not initialized"`.

- **Condition**: The finder encounters insufficient read permissions when accessing `..`.
  **Expected**: The finder immediately returns an error: `"spectra not initialized"`.

- **Condition**: `StartDir` is the filesystem root (`/` on Unix or `C:\` on Windows) itself, and no `.spectra` exists there.
  **Expected**: The finder returns an error: `"spectra not initialized"`.

- **Condition**: `StartDir` is provided as a relative path.
  **Expected**: The finder resolves it to an absolute path before starting the search.

- **Condition**: `StartDir` does not exist or is not a directory.
  **Expected**: The finder returns an error: `"invalid start directory: <path>"`.

## Related

- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
- [StorageLayout](./storage_layout.md) - Defines the structure of `.spectra/` directory
