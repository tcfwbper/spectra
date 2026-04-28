# Test Specification: `file_accessor.go`

## Source File Under Test
`storage/file_accessor.go`

## Test File
`storage/file_accessor_test.go`

---

## `FileAccessor`

### Happy Path — File Exists

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_FileExists` | `unit` | Returns path immediately when file exists without invoking callback. | Temporary test directory created programmatically within test fixture; file `test.txt` created inside test fixture | `FilePath=<test-fixture>/test.txt`, `PrepareCallback=<mock-callback>` | Returns `<test-fixture>/test.txt`; callback not invoked |
| `TestFileAccessor_FileExistsCallbackNotCalled` | `unit` | Verifies callback is never invoked when file exists. | Temporary test directory created programmatically within test fixture; file `existing.txt` created inside test fixture | `FilePath=<test-fixture>/existing.txt`, `PrepareCallback=<counting-callback>` | Returns path; callback invocation count is 0 |

### Happy Path — File Does Not Exist, Callback Creates It

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_CallbackCreatesFile` | `unit` | Invokes callback when file does not exist; returns path after successful creation. | Temporary test directory created programmatically within test fixture; file `new.txt` does not exist in test fixture | `FilePath=<test-fixture>/new.txt`, `PrepareCallback=<creates-file>` | Callback invoked; file created by callback; returns `<test-fixture>/new.txt` |
| `TestFileAccessor_CallbackCreatesParentDirs` | `unit` | Callback creates parent directories and file. | Temporary test directory created programmatically within test fixture; nested path `a/b/c/file.txt` does not exist in test fixture (parents also missing) | `FilePath=<test-fixture>/a/b/c/file.txt`, `PrepareCallback=<creates-dirs-and-file>` | Callback invoked; parents and file created; returns `<test-fixture>/a/b/c/file.txt` |

### Validation Failures — Callback Returns Error

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_CallbackReturnsError` | `unit` | Returns wrapped error when callback fails. | Temporary test directory created programmatically within test fixture; file `fail.txt` does not exist in test fixture | `FilePath=<test-fixture>/fail.txt`, `PrepareCallback=<returns-error>` | Returns error matching `/failed to prepare file.*fail\.txt:.*<callback-error>/i`; file not created |
| `TestFileAccessor_CallbackErrorParentMissing` | `unit` | Callback returns error indicating parent directory missing. | Temporary test directory created programmatically within test fixture; path `missing/file.txt` does not exist in test fixture; callback designed to fail if parent missing | `FilePath=<test-fixture>/missing/file.txt`, `PrepareCallback=<fails-on-missing-parent>` | Returns error matching `/failed to prepare file.*missing\/file\.txt:.*parent directory/i` |

### Validation Failures — Callback Succeeds but File Not Created

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_CallbackNilButFileNotCreated` | `unit` | Returns error when callback returns nil but file still does not exist. | Temporary test directory created programmatically within test fixture; file `buggy.txt` does not exist in test fixture | `FilePath=<test-fixture>/buggy.txt`, `PrepareCallback=<returns-nil-without-creating-file>` | Returns error matching `/file preparation succeeded but file was not created:.*buggy\.txt/i` |

### Validation Failures — Stat Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_PermissionDeniedBeforeCallback` | `unit` | Returns permission error without invoking callback. | Temporary test directory created programmatically within test fixture; file `restricted.txt` created inside test fixture with permissions `0000` (no read access) | `FilePath=<test-fixture>/restricted.txt`, `PrepareCallback=<mock-callback>` | Returns error matching `/permission denied:.*restricted\.txt/i`; callback not invoked |
| `TestFileAccessor_PermissionDeniedAfterCallback` | `unit` | Returns permission error from post-callback stat. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture; callback creates file but changes permissions during creation | `FilePath=<test-fixture>/denied.txt`, `PrepareCallback=<creates-file-then-removes-permissions>` | Callback invoked; returns error matching `/permission denied:.*denied\.txt/i` |
| `TestFileAccessor_PathIsDirectory` | `unit` | Stat succeeds for directory; returns path (caller will fail when opening as file). | Temporary test directory created programmatically within test fixture; directory `dir/` created inside test fixture | `FilePath=<test-fixture>/dir`, `PrepareCallback=<mock-callback>` | Returns `<test-fixture>/dir`; callback not invoked (directory exists) |

