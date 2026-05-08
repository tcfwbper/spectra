# Storage Constants

## Overview

Defines package-level constants shared across the storage package. These constants establish system-wide limits that multiple storage components reference.

## Boundaries

- Owns: definition of shared size limit constants for the storage package.
- Must not: define constants that belong to other packages (entities, runtime).
- Must not: contain any logic, functions, or types.

## Dependencies

None.

## Behavior

1. Defines `MaxPayloadSize` as a package-level constant representing the maximum allowed size (in bytes) for any single serialized payload persisted by storage components.
2. The value is `10 * 1024 * 1024` (10 MB).
3. Both EventStore and SessionMetadataStore reference this constant to enforce size limits before writing. Other storage components that enforce per-message size limits may also reference this constant.

## Inputs

None (constants only).

## Outputs

| Constant | Type | Value | Description |
|----------|------|-------|-------------|
| `MaxPayloadSize` | int | `10 * 1024 * 1024` (10,485,760) | Maximum serialized payload size in bytes |

## Invariants

1. **Single Source of Truth**: All storage components must reference `MaxPayloadSize` rather than hardcoding their own size limits.
2. **Immutable**: Constants are compile-time values and cannot be modified at runtime.

## Edge Cases

None.

## Related

- [EventStore](./event_store.md) — References MaxPayloadSize for per-event size enforcement
- [SessionMetadataStore](./session_metadata_store.md) — References MaxPayloadSize for per-write size enforcement
