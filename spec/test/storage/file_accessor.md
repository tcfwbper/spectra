# Test Specification: `file_accessor_test.go`

## Source File Under Test
`storage/file_accessor.go`

## Test File
`storage/file_accessor_test.go`

---

## `FileAccessor`

### Happy Path — FileAccessor

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_FileExists` | `unit` | Returns path immediately when file already exists. | Create a temporary file in a test fixture directory. | `filePath=<temp file path>`, `prepare=<callback that should not be called>` | Returns the file path and nil error; callback is never invoked |
| `TestFileAccessor_FileNotExistsCallbackCreates` | `unit` | Invokes callback when file does not exist; returns path after callback creates file. | Create a temporary directory (no target file). Callback creates the file at the given path. | `filePath=<non-existent path in temp dir>`, `prepare=<callback that creates the file>` | Returns the file path and nil error |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_CallbackReturnsError` | `unit` | Returns wrapped error when callback fails. | Create a temporary directory (no target file). Callback returns a known error. | `filePath=<non-existent path in temp dir>`, `prepare=<callback returning errors.New("disk full")>` | Returns error matching `"failed to prepare file <path>: disk full"` |
| `TestFileAccessor_CallbackSucceedsButFileNotCreated` | `unit` | Returns error when callback returns nil but file still does not exist. | Create a temporary directory (no target file). Callback is a no-op (returns nil without creating file). | `filePath=<non-existent path in temp dir>`, `prepare=<no-op callback>` | Returns error matching `"file preparation succeeded but file was not created: <path>"` |
| `TestFileAccessor_InitialStatPermissionDenied` | `unit` | Returns stat error immediately when initial stat fails with non-ErrNotExist. | Create a temporary directory with permissions `0000` so stat on a child path fails with permission denied. | `filePath=<path inside restricted dir>`, `prepare=<callback that should not be called>` | Returns a permission error; callback is never invoked |
| `TestFileAccessor_PostCallbackStatPermissionDenied` | `unit` | Returns stat error when post-callback stat fails with non-ErrNotExist. | Create a temporary directory. Callback creates the file, then changes parent directory permissions to `0000` before returning. | `filePath=<path in temp dir>`, `prepare=<callback that creates file then restricts parent>` | Returns error wrapping the permission error with path context |
| `TestFileAccessor_CallbackPanics` | `unit` | Panic in callback propagates to caller without recovery. | Create a temporary directory (no target file). Callback panics with a known value. Test uses `recover` to capture the panic. | `filePath=<non-existent path in temp dir>`, `prepare=<callback that panics with "boom">` | Panic propagates; recovered value equals `"boom"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_EmptyFilePath` | `unit` | Returns stat error when filePath is empty string. | | `filePath=""`, `prepare=<no-op callback>` | Returns an error from stat (e.g., "no such file or directory"); callback may or may not be invoked depending on stat error type |
| `TestFileAccessor_NilCallbackFileNotExists` | `unit` | Panics when prepare is nil and file does not exist. | Create a temporary directory (no target file). | `filePath=<non-existent path>`, `prepare=nil` | Panics with nil function call |

### Boundary Values — filePath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_FilePathIsDirectory` | `unit` | Returns path when filePath points to an existing directory. | Create a temporary directory. | `filePath=<temp directory path>`, `prepare=<callback that should not be called>` | Returns the directory path and nil error (stat succeeds on directories); callback is never invoked |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFileAccessor_IndependentInvocations` | `unit` | Each call performs fresh stat; no state carried between calls. | Create a temporary directory. First call: file does not exist, callback creates it. Second call: file now exists. | Two sequential calls with same `filePath` and a counting callback | First call invokes callback once. Second call does not invoke callback. Both return the path. |
