# Test Specification: `directory_creator_test.go`

## Source File Under Test

`internal/cmd/spectra/directory_creator.go`

## Test File

`internal/cmd/spectra/directory_creator_test.go`

---

## `DirectoryCreator`

### Happy Path — CreateAll

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDirectoryCreator_CreateAll_AllNew` | `unit` | Creates all directories when none exist. | Create a temporary directory as `projectRoot` using `t.TempDir()`. | `projectRoot` = temp dir path | Returns `nil`; all 7 directories exist: `.spectra/`, `.spectra/sessions/`, `.spectra/workflows/`, `.spectra/agents/`, `spec/`, `spec/logic/`, `spec/test/`; each has permissions `0755` |
| `TestDirectoryCreator_CreateAll_PartialExist` | `unit` | Creates only missing directories when some already exist. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Pre-create `.spectra/` and `spec/` subdirectories. | `projectRoot` = temp dir path | Returns `nil`; pre-existing directories unchanged; missing directories created with `0755` |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDirectoryCreator_CreateAll_AllExist` | `unit` | Returns nil without modification when all directories already exist. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Pre-create all 7 directories. | `projectRoot` = temp dir path | Returns `nil`; no modifications to existing directories |
| `TestDirectoryCreator_CreateAll_CalledTwice` | `unit` | Second call is idempotent after first successful call. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Call `CreateAll` once successfully. | `projectRoot` = temp dir path (second call) | Returns `nil`; all directories still exist with correct permissions |

### Ordering — Directory Creation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDirectoryCreator_CreateAll_ParentBeforeChild` | `unit` | Parent directories are created before their children (e.g., `.spectra/` before `.spectra/sessions/`). | Create a temporary directory as `projectRoot` using `t.TempDir()`. | `projectRoot` = temp dir path | Returns `nil`; `.spectra/sessions/`, `.spectra/workflows/`, `.spectra/agents/` exist inside `.spectra/`; `spec/logic/`, `spec/test/` exist inside `spec/` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDirectoryCreator_CreateAll_PathExistsAsFile` | `unit` | Returns error when a target directory path exists as a regular file. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create a regular file at `<projectRoot>/.spectra`. | `projectRoot` = temp dir path | Returns error containing `"failed to create directory '.spectra'"` |
| `TestDirectoryCreator_CreateAll_PermissionDenied` | `unit` | Returns error when parent directory is not writable. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Set permissions on `projectRoot` to `0555` (read-only). Cleanup: restore permissions in `t.Cleanup`. | `projectRoot` = temp dir path | Returns error containing `"failed to create directory '.spectra'"` |
| `TestDirectoryCreator_CreateAll_FailFastStopsProcessing` | `unit` | Stops creating subsequent directories after first failure. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create a regular file at `<projectRoot>/.spectra` to block first directory creation. | `projectRoot` = temp dir path | Returns error for `.spectra`; subsequent directories (`.spectra/sessions/`, etc.) do not exist |

### Boundary Values — projectRoot

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestDirectoryCreator_CreateAll_NestedChildMissing` | `unit` | Creates child directory when parent exists but child does not (e.g., `.spectra/` exists, `.spectra/sessions/` does not). | Create a temporary directory as `projectRoot` using `t.TempDir()`. Pre-create only `.spectra/`. | `projectRoot` = temp dir path | Returns `nil`; `.spectra/sessions/`, `.spectra/workflows/`, `.spectra/agents/` created with `0755` |
