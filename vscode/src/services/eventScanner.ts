/**
 * EventScanner — reads and parses the events.jsonl file for a given session,
 * returning an array of event summary objects.
 *
 * Logic spec: spec/logic/vscode/src/services/eventScanner.md
 */

/**
 * Represents a single parsed event from the events.jsonl file.
 */
export interface EventSummary {
  Type: string;
  EmittedBy: string;
  Message: string;
}

/**
 * Minimal filesystem interface required by EventScanner.
 * In production this is satisfied by Node.js fs/promises; in tests a stub is passed.
 */
export interface EventScannerFs {
  access(path: string): Promise<void>;
  readFile(path: string, encoding: string): Promise<string>;
}

/**
 * Provides a static method that reads and parses the events.jsonl file for a
 * given session, returning an array of event objects each containing Type,
 * EmittedBy, and Message.
 *
 * - Owns: reading the events.jsonl file, line-by-line JSON parsing, extracting
 *   the three required keys from each line.
 * - Must not: write, create, or delete any file or directory.
 * - Must not: throw on missing file or malformed lines.
 */
export class EventScanner {
  /**
   * Scans the events.jsonl file for the given session and returns parsed event summaries.
   *
   * @param projectRoot - Absolute path to the project root.
   * @param sessionId - Session identifier (directory name).
   * @param logger - Logger with a warn method for reporting issues.
   * @param fs - Injectable filesystem dependency (defaults to Node.js fs/promises).
   * @returns Array of event summaries in file order.
   */
  static async scanEvents(
    projectRoot: string,
    sessionId: string,
    logger: { warn(msg: string): void },
    fs?: EventScannerFs,
  ): Promise<EventSummary[]> {
    const fsImpl = fs ?? (await import("fs/promises"));
    const path = `${projectRoot}/.spectra/sessions/${sessionId}/events.jsonl`;

    // Check file existence
    try {
      await fsImpl.access(path);
    } catch {
      logger.warn(`Events file not found: ${path}`);
      return [];
    }

    // Read file content
    let content: string;
    try {
      content = await fsImpl.readFile(path, "utf-8");
    } catch {
      logger.warn(`Cannot read events file: ${path}`);
      return [];
    }

    // Parse lines
    const lines = content.split("\n");
    const results: EventSummary[] = [];

    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed === "") {
        continue;
      }

      let parsed: Record<string, unknown>;
      try {
        parsed = JSON.parse(trimmed);
      } catch {
        logger.warn(`Invalid JSON in events file: ${trimmed}`);
        continue;
      }

      if (
        parsed.Type === undefined ||
        parsed.EmittedBy === undefined ||
        parsed.Message === undefined
      ) {
        logger.warn(`Missing required key in event line`);
        continue;
      }

      results.push({
        Type: parsed.Type as string,
        EmittedBy: parsed.EmittedBy as string,
        Message: parsed.Message as string,
      });
    }

    return results;
  }
}