### Validation Failures — Invalid Inputs

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_EmptyFilePath` | `unit` | Returns stat error for empty file path. | | `FilePath=""`, `PrepareCallback=<mock-callback>` | Returns error (stat fails for empty path); callback may or may not be invoked depending on stat behavior |
| `TestFileAccessor_NilCallback` | `unit` | Panics when callback is nil and file does not exist. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture | `FilePath=<test-fixture>/test.txt`, `PrepareCallback=nil` | Panics when attempting to invoke nil callback |

### Boundary Values — Race Conditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_FileCreatedBetweenStatAndCallback` | `unit` | File created by another process between stat checks. | Temporary test directory created programmatically within test fixture; file does not exist initially in test fixture; test simulates external process creating file after initial stat but before callback | `FilePath=<test-fixture>/race.txt`, `PrepareCallback=<mock-callback>` | Initial stat returns not-exist; callback invoked; post-callback stat succeeds; returns path (even though callback didn't create it) |
| `TestFileAccessor_FileDeletedBetweenCallbackAndPostStat` | `unit` | File created by callback but deleted before post-callback stat. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture; callback creates file; test simulates external process deleting file before post-callback stat | `FilePath=<test-fixture>/deleted.txt`, `PrepareCallback=<creates-file>` | Callback succeeds; post-callback stat fails; returns error matching `/file preparation succeeded but file was not created/i` |

### Idempotency — Callback Invocation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_CallbackIdempotent` | `unit` | Callback can be invoked multiple times safely. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture | Call FileAccessor twice with same path; callback checks if file exists before creating | First call invokes callback, creates file; second call sees file exists, skips callback; both calls succeed |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_ConcurrentAccessSameFile` | `race` | Multiple goroutines call FileAccessor for same non-existent file. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture | 5 goroutines call FileAccessor with same `FilePath` and callback that creates file | Multiple callbacks may execute; first to create file wins; subsequent callbacks may fail or succeed depending on timing; FileAccessor returns path for those where post-stat succeeds |
| `TestFileAccessor_ConcurrentAccessDifferentFiles` | `race` | Multiple goroutines call FileAccessor for different files. | Temporary test directory created programmatically within test fixture; multiple files do not exist in test fixture | 10 goroutines call FileAccessor with different file paths and callbacks | All calls succeed; all files created; no data races |

### Happy Path — No Content Access

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_NeverReadsFileContent` | `unit` | FileAccessor only stats the file, never opens or reads it. | Temporary test directory created programmatically within test fixture; file `content.txt` created inside test fixture with content "secret data" | `FilePath=<test-fixture>/content.txt`, `PrepareCallback=<mock-callback>` | Returns path; file content remains unread; no file handle opened |
| `TestFileAccessor_NeverWritesFileContent` | `unit` | FileAccessor never writes to the file. | Temporary test directory created programmatically within test fixture; file `readonly.txt` created inside test fixture with content "original" | `FilePath=<test-fixture>/readonly.txt`, `PrepareCallback=<mock-callback>` | Returns path; file content unchanged |

### Happy Path — Error Wrapping

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_WrapsCallbackError` | `unit` | Callback errors are wrapped with file path context. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture | `FilePath=<test-fixture>/fail.txt`, `PrepareCallback=<returns-error-"disk full">` | Returns error matching `/failed to prepare file.*fail\.txt:.*disk full/i` |

### Happy Path — Callback Panics

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_CallbackPanics` | `unit` | Panic from callback propagates to caller. | Temporary test directory created programmatically within test fixture; file does not exist in test fixture | `FilePath=<test-fixture>/panic.txt`, `PrepareCallback=<panics>` | FileAccessor panics; no recovery |

### Boundary Values — Path Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_AbsolutePath` | `unit` | Handles absolute paths correctly. | Temporary test directory created programmatically within test fixture; file created inside test fixture | `FilePath=<absolute-test-fixture-path>/file.txt`, file exists | Returns absolute path |
| `TestFileAccessor_RelativePath` | `unit` | Passes relative paths to stat as-is; no validation or conversion to absolute. | Temporary test directory created programmatically within test fixture; relative path provided; current working directory is test fixture | `FilePath="./relative.txt"`, file exists in current directory | Returns `./relative.txt` as-is (no conversion to absolute) |
