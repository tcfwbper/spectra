# ProjectRootResolver

## Overview

Provides static methods that resolve the effective project root path for the Spectra extension and check whether the project has been initialized. Combines the VS Code workspace folder with the optional `spectra.projectRoot` configuration value. Does not create directories or mutate state.

## Boundaries

- Owns: Computing the resolved project root path from workspace state and extension configuration; checking whether a `.spectra/` directory exists at the resolved path.
- Delegates: Workspace folder lookup to the VS Code workspace API (`vscode.workspace.workspaceFolders`).
- Delegates: Configuration reading to the VS Code configuration API (`vscode.workspace.getConfiguration`).
- Delegates: Filesystem existence check to `vscode.workspace.fs.stat`.
- Must not: Create, write, or delete any file or directory.
- Must not: Mutate or persist any state.
- Must not: Handle multi-root workspace selection — always uses the first workspace folder.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | Workspace state provider | Read `workspaceFolders[0].uri.fsPath` | Must not modify workspace state |
| `vscode.workspace` | Configuration provider | Read `getConfiguration('spectra').get<string>('projectRoot')` | Must not write configuration |
| `vscode.workspace.fs` | Filesystem stat | `stat(uri)` for existence check | Must not create, write, or delete |
| `vscode.Uri` | URI construction | `vscode.Uri.file(path)` | — |
| Node.js `path` | Path utility | `path.join` | Must not use `path.resolve` or other methods that could override the workspace base |

## Behavior

1. Reads the first workspace folder path from `vscode.workspace.workspaceFolders[0].uri.fsPath`.
2. If the workspace value is falsy (`undefined`, `null`, or empty string), returns `undefined`.
3. Reads the `spectra.projectRoot` configuration value via `vscode.workspace.getConfiguration('spectra').get<string>('projectRoot')`. The declared default for this setting is `undefined`.
4. If the configuration value is falsy (`undefined`, `null`, or empty string), returns the workspace path as-is.
5. Joins the workspace path and the configuration value using `path.join(workspace, configValue)` and returns the result.

### isInitialized(projectRoot)

6. Constructs a URI for `path.join(projectRoot, '.spectra')`.
7. Calls `vscode.workspace.fs.stat(uri)`.
8. If `stat` succeeds (no error thrown), returns `true`.
9. If `stat` throws (file/directory does not exist), returns `false`.

## Inputs

| Field | Type | Constraints | Required? |
|---|---|---|---|
| *(none — resolve is a static method with no parameters)* | — | — | — |
| projectRoot (isInitialized) | string | Non-empty, absolute path | Yes |

The `resolve` method reads its inputs from VS Code APIs, not from caller-supplied arguments.

## Outputs

| Field | Type | Description |
|---|---|---|
| resolve result | `string \| undefined` | The resolved absolute path to the project root, or `undefined` if no workspace is open. |
| isInitialized result | `boolean` | `true` if `.spectra/` exists at the given project root, `false` otherwise. |

## Invariants

- `resolve()` must always return `undefined` when no workspace folder is available.
- The resolve result is always derived from the workspace folder via `path.join`; no independent root is introduced. Path traversal via `..` in the configuration value may produce a result outside the workspace subtree — this is acceptable as the method is not a security boundary.
- Must treat `undefined`, `null`, and empty string equivalently as "no value" for both workspace and configuration inputs.
- `resolve()` must be a pure computation with no side effects.
- `isInitialized()` must only perform a read (stat) — never create the directory.
- Both must be static methods (no instance state required).

## Edge Cases

- Condition: VS Code has no workspace open (`workspaceFolders` is `undefined` or empty array).
  Expected: Returns `undefined`.

- Condition: `spectra.projectRoot` is explicitly set to an empty string by the user.
  Expected: Treats it as no value; returns the workspace path.

- Condition: `spectra.projectRoot` contains path segments with `..` (e.g., `../sibling`).
  Expected: Joins as-is without validation; `path.join` normalizes the segments.

- Condition: `workspaceFolders` exists but the first entry has an empty `fsPath`.
  Expected: Treats it as falsy; returns `undefined`.

## Related

- VS Code extension structure: see `CONVENTIONS.md` → TypeScript (spectra-vscode extension) section.
- **Security note for future authors:** This resolver does not guarantee workspace containment. If a future caller requires that the resolved path stays within the workspace subtree, that validation must be owned by the caller, not by this unit.
