# NopLogger — No-Op Implementation

## Overview

NopLogger is a concrete implementation of the Logger interface that discards all log output. It is intended for use in tests and situations where logging is not desired. All methods are no-ops that return immediately.

## Boundaries

- Owns: providing a valid Logger implementation that produces no output.
- Must not: write to any output destination (stdout, stderr, files, buffers).
- Must not: allocate memory or perform any work beyond returning.

## Dependencies

None.

Construction constraint: Must be constructed via `NewNopLogger() Logger`. Direct struct literal is forbidden.

## Behavior

1. `NewNopLogger()` returns a NopLogger instance satisfying the Logger interface.
2. `Debug(msg, args...)` does nothing and returns immediately.
3. `Info(msg, args...)` does nothing and returns immediately.
4. `Warn(msg, args...)` does nothing and returns immediately.
5. `Error(msg, args...)` does nothing and returns immediately.

## Inputs

### For Construction

None.

### For Log Methods

Same as Logger interface: `msg string, args ...any`. All inputs are ignored.

## Outputs

### For Construction

| Field | Type | Description |
|-------|------|-------------|
| logger | Logger | NopLogger instance satisfying the Logger interface |

No error — constructor cannot fail.

### For Log Methods

None.

## Invariants

1. **Interface Compliance**: NopLogger satisfies the Logger interface at compile time.
2. **Zero Side Effects**: No method produces any observable side effect (no I/O, no allocation beyond the receiver).
3. **Concurrent Safety**: Safe for concurrent use (no internal state is read or written).
4. **No Constructor Bypass**: Must be constructed via `NewNopLogger`. Direct struct literal is forbidden.

## Edge Cases

- Condition: Any combination of `msg` and `args` values.
  Expected: All inputs are silently discarded. No panic regardless of input.

## Related

- [Logger](./logger.md) — Interface that NopLogger implements
- [SlogLogger](./slog_logger.md) — Alternative implementation for production use
