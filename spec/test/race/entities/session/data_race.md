# Test Specification: `data_race_test.go`

## Source File Under Test
`entities/session/data.go`

## Test File
`test/race/entities/session/data_race_test.go`

---

## `UpdateSessionDataSafe` / `GetSessionDataSafe`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateSessionDataSafe_ConcurrentWrites` | `race` | Concurrent writes to same key are serialized without data race. | Construct session | Launch multiple goroutines each calling `UpdateSessionDataSafe("k", distinctValue)` | No data race (run with `-race`); final value is one of the written values |
| `TestGetSessionDataSafe_ConcurrentReadDuringWrite` | `race` | Concurrent reads during writes do not race. | Construct session; pre-populate key | Launch reader and writer goroutines concurrently | No data race (run with `-race`); reader always gets a consistent value |
