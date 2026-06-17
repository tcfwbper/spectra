/**
 * ProjectRootResolver — resolves the effective project root path for the
 * Spectra extension by combining the VS Code workspace folder with the
 * optional `spectra.projectRoot` configuration value. Also checks whether
 * the project has been initialized (.spectra/ directory exists).
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
 * Minimal interface for the filesystem stat provider used by isInitialized.
 * In production this is satisfied by vscode.workspace.fs; in tests a stub is passed.
 */
export interface FsProvider {
  stat(uri: any): Promise<any>;
}

/**
 * Minimal interface for URI construction used by isInitialized.
 * In production this is satisfied by vscode.Uri.file; in tests a stub is passed.
 */
export interface UriFileProvider {
  (path: string): any;
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

  /**
   * Checks whether the project has been initialized by verifying
   * that the .spectra/ directory exists at the given project root.
   *
   * @param projectRoot - The resolved project root path.
   * @param fsProvider - Injectable filesystem stat provider (for testing).
   * @returns True if .spectra/ directory exists, false otherwise.
   */
  static async isInitialized(
    projectRoot: string,
    fsProvider: { stat: FsProvider["stat"]; uriFile: UriFileProvider },
  ): Promise<boolean> {
    // Step 6: Construct URI for .spectra directory
    const spectraPath = path.join(projectRoot, ".spectra");
    const uri = fsProvider.uriFile(spectraPath);

    try {
      // Step 7: Call stat
      await fsProvider.stat(uri);
      // Step 8: If stat succeeds, return true
      return true;
    } catch {
      // Step 9: If stat throws, return false
      return false;
    }
  }
}
