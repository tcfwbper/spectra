# Test Specification: `session_directory_manager_test.go`

## Source File Under Test
`storage/session_directory_manager.go`

## Test File
`storage/session_directory_manager_test.go`

---

## `SessionDirectoryManager`

### Happy Path — EnsureSessionsDirectory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEnsureSessionsDirectory_CreatesWhenMissing` | `unit` | Creates `.spectra/sessions/` when it does not exist. | Create temp directory with `.spectra/` subdirectory (no `sessions/` inside). | `projectRoot=<tmpdir>` | Returns nil; `.spectra/sessions/` exists with permissions `0755` |
| `TestEnsureSessionsDirectory_AlreadyExists` | `unit` | Returns nil when `.spectra/sessions/` already exists. | Create temp directory with `.spectra/sessions/` already present. | `projectRoot=<tmpdir>` | Returns nil; directory remains unchanged |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEnsureSessionsDirectory_Idempotent` | `unit` | Multiple calls produce same result without error. | Create temp directory with `.spectra/` subdirectory. | Call `EnsureSessionsDirectory` twice with same `projectRoot` | Both calls return nil; `.spectra/sessions/` exists |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEnsureSessionsDirectory_SpectraDirMissing` | `unit` | Returns ErrNotInitialized when `.spectra/` does not exist. | Create temp directory with no `.spectra/` subdirectory. | `projectRoot=<tmpdir>` | Returns `ErrNotInitialized` |
| `TestEnsureSessionsDirectory_SpectraDirIsFile` | `unit` | Returns ErrNotInitialized when `.spectra` is a file not a directory. | Create temp directory with `.spectra` as a regular file. | `projectRoot=<tmpdir>` | Returns `ErrNotInitialized` |
| `TestEnsureSessionsDirectory_MkdirPermissionDenied` | `unit` | Returns wrapped error when mkdir fails due to permissions. | Create temp directory with `.spectra/` set to permission `0555` (read-only, cannot create subdirectories). | `projectRoot=<tmpdir>` | Returns error matching `"failed to create sessions directory: permission denied"` |

### Happy Path — CreateSessionDirectory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_Success` | `unit` | Creates session directory within `.spectra/sessions/`. | Create temp directory with `.spectra/sessions/` already present. | `projectRoot=<tmpdir>`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns nil; `.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/` exists with permissions `0755` |
| `TestCreateSessionDirectory_EnsuresParent` | `unit` | Automatically creates `.spectra/sessions/` if only `.spectra/` exists. | Create temp directory with `.spectra/` but no `sessions/` subdirectory. | `projectRoot=<tmpdir>`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns nil; both `.spectra/sessions/` and session directory exist |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_AlreadyExists` | `unit` | Returns ErrSessionDirExists when session directory already exists. | Create temp directory with `.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/` already present. | `projectRoot=<tmpdir>`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns `ErrSessionDirExists` |
| `TestCreateSessionDirectory_SpectraDirMissing` | `unit` | Propagates ErrNotInitialized from EnsureSessionsDirectory. | Create temp directory with no `.spectra/`. | `projectRoot=<tmpdir>`, `sessionUUID="some-uuid"` | Returns `ErrNotInitialized` |
| `TestCreateSessionDirectory_MkdirFails` | `unit` | Returns wrapped error when session directory mkdir fails. | Create temp directory with `.spectra/sessions/` set to permission `0555`. | `projectRoot=<tmpdir>`, `sessionUUID="new-session-uuid"` | Returns error matching `"failed to create session directory:"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_EmptyUUID` | `unit` | Accepts empty UUID without validation. | Create temp directory with `.spectra/sessions/`. | `projectRoot=<tmpdir>`, `sessionUUID=""` | Does not panic; returns nil or an `os.PathError` (no UUID validation performed) |

### Boundary Values — sessionUUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_UUIDWithPathSeparator` | `unit` | UUID containing path traversal characters is passed through without validation. | Create temp directory with `.spectra/sessions/`. | `projectRoot=<tmpdir>`, `sessionUUID="../escape"` | Does not panic; does not return a validation error; path traversal is not prevented by SessionDirectoryManager |
