# ProjectRootResolver

## Overview

Provides a static method that resolves the effective project root path for the Spectra extension. It combines the VS Code workspace folder with the optional `spectra.projectRoot` configuration value. Does not perform file-system validation or create directories.

## Boundaries

- Owns: Computing the resolved project root path from workspace state and extension configuration.
- Delegates: Workspace folder lookup to the VS Code workspace API (`vscode.workspace.workspaceFolders`).
- Delegates: Configuration reading to the VS Code configuration API (`vscode.workspace.getConfiguration`).
- Must not: Perform any file-system I/O (no existence checks, no directory creation).
- Must not: Mutate or persist any state.
- Must not: Handle multi-root workspace selection — always uses the first workspace folder.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.workspace` | Workspace state provider | Read `workspaceFolders[0].uri.fsPath` | Must not modify workspace state |
| `vscode.workspace` | Configuration provider | Read `getConfiguration('spectra').get<string>('projectRoot')` | Must not write configuration |
| Node.js `path` | Path utility | `path.join` | Must not use `path.resolve` or other methods that could override the workspace base |

## Behavior

1. Reads the first workspace folder path from `vscode.workspace.workspaceFolders[0].uri.fsPath`.
2. If the workspace value is falsy (`undefined`, `null`, or empty string), returns `undefined`.
3. Reads the `spectra.projectRoot` configuration value via `vscode.workspace.getConfiguration('spectra').get<string>('projectRoot')`. The declared default for this setting is `undefined`.
4. If the configuration value is falsy (`undefined`, `null`, or empty string), returns the workspace path as-is.
5. Joins the workspace path and the configuration value using `path.join(workspace, configValue)` and returns the result.

## Inputs

| Field | Type | Constraints | Required? |
|---|---|---|---|
| *(none — static method with no parameters)* | — | — | — |

The method reads its inputs from VS Code APIs, not from caller-supplied arguments.

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `string \| undefined` | The resolved absolute path to the project root, or `undefined` if no workspace is open. |

## Invariants

- Must always return `undefined` when no workspace folder is available.
- The result is always derived from the workspace folder via `path.join`; no independent root is introduced. Path traversal via `..` in the configuration value may produce a result outside the workspace subtree — this is acceptable as the method is not a security boundary.
- Must treat `undefined`, `null`, and empty string equivalently as "no value" for both workspace and configuration inputs.
- Must be a pure computation with no side effects.
- Must be a static method (no instance state required).

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
