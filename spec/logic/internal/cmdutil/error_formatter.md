# ErrorFormatter

## Overview

Provides helper functions for formatting CLI error and warning messages to stderr. Ensures all user-facing error output follows a consistent format across all commands. Does not perform I/O directly — returns formatted strings that the caller prints.

## Boundaries

- Owns: error message string formatting (prefix rules, newline handling).
- Must not: write to stderr or stdout directly (caller is responsible for I/O).
- Must not: determine exit codes (that is the caller's responsibility using exit code constants).

## Dependencies

None. Uses only Go standard library (`fmt`).

## Behavior

1. `FormatError(msg string) string` — returns `"Error: <msg>"`.
2. `FormatWarning(msg string) string` — returns `"Warning: <msg>"`.

## Inputs

| Function | Parameter | Type | Constraints | Required |
|----------|-----------|------|-------------|----------|
| `FormatError` | msg | string | Non-empty string | Yes |
| `FormatWarning` | msg | string | Non-empty string | Yes |

## Outputs

| Function | Return Type | Format |
|----------|-------------|--------|
| `FormatError` | string | `"Error: <msg>"` |
| `FormatWarning` | string | `"Warning: <msg>"` |

## Invariants

1. **Error Prefix**: All error messages must be prefixed with exactly `"Error: "`.
2. **Warning Prefix**: All warning messages must be prefixed with exactly `"Warning: "`.
3. **No I/O**: Functions must not perform any I/O. They are pure string transformations.

## Edge Cases

- Condition: `msg` is an empty string.
  Expected: Returns `"Error: "` or `"Warning: "` (prefix only). Caller is responsible for ensuring non-empty input.

- Condition: `msg` already contains an `"Error: "` prefix.
  Expected: Returns `"Error: Error: ..."` (no deduplication). Caller is responsible for not double-prefixing.

## Related

- [ExitCodes](./exit_codes.md) — defines which exit code accompanies error output
- [SendAndHandle](./send_and_handle.md) — uses FormatError when building error output
