# Test Specification: `slog_logger_race_test.go`

## Source File Under Test
`logger/slog_logger.go`

## Test File
`test/race/logger/slog_logger_race_test.go`

---

## `SlogLogger`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_ConcurrentCalls` | `race` | Concurrent calls from multiple goroutines do not race. | `logger := NewSlogLogger(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))`; launch multiple goroutines calling all methods concurrently | Multiple goroutines call `Debug`, `Info`, `Warn`, `Error` simultaneously | No data race detected (pass with `-race` flag) |
