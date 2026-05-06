# Test Specification: `storage_layout_race_test.go`

## Source File Under Test
`storage/storage_layout.go`

## Test File
`test/race/storage/storage_layout_race_test.go`

---

## `StorageLayout`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestStorageLayout_ConcurrentAccess` | `race` | Multiple goroutines call path composition functions concurrently without data races. | Launch multiple goroutines calling various Get* functions simultaneously | Various valid inputs (`projectRoot="/home/user/project"`, assorted UUIDs and names) | All calls return correct paths; no race condition detected by `-race` flag |
