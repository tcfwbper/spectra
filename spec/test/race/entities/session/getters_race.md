# Test Specification: `getters_race_test.go`

## Source File Under Test
`entities/session/getters.go`

## Test File
`test/race/entities/session/getters_race_test.go`

---

## `GetStatusSafe` / `GetCurrentStateSafe` / `GetErrorSafe` / `GetMetadataSnapshotSafe`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetters_ConcurrentReadDuringWrite` | `race` | Concurrent getter calls during mutations do not race. | Construct session | Launch goroutines calling `Run()`, `UpdateSessionDataSafe`, `UpdateCurrentStateSafe` concurrently with goroutines calling all getters | No data race (run with `-race`) |
