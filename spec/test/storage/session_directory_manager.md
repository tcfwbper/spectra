# Test Specification: `session_directory_manager.go`

## Source File Under Test
`storage/session_directory_manager.go`

## Test File
`storage/session_directory_manager_test.go`

---

## `SessionDirectoryManager`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionDirectoryManager_ValidConstruction` | `unit` | Creates manager with valid project root. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists inside test fixture | `ProjectRoot=<test-fixture-path>` | Returns valid SessionDirectoryManager instance; no error |

### Happy Path — CreateSessionDirectory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_ValidUUID` | `unit` | Creates session directory with valid UUID. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists inside test fixture | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns nil; directory created at `<test-fixture>/.spectra/sessions/123e4567-e89b-12d3-a456-426614174000` with permissions `0775` |
| `TestCreateSessionDirectory_Permissions0775` | `unit` | Created directory has permissions 0775. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists inside test fixture | `SessionUUID="a1b2c3d4-e5f6-7890-abcd-ef1234567890"` | Directory created; stat shows permissions `0775` (owner rwx, group rwx, others rx) |

### Validation Failures — Parent Directory Missing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_SessionsParentMissing` | `unit` | Returns error when .spectra/sessions/ does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/` exists but `.spectra/sessions/` does not exist | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns error matching `/sessions directory does not exist:.*\.spectra\/sessions.*Run 'spectra init'/i` |
| `TestCreateSessionDirectory_SpectraDirMissing` | `unit` | Returns error when .spectra/ does not exist. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory exists | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns error matching `/sessions directory does not exist:.*\.spectra\/sessions.*Run 'spectra init'/i` |

### Validation Failures — Directory Already Exists

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_AlreadyExists` | `unit` | Returns error when session directory already exists. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists; session directory `123e4567-e89b-12d3-a456-426614174000` already exists | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns error matching `/session directory already exists:.*123e4567-e89b-12d3-a456-426614174000.*UUID collision/i` |

### Validation Failures — Filesystem Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_PermissionDenied` | `unit` | Returns error when directory creation fails due to permission denied. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists with permissions `0555` (read-only) | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns error matching `/failed to create session directory:.*permission denied/i` |
| `TestCreateSessionDirectory_DiskFull` | `unit` | Returns error when directory creation fails due to disk full. | Simulated environment with full disk | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns error matching `/failed to create session directory:.*no space left on device/i` |

### Boundary Values — Empty UUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_EmptyUUID` | `unit` | Attempts to create directory with empty UUID (malformed path). | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists inside test fixture | `SessionUUID=""` | Behavior depends on OS; may create directory at `.spectra/sessions/` or fail; no UUID validation performed |

### Boundary Values — UUID with Path Separators

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_UUIDWithPathSeparator` | `unit` | No validation of UUID format; potentially dangerous paths passed through. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists inside test fixture | `SessionUUID="../malicious"` | Attempts to create directory at `.spectra/sessions/../malicious` (potentially outside sessions dir); may succeed or fail depending on filesystem state |

### Boundary Values — Relative Project Root

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionDirectoryManager_RelativeProjectRoot` | `unit` | Uses relative path if ProjectRoot is relative. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists; current working directory is test directory | `ProjectRoot="."`, `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Creates directory at `./.spectra/sessions/123e4567-e89b-12d3-a456-426614174000` (relative to current directory) |

### Boundary Values — Path Length Limits

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_PathExceedsLimit` | `unit` | Returns error when full path exceeds platform maximum. | Temporary test directory created programmatically within test fixture with very long path; `.spectra/sessions/` exists | `SessionUUID=<very-long-uuid-string>` causing total path length to exceed limit | Returns error from OS (e.g., "file name too long"); manager propagates error |

### Idempotency — Stateless Operations

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionDirectoryManager_NoStateCaching` | `unit` | Manager is stateless; each call performs fresh filesystem checks. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists initially; manager instance created; delete `.spectra/sessions/` directory | Call `CreateSessionDirectory` after deletion | Returns error (sessions directory does not exist); no cached state from initialization |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_ConcurrentSameUUID` | `race` | Two goroutines attempt to create same session directory simultaneously. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists | Both goroutines call `CreateSessionDirectory` with same `SessionUUID` | One succeeds, one fails with "already exists" or "file exists" error; no data corruption |
| `TestCreateSessionDirectory_ConcurrentDifferentUUIDs` | `race` | Multiple goroutines create different session directories simultaneously. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists | 10 goroutines call `CreateSessionDirectory` with different UUIDs | All succeed; all 10 directories created with correct permissions; no data races |

### Happy Path — No Parent Directory Creation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionDirectoryManager_DoesNotCreateParents` | `unit` | Manager never creates .spectra/ or .spectra/sessions/ directories. | Temporary test directory created programmatically within test fixture; only `.spectra/` exists (no `sessions/` subdirectory) | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns error; does not create `.spectra/sessions/` directory |

### Happy Path — Path Composition Delegation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionDirectoryManager_UsesStorageLayout` | `unit` | Manager delegates all path composition to StorageLayout. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Created directory path matches `StorageLayout.GetSessionDir(ProjectRoot, SessionUUID)` exactly |

### Concurrent Behaviour — Race with External Process

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_ExternalProcessCreates` | `unit` | Another process creates directory between existence check and mkdir call. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists; simulate external process creating directory after existence check | `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | `os.Mkdir` fails with "file exists"; manager returns error matching `/failed to create session directory:.*file exists/i` |

### Happy Path — Thread Safety

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionDirectoryManager_ThreadSafe` | `race` | Multiple goroutines use same manager instance concurrently. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` exists; single SessionDirectoryManager instance | 10 goroutines call `CreateSessionDirectory` on same manager with different UUIDs | All calls succeed; no data races detected on manager instance |
