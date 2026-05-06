# Test Specification: `nop_logger_race_test.go`

## Source File Under Test
`logger/nop_logger.go`

## Test File
`test/race/logger/nop_logger_race_test.go`

---

## `NopLogger`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNopLogger_ConcurrentCalls` | `race` | Concurrent calls from multiple goroutines do not race. | `logger := NewNopLogger()`; launch multiple goroutines calling all methods concurrently | Multiple goroutines call `Debug`, `Info`, `Warn`, `Error` simultaneously | No data race detected (pass with `-race` flag) |
