# Signal Exit Codes

## Overview

Defines exit code constants for OS signal-based termination, following the standard Unix convention of 128 + signal number. These constants are used by `spectra run` to map signal-terminated sessions to appropriate process exit codes.

## Boundaries

- Owns: numeric exit code constant definitions for signal-based termination.
- Must not: perform any I/O or contain logic beyond constant declarations.
- Must not: define error messages (that is the responsibility of the calling command).
- Must not: define non-signal exit codes (those belong in exit_codes.md).

## Dependencies

None.

## Behavior

1. Declares package-level integer constants for signal-based exit codes.
2. Each constant is exported and documented with its semantic meaning.

## Inputs

None (constants only).

## Outputs

| Constant | Value | Semantic Meaning |
|----------|-------|------------------|
| `ExitSignalINT` | 130 | Process terminated by SIGINT (128 + 2). Standard Unix convention for Ctrl+C. |
| `ExitSignalTERM` | 143 | Process terminated by SIGTERM (128 + 15). Standard Unix convention for kill default signal. |

## Invariants

1. **Fixed Values**: Exit code values must not change. They form a public contract between spectra CLI and its callers (scripts, CI, etc.).
2. **Unix Convention**: Values follow 128 + signal number (SIGINT=2, SIGTERM=15).
3. **No Overlap**: Must not overlap with exit codes defined in exit_codes.md (0, 1, 2, 3).

## Edge Cases

None (constants only).

## Related

- [ExitCodes](./exit_codes.md) - Non-signal exit code constants (0, 1, 2, 3)
- [spectra run](../cmd/spectra/run.md) - Primary consumer of signal exit codes
