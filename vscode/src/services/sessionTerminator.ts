/**
 * SessionTerminator — gracefully terminates a Spectra runtime process
 * identified by its PID using a SIGTERM → grace period → SIGKILL
 * escalation strategy.
 *
 * Logic spec: spec/logic/vscode/src/services/sessionTerminator.md
 */

import * as path from "path";

/**
 * Logger interface required by SessionTerminator.
 */
export interface SessionTerminatorLogger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * Injectable dependencies for SessionTerminator (testability seam).
 */
export interface SessionTerminatorDeps {
  getConfiguration: () => { get(key: string): string | undefined };
  processKill: (pid: number, signal: string | number) => void;
  execFile: (
    cmd: string,
    args: string[],
    callback: (err: Error | null, stdout: string) => void,
  ) => void;
}

/**
 * Structured result describing the outcome of a termination attempt.
 */
export interface TerminationResult {
  terminated: boolean;
  method: "sigterm" | "sigkill" | "already_dead" | "not_spectra";
  error?: string;
}

/** Grace period in milliseconds before escalating to SIGKILL. */
const GRACE_PERIOD_MS = 5000;

/** Poll interval in milliseconds for liveness checks. */
const POLL_INTERVAL_MS = 500;

/**
 * Gracefully terminates a Spectra runtime process.
 */
export class SessionTerminator {
  /**
   * Terminates a process identified by PID after verifying it belongs to Spectra.
   *
   * @param pid - The process ID to terminate
   * @param logger - Logger for diagnostic output
   * @param deps - Injectable dependencies (for testing)
   */
  static terminate(
    pid: number,
    logger: SessionTerminatorLogger,
    deps: SessionTerminatorDeps,
  ): Promise<TerminationResult> {
    return new Promise<TerminationResult>((resolve) => {
      try {
        const config = deps.getConfiguration();
        const binaryPath = config.get("binaryPath") || "spectra";
        const expectedBasename = path.basename(binaryPath);

        // Step 1: Check if process is alive
        if (!SessionTerminator.isAlive(pid, deps)) {
          resolve({ terminated: true, method: "already_dead" });
          return;
        }

        // Step 2: Verify command name via ps (callback-based, no await)
        deps.execFile("ps", ["-p", String(pid), "-o", "comm="], (err, stdout) => {
          try {
            if (err) {
              logger.error(`Failed to verify process ${pid}: ${err.message}`);
              resolve({
                terminated: false,
                method: "not_spectra",
                error: err.message,
              });
              return;
            }

            // Step 3: Match command name against expected binary
            const trimmed = stdout.trim();
            if (trimmed !== expectedBasename && trimmed !== "spectra") {
              resolve({ terminated: false, method: "not_spectra" });
              return;
            }

            // Step 4: Send SIGTERM
            try {
              deps.processKill(pid, "SIGTERM");
            } catch (killErr: any) {
              if (killErr.code === "ESRCH") {
                resolve({ terminated: true, method: "already_dead" });
                return;
              }
              if (killErr.code === "EPERM") {
                resolve({
                  terminated: false,
                  method: "sigterm",
                  error: killErr.message,
                });
                return;
              }
              resolve({
                terminated: false,
                method: "sigterm",
                error: killErr.message,
              });
              return;
            }

            logger.info(`Sent SIGTERM to process ${pid}`);

            // Step 5: Poll for death during grace period
            SessionTerminator.pollForDeath(pid, GRACE_PERIOD_MS, deps, (died) => {
              if (died) {
                resolve({ terminated: true, method: "sigterm" });
                return;
              }

              // Step 6: Escalate to SIGKILL
              logger.warn(
                `Process ${pid} did not respond to SIGTERM, sending SIGKILL`,
              );

              try {
                deps.processKill(pid, "SIGKILL");
              } catch (killErr: any) {
                if (killErr.code === "ESRCH") {
                  resolve({ terminated: true, method: "sigterm" });
                  return;
                }
              }

              // Step 7: Wait briefly and confirm death
              setTimeout(() => {
                resolve({ terminated: true, method: "sigkill" });
              }, POLL_INTERVAL_MS);
            });
          } catch (innerErr: any) {
            logger.error(
              `Unexpected error terminating process ${pid}: ${innerErr.message}`,
            );
            resolve({
              terminated: false,
              method: "sigterm",
              error: innerErr.message,
            });
          }
        });
      } catch (err: any) {
        logger.error(
          `Unexpected error terminating process ${pid}: ${err.message}`,
        );
        resolve({
          terminated: false,
          method: "sigterm",
          error: err.message,
        });
      }
    });
  }

  /**
   * Checks whether a process is alive using signal 0.
   */
  private static isAlive(pid: number, deps: SessionTerminatorDeps): boolean {
    try {
      deps.processKill(pid, 0);
      return true;
    } catch (err: any) {
      if (err.code === "ESRCH") {
        return false;
      }
      throw err;
    }
  }

  /**
   * Polls liveness every POLL_INTERVAL_MS until the process is dead or
   * the grace period expires. Uses a callback to avoid async/await boundary.
   */
  private static pollForDeath(
    pid: number,
    graceMs: number,
    deps: SessionTerminatorDeps,
    done: (died: boolean) => void,
  ): void {
    let elapsed = 0;

    const check = () => {
      elapsed += POLL_INTERVAL_MS;

      if (!SessionTerminator.isAlive(pid, deps)) {
        done(true);
        return;
      }

      if (elapsed >= graceMs) {
        done(false);
        return;
      }

      setTimeout(check, POLL_INTERVAL_MS);
    };

    setTimeout(check, POLL_INTERVAL_MS);
  }
}
