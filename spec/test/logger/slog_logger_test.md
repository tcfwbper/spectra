# Test Specification: `slog_logger_test.go`

## Source File Under Test
`logger/slog_logger.go`

## Test File
`logger/slog_logger_test.go`

---

## `SlogLogger`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSlogLogger_WithValidLogger` | `unit` | Constructor accepts a valid *slog.Logger and returns Logger. | Create `*slog.Logger` with `slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))` | `NewSlogLogger(slogger)` | Returned value is non-nil and assignable to `Logger` interface variable |
| `TestNewSlogLogger_WithNilFallsBackToDefault` | `unit` | Constructor accepts nil and falls back to slog.Default(). | | `NewSlogLogger(nil)` | Returned value is non-nil and assignable to `Logger`; subsequent calls do not panic |

### Happy Path — Debug

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_Debug_DelegatesToSlog` | `unit` | Debug forwards msg and args to underlying slog.Logger.Debug. | Create `*slog.Logger` with a `slog.NewTextHandler` writing to a `*bytes.Buffer` at `slog.LevelDebug` | `logger.Debug("test event", "key", "value")` | Buffer contains output with level=DEBUG, msg="test event", key=value |

### Happy Path — Info

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_Info_DelegatesToSlog` | `unit` | Info forwards msg and args to underlying slog.Logger.Info. | Create `*slog.Logger` with a `slog.NewTextHandler` writing to a `*bytes.Buffer` at `slog.LevelDebug` | `logger.Info("info event", "count", 42)` | Buffer contains output with level=INFO, msg="info event", count=42 |

### Happy Path — Warn

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_Warn_DelegatesToSlog` | `unit` | Warn forwards msg and args to underlying slog.Logger.Warn. | Create `*slog.Logger` with a `slog.NewTextHandler` writing to a `*bytes.Buffer` at `slog.LevelDebug` | `logger.Warn("warn event", "detail", "something")` | Buffer contains output with level=WARN, msg="warn event", detail=something |

### Happy Path — Error

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_Error_DelegatesToSlog` | `unit` | Error forwards msg and args to underlying slog.Logger.Error. | Create `*slog.Logger` with a `slog.NewTextHandler` writing to a `*bytes.Buffer` at `slog.LevelDebug` | `logger.Error("error event", "err", "timeout")` | Buffer contains output with level=ERROR, msg="error event", err=timeout |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_EmptyMessage` | `unit` | Passes empty string message to slog without panic. | Create `*slog.Logger` with a buffer handler at `slog.LevelDebug` | `logger.Info("")` | Buffer contains output with empty msg; no panic |
| `TestSlogLogger_NoArgs` | `unit` | Passes message with no variadic args to slog without panic. | Create `*slog.Logger` with a buffer handler at `slog.LevelDebug` | `logger.Info("msg")` | Buffer contains output with msg="msg" and no extra fields; no panic |
| `TestSlogLogger_OddArgs` | `unit` | Passes odd number of args to slog (implementation-defined behavior). | Create `*slog.Logger` with a buffer handler at `slog.LevelDebug` | `logger.Info("msg", "orphan_key")` | No panic; slog handles gracefully (e.g., `!BADKEY` marker) |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_ImplementsLogger` | `unit` | SlogLogger satisfies Logger interface at compile time. | | Compile-time assignment: `var _ Logger = NewSlogLogger(nil)` | Compiles without error |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSlogLogger_PassThroughNoTransformation` | `unit` | Args are forwarded to slog without transformation. | Create `*slog.Logger` with a buffer handler at `slog.LevelDebug` | `logger.Info("evt", "a", 1, "b", "two")` | Buffer output contains both key-value pairs exactly as provided: a=1, b=two |

