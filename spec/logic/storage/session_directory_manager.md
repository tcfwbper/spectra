# SessionDirectoryManager

## Overview

SessionDirectoryManager provides methods to create and validate session directories within `.spectra/sessions/`. It creates session-specific directories with appropriate permissions and validates that parent directories exist. SessionDirectoryManager does not delete directories or manage directory contents (files within the directory).

## Behavior

1. SessionDirectoryManager is initialized with a ProjectRoot path (absolute path to the directory containing `.spectra`).
2. SessionDirectoryManager uses StorageLayout to compose absolute paths to session directories.
3. When `CreateSessionDirectory(sessionUUID)` is called, SessionDirectoryManager composes the session directory path using `StorageLayout.GetSessionDir(ProjectRoot, sessionUUID)`.
4. SessionDirectoryManager validates that the parent directory `.spectra/sessions/` exists. If it does not exist, it returns an error: `"sessions directory does not exist: <path>. Run 'spectra init' to initialize the project."`.
5. SessionDirectoryManager checks if the session directory already exists. If it exists, it returns an error: `"session directory already exists: <path>. This indicates a UUID collision or a previous session was not cleaned up properly."`.
6. SessionDirectoryManager creates the session directory using `os.Mkdir()` with permissions `0775` (owner read/write/execute, group read/write/execute, others read/execute).
7. If directory creation fails (e.g., permission denied, disk full), SessionDirectoryManager returns an error: `"failed to create session directory: <error>"`.
8. If directory creation succeeds, SessionDirectoryManager returns nil (success).
9. SessionDirectoryManager does not create parent directories (`.spectra/` or `.spectra/sessions/`). These must exist before creating session directories.
10. SessionDirectoryManager is stateless and does not cache directory existence checks.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### For CreateSessionDirectory Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| SessionUUID | string (UUID) | Valid UUID v4 format | Yes |

## Outputs

### For CreateSessionDirectory Method

**Success Case**: No return value (void / nil error in Go).

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"sessions directory does not exist: <path>. Run 'spectra init' to initialize the project."` | Parent directory `.spectra/sessions/` does not exist |
| `"session directory already exists: <path>. This indicates a UUID collision or a previous session was not cleaned up properly."` | Session directory already exists at the target path |
| `"failed to create session directory: <error>"` | Directory creation failed (e.g., permission denied, disk full) |

## Invariants

1. **No Parent Directory Creation**: SessionDirectoryManager must not create `.spectra/` or `.spectra/sessions/` directories. It must return an error if they do not exist.

2. **Permission 0775**: Newly created session directories must have permissions `0775`.

3. **UUID Collision Detection**: SessionDirectoryManager must check for existing directories before creation and return an error if the directory already exists.

4. **Atomic Directory Creation**: Directory creation uses `os.Mkdir()`, which is atomic at the filesystem level. If creation fails, no partial directory is left.

5. **Stateless**: SessionDirectoryManager must not cache directory existence. Each `CreateSessionDirectory()` call performs fresh filesystem checks.

6. **Thread-Safe**: SessionDirectoryManager must be safe to call concurrently from multiple goroutines, but race conditions between existence check and directory creation are possible (filesystem-level).

7. **Path Composition Delegation**: SessionDirectoryManager must delegate all path composition to StorageLayout. It must not construct paths manually.

## Edge Cases

- **Condition**: `.spectra/sessions/` directory does not exist.
  **Expected**: SessionDirectoryManager returns an error: `"sessions directory does not exist: <path>. Run 'spectra init' to initialize the project."`.

- **Condition**: Session directory already exists at the target path.
  **Expected**: SessionDirectoryManager returns an error: `"session directory already exists: <path>. This indicates a UUID collision or a previous session was not cleaned up properly."`.

- **Condition**: Session directory creation fails due to permission denied.
  **Expected**: SessionDirectoryManager returns an error: `"failed to create session directory: permission denied"`.

- **Condition**: Session directory creation fails due to disk full.
  **Expected**: SessionDirectoryManager returns an error: `"failed to create session directory: no space left on device"`.

- **Condition**: `SessionUUID` is an empty string.
  **Expected**: StorageLayout produces a malformed path (e.g., `.spectra/sessions//`). Directory creation may succeed or fail depending on OS behavior. SessionDirectoryManager does not validate UUID format.

- **Condition**: `SessionUUID` contains path separators (e.g., `"../malicious"`).
  **Expected**: StorageLayout composes the path as-is, potentially creating a directory outside `.spectra/sessions/`. SessionDirectoryManager does not validate UUID content. Caller is responsible for UUID validation.

- **Condition**: `ProjectRoot` is a relative path (e.g., `"./project"`).
  **Expected**: Path composition produces a relative path. Directory creation may succeed or fail depending on the current working directory. Caller is responsible for providing an absolute ProjectRoot.

- **Condition**: `ProjectRoot` does not contain a `.spectra/` directory.
  **Expected**: The `.spectra/sessions/` check fails. SessionDirectoryManager returns an error: `"sessions directory does not exist: <path>. Run 'spectra init' to initialize the project."`.

- **Condition**: Two goroutines call `CreateSessionDirectory()` with the same `SessionUUID` simultaneously.
  **Expected**: Both goroutines perform existence checks. If the directory does not exist initially, both may attempt to create it. The first `os.Mkdir()` call succeeds; the second fails with "file exists" error, which SessionDirectoryManager wraps as `"failed to create session directory: file exists"`. This is acceptable behavior; UUID collision should be rare.

- **Condition**: Another process creates the session directory between the existence check and the `os.Mkdir()` call.
  **Expected**: `os.Mkdir()` fails with "file exists" error. SessionDirectoryManager returns: `"failed to create session directory: file exists"`.

- **Condition**: Session directory path exceeds platform maximum path length.
  **Expected**: Directory creation fails with a filesystem error (e.g., "file name too long"). SessionDirectoryManager propagates the error.

## Related

- [StorageLayout](./storage_layout.md) - Provides session directory paths
- [SessionMetadataStore](./session_metadata_store.md) - Requires session directory to exist before writing metadata
- [EventStore](./event_store.md) - Requires session directory to exist before writing events
- [RuntimeSocketManager](./runtime_socket_manager.md) - Requires session directory to exist before creating socket
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Session lifecycle and storage structure
