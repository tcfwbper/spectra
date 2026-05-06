# SlogLogger — Default Implementation

## Overview

SlogLogger is a concrete implementation of the Logger interface that delegates all logging to Go's standard library `log/slog`. It wraps a `*slog.Logger` instance and maps each Logger method to the corresponding slog level. SlogLogger does not own the underlying slog handler configuration (output destination, format, minimum level) — those are determined by the `*slog.Logger` passed at construction time.

## Boundaries

- Owns: mapping Logger interface methods to `slog.Logger` method calls at the correct level.
- Delegates: output formatting, filtering, and destination to the injected `*slog.Logger`.
- Must not: configure slog handlers internally (no hardcoded stdout, JSON format, or level filtering).
- Must not: add additional log levels or behaviors beyond what the Logger interface defines.
- Must not: wrap or transform errors — it only logs them.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `*slog.Logger` | Underlying logging engine | Call `Debug`, `Info`, `Warn`, `Error` methods | Must not reconfigure handler or set global default |

Construction constraint: Must be constructed via `NewSlogLogger(slogger *slog.Logger) Logger`. Direct struct literal is forbidden. If `slogger` is nil, the constructor uses `slog.Default()` as fallback.

## Behavior

1. `NewSlogLogger(slogger)` stores the provided `*slog.Logger` internally. If `slogger` is nil, uses `slog.Default()`.
2. `Debug(msg, args...)` calls the underlying slog logger's `Debug` method with the same `msg` and `args`.
3. `Info(msg, args...)` calls the underlying slog logger's `Info` method with the same `msg` and `args`.
4. `Warn(msg, args...)` calls the underlying slog logger's `Warn` method with the same `msg` and `args`.
5. `Error(msg, args...)` calls the underlying slog logger's `Error` method with the same `msg` and `args`.
6. All method calls are direct pass-through with no transformation of `msg` or `args`.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| slogger | *slog.Logger | Nil allowed (falls back to slog.Default()) | Yes (parameter present, value may be nil) |

### For Log Methods

Same as Logger interface: `msg string, args ...any`.

## Outputs

### For Construction

| Field | Type | Description |
|-------|------|-------------|
| logger | Logger | SlogLogger instance satisfying the Logger interface |

No error — constructor cannot fail.

### For Log Methods

None. Fire-and-forget.

## Invariants

1. **Interface Compliance**: SlogLogger satisfies the Logger interface at compile time.
2. **Pass-Through**: No transformation is applied to `msg` or `args`. They are forwarded as-is to slog.
3. **Nil Safety**: Constructor handles nil `*slog.Logger` by falling back to `slog.Default()`.
4. **Concurrent Safety**: Inherits concurrent safety from `*slog.Logger` (which is documented as safe for concurrent use).
5. **No Constructor Bypass**: Must be constructed via `NewSlogLogger`. Direct struct literal is forbidden.

## Edge Cases

- Condition: `slogger` is nil at construction.
  Expected: Uses `slog.Default()` as the underlying logger. No panic or error.

- Condition: Underlying slog handler filters out Debug-level messages.
  Expected: `Debug` calls are still made to slog; slog handles the filtering. SlogLogger does not pre-filter.

## Related

- [Logger](./logger.md) — Interface that SlogLogger implements
- [NopLogger](./nop_logger.md) — Alternative no-op implementation for testing
