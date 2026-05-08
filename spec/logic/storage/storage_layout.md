# StorageLayout

## Overview

StorageLayout defines the directory structure and path constants for `.spectra/` persistent storage. It provides package-level constants for relative paths and exported functions that compose absolute file paths by joining a project root with the appropriate relative path. StorageLayout is pure computation — it performs no I/O and does not validate whether paths exist on the filesystem.

## Boundaries

- Owns: definition of all `.spectra/` relative path constants and path composition logic.
- Owns: platform-correct path separator handling via `filepath.Join`.
- Delegates: project root discovery to SpectraFinder.
- Delegates: filesystem validation (existence, permissions) to consuming I/O modules.
- Must not: perform any filesystem operations (stat, read, write, create, delete).
- Must not: validate inputs (empty strings, UUID format, path existence).
- Must not: hold any state or require instantiation.

## Dependencies

None. Depends only on Go standard library (`path/filepath`).

## Behavior

1. Defines package-level string constants for all relative paths within `.spectra/`.
2. Provides exported functions that accept a project root (and optionally a session UUID, workflow name, or agent role) and return an absolute path string via `filepath.Join`.
3. Path composition functions do not validate inputs. Empty strings, invalid UUIDs, or nonexistent paths produce malformed output without error.
4. All functions are pure (same inputs always produce same output, no side effects).
5. Constants do not include trailing path separators.

## Inputs

### For Path Composition Functions

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| sessionUUID | string | UUID v4 format | Conditional (session-specific paths only) |
| workflowName | string | Non-empty, PascalCase identifier | Conditional (workflow-specific paths only) |
| agentRole | string | Non-empty, PascalCase identifier | Conditional (agent-specific paths only) |

## Outputs

### Path Constants (Relative Paths)

| Constant Name | Value | Description |
|---------------|-------|-------------|
| `SpectraDir` | `.spectra` | Base directory name |
| `SessionsDir` | `.spectra/sessions` | Sessions storage directory |
| `WorkflowsDir` | `.spectra/workflows` | Workflow definitions directory |
| `AgentsDir` | `.spectra/agents` | Agent definitions directory |
| `SessionMetadataFile` | `session.json` | Session metadata filename (relative to session dir) |
| `EventHistoryFile` | `events.jsonl` | Event history filename (relative to session dir) |
| `RuntimeSocketFile` | `runtime.sock` | Runtime socket filename (relative to session dir) |

### Path Composition Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `GetSpectraDir` | `(projectRoot string) string` | Absolute path to `.spectra/` |
| `GetSessionsDir` | `(projectRoot string) string` | Absolute path to `.spectra/sessions/` |
| `GetWorkflowsDir` | `(projectRoot string) string` | Absolute path to `.spectra/workflows/` |
| `GetAgentsDir` | `(projectRoot string) string` | Absolute path to `.spectra/agents/` |
| `GetSessionDir` | `(projectRoot, sessionUUID string) string` | Absolute path to `.spectra/sessions/<UUID>/` |
| `GetSessionMetadataPath` | `(projectRoot, sessionUUID string) string` | Absolute path to `.spectra/sessions/<UUID>/session.json` |
| `GetEventHistoryPath` | `(projectRoot, sessionUUID string) string` | Absolute path to `.spectra/sessions/<UUID>/events.jsonl` |
| `GetRuntimeSocketPath` | `(projectRoot, sessionUUID string) string` | Absolute path to `.spectra/sessions/<UUID>/runtime.sock` |
| `GetWorkflowPath` | `(projectRoot, workflowName string) string` | Absolute path to `.spectra/workflows/<WorkflowName>.yaml` |
| `GetAgentPath` | `(projectRoot, agentRole string) string` | Absolute path to `.spectra/agents/<AgentRole>.yaml` |

### Error Cases

None. Path composition functions do not return errors.

## Invariants

1. **No Trailing Separators**: Directory path constants must not end with a path separator.
2. **Platform-Agnostic Composition**: All path composition uses `filepath.Join` for correct separators on all platforms.
3. **Absolute Path Output**: When given an absolute `projectRoot`, all functions return absolute paths.
4. **UUID Passed As-Is**: Session-specific paths interpolate the UUID without modification (no lowercasing, no dash removal).
5. **Idempotent Composition**: Same inputs always produce the same output string.
6. **No I/O**: StorageLayout must never perform filesystem operations.
7. **No State**: All functions are stateless package-level functions. No struct instantiation required.

## Edge Cases

- Condition: `projectRoot` is a relative path (e.g., `./project`).
  Expected: Returns a relative path. Caller is responsible for providing absolute paths.

- Condition: `projectRoot` ends with a trailing slash (e.g., `/home/user/project/`).
  Expected: `filepath.Join` normalizes correctly without double slashes.

- Condition: `sessionUUID` is an empty string.
  Expected: Returns a malformed path (e.g., `.spectra/sessions//session.json`). No error.

- Condition: `workflowName` or `agentRole` contains path separators (e.g., `../malicious`).
  Expected: `filepath.Join` resolves the path as-is, potentially escaping `.spectra/`. Caller is responsible for input validation.

- Condition: Multiple goroutines call path composition functions concurrently.
  Expected: All calls succeed. Functions are stateless and safe for concurrent use.

- Condition: Runtime socket path exceeds Unix domain socket path limit (~108 characters).
  Expected: Path is returned without validation. Socket creation at the I/O layer will fail with a system error.

## Related

- [SpectraFinder](./spectra_finder.md) — Provides the `projectRoot` input
- [FileAccessor](./file_accessor.md) — Uses StorageLayout paths to access files
- [SessionDirectoryManager](./session_directory_manager.md) — Uses StorageLayout for session directory paths
