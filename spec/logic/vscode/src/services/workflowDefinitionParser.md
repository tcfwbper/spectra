# WorkflowDefinitionParser

## Overview

Provides a static method that reads and parses a workflow definition YAML file at `<projectRoot>/.spectra/workflows/<workflowName>.yaml`, extracts all `eventType` values from transitions whose `fromNode` equals the workflow's `entryNode`, deduplicates them, and returns the resulting array. This is a read-only, stateless, lightweight parser that does not perform full workflow validation (no graph integrity, reachability, or agent-role checks).

## Boundaries

- Owns: reading a single workflow YAML file from disk.
- Owns: validating presence of required top-level keys (`entryNode`, `transitions`).
- Owns: validating that each transition dict contains required keys (`fromNode`, `eventType`).
- Owns: filtering transitions by `fromNode === entryNode` and collecting unique `eventType` values.
- Owns: graceful degradation — returns empty array on file-not-found or validation failure after logging a warning.
- Delegates: project root resolution to the caller.
- Delegates: warning output to the injected logger.
- Must not: write, create, or delete any file or directory.
- Must not: perform cross-component structural validation (uniqueness, reachability, determinism, agent-role existence).
- Must not: reject unknown keys in the YAML — only validates presence of required keys.
- Must not: throw exceptions to the caller.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| Node.js `fs` (promises) | Filesystem reader | `readFile` | Must not write, mkdir, or unlink |
| Node.js `path` | Path utility | `join` | — |
| YAML parsing library | YAML deserialization | Parse YAML string to JS object | — |
| Logger (`{ warn(msg: string): void }`) | Warning output | `warn()` | Must not use for info/debug/error |

Construction constraint: WorkflowDefinitionParser is a class with a single static async method. No instantiation required.

## Behavior

1. Receives `projectRoot`, `workflowName`, and `logger` as parameters.
2. Constructs the file path: `<projectRoot>/.spectra/workflows/<workflowName>.yaml`.
3. Attempts to read the file.
4. If the file does not exist, calls `logger.warn` with a descriptive message and returns the failure result `{ entryNode: '', eventTypes: [] }`.
5. Parses the file content as YAML.
6. If YAML parsing fails (syntax error), calls `logger.warn` with a descriptive message and returns the failure result `{ entryNode: '', eventTypes: [] }`.
7. Validates that the parsed result contains a top-level `entryNode` key with a string value. If missing or not a string, calls `logger.warn` and returns the failure result `{ entryNode: '', eventTypes: [] }`.
8. Validates that the parsed result contains a top-level `transitions` key with an array value. If missing or not an array, calls `logger.warn` and returns the failure result `{ entryNode: '', eventTypes: [] }`.
9. Iterates over each element in `transitions`. For each element, validates that it is an object containing both `fromNode` (string) and `eventType` (string). If any element fails this check, calls `logger.warn` and returns the failure result `{ entryNode: '', eventTypes: [] }` (fail-fast).
10. Filters transitions where `fromNode` equals `entryNode`.
11. Collects the `eventType` values from matching transitions.
12. Deduplicates the collected `eventType` values.
13. Returns a `WorkflowParseResult` object containing both the `entryNode` string and the deduplicated `eventTypes` array.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |
| workflowName | string | Non-empty, corresponds to `<workflowName>.yaml` filename | Yes |
| logger | `{ warn(msg: string): void }` | Must provide a `warn` method | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<WorkflowParseResult>` | Object containing `entryNode` and deduplicated `eventTypes`. Returns `{ entryNode: '', eventTypes: [] }` on any failure. |

### WorkflowParseResult Type

| Field | Type | Description |
|---|---|---|
| entryNode | string | The workflow's entry node name. Empty string on failure. |
| eventTypes | string[] | Deduplicated array of `eventType` strings from transitions whose `fromNode` equals `entryNode`. Empty array on failure or no matches. |

## Invariants

- Must never throw an exception to the caller.
- Must always return a `WorkflowParseResult` object (never `undefined` or `null`).
- `entryNode` is always a string (empty on failure, non-empty on success).
- `eventTypes` is always an array (never `undefined` or `null`).
- Must be a static async method (no instance state required).
- Must not modify the filesystem.
- The returned `eventTypes` array contains no duplicate values.
- Unknown keys in the YAML are silently ignored at all levels (top-level and within transition dicts).
- Any single transition dict missing `fromNode` or `eventType` causes immediate failure (warning + empty array), not silent skip.

## Edge Cases

- Condition: `<projectRoot>/.spectra/workflows/<workflowName>.yaml` does not exist.
  Expected: Logs a warning via `logger.warn` and returns the failure result `{ entryNode: '', eventTypes: [] }`.

- Condition: The file exists but contains invalid YAML syntax.
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }`.

- Condition: The YAML is valid but missing the `entryNode` key.
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }`.

- Condition: The YAML is valid but missing the `transitions` key.
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }`.

- Condition: `entryNode` is present but its value is not a string (e.g., a number or object).
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }`.

- Condition: `transitions` is present but its value is not an array (e.g., a string or object).
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }`.

- Condition: A transition dict in the array is missing `fromNode`.
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }` (fail-fast).

- Condition: A transition dict in the array is missing `eventType`.
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }` (fail-fast).

- Condition: A transition dict has `fromNode` or `eventType` with a non-string value.
  Expected: Logs a warning and returns the failure result `{ entryNode: '', eventTypes: [] }` (fail-fast).

- Condition: No transition has `fromNode` equal to `entryNode`.
  Expected: Returns an empty array without logging a warning (valid scenario, just no matches).

- Condition: Multiple transitions have the same `fromNode === entryNode` and the same `eventType`.
  Expected: Returns a deduplicated array containing that `eventType` only once.

- Condition: The YAML contains additional keys beyond `entryNode` and `transitions` (e.g., `nodes`, `description`).
  Expected: Extra keys are silently ignored; parsing proceeds normally.

- Condition: A transition dict contains additional keys beyond `fromNode` and `eventType` (e.g., `toNode`).
  Expected: Extra keys are silently ignored; only `fromNode` and `eventType` are validated and used.

- Condition: `transitions` is an empty array.
  Expected: Returns an empty array without logging a warning (valid structure, just no transitions).

## Related

- [WorkflowScanner](./workflowScanner.md) — Lists available workflow names from the same directory this parser reads from.
- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
