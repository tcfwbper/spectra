/**
 * ProjectRootResolver — resolves the effective project root path for the
 * Spectra extension by combining the VS Code workspace folder with the
 * optional `spectra.projectRoot` configuration value.
 *
 * Logic spec: spec/logic/vscode/src/services/projectRootResolver.md
 */
import * as path from "path";

/**
 * Minimal interface for the workspace dependency consumed by ProjectRootResolver.
 * In production this is satisfied by `vscode.workspace`; in tests a stub is passed.
 */
export interface WorkspaceProvider {
  workspaceFolders:
    | ReadonlyArray<{ uri: { fsPath: string } }>
    | undefined;
  getConfiguration(section: string): { get<T>(key: string): T | undefined };
}

/**
 * Provides a static method that resolves the effective project root path.
 *
 * - Owns: Computing the resolved project root path from workspace state and extension configuration.
 * - Must not: Perform any file-system I/O (no existence checks, no directory creation).
 * - Must not: Mutate or persist any state.
 * - Must not: Handle multi-root workspace selection — always uses the first workspace folder.
 */
export class ProjectRootResolver {
  /**
   * Resolves the project root path.
   *
   * @param workspace - The workspace provider (injectable for testing).
   * @returns The resolved absolute path, or `undefined` if no workspace is open.
   */
  static resolve(workspace: WorkspaceProvider): string | undefined {
    // Step 1: Read the first workspace folder path.
    const folders = workspace.workspaceFolders;
    if (!folders || folders.length === 0) {
      return undefined;
    }

    const workspacePath = folders[0].uri.fsPath;

    // Step 2: If workspace path is falsy, return undefined.
    if (!workspacePath) {
      return undefined;
    }

    // Step 3: Read the spectra.projectRoot configuration value.
    const config = workspace.getConfiguration("spectra");
    const projectRoot = config.get<string>("projectRoot");

    // Step 4: If configuration value is falsy, return workspace path as-is.
    if (!projectRoot) {
      return workspacePath;
    }

    // Step 5: Join workspace path with configuration value.
    return path.join(workspacePath, projectRoot);
  }
}
