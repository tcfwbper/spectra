# SessionScanner

## Overview

Provides a static method that scans all session directories under `<projectRoot>/.spectra/sessions/` and returns a sorted array of session summary objects. Each session is read from its `session.json` file and must contain all required fields to be included. This is a read-only, stateless operation that performs no mutation.

## Boundaries

- Owns: discovering session directories, reading and validating `session.json` files, assembling and sorting session summary objects.
- Delegates: project root resolution to the caller.
- Delegates: warning output to the injected logger.
- Must not: write, create, or delete any file or directory.
- Must not: validate the semantic correctness of field values (e.g., status enum membership at runtime, UUID format).
- Must not: throw on missing directories or malformed JSON — logs a warning and skips or returns an empty array.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| Node.js `fs` (promises) | Filesystem reader | `readdir`, `readFile`, `stat`/`access` | Must not write, mkdir, or unlink |
| Node.js `path` | Path utility | `join` | — |
| Logger (`{ warn(msg: string): void }`) | Warning output | `warn()` | Must not use for info/debug/error |

Construction constraint: SessionScanner is a class with a single static async method. No instantiation required.

## Behavior

1. Receives `projectRoot` and `logger` as parameters.
2. Constructs the sessions directory path: `<projectRoot>/.spectra/sessions`.
3. Checks whether the sessions directory exists.
4. If the directory does not exist, calls `logger.warn` with a descriptive message and returns an empty array.
5. Reads the directory listing to discover all session subdirectories.
6. For each session subdirectory, attempts to read `<sessionDir>/session.json`.
7. If the subdirectory does not exist, the `session.json` file does not exist, or the file content is not valid JSON, calls `logger.warn` with a descriptive message and skips this session.
8. After successful JSON parse, validates that all six required keys are present: `workflowName`, `createdAt`, `id`, `pid`, `status`, `currentState`. If any key is missing, calls `logger.warn` and skips this session.
9. Constructs a session summary object from the six required fields.
10. After processing all session directories, sorts the resulting array by `createdAt` in descending order (most recent first).
11. Returns the sorted array.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |
| logger | `{ warn(msg: string): void }` | Must provide a `warn` method | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<SessionSummary[]>` | Sorted array of session summaries (descending by `createdAt`). Empty array if directory does not exist or no valid sessions found. |

### SessionSummary Type

| Field | Type | Description |
|---|---|---|
| id | string | Session identifier |
| workflowName | string | Name of the workflow being executed |
| createdAt | number | POSIX timestamp in seconds |
| pid | number | OS process ID |
| status | `'initializing' \| 'running' \| 'completed' \| 'failed'` | Current execution status |
| currentState | string | Active node in the workflow state machine |

## Invariants

- Must never throw an exception to the caller under normal filesystem conditions (missing directory, malformed JSON, missing fields).
- Must always return an array (never `undefined` or `null`).
- Must be a static async method (no instance state required).
- The returned array must be sorted by `createdAt` descending. If two sessions share the same `createdAt`, their relative order is unspecified.
- Each returned object must contain exactly the six required fields — no extra keys from the JSON are included.
- A session with any missing required field is never included in the result.

## Edge Cases

- Condition: `<projectRoot>/.spectra/sessions` directory does not exist.
  Expected: Logs a warning via `logger.warn` and returns an empty array.

- Condition: The sessions directory exists but is empty.
  Expected: Returns an empty array without logging a warning.

- Condition: A session subdirectory exists but contains no `session.json` file.
  Expected: Logs a warning and skips this session.

- Condition: A `session.json` file contains valid JSON but is missing one or more required keys.
  Expected: Logs a warning and skips this session.

- Condition: A `session.json` file contains extra keys beyond the six required ones.
  Expected: Extra keys are ignored; only the six required fields are extracted.

- Condition: A `session.json` file contains malformed JSON (not parseable).
  Expected: Logs a warning and skips this session.

- Condition: Multiple sessions have the same `createdAt` value.
  Expected: Both are included; their relative order is unspecified (stable sort not required).

- Condition: The sessions directory contains regular files (not subdirectories) at the top level.
  Expected: These entries are skipped (only subdirectories are traversed).

## Related

- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
- Storage layout reference: `.spectra/sessions/<UUID>/session.json` stores session metadata.
