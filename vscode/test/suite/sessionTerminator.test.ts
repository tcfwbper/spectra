/**
 * Unit tests for SessionTerminator.
 *
 * Test spec: spec/test/vscode/src/services/sessionTerminator.md
 * Source under test: vscode/src/services/sessionTerminator.ts
 *
 * Scaffolded: The production module (sessionTerminator.ts) does not yet exist.
 * All tests are structured with full setup/assertions but import is deferred.
 * Once the production surface is created with a dependency-injectable static method,
 * remove the t.skip markers and wire the import.
 *
 * Expected production surface:
 *   SessionTerminator.terminate(pid, logger, deps?)
 *   where deps = { getConfiguration, processKill, execFile }
 *   Returns Promise<TerminationResult>
 *   TerminationResult = { terminated: boolean, method: string, error?: string }
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockServiceLogger,
  createSessionTerminatorDeps,
  makeProcessNotFound,
  makeProcessPermDenied,
  makePsReturn,
  makePsFail,
  type MockLogger,
  type SessionTerminatorDeps,
} from "./helpers/processStubs";

// Scaffolded import — uncomment when production module exists:
// import { SessionTerminator } from '../../src/services/sessionTerminator';
// import type { TerminationResult } from '../../src/services/sessionTerminator';

describe("SessionTerminator", function () {
  let sandbox: sinon.SinonSandbox;
  let clock: sinon.SinonFakeTimers;
  let logger: MockLogger;
  let deps: SessionTerminatorDeps;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    clock = sinon.useFakeTimers();
    logger = createMockServiceLogger();
    deps = createSessionTerminatorDeps("spectra");
  });

  afterEach(function () {
    clock.restore();
    sandbox.restore();
  });

  describe("Happy Path — terminate", function () {
    it("should return terminated with sigterm when process responds to SIGTERM", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // // Process is alive initially
      // deps.processKill.withArgs(1234, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(1234, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(1234, logger, deps);
      //
      // // After first poll (500ms), process is dead
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(1234, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result).to.deep.equal({ terminated: true, method: 'sigterm' });
    });

    it("should return terminated with sigkill when process ignores SIGTERM", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // // Process stays alive during entire grace period
      // deps.processKill.withArgs(5678, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(5678, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(5678, logger, deps);
      //
      // // Advance through 5s grace period — process stays alive
      // clock.tick(5000);
      //
      // // After SIGKILL, process dies
      // deps.processKill.withArgs(5678, 'SIGKILL').returns(undefined);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(5678, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result).to.deep.equal({ terminated: true, method: 'sigkill' });
    });

    it("should log info when SIGTERM is sent", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, deps);
      //
      // // Process dies on first poll
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // await promise;
      // expect(logger.info.called).to.be.true;
      // expect(logger.info.firstCall.args[0]).to.be.a('string');
    });

    it("should log warn when escalating to SIGKILL", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, deps);
      //
      // // Process stays alive through grace period
      // clock.tick(5000);
      // deps.processKill.withArgs(100, 'SIGKILL').returns(undefined);
      // clock.tick(500);
      //
      // await promise;
      // expect(logger.warn.called).to.be.true;
    });
  });

  describe("Happy Path — already dead", function () {
    it("should return already_dead when process does not exist", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // makeProcessNotFound(deps.processKill, 9999);
      //
      // const result = await SessionTerminator.terminate(9999, logger, deps);
      // expect(result).to.deep.equal({ terminated: true, method: 'already_dead' });
    });

    it("should return already_dead when process dies between check and SIGTERM", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // // First kill(pid, 0) succeeds — process is alive
      // deps.processKill.withArgs(4321, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // // SIGTERM throws ESRCH — process died between check and signal
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(4321, 'SIGTERM').throws(esrch);
      //
      // const result = await SessionTerminator.terminate(4321, logger, deps);
      // expect(result).to.deep.equal({ terminated: true, method: 'already_dead' });
    });
  });

  describe("Happy Path — configuration default", function () {
    it("should default to spectra when config is undefined", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // const undefinedDeps = createSessionTerminatorDeps(undefined);
      // undefinedDeps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(undefinedDeps.execFile, 'spectra');
      // undefinedDeps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, undefinedDeps);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // undefinedDeps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result.terminated).to.be.true;
    });

    it("should default to spectra when config is empty string", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // const emptyDeps = createSessionTerminatorDeps('' as any);
      // emptyDeps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(emptyDeps.execFile, 'spectra');
      // emptyDeps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, emptyDeps);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // emptyDeps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result.terminated).to.be.true;
    });
  });

  describe("Boundary Values — command name matching", function () {
    it("should match when ps reports basename of configured path", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // const pathDeps = createSessionTerminatorDeps('/usr/local/bin/spectra');
      // pathDeps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(pathDeps.execFile, 'spectra');
      // pathDeps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, pathDeps);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // pathDeps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result).to.deep.equal({ terminated: true, method: 'sigterm' });
    });

    it("should match when ps reports literal spectra regardless of config", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // const customDeps = createSessionTerminatorDeps('/opt/custom/spectra-dev');
      // customDeps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(customDeps.execFile, 'spectra');
      // customDeps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, customDeps);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // customDeps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result.terminated).to.be.true;
    });

    it("should match custom binary name from config basename", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // const customDeps = createSessionTerminatorDeps('/opt/bin/spectra-dev');
      // customDeps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(customDeps.execFile, 'spectra-dev');
      // customDeps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, customDeps);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // customDeps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result).to.deep.equal({ terminated: true, method: 'sigterm' });
    });

    it("should return not_spectra when command name does not match", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'node');
      //
      // const result = await SessionTerminator.terminate(100, logger, deps);
      // expect(result).to.deep.equal({ terminated: false, method: 'not_spectra' });
    });
  });

  describe("Error Propagation", function () {
    it("should return error result when SIGTERM fails with EPERM", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // makeProcessPermDenied(deps.processKill, 100, 'SIGTERM');
      //
      // const result = await SessionTerminator.terminate(100, logger, deps);
      // expect(result.terminated).to.be.false;
      // expect(result.method).to.equal('sigterm');
      // expect(result.error).to.be.a('string');
    });

    it("should return not_spectra with error when ps command fails", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsFail(deps.execFile, 'ps command not found');
      //
      // const result = await SessionTerminator.terminate(100, logger, deps);
      // expect(logger.error.called).to.be.true;
      // expect(result.terminated).to.be.false;
      // expect(result.method).to.equal('not_spectra');
      // expect(result.error).to.be.a('string');
    });

    it("should never throw to caller", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // // Unexpected error (not ESRCH) on liveness check
      // const unexpected = new Error('unexpected kill error');
      // deps.processKill.withArgs(100, 0).throws(unexpected);
      //
      // // Should resolve, not reject
      // const result = await SessionTerminator.terminate(100, logger, deps);
      // expect(result.error).to.be.a('string');
    });
  });

  describe("Mock / Dependency Interaction", function () {
    it("should poll liveness every 500ms during grace period", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, deps);
      //
      // // Keep alive for 4 polls, die on 5th (2500ms)
      // clock.tick(500); // poll 1
      // clock.tick(500); // poll 2
      // clock.tick(500); // poll 3
      // clock.tick(500); // poll 4
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500); // poll 5 — dead
      //
      // const result = await promise;
      // expect(result).to.deep.equal({ terminated: true, method: 'sigterm' });
      // // Verify kill(pid, 0) was called ~5 times during polling
      // const livenessCalls = deps.processKill.getCalls()
      //   .filter(c => c.args[1] === 0);
      // expect(livenessCalls.length).to.be.at.least(5);
    });

    it("should use 5 second grace period before SIGKILL", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, deps);
      //
      // // Just before 5000ms — no SIGKILL yet
      // clock.tick(4999);
      // const killCallsBefore = deps.processKill.getCalls()
      //   .filter(c => c.args[1] === 'SIGKILL');
      // expect(killCallsBefore).to.have.lengthOf(0);
      //
      // // At 5000ms — SIGKILL sent
      // clock.tick(1);
      // const killCallsAfter = deps.processKill.getCalls()
      //   .filter(c => c.args[1] === 'SIGKILL');
      // expect(killCallsAfter).to.have.lengthOf(1);
      //
      // // Cleanup
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      // await promise;
    });

    it("should wait 500ms after SIGKILL to confirm death", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const promise = SessionTerminator.terminate(100, logger, deps);
      //
      // // Grace period expires
      // clock.tick(5000);
      // deps.processKill.withArgs(100, 'SIGKILL').returns(undefined);
      //
      // // After SIGKILL, wait 500ms for confirmation
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      //
      // const result = await promise;
      // expect(result).to.deep.equal({ terminated: true, method: 'sigkill' });
    });

    it("should read configuration on every invocation", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // // First invocation
      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'spectra');
      // deps.processKill.withArgs(100, 'SIGTERM').returns(undefined);
      //
      // const p1 = SessionTerminator.terminate(100, logger, deps);
      // const esrch: any = new Error('ESRCH'); esrch.code = 'ESRCH';
      // deps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      // await p1;
      //
      // // Reset stubs for second invocation
      // deps.processKill.withArgs(100, 0).returns(undefined);
      // const p2 = SessionTerminator.terminate(100, logger, deps);
      // deps.processKill.withArgs(100, 0).throws(esrch);
      // clock.tick(500);
      // await p2;
      //
      // expect(deps.getConfiguration.callCount).to.be.at.least(2);
    });

    it("should not send any signal when process is not_spectra", function () {
      // Scaffolded: awaiting SessionTerminator production surface
      this.skip(); // Missing: SessionTerminator.terminate static method in sessionTerminator.ts

      // deps.processKill.withArgs(100, 0).returns(undefined);
      // makePsReturn(deps.execFile, 'nginx');
      //
      // const result = await SessionTerminator.terminate(100, logger, deps);
      //
      // // Only signal 0 (liveness check) should have been called
      // const signalCalls = deps.processKill.getCalls()
      //   .filter(c => c.args[1] === 'SIGTERM' || c.args[1] === 'SIGKILL');
      // expect(signalCalls).to.have.lengthOf(0);
      // expect(result).to.deep.equal({ terminated: false, method: 'not_spectra' });
    });
  });
});
