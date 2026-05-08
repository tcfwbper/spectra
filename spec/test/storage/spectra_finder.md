# Test Specification: `spectra_finder_test.go`

## Source File Under Test
`storage/spectra_finder.go`

## Test File
`storage/spectra_finder_test.go`

---

## `SpectraFinder`

### Happy Path — SpectraFinder

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_FoundInStartDir` | `unit` | Returns start directory when `.spectra` directory exists there. | Create temp directory structure: `<tmpdir>/.spectra/` (as a directory). | `startDir=<tmpdir>` | Returns `<tmpdir>` as projectRoot; no error |
| `TestSpectraFinder_FoundInParent` | `unit` | Returns ancestor directory when `.spectra` is in a parent. | Create temp directory structure: `<tmpdir>/.spectra/` and `<tmpdir>/sub/deep/`. | `startDir=<tmpdir>/sub/deep` | Returns `<tmpdir>` as projectRoot; no error |
| `TestSpectraFinder_FoundNearestAncestor` | `unit` | Returns the nearest (deepest) directory containing `.spectra`. | Create: `<tmpdir>/.spectra/`, `<tmpdir>/inner/.spectra/`, `<tmpdir>/inner/child/`. | `startDir=<tmpdir>/inner/child` | Returns `<tmpdir>/inner` as projectRoot |
| `TestSpectraFinder_EmptyStartDirUsesCwd` | `unit` | Uses current working directory when startDir is empty. | Create temp directory with `.spectra/` inside. Change working directory to that temp directory using test fixture. | `startDir=""` | Returns the temp directory as projectRoot; no error |
| `TestSpectraFinder_RelativeStartDir` | `unit` | Resolves relative path before starting traversal. | Create temp directory structure with `.spectra/` at the root. Change working directory so that relative path resolves correctly. | `startDir=<relative path to temp dir>` | Returns absolute path of the directory containing `.spectra`; no error |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_NotFoundReturnsErrNotInitialized` | `unit` | Returns ErrNotInitialized when `.spectra` is not found up to root. | Create a temp directory with no `.spectra` in it or any ancestor up to a controlled boundary. Use a deeply nested temp directory without `.spectra`. | `startDir=<tmpdir>/a/b/c` (no `.spectra` anywhere in `<tmpdir>`) | Returns `ErrNotInitialized` |
| `TestSpectraFinder_PermissionDeniedReturnsErrNotInitialized` | `unit` | Returns ErrNotInitialized when parent directory has no read permission. | Create temp directory structure: `<tmpdir>/restricted/child/`. Set `<tmpdir>/restricted/` to permission `0000` after creating child. Note: `.spectra` does not exist in child. | `startDir=<tmpdir>/restricted/child` | Returns `ErrNotInitialized` |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_NonExistentStartDir` | `unit` | Returns error when startDir path does not exist. | | `startDir="/tmp/non-existent-path-abc123"` | Returns error matching `"invalid start directory: /tmp/non-existent-path-abc123"` |
| `TestSpectraFinder_StartDirIsFile` | `unit` | Returns error when startDir is a regular file. | Create a temporary file (not a directory). | `startDir=<path to temp file>` | Returns error matching `"invalid start directory: <path>"` |
| `TestSpectraFinder_GetwdFails` | `unit` | Returns error when startDir is empty and os.Getwd fails. | Override or simulate Getwd failure (e.g., change to a directory then remove it, or use a test seam if available). | `startDir=""` | Returns error matching `"failed to get working directory: <error>"` |

### Boundary Values — startDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_StartDirIsRoot` | `unit` | Returns ErrNotInitialized when starting from filesystem root with no `.spectra`. | Ensure no `.spectra` directory exists at filesystem root (test should not create one at root). | `startDir="/"` | Returns `ErrNotInitialized` |
| `TestSpectraFinder_SpectraExistsAsFile` | `unit` | Ignores `.spectra` when it is a regular file and continues upward. | Create: `<tmpdir>/parent/.spectra` (as directory), `<tmpdir>/parent/child/.spectra` (as a regular file). | `startDir=<tmpdir>/parent/child` | Returns `<tmpdir>/parent` as projectRoot (skips the file, finds the directory) |
| `TestSpectraFinder_SymlinkLoop` | `unit` | Returns ErrNotInitialized when symlink loop is detected. | Create temp directory with a symlink that creates a loop (e.g., `<tmpdir>/a/link -> <tmpdir>/a`). No `.spectra` in the structure. | `startDir=<tmpdir>/a/link` | Returns `ErrNotInitialized` |
