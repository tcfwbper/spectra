# WorkflowScanner

## Overview

Provides a static method that scans the `<projectRoot>/.spectra/workflows/` directory and returns an array of workflow names (filenames without the `.yaml` extension). This is a read-only, stateless operation that performs no validation of file contents.

## Boundaries

- Owns: listing files in the workflows directory and extracting base names from `*.yaml` files.
- Delegates: project root resolution to the caller.
- Delegates: warning output to the injected logger.
- Must not: read or parse the content of any YAML file.
- Must not: write, create, or delete any file or directory.
- Must not: throw on missing directory — returns an empty array after logging a warning.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| Node.js `fs` (promises) | Filesystem reader | `readdir` | Must not write, mkdir, or unlink |
| Node.js `path` | Path utility | `join`, `extname`, `basename` | — |
| Logger (`{ warn(msg: string): void }`) | Warning output | `warn()` | Must not use for info/debug/error |

Construction constraint: WorkflowScanner is a class with a single static async method. No instantiation required.

## Behavior

1. Receives `projectRoot` and `logger` as parameters.
2. Constructs the workflows directory path: `<projectRoot>/.spectra/workflows`.
3. Checks whether the workflows directory exists.
4. If the directory does not exist, calls `logger.warn` with a descriptive message and returns an empty array.
5. Reads the directory listing.
6. Filters entries to include only files whose extension is `.yaml`.
7. Extracts the base name (filename without `.yaml` extension) from each matching file.
8. Returns the resulting array of workflow name strings.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| projectRoot | string | Non-empty, absolute path | Yes |
| logger | `{ warn(msg: string): void }` | Must provide a `warn` method | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `Promise<string[]>` | Array of workflow names (filenames without `.yaml` extension). Empty array if directory does not exist or contains no `.yaml` files. |

## Invariants

- Must never throw an exception to the caller under normal filesystem conditions (missing directory, empty directory).
- Must always return an array (never `undefined` or `null`).
- Must be a static async method (no instance state required).
- Must not include the `.yaml` extension in any returned name.
- Must not include non-`.yaml` files in the result.
- Must not recurse into subdirectories.

## Edge Cases

- Condition: `<projectRoot>/.spectra/workflows` directory does not exist.
  Expected: Logs a warning via `logger.warn` and returns an empty array.

- Condition: The workflows directory exists but contains no `.yaml` files (only other file types or is empty).
  Expected: Returns an empty array without logging a warning.

- Condition: The workflows directory contains files like `my-workflow.yaml` and `README.md`.
  Expected: Returns `['my-workflow']` — only `.yaml` files are included.

- Condition: A file is named `.yaml` (no base name before the extension).
  Expected: Returns an empty string element. The scanner does not validate name content.

- Condition: The directory contains subdirectories with `.yaml` in their name.
  Expected: Only regular files are included; directories are excluded from the result.

## Related

- [ProjectRootResolver](./projectRootResolver.md) — Caller uses this to obtain the `projectRoot` value.
- Storage layout reference: `.spectra/workflows/` contains `*.yaml` workflow definition files.
