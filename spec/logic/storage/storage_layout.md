# StorageLayout

## Overview

StorageLayout defines the directory structure and path constants for the `.spectra/` persistent storage. It provides methods to compose absolute file paths by combining the project root (from SpectraFinder) with relative paths within `.spectra/`. It does not perform I/O operations or validate file existence.

## Behavior

1. StorageLayout defines relative path constants for all storage locations within `.spectra/`.
2. Each path constant includes the `.spectra/` prefix and follows the directory structure defined in the architecture.
3. StorageLayout provides path composition methods that accept a project root directory and return absolute paths.
4. Path composition methods join the project root with the relative path constant using platform-appropriate path separators.
5. Path composition methods do not validate whether the resulting path exists on the filesystem.
6. Path composition methods for session-specific files accept a session UUID parameter and interpolate it into the path.
7. Session-specific files include session metadata (`session.json`), event history (`events.jsonl`), and runtime socket (`runtime.sock`).
8. The runtime socket file (`runtime.sock`) is a Unix domain socket created by the workflow runtime for each active session, enabling spectra-agent to communicate with the runtime process.
9. On Windows, the runtime may use named pipes instead of Unix domain sockets. The path composition methods remain the same, but the underlying transport differs.
10. All relative paths are defined as constants to ensure consistency across the codebase.
11. StorageLayout does not create, read, modify, or delete files or directories.

## Inputs

### For Path Composition Methods

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` (e.g., `/home/user/project`) | Yes |
| SessionUUID | string (UUID) | Valid UUID v4 format | Conditional (required only for session-specific paths) |
| WorkflowName | string | Non-empty, PascalCase identifier | Conditional (required only for workflow-specific paths) |
| AgentRole | string | Non-empty, PascalCase identifier | Conditional (required only for agent-specific paths) |

## Outputs

### Path Constants (Relative Paths)

These constants are defined at the package level:

| Constant Name | Type | Value | Description |
|---------------|------|-------|-------------|
| `SpectraDir` | string | `.spectra` | Base directory name |
| `SessionsDir` | string | `.spectra/sessions` | Sessions storage directory |
| `WorkflowsDir` | string | `.spectra/workflows` | Workflow definitions directory |
| `AgentsDir` | string | `.spectra/agents` | Agent definitions directory |
| `SessionMetadataFile` | string | `session.json` | Session metadata filename (relative to session directory) |
| `EventHistoryFile` | string | `events.jsonl` | Event history filename (relative to session directory) |
| `RuntimeSocketFile` | string | `runtime.sock` | Unix domain socket filename (relative to session directory) for spectra-agent communication |

### Path Composition Methods

Each method returns an absolute path string:

| Method | Returns | Description |
|--------|---------|-------------|
| `GetSpectraDir(ProjectRoot)` | string | Absolute path to `.spectra/` directory |
| `GetSessionsDir(ProjectRoot)` | string | Absolute path to `sessions/` directory |
| `GetWorkflowsDir(ProjectRoot)` | string | Absolute path to `workflows/` directory |
| `GetAgentsDir(ProjectRoot)` | string | Absolute path to `agents/` directory |
| `GetSessionDir(ProjectRoot, SessionUUID)` | string | Absolute path to specific session directory |
| `GetSessionMetadataPath(ProjectRoot, SessionUUID)` | string | Absolute path to `session.json` for a specific session |
| `GetEventHistoryPath(ProjectRoot, SessionUUID)` | string | Absolute path to `events.jsonl` for a specific session |
| `GetRuntimeSocketPath(ProjectRoot, SessionUUID)` | string | Absolute path to `runtime.sock` for a specific session |
| `GetWorkflowPath(ProjectRoot, WorkflowName)` | string | Absolute path to workflow YAML file (`.spectra/workflows/<WorkflowName>.yaml`) |
| `GetAgentPath(ProjectRoot, AgentRole)` | string | Absolute path to agent YAML file (`.spectra/agents/<AgentRole>.yaml`) |

### Error Cases

Path composition methods do not return errors. Invalid inputs (e.g., empty strings, invalid UUIDs) result in malformed paths that will fail at the I/O layer when used.

## Invariants

1. **Relative Path Prefix**: All relative path constants must start with `.spectra/` (except base filenames like `session.json`).

2. **No Trailing Separators**: Directory path constants must not end with a path separator (`/` or `\`).

3. **Platform-Agnostic Composition**: Path composition methods must use the Go standard library's `filepath.Join` to ensure correct path separators on all platforms.

4. **Absolute Path Output**: All path composition methods must return absolute paths when given an absolute `ProjectRoot`.

5. **UUID Interpolation Format**: Session-specific paths must interpolate the UUID as-is without modification (no lowercasing, no dashes removed).

6. **No I/O Operations**: StorageLayout must never perform file or directory operations (stat, read, write, create, delete).

7. **Idempotent Composition**: Calling the same path composition method with the same inputs must always return the same path string.

## Edge Cases

- **Condition**: `ProjectRoot` is a relative path (e.g., `./project`).
  **Expected**: Path composition methods return a path that is relative, not absolute. The caller is responsible for providing an absolute `ProjectRoot`.

- **Condition**: `ProjectRoot` ends with a trailing slash (e.g., `/home/user/project/`).
  **Expected**: Path composition methods correctly join paths without introducing double slashes.

- **Condition**: `SessionUUID` contains uppercase letters.
  **Expected**: The UUID is used as-is without case conversion. Path composition returns a path with the UUID exactly as provided.

- **Condition**: `SessionUUID` is an empty string.
  **Expected**: Path composition methods return a malformed path (e.g., `.spectra/sessions//session.json`). The method does not validate the UUID.

- **Condition**: `WorkflowName` or `AgentRole` contains path separators (e.g., `../malicious/workflow`).
  **Expected**: Path composition methods join the name as-is, potentially creating a path outside `.spectra/`. The caller (WorkflowDefinitionLoader or AgentDefinitionLoader) is responsible for validating input format (PascalCase).

- **Condition**: `ProjectRoot` does not actually contain a `.spectra/` directory.
  **Expected**: Path composition methods return the path as if `.spectra/` exists. No validation is performed.

- **Condition**: Multiple goroutines call path composition methods concurrently.
  **Expected**: All calls succeed without data races. Path constants are immutable and composition is stateless.

- **Condition**: Caller requests the runtime socket path for a session that is not currently running.
  **Expected**: Path composition methods return the path without validation. The socket file may not exist if the session is not active.

- **Condition**: Runtime socket path exceeds platform limits (e.g., Unix domain socket path limit of ~108 characters on some systems).
  **Expected**: Path composition methods return the full path without validation. Socket creation at the I/O layer will fail with an appropriate system error. Consider using shorter project root paths or session UUID abbreviations if this becomes an issue.

- **Condition**: On Windows, runtime uses named pipes instead of Unix domain sockets.
  **Expected**: Path composition methods return the same path format. The I/O layer is responsible for adapting to the platform-specific transport mechanism.

## Related

- [SpectraFinder](./spectra_finder.md) - Provides the `ProjectRoot` input for path composition
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Defines the `.spectra/` directory structure
- [FileAccessor](./file_accessor.md) - Uses StorageLayout paths to access files
