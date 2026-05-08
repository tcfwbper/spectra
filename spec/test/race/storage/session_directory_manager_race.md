# Test Specification: `session_directory_manager_race_test.go`

## Source File Under Test
`storage/session_directory_manager.go`

## Test File
`test/race/storage/session_directory_manager_race_test.go`

---

## `SessionDirectoryManager`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCreateSessionDirectory_ConcurrentSameUUID` | `race` | Two goroutines creating same session directory; one succeeds, one gets error. | Create temp directory with `.spectra/sessions/`. Launch two goroutines concurrently. | Both goroutines call `CreateSessionDirectory` with same `projectRoot` and `sessionUUID` | One returns nil; the other returns `ErrSessionDirExists` or a wrapped "file exists" error from `os.Mkdir` |
