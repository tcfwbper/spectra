/**
 * Shared fake timer helper for controller tests that involve timer scheduling.
 *
 * Uses sinon's fake timer facility to control `setTimeout`/`clearTimeout`
 * deterministically. Provides a small wrapper for setup/teardown and
 * idiomatic time advancement.
 */
import * as sinon from "sinon";

/**
 * Context returned by createFakeTimerContext.
 * Holds the sinon clock and provides convenience methods.
 */
export interface FakeTimerContext {
  /** The underlying sinon fake clock. */
  clock: sinon.SinonFakeTimers;
  /** Advance the clock by `ms` milliseconds (synchronous tick). */
  tick(ms: number): void;
  /** Restore real timers. Must be called in afterEach. */
  restore(): void;
}

/**
 * Creates and installs fake timers (setTimeout, clearTimeout).
 * Call `restore()` when done (typically in afterEach).
 */
export function createFakeTimerContext(): FakeTimerContext {
  const clock = sinon.useFakeTimers({
    toFake: ["setTimeout", "clearTimeout"],
  });

  return {
    clock,
    tick(ms: number) {
      clock.tick(ms);
    },
    restore() {
      clock.restore();
    },
  };
}
