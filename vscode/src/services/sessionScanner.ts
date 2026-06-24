/**
 * SessionScanner — scans all session directories and returns a sorted array
 * of session summary objects.
 *
 * Logic spec: spec/logic/vscode/src/services/sessionScanner.md
 */

/**
 * Represents a single session summary parsed from session.json.
 */
export interface SessionSummary {
  id: string;
  workflowName: string;
  createdAt: number;
  pid: number;
  status: string;
  currentState: string;
}

/**
 * Minimal directory entry interface for readdir results.
 */
export interface DirEntry {
  name: string;
  isDirectory(): boolean;
  isFile(): boolean;
}

/**
 * Minimal filesystem interface required by SessionScanner.
 * In production this is satisfied by Node.js fs/promises; in tests a stub is passed.
 */
export interface SessionScannerFs {
  access(path: string): Promise<void>;
  readdir(
    path: string,
    options: { withFileTypes: true },
  ): Promise<DirEntry[]>;
  readFile(path: string, encoding: string): Promise<string>;
}

/**
 * Provides a static method that scans all session directories under
 * <projectRoot>/.spectra/sessions/ and returns a sorted array of session
 * summary objects.
 *
 * - Owns: discovering session directories, reading and validating session.json
 *   files, assembling and sorting session summary objects.
 * - Must not: write, create, or delete any file or directory.
 * - Must not: throw on missing directories or malformed JSON.
 */
export class SessionScanner {
  /**
   * Scans all sessions and returns summaries sorted by createdAt descending.
   *
   * @param projectRoot - Absolute path to the project root.
   * @param logger - Logger with a warn method for reporting issues.
   * @param fs - Injectable filesystem dependency (defaults to Node.js fs/promises).
   * @returns Sorted array of session summaries (most recent first).
   */
  static async scanSessions(
    projectRoot: string,
    logger: { warn(msg: string): void },
    fs?: SessionScannerFs,
  ): Promise<SessionSummary[]> {
    const fsImpl = fs ?? ((await import("fs/promises")) as unknown as SessionScannerFs);
    const sessionsDir = `${projectRoot}/.spectra/sessions`;

    // Check if sessions directory exists
    try {
      await fsImpl.access(sessionsDir);
    } catch {
      logger.warn(`Sessions directory not found: ${sessionsDir}`);
      return [];
    }

    // Read directory entries
    let entries: DirEntry[];
    try {
      entries = await fsImpl.readdir(sessionsDir, { withFileTypes: true });
    } catch {
      logger.warn(`Cannot read sessions directory: ${sessionsDir}`);
      return [];
    }

    // Process each subdirectory
    const results: SessionSummary[] = [];
    const requiredKeys = [
      "id",
      "workflowName",
      "createdAt",
      "pid",
      "status",
      "currentState",
    ];

    for (const entry of entries) {
      if (!entry.isDirectory()) {
        continue;
      }

      const sessionJsonPath = `${sessionsDir}/${entry.name}/session.json`;

      let content: string;
      try {
        content = await fsImpl.readFile(sessionJsonPath, "utf-8");
      } catch {
        logger.warn(`Cannot read session.json: ${sessionJsonPath}`);
        continue;
      }

      let parsed: Record<string, unknown>;
      try {
        parsed = JSON.parse(content);
      } catch {
        logger.warn(`Invalid JSON in session.json: ${sessionJsonPath}`);
        continue;
      }

      // Validate required keys
      const missing = requiredKeys.filter((k) => parsed[k] === undefined);
      if (missing.length > 0) {
        logger.warn(
          `Missing required keys in session.json: ${missing.join(", ")}`,
        );
        continue;
      }

      results.push({
        id: parsed.id as string,
        workflowName: parsed.workflowName as string,
        createdAt: parsed.createdAt as number,
        pid: parsed.pid as number,
        status: parsed.status as string,
        currentState: parsed.currentState as string,
      });
    }

    // Sort by createdAt descending
    results.sort((a, b) => b.createdAt - a.createdAt);

    return results;
  }
}
