# SessionDirectoryManager

## Overview

SessionDirectoryManager provides methods to ensure the sessions parent directory exists and to create individual session directories within `.spectra/sessions/`. It uses StorageLayout for all path composition. SessionDirectoryManager does not delete directories, manage directory contents, or validate UUID formats.

## Boundaries

- Owns: ensuring `.spectra/sessions/` directory exists (idempotent creation).
- Owns: creating session-specific directories with correct permissions.
- Owns: UUID collision detection (session directory already exists).
- Delegates: path composition to StorageLayout.
- Delegates: `.spectra/` directory creation to the init flow (SpectraFinder guarantees it exists).
- Delegates: UUID format validation to upstream callers (Session entity constructor).
- Must not: create `.spectra/` directory (owned by init).
- Must not: delete directories or manage files within session directories.
- Must not: validate UUID format or content.
- Must not: hold state or cache filesystem checks between calls.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| StorageLayout | Path composition | `GetSpectraDir`, `GetSessionsDir`, `GetSessionDir` | Must not bypass StorageLayout for path construction |

Construction constraint: SessionDirectoryManager is a set of exported functions (no struct). Requires `projectRoot` as parameter on each call.

## Behavior

### EnsureSessionsDirectory

1. Accepts `projectRoot` (absolute path).
2. Composes the `.spectra/` path via `StorageLayout.GetSpectraDir(projectRoot)`.
3. Calls `os.Stat` on the `.spectra/` path. If it does not exist or is not a directory, returns `ErrNotInitialized` (same sentinel as SpectraFinder).
4. Composes the `.spectra/sessions/` path via `StorageLayout.GetSessionsDir(projectRoot)`.
5. Calls `os.Stat` on `.spectra/sessions/`. If it already exists and is a directory, returns nil (idempotent success).
6. If `.spectra/sessions/` does not exist, creates it using `os.Mkdir` with permissions `0755`.
7. If creation fails, returns wrapped error: `"failed to create sessions directory: <error>"`.
8. If stat returns an error other than `os.ErrNotExist`, returns wrapped error.

### CreateSessionDirectory

1. Accepts `projectRoot` (absolute path) and `sessionUUID` (string).
2. Calls `EnsureSessionsDirectory(projectRoot)` internally to guarantee the parent directory exists.
3. If `EnsureSessionsDirectory` returns an error, propagates it immediately.
4. Composes the session directory path via `StorageLayout.GetSessionDir(projectRoot, sessionUUID)`.
5. Calls `os.Stat` on the session directory. If it already exists, returns `ErrSessionDirExists` (sentinel error wrapping the path).
6. Creates the session directory using `os.Mkdir` with permissions `0755`.
7. If creation fails, returns wrapped error: `"failed to create session directory: <error>"`.
8. On success, returns nil.

## Inputs

### EnsureSessionsDirectory

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### CreateSessionDirectory

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| sessionUUID | string | Expected to be valid UUID v4 (not validated here) | Yes |

## Outputs

### EnsureSessionsDirectory

| Output | Type | Description |
|--------|------|-------------|
| error | error | nil on success |

**Error Cases**:

| Error | Description |
|-------|-------------|
| `ErrNotInitialized` | `.spectra/` does not exist or is not a directory |
| `"failed to create sessions directory: <error>"` | `os.Mkdir` failed (permission denied, disk full, etc.) |

### CreateSessionDirectory

| Output | Type | Description |
|--------|------|-------------|
| error | error | nil on success |

**Error Cases**:

| Error | Description |
|-------|-------------|
| `ErrNotInitialized` | Propagated from `EnsureSessionsDirectory` |
| `ErrSessionDirExists` | Session directory already exists (UUID collision or stale session) |
| `"failed to create session directory: <error>"` | `os.Mkdir` failed |

## Invariants

1. **No .spectra/ Creation**: Must not create `.spectra/` directory. Returns `ErrNotInitialized` if it is absent.
2. **Idempotent EnsureSessionsDirectory**: Calling `EnsureSessionsDirectory` multiple times with the same input is safe. If `.spectra/sessions/` already exists, returns nil.
3. **Non-Idempotent CreateSessionDirectory**: If the session directory already exists, returns `ErrSessionDirExists` (preserves UUID collision detection).
4. **CreateSessionDirectory Calls EnsureSessionsDirectory**: `CreateSessionDirectory` always ensures the parent directory exists before attempting session directory creation.
5. **Permission 0755**: Newly created directories (both `sessions/` and `sessions/<UUID>/`) use permissions `0755`.
6. **Path Composition Delegation**: All paths are composed via StorageLayout. Must not construct paths manually.
7. **No UUID Validation**: Does not validate UUID format. Trusts upstream callers.
8. **No State**: Stateless functions. Each call performs fresh filesystem checks.
9. **Atomic Directory Creation**: Uses `os.Mkdir` (not `os.MkdirAll`) for individual directory creation, which is atomic at the filesystem level.

## Edge Cases

- Condition: `.spectra/` does not exist.
  Expected: `EnsureSessionsDirectory` returns `ErrNotInitialized`. `CreateSessionDirectory` propagates the same error.

- Condition: `.spectra/sessions/` already exists.
  Expected: `EnsureSessionsDirectory` returns nil (idempotent). No error.

- Condition: `.spectra/sessions/` does not exist but `.spectra/` does.
  Expected: `EnsureSessionsDirectory` creates `.spectra/sessions/` and returns nil.

- Condition: Session directory already exists.
  Expected: `CreateSessionDirectory` returns `ErrSessionDirExists`.

- Condition: `os.Mkdir` for sessions directory fails with permission denied.
  Expected: Returns `"failed to create sessions directory: permission denied"`.

- Condition: `os.Mkdir` for session directory fails with disk full.
  Expected: Returns `"failed to create session directory: no space left on device"`.

- Condition: `sessionUUID` is empty string.
  Expected: StorageLayout produces malformed path. `os.Mkdir` may succeed or fail depending on OS. No UUID validation is performed.

- Condition: `sessionUUID` contains path separators (e.g., `../malicious`).
  Expected: StorageLayout joins as-is. May create directory outside `.spectra/sessions/`. Caller is responsible for UUID validation.

- Condition: Two goroutines call `CreateSessionDirectory` with the same UUID simultaneously.
  Expected: One succeeds; the other gets `ErrSessionDirExists` or a filesystem "file exists" error from `os.Mkdir`.

- Condition: Another process creates the session directory between stat and mkdir.
  Expected: `os.Mkdir` fails with "file exists". Wrapped as `"failed to create session directory: file exists"`.

- Condition: `.spectra` exists as a file (not a directory).
  Expected: `EnsureSessionsDirectory` stat check detects it is not a directory, returns `ErrNotInitialized`.

## Related

- [StorageLayout](./storage_layout.md) — Provides all path composition
- [SpectraFinder](./spectra_finder.md) — Provides `projectRoot`; guarantees `.spectra/` exists before downstream calls
- [FileAccessor](./file_accessor.md) — Sibling I/O utility for file access patterns
