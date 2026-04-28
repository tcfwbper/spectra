# Test Specification: `spectra_finder.go`

## Source File Under Test
`storage/spectra_finder.go`

## Test File
`storage/spectra_finder_test.go`

---

## `SpectraFinder`

### Happy Path — Find in Current Directory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_FindInCurrentDir` | `unit` | Finds .spectra in the current directory immediately. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture directory; test changes working directory to test fixture | `StartDir` omitted (defaults to current directory) | Returns absolute path to test fixture directory; no upward traversal |
| `TestSpectraFinder_FindWithExplicitStartDir` | `unit` | Finds .spectra when StartDir explicitly provided. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture directory | `StartDir=<test-fixture-directory-path>` | Returns absolute path to test fixture directory |

### Happy Path — Find in Parent Directory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_FindInParent` | `unit` | Searches upward and finds .spectra in parent directory. | Temporary test directory created programmatically within test fixture; `parent/.spectra/` directory and `parent/child/` directory created inside test fixture; start from `parent/child/` | `StartDir="<test-fixture>/parent/child"` | Returns absolute path to `<test-fixture>/parent/` |
| `TestSpectraFinder_FindMultipleLevelsUp` | `unit` | Searches upward through multiple directories. | Temporary test directory created programmatically within test fixture; `root/.spectra/` directory and nested directories `root/a/b/c/` created inside test fixture; start from `root/a/b/c/` | `StartDir="<test-fixture>/root/a/b/c"` | Returns absolute path to `<test-fixture>/root/` |

### Happy Path — Nearest .spectra Wins

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_NearestSpectraWins` | `unit` | Returns nearest .spectra when multiple exist in hierarchy. | Temporary test directory created programmatically within test fixture; `root/.spectra/`, `root/project/.spectra/`, and `root/project/subdir/` directories all created inside test fixture | `StartDir="<test-fixture>/root/project/subdir"` | Returns absolute path to `<test-fixture>/root/project/` (not `root/`) |

### Happy Path — Symbolic Link Resolution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_FollowsSymlinks` | `unit` | Follows symbolic links during upward traversal. | Temporary test directory created programmatically within test fixture; `real/.spectra/` directory and `real/subdir/` created inside test fixture; symlink `link -> real/subdir` created inside test fixture | `StartDir="<test-fixture>/link"` | Returns absolute path to `<test-fixture>/real/` (resolved from symlink) |

### Validation Failures — Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_NotFoundReachesRoot` | `unit` | Returns error when .spectra not found after reaching filesystem root. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created in test fixture; test ensures no `.spectra/` exists in any parent up to filesystem root | `StartDir=<test-fixture-path>` | Returns error matching `/spectra not initialized/i` |
| `TestSpectraFinder_SpectraIsFile` | `unit` | Continues searching upward if .spectra is a file, not a directory. | Temporary test directory created programmatically within test fixture; `parent/.spectra` created as a regular file (not directory) inside test fixture; `parent/child/` directory created inside test fixture | `StartDir="<test-fixture>/parent/child"` | Continues searching upward past `parent/`; returns error `/spectra not initialized/i` if no directory found |

### Validation Failures — Invalid Start Directory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_StartDirNotExist` | `unit` | Returns error when StartDir does not exist. | Temporary test directory created programmatically within test fixture | `StartDir="<test-fixture>/nonexistent/directory"` | Returns error matching `/invalid start directory:.*nonexistent/i` |
| `TestSpectraFinder_StartDirIsFile` | `unit` | Returns error when StartDir is a file, not a directory. | Temporary test directory created programmatically within test fixture; regular file `file.txt` created inside test fixture | `StartDir=<test-fixture>/file.txt` | Returns error matching `/invalid start directory:.*file\.txt/i` |

### Validation Failures — Permission Denied

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_PermissionDeniedOnParent` | `unit` | Returns error when parent directory is not readable. | Temporary test directory created programmatically within test fixture; `parent/child/` directories created inside test fixture; `parent/` permissions set to `0000` (no read access) within test fixture | `StartDir="<test-fixture>/parent/child"` | Returns error matching `/spectra not initialized/i` (permission error during traversal) |

### Validation Failures — Symbolic Link Loop

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_SymlinkLoop` | `unit` | Returns error when encountering symbolic link loop during traversal. | Temporary test directory created programmatically within test fixture; symlink loop `a -> b` and `b -> a` created inside test fixture | `StartDir=<test-fixture>/a` | Returns error matching `/spectra not initialized/i` (loop causes traversal to fail or timeout) |

### Boundary Values — Filesystem Root

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_StartAtFilesystemRoot` | `unit` | Returns error when starting at filesystem root with no .spectra. | Temporary test directory created programmatically within test fixture; test verifies behavior starting from `/` (read-only operation, no filesystem modification) | `StartDir="/"` | Returns error matching `/spectra not initialized/i` |

### Boundary Values — Relative Path Resolution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_RelativeStartDir` | `unit` | Resolves relative StartDir to absolute path before searching. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `StartDir="."` | Returns absolute path to test fixture directory |
| `TestSpectraFinder_RelativeStartDirWithParent` | `unit` | Resolves relative path with parent references. | Temporary test directory created programmatically within test fixture; `root/.spectra/` directory and `root/a/b/` directories created inside test fixture; test changes working directory to `<test-fixture>/root/a/b/` | `StartDir=".."` | Returns absolute path to `<test-fixture>/root/` |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_RepeatedSearch` | `unit` | Multiple searches from same directory return identical results. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture | Call finder three times with same `StartDir=<test-fixture-path>` | All three calls return identical absolute path |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_ConcurrentSearches` | `race` | Multiple goroutines search for .spectra simultaneously. | Temporary test directory created programmatically within test fixture; `.spectra/` and nested directories created inside test fixture | 10 goroutines each call finder with same or different start directories within test fixture hierarchy | All calls succeed; no data races; all return correct absolute paths |

### Happy Path — Absolute Path Return

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_ReturnsAbsolutePath` | `unit` | Returned ProjectRoot is always an absolute path. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `StartDir=<relative-path-within-test-fixture>` | Returned path starts with `/` (Unix) or drive letter (Windows); is an absolute path |

### Happy Path — No Caching

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpectraFinder_NoCaching` | `unit` | Finder performs fresh search on each invocation, no caching. | Temporary test directory created programmatically within test fixture; no `.spectra/` initially; first search fails; `.spectra/` directory then created inside test fixture; second search from same directory | First call with `StartDir=<test-fixture>`, then create `.spectra/` inside test fixture, then second call with same `StartDir` | First call returns error; second call succeeds and returns test fixture path |
