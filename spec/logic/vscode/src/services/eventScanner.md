# EventScanner

## Overview

Provides a static method that reads and parses the `events.jsonl` file for a given session, returning an array of event objects each containing `Type`, `EmittedBy`, and `Message`. This is a read-only, stateless operation that performs no mutation. Lines that fail JSON parsing are skipped; empty lines are silently ignored.

## Boundaries

- Owns: reading the events.jsonl file, line-by-line JSON parsing, extracting the three required keys from each line.
- Delegates: project root and session ID provision to the caller.
- Delegates: warning output to the injected logger.
- Must not: write, create, or delete any file or directory.
- Must not: validate the semantic correctness of field values (e.g., event type format, EmittedBy matching a known node).
- Must not: throw on missing file or malformed lines — logs a warning or silently skips as appropriate.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| Node.js `fs` (promises) | Filesystem reader | `readFile`, `access` | Must not write, mkdir, or unlink |
| Node.js `path` | Path utility | `join` | — |
| Logger (`{ warn(msg: string): void }`) | Warning output | `warn()` | Must not use for info/debug/error |

Construction constraint: EventScanner is a class with a single static async method. No instantiation required.

## Behavior

1. Receives `projectRoot`, `sessionId`, and `logger` as parameters.
2. Constructs the events file path: `<projectRoot>/.spectra/sessions/<sessionId>/events.jsonl`.
3. Checks whether the events file exists.
4. If the file does not exist, calls `logger.warn` with a descriptive message and returns an empty array.
5. Reads the entire file content as a UTF-8 string.
6. Splits the content into lines.
7. Iterates over lines from first to last:
   - If a line is empty after trimming whitespace, silently skips it (no warning).
   - Attempts to parse the line as JSON.
   - If JSON parsing fails, calls `logger.warn` with a descriptive message (including line context) and skips this line.
   - If JSON parsing succeeds, extracts the three required keys: `Type`, `EmittedBy`, `Message`.
   - If any required key is missing, calls `logger.warn` and skips this line.
   - Constructs an event summary object from the three fields.
8. Returns the accumulated array of event summary objects in file order (first line first).

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |
| sessionId | string | Non-empty, identifies a session directory | Yes |
| logger | `{ warn(msg: string): void }` | Must provide a `warn` method | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<EventSummary[]>` | Array of event summaries in file order. Empty array if file does not exist or contains no valid lines. |

### EventSummary Type

| Field | Type | Description |
|---|---|---|
| Type | string | Event type identifier |
| EmittedBy | string | Node name that emitted the event |
| Message | string | Event message content |

## Invariants

- Must never throw an exception to the caller under normal filesystem conditions (missing file, malformed lines).
- Must always return an array (never `undefined` or `null`).
- Must be a static async method (no instance state required).
- The returned array preserves file order (first line parsed successfully appears first in the array).
- Each returned object contains exactly the three required fields — no extra keys from the JSON line are included.
- Empty lines (whitespace-only after trim) are silently ignored without warning.
- Non-empty lines that fail JSON parse produce a warning and are skipped.
- Non-empty lines that parse successfully but lack a required key produce a warning and are skipped.

## Edge Cases

- Condition: `<projectRoot>/.spectra/sessions/<sessionId>/events.jsonl` does not exist.
  Expected: Logs a warning via `logger.warn` and returns an empty array.

- Condition: The file exists but is completely empty (zero bytes).
  Expected: Returns an empty array without logging a warning.

- Condition: The file contains only empty lines or whitespace-only lines.
  Expected: Returns an empty array without logging a warning.

- Condition: A line contains valid JSON but is missing one or more of the three required keys.
  Expected: Logs a warning and skips this line.

- Condition: A line contains valid JSON with extra keys beyond the three required ones.
  Expected: Extra keys are ignored; only `Type`, `EmittedBy`, `Message` are extracted.

- Condition: A line is not valid JSON (e.g., truncated, malformed).
  Expected: Logs a warning and skips this line.

- Condition: The file ends with a trailing newline (producing one empty line at the end).
  Expected: The trailing empty line is silently ignored.

- Condition: The file contains a single valid line with no trailing newline.
  Expected: That line is parsed and the single-element array is returned.

## Related

- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
- [SessionScanner](./sessionScanner.md) — Caller may use SessionScanner to discover valid `sessionId` values.
- Storage layout reference: `.spectra/sessions/<UUID>/events.jsonl` stores the event history.
