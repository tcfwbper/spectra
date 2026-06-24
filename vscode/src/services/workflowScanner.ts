/**
 * WorkflowScanner — scans the workflows directory and returns an array of
 * workflow names (filenames without the .yaml extension).
 *
 * Logic spec: spec/logic/vscode/src/services/workflowScanner.md
 */

/**
 * Minimal directory entry interface for readdir results.
 */
export interface WfDirEntry {
  name: string;
  isDirectory(): boolean;
  isFile(): boolean;
}

/**
 * Minimal filesystem interface required by WorkflowScanner.
 * In production this is satisfied by Node.js fs/promises; in tests a stub is passed.
 */
export interface WorkflowScannerFs {
  readdir(
    path: string,
    options: { withFileTypes: true },
  ): Promise<WfDirEntry[]>;
}

/**
 * Provides a static method that scans the workflows directory and returns
 * an array of workflow names (filenames without the .yaml extension).
 *
 * - Owns: listing files in the workflows directory and extracting base names.
 * - Must not: read or parse the content of any YAML file.
 * - Must not: write, create, or delete any file or directory.
 * - Must not: throw on missing directory.
 */
export class WorkflowScanner {
  /**
   * Scans the workflows directory and returns workflow names.
   *
   * @param projectRoot - Absolute path to the project root.
   * @param logger - Logger with a warn method for reporting issues.
   * @param fs - Injectable filesystem dependency (defaults to Node.js fs/promises).
   * @returns Array of workflow names (without .yaml extension).
   */
  static async scanWorkflows(
    projectRoot: string,
    logger: { warn(msg: string): void },
    fs?: WorkflowScannerFs,
  ): Promise<string[]> {
    const fsImpl = fs ?? ((await import("fs/promises")) as unknown as WorkflowScannerFs);
    const workflowsDir = `${projectRoot}/.spectra/workflows`;

    // Read directory entries
    let entries: WfDirEntry[];
    try {
      entries = await fsImpl.readdir(workflowsDir, { withFileTypes: true });
    } catch {
      logger.warn(`Workflows directory not found: ${workflowsDir}`);
      return [];
    }

    // Filter to .yaml files only (regular files only)
    const results: string[] = [];
    for (const entry of entries) {
      if (!entry.isFile()) {
        continue;
      }

      const name = entry.name;
      if (!name.endsWith(".yaml")) {
        continue;
      }

      // Strip .yaml extension
      const baseName = name.slice(0, -5);
      results.push(baseName);
    }

    return results;
  }
}
