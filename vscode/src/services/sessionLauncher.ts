/**
 * SessionLauncher — launches a new Spectra workflow session by spawning
 * a detached CLI process.
 *
 * Logic spec: spec/logic/vscode/src/services/sessionLauncher.md
 */

/**
 * Logger interface required by SessionLauncher.
 */
export interface SessionLauncherLogger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * Injectable dependencies for SessionLauncher (testability seam).
 */
export interface SessionLauncherDeps {
  getConfiguration: () => { get(key: string): string | undefined };
  spawn: (binary: string, args: string[], options: any) => any;
  randomUUID: () => string;
}

/**
 * Launches a new Spectra workflow session by spawning a detached process.
 * The spawned process survives extension reload and VS Code restart.
 */
export class SessionLauncher {
  /**
   * Launches a new workflow session.
   *
   * @param workflowName - The workflow to run
   * @param logger - Logger for diagnostic output
   * @param deps - Injectable dependencies (for testing)
   */
  static async launch(
    workflowName: string,
    logger: SessionLauncherLogger,
    deps: SessionLauncherDeps,
  ): Promise<void> {
    const config = deps.getConfiguration();
    const binaryPath = config.get("binaryPath") || "spectra";

    const sessionId = deps.randomUUID();

    const args = [
      "run",
      "--workflow",
      workflowName,
      "--session-id",
      sessionId,
    ];

    const child = deps.spawn(binaryPath, args, {
      detached: true,
      stdio: "ignore",
    });

    child.unref();

    logger.info(
      `Launched workflow "${workflowName}" with session ${sessionId}`,
    );

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

      process.nextTick(() => {
        if (!settled) {
          settled = true;
          resolve();
        }
      });
    });
  }
}
