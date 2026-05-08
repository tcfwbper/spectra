# Exit Codes

## Overview

Defines the standard exit code constants shared by all CLI commands (`spectra-agent` and future `cmd/` packages). Each constant represents a distinct failure category. This module is purely declarative — it defines constants only.

## Boundaries

- Owns: numeric exit code constant definitions and their semantic categories.
- Must not: perform any I/O or contain logic beyond constant declarations.
- Must not: define error messages (that is the responsibility of the error formatter or individual commands).

## Dependencies

None.

## Behavior

1. Declares package-level integer constants for CLI exit codes.
2. Each constant is exported and documented with its semantic meaning.

## Inputs

None (constants only).

## Outputs

| Constant | Value | Semantic Meaning |
|----------|-------|------------------|
| `ExitSuccess` | 0 | Operation completed successfully |
| `ExitInvocationError` | 1 | Missing required argument/flag, invalid flag value, invalid JSON, `.spectra` directory not found, unknown subcommand |
| `ExitTransportError` | 2 | Socket file not found, connection refused, connection timeout, I/O error during send/receive |
| `ExitRuntimeError` | 3 | Runtime responded with error status, malformed response JSON, response missing required fields |

## Invariants

1. **Fixed Values**: Exit code values must not change. They form a public contract between spectra-agent and its callers.
2. **No Overlap**: Each exit code maps to exactly one failure category.

## Edge Cases

None (constants only).

## Related

- [SendAndHandle](./send_and_handle.md) — uses exit codes when classifying responses
- [SocketClient](./socket_client.md) — returns exit codes from transport operations
