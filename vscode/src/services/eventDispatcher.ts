/**
 * EventDispatcher — dispatches events to the Spectra runtime by spawning
 * an external `spectra-agent event emit` CLI process.
 *
 * Logic spec: spec/logic/vscode/src/services/eventDispatcher.md
 */

/**
 * Logger interface required by EventDispatcher.
 */
export interface EventDispatcherLogger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * Injectable dependencies for EventDispatcher (testability seam).
 */
export interface EventDispatcherDeps {
  getConfiguration: () => { get(key: string): string | undefined };
  execFile: (binary: string, args: string[], options?: any) => any;
}

/**
 * Dispatches events to the Spectra runtime by spawning an external CLI process.
 * This is a fire-and-forget operation on the happy path — the method does not
 * wait for the child process to exit.
 */
export class EventDispatcher {
  /**
   * Dispatches an event by spawning the spectra-agent CLI.
   *
   * @param eventType - The event type to dispatch
   * @param sessionId - The session identifier
   * @param message - The event message
   * @param projectRoot - The project root directory (used as cwd)
   * @param logger - Logger for diagnostic output
   * @param deps - Injectable dependencies (for testing)
   */
  static async dispatch(
    eventType: string,
    sessionId: string,
    message: string,
    projectRoot: string,
    logger: EventDispatcherLogger,
    deps: EventDispatcherDeps,
  ): Promise<void> {
    const config = deps.getConfiguration();
    const binaryPath = config.get("agentBinaryPath") || "spectra-agent";

    const args = [
      "event",
      "emit",
      eventType,
      "--session-id",
      sessionId,
      "--message",
      message,
    ];

    const child = deps.execFile(binaryPath, args, { cwd: projectRoot });

    logger.info(
      `Dispatching event "${eventType}" for session ${sessionId}`,
    );

    // Register exit handler for diagnostic logging (non-zero exit warning)
    child.on("exit", (code: number | null) => {
      if (code !== null && code !== 0) {
        logger.warn(`spectra-agent exited with code ${code}`);
      }
    });

    // Wait one tick to detect spawn failures (ENOENT, EACCES).
    // On success, resolve immediately without waiting for process exit.
    return new Promise<void>((resolve, reject) => {
      let settled = false;

      child.on("error", (err: any) => {
        if (!settled) {
          settled = true;
          reject(
            new Error(`Failed to spawn ${binaryPath}: ${err.message}`),
          );
        }
      });

      // Use nextTick to give the error event a chance to fire first
      process.nextTick(() => {
        if (!settled) {
          settled = true;
          resolve();
        }
      });
    });
  }
}
