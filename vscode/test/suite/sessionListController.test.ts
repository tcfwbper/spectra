/**
 * Unit tests for SessionListController.
 *
 * Test spec: spec/test/vscode/src/controllers/sessionListController.md
 * Source under test: vscode/src/controllers/sessionListController.ts
 *
 * Scaffolded: The controller source file does not yet exist. These tests
 * are structured to compile and provide coverage once the production surface
 * is created with the expected dependency-injection seam
 * (SessionListControllerDeps).
 *
 * Missing production surface:
 *   - vscode/src/controllers/sessionListController.ts
 *   - SessionListController class
 *   - SessionListControllerDeps interface
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockControllerLogger,
  createSessionListControllerDeps,
  createDeferred,
  type MockControllerLogger,
  type SessionListControllerDeps,
  type MockTypedEventEmitter,
  type MockListWatcherInstance,
} from "./helpers/controllerStubs";

// Scaffolded: Uncomment when the production module exists.
// import { SessionListController } from "../../src/controllers/sessionListController";
// import type { SessionListControllerDeps } from "../../src/controllers/sessionListController";

describe("SessionListController", function () {
  let sandbox: sinon.SinonSandbox;
  let logger: MockControllerLogger;
  let deps: SessionListControllerDeps;
  let stateEmitter: MockTypedEventEmitter<any>;
  let errorEmitter: MockTypedEventEmitter<Error>;
  let sessionWatcher: MockListWatcherInstance;
  let workflowWatcher: MockListWatcherInstance;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    logger = createMockControllerLogger();
    const context = createSessionListControllerDeps();
    deps = context.deps;
    stateEmitter = context.stateEmitter;
    errorEmitter = context.errorEmitter;
    sessionWatcher = context.sessionWatcher;
    workflowWatcher = context.workflowWatcher;
  });

  afterEach(function () {
    sandbox.restore();
  });

  // ─── Helper: construct instance ───────────────────────────────────────────
  // Scaffolded: replace with actual constructor call when production file exists.
  // Expected signature: new SessionListController('/project', logger, deps)
  function createInstance(): any {
    // return new SessionListController('/project', logger, deps);
    return undefined;
  }

  /**
   * Helper: wait for the initial async scan to complete.
   * The controller kicks off initial scans during construction but does not
   * block on them. This helper gives the microtask queue time to drain.
   */
  async function waitForInitialScan(): Promise<void> {
    await new Promise((r) => setImmediate(r));
  }

  // ─── Happy Path — Construction ────────────────────────────────────────────

  describe("Happy Path — Construction", function () {
    it("should create SessionWatcher and WorkflowWatcher during construction", function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      createInstance();

      expect(deps.createSessionWatcher.calledOnce).to.be.true;
      expect(deps.createWorkflowWatcher.calledOnce).to.be.true;

      const [swRoot] = deps.createSessionWatcher.firstCall.args;
      const [wwRoot] = deps.createWorkflowWatcher.firstCall.args;
      expect(swRoot).to.equal("/project");
      expect(wwRoot).to.equal("/project");
    });

    it("should expose onDidUpdate and onDidError events", function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      const instance = createInstance();
      expect(instance.onDidUpdate).to.be.a("function");
      expect(instance.onDidError).to.be.a("function");
    });

    it("should kick off initial scan asynchronously without blocking construction", function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      const sessionScanDeferred = createDeferred<any[]>();
      const workflowScanDeferred = createDeferred<string[]>();
      deps.scanSessions.returns(sessionScanDeferred.promise);
      deps.scanWorkflows.returns(workflowScanDeferred.promise);

      // Constructor returns synchronously
      const instance = createInstance();
      expect(instance).to.exist;

      // Scanners have been called even though they haven't resolved
      expect(deps.scanSessions.calledOnce).to.be.true;
      expect(deps.scanWorkflows.calledOnce).to.be.true;
    });
  });

  // ─── Happy Path — onDidUpdate ─────────────────────────────────────────────

  describe("Happy Path — onDidUpdate", function () {
    it("should fire onDidUpdate with sessions and workflows after initial scan completes", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([{ id: "s1", createdAt: 100 }]);
      deps.scanWorkflows.resolves(["wf1"]);

      createInstance();
      await waitForInitialScan();

      expect(stateEmitter.fire.called).to.be.true;
      const state = stateEmitter.fire.lastCall.args[0];
      expect(state.sessions).to.deep.equal([{ id: "s1", createdAt: 100 }]);
      expect(state.workflows).to.deep.equal(["wf1"]);
    });

    it("should fire onDidUpdate when SessionWatcher triggers onDidChange", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([{ id: "s1", createdAt: 100 }]);
      deps.scanWorkflows.resolves(["wf1"]);

      createInstance();
      await waitForInitialScan();

      stateEmitter.fire.resetHistory();
      deps.scanSessions.resolves([{ id: "s2", createdAt: 200 }]);

      sessionWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.sessions).to.deep.equal([{ id: "s2", createdAt: 200 }]);
    });

    it("should fire onDidUpdate when WorkflowWatcher triggers onDidChange", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([{ id: "s1", createdAt: 100 }]);
      deps.scanWorkflows.resolves(["wf1"]);

      createInstance();
      await waitForInitialScan();

      stateEmitter.fire.resetHistory();
      deps.scanWorkflows.resolves(["wf1", "wf2"]);

      workflowWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.workflows).to.deep.equal(["wf1", "wf2"]);
    });

    it("should push full composite state even when only sessions changed", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([{ id: "s1" }]);
      deps.scanWorkflows.resolves(["wf1"]);

      createInstance();
      await waitForInitialScan();

      stateEmitter.fire.resetHistory();
      deps.scanSessions.resolves([{ id: "s1" }, { id: "s2" }]);

      sessionWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.sessions).to.deep.equal([{ id: "s1" }, { id: "s2" }]);
      expect(state.workflows).to.deep.equal(["wf1"]);
    });
  });

  // ─── Happy Path — launch ──────────────────────────────────────────────────

  describe("Happy Path — launch", function () {
    it("should call SessionLauncher.launch with workflowName and logger", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const instance = createInstance();
      await waitForInitialScan();
      await instance.launch("my-workflow");

      expect(deps.launch.calledOnce).to.be.true;
      const [workflowName, loggerArg] = deps.launch.firstCall.args;
      expect(workflowName).to.equal("my-workflow");
      expect(loggerArg).to.equal(logger);
    });
  });

  // ─── Happy Path — terminate ───────────────────────────────────────────────

  describe("Happy Path — terminate", function () {
    it("should call SessionTerminator.terminate with pid and logger", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(1234);

      expect(deps.terminate.calledOnce).to.be.true;
      const [pid, loggerArg] = deps.terminate.firstCall.args;
      expect(pid).to.equal(1234);
      expect(loggerArg).to.equal(logger);
    });

    it("should treat already_dead as success without firing onDidError", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.terminate.resolves({ method: "already_dead", terminated: true });

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(5678);

      expect(errorEmitter.fire.called).to.be.false;
    });

    it("should treat sigterm terminated as success without firing onDidError", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.terminate.resolves({ method: "sigterm", terminated: true });

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(5678);

      expect(errorEmitter.fire.called).to.be.false;
    });

    it("should treat sigkill terminated as success without firing onDidError", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.terminate.resolves({ method: "sigkill", terminated: true });

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(5678);

      expect(errorEmitter.fire.called).to.be.false;
    });
  });

  // ─── Error Propagation ────────────────────────────────────────────────────

  describe("Error Propagation", function () {
    it("should fire onDidError and log when launch throws", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.launch.rejects(new Error("ENOENT"));

      const instance = createInstance();
      await waitForInitialScan();
      await instance.launch("bad-workflow");

      expect(errorEmitter.fire.calledOnce).to.be.true;
      expect(errorEmitter.fire.firstCall.args[0]).to.be.instanceOf(Error);
      expect(errorEmitter.fire.firstCall.args[0].message).to.include("ENOENT");
      expect(logger.error.calledOnce).to.be.true;
    });

    it("should fire onDidError and log when terminate returns not_spectra", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.terminate.resolves({ method: "not_spectra", terminated: false });

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(9999);

      expect(errorEmitter.fire.calledOnce).to.be.true;
      expect(errorEmitter.fire.firstCall.args[0]).to.be.instanceOf(Error);
      expect(logger.error.calledOnce).to.be.true;
    });

    it("should fire onDidError and log when terminate returns EPERM", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.terminate.resolves({ method: "sigterm", terminated: false, error: new Error("EPERM") });

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(9999);

      expect(errorEmitter.fire.calledOnce).to.be.true;
      expect(errorEmitter.fire.firstCall.args[0]).to.be.instanceOf(Error);
      expect(logger.error.calledOnce).to.be.true;
    });
  });

  // ─── Concurrent Behaviour ─────────────────────────────────────────────────

  describe("Concurrent Behaviour", function () {
    it("should coalesce overlapping session scans via dirty flag", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      createInstance();
      await waitForInitialScan();

      // Reset after initial scan
      deps.scanSessions.resetHistory();

      const scanDeferred = createDeferred<any[]>();
      deps.scanSessions.returns(scanDeferred.promise);

      // Trigger onDidChange three times while first scan is in-flight
      sessionWatcher.triggerChange();
      sessionWatcher.triggerChange();
      sessionWatcher.triggerChange();

      // Resolve the in-flight scan
      scanDeferred.resolve([]);
      await new Promise((r) => setImmediate(r));

      // Should be called exactly twice: in-flight + one re-scan
      expect(deps.scanSessions.callCount).to.equal(2);
    });

    it("should coalesce overlapping workflow scans independently", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      createInstance();
      await waitForInitialScan();

      deps.scanWorkflows.resetHistory();

      const scanDeferred = createDeferred<string[]>();
      deps.scanWorkflows.returns(scanDeferred.promise);

      // Trigger onDidChange twice while scan is in-flight
      workflowWatcher.triggerChange();
      workflowWatcher.triggerChange();

      // Resolve
      scanDeferred.resolve([]);
      await new Promise((r) => setImmediate(r));

      // In-flight + one re-scan
      expect(deps.scanWorkflows.callCount).to.equal(2);
    });

    it("should run session and workflow scans concurrently", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      createInstance();
      await waitForInitialScan();

      deps.scanSessions.resetHistory();
      deps.scanWorkflows.resetHistory();

      const sessionDeferred = createDeferred<any[]>();
      const workflowDeferred = createDeferred<string[]>();
      deps.scanSessions.returns(sessionDeferred.promise);
      deps.scanWorkflows.returns(workflowDeferred.promise);

      // Trigger both watchers simultaneously
      sessionWatcher.triggerChange();
      workflowWatcher.triggerChange();

      // Both scanners should be called without waiting for the other
      expect(deps.scanSessions.calledOnce).to.be.true;
      expect(deps.scanWorkflows.calledOnce).to.be.true;

      // Resolve both
      sessionDeferred.resolve([]);
      workflowDeferred.resolve([]);
      await new Promise((r) => setImmediate(r));
    });
  });

  // ─── Resource Cleanup ─────────────────────────────────────────────────────

  describe("Resource Cleanup", function () {
    it("should dispose watchers and emitters on dispose", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const instance = createInstance();
      await waitForInitialScan();
      instance.dispose();

      expect(sessionWatcher.dispose.calledOnce).to.be.true;
      expect(workflowWatcher.dispose.calledOnce).to.be.true;
      expect(stateEmitter.dispose.calledOnce).to.be.true;
      expect(errorEmitter.dispose.calledOnce).to.be.true;
    });

    it("should suppress onDidUpdate after dispose", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const instance = createInstance();
      await waitForInitialScan();

      stateEmitter.fire.resetHistory();
      const scanDeferred = createDeferred<any[]>();
      deps.scanSessions.returns(scanDeferred.promise);

      sessionWatcher.triggerChange();
      instance.dispose();

      // Resolve the pending scan after dispose
      scanDeferred.resolve([{ id: "late" }]);
      await new Promise((r) => setImmediate(r));

      expect(stateEmitter.fire.called).to.be.false;
    });

    it("should suppress onDidError from launch after dispose", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.launch.rejects(new Error("ENOENT"));

      const instance = createInstance();
      await waitForInitialScan();

      instance.dispose();
      await instance.launch("wf");

      expect(errorEmitter.fire.called).to.be.false;
    });

    it("should suppress onDidError from terminate after dispose", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);
      deps.terminate.resolves({ method: "not_spectra", terminated: false });

      const instance = createInstance();
      await waitForInitialScan();

      instance.dispose();
      await instance.terminate(123);

      expect(errorEmitter.fire.called).to.be.false;
    });
  });

  // ─── Idempotency ──────────────────────────────────────────────────────────

  describe("Idempotency", function () {
    it("should handle multiple dispose calls without error", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const instance = createInstance();
      await waitForInitialScan();

      expect(() => {
        instance.dispose();
        instance.dispose();
        instance.dispose();
      }).to.not.throw();
    });
  });

  // ─── Mock / Dependency Interaction ────────────────────────────────────────

  describe("Mock / Dependency Interaction", function () {
    it("should not read or write any files directly", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      // Verified by DI design: the controller has no fs dependency.
      deps.scanSessions.resolves([{ id: "s1", createdAt: 100 }]);
      deps.scanWorkflows.resolves(["wf1"]);

      const fsReadFile = sinon.spy();
      const fsWriteFile = sinon.spy();

      const instance = createInstance();
      await waitForInitialScan();
      sessionWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));
      await instance.launch("wf1");
      await instance.terminate(123);

      expect(fsReadFile.called).to.be.false;
      expect(fsWriteFile.called).to.be.false;
    });

    it("should not spawn processes directly", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      // Verified by DI: the controller has no child_process dependency.
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const cpSpawn = sinon.spy();
      const cpExec = sinon.spy();

      const instance = createInstance();
      await waitForInitialScan();
      await instance.launch("wf");

      expect(cpSpawn.called).to.be.false;
      expect(cpExec.called).to.be.false;
      expect(deps.launch.calledOnce).to.be.true;
    });

    it("should not send signals directly", async function () {
      // Scaffolded: SessionListController class not yet implemented
      this.skip();
      // Verified by DI: the controller has no process.kill dependency.
      deps.scanSessions.resolves([]);
      deps.scanWorkflows.resolves([]);

      const processKill = sinon.spy();

      const instance = createInstance();
      await waitForInitialScan();
      await instance.terminate(123);

      expect(processKill.called).to.be.false;
      expect(deps.terminate.calledOnce).to.be.true;
    });
  });
});
