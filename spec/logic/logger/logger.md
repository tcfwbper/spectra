# Logger — Interface

## Overview

Logger is a structured logging interface that provides four standard log levels (Debug, Info, Warn, Error). It follows the `log/slog` key-value pair convention for structured fields. Logger decouples all application modules from concrete logging implementations, enabling testability and flexibility in logging backends.

Logger does not own log output destination, formatting, or filtering — those are implementation concerns.

## Boundaries

- Owns: interface contract definition for structured logging across the entire application.
- Delegates: log output formatting, filtering, and destination to concrete implementations.
- Must not: prescribe a concrete implementation (stdlib `log/slog`, third-party loggers, etc.).
- Must not: provide a concrete implementation within this file (implementations are in separate units).
- Must not: include Fatal or Panic level methods (library code must not call os.Exit or panic).
- Must not: accept `context.Context` parameters.

## Dependencies

None. This is a leaf interface with no collaborators.

## Behavior

1. Defines a four-method interface:

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

2. `msg` is a human-readable description of the log event.
3. `args` are structured key-value pairs following `log/slog` conventions: alternating key (string) and value (any). Implementations format these as structured fields.
4. All methods are safe for concurrent use. Implementations must handle concurrent calls from multiple goroutines.
5. All methods are fire-and-forget. They do not return values or errors.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| msg | string | Human-readable log event description | Yes |
| args | ...any | Structured key-value pairs (alternating key string, value any) | No |

## Outputs

None. All methods are void. Logging failures are silently ignored by implementations.

## Invariants

1. **Interface Only**: This unit is an interface definition. No concrete implementation is provided here.
2. **Four Levels Only**: Only Debug, Info, Warn, Error are exposed. No Fatal, Panic, or Trace.
3. **Concurrent Safety**: Implementations must be safe for concurrent use from multiple goroutines.
4. **No Return Value**: No method returns errors or values. Logging is fire-and-forget.
5. **No Context**: Methods do not accept `context.Context`. Context-aware logging is out of scope.
6. **slog Key-Value Convention**: `args` follows the `log/slog` alternating key-value convention. Implementations should handle odd-length args gracefully (implementation-defined behavior).

## Edge Cases

- Condition: `msg` is an empty string.
  Expected: Implementation-defined behavior. The interface does not constrain empty messages.

- Condition: `args` contain an odd number of elements (missing value for last key).
  Expected: Implementation-defined behavior (e.g., `log/slog` appends a `!BADKEY` marker).

- Condition: `args` contain a non-string key.
  Expected: Implementation-defined behavior (e.g., `log/slog` converts to string representation).

## Related

- [SlogLogger](./slog_logger.md) — Default concrete implementation based on `log/slog`
- [NopLogger](./nop_logger.md) — No-op implementation for testing
