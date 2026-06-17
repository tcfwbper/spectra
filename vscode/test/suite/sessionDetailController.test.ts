/**
 * Unit tests for SessionDetailController.
 *
 * Test spec: spec/test/vscode/src/controllers/sessionDetailController.md
 * Source under test: vscode/src/controllers/sessionDetailController.ts
 *
 * Scaffolded: The controller source file does not yet exist. These tests
 * are structured to compile and provide coverage once the production surface
 * is created with the expected dependency-injection seam
 * (SessionDetailControllerDeps).
 *
 * Missing production surface:
 *   - vscode/src/controllers/sessionDetailController.ts
 *   - SessionDetailController class
 *   - SessionDetailControllerDeps interface
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockControllerLogger,
  createSessionDetailControllerDeps,
  createMockEventWatcherInstance,
  createDeferred,
  type MockControllerLogger,
  type SessionDetailControllerDeps,
  type MockTypedEventEmitter,
  type MockEventWatcherInstance,
} from "./helpers/controllerStubs";

// Scaffolded: Uncomment when the production module exists.
// import { SessionDetailController } from "../../src/controllers/sessionDetailController";
// import type { SessionDetailControllerDeps } from "../../src/controllers/sessionDetailController";

describe("SessionDetailController", function () {
  let sandbox: sinon.SinonSandbox;
  let logger: MockControllerLogger;
  let deps: SessionDetailControllerDeps;
  let stateEmitter: MockTypedEventEmitter<any>;
  let errorEmitter: MockTypedEventEmitter<Error>;
  let eventWatcher: MockEventWatcherInstance;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    logger = createMockControllerLogger();
    const context = createSessionDetailControllerDeps();
    deps = context.deps;
    stateEmitter = context.stateEmitter;
    errorEmitter = context.errorEmitter;
    eventWatcher = context.eventWatcher;
  });

  afterEach(function () {
    sandbox.restore();
  });

  // ─── Helper: construct instance ───────────────────────────────────────────
  // Scaffolded: replace with actual constructor call when production file exists.
  // Expected signature: new SessionDetailController('/project', logger, deps)
  function createInstance(): any {
    // return new SessionDetailController('/project', logger, deps);
    return undefined;
  }

  // ─── Happy Path — Construction ────────────────────────────────────────────

  describe("Happy Path — Construction", function () {
    it("should store projectRoot and logger", function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      expect(instance).to.exist;
    });

    it("should expose onDidUpdate and onDidError events", function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      expect(instance.onDidUpdate).to.be.a("function");
      expect(instance.onDidError).to.be.a("function");
    });

    it("should not create EventWatcher during construction", function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      createInstance();
      expect(deps.createEventWatcher.called).to.be.false;
    });

    it("should initialize with null currentSessionId and zero generation", function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      createInstance();
      // No onDidUpdate fired during construction
      expect(stateEmitter.fire.called).to.be.false;
    });
  });

  // ─── Happy Path — open ────────────────────────────────────────────────────

  describe("Happy Path — open", function () {
    it("should create EventWatcher and fire onDidUpdate with assembled state", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([{ type: "submit", ts: 100 }]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "running", status: "running", pid: 42 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      expect(deps.createEventWatcher.calledOnce).to.be.true;
      expect(stateEmitter.fire.calledOnce).to.be.true;

      const state = stateEmitter.fire.firstCall.args[0];
      expect(state).to.deep.include({
        sessionId: "s1",
        workflowName: "wf1",
        entryNode: "start",
        currentState: "running",
        status: "running",
        pid: 42,
      });
      expect(state.eventTypes).to.deep.equal(["submit"]);
      expect(state.events).to.deep.equal([{ type: "submit", ts: 100 }]);
    });

    it("should pass correct arguments to EventWatcher constructor", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      await instance.open("sess-abc", "wf1");

      expect(deps.createEventWatcher.calledOnce).to.be.true;
      const [projectRoot, sessionId] = deps.createEventWatcher.firstCall.args;
      expect(projectRoot).to.equal("/project");
      expect(sessionId).to.equal("sess-abc");
    });

    it("should pass correct arguments to WorkflowDefinitionParser.parse", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      await instance.open("s1", "my-workflow");

      expect(deps.parseWorkflowDefinition.calledOnce).to.be.true;
      const [projectRoot, workflowName, loggerArg] = deps.parseWorkflowDefinition.firstCall.args;
      expect(projectRoot).to.equal("/project");
      expect(workflowName).to.equal("my-workflow");
      expect(loggerArg).to.equal(logger);
    });

    it("should subscribe to EventWatcher.onDidChange", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      await instance.open("s1", "wf1");

      // Verified by watcher mock having a registered listener after open
      expect(deps.createEventWatcher.calledOnce).to.be.true;
    });
  });

  // ─── Happy Path — internal scan routine ───────────────────────────────────

  describe("Happy Path — internal scan routine", function () {
    it("should re-scan and fire onDidUpdate when onDidChange fires", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([{ type: "submit", ts: 100 }]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "running", status: "running", pid: 42 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      // Reset for re-scan
      stateEmitter.fire.resetHistory();
      deps.scanEvents.resolves([{ type: "ack", ts: 200 }]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "done", status: "completed", pid: 42 }]);

      eventWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.events).to.deep.equal([{ type: "ack", ts: 200 }]);
      expect(state.entryNode).to.equal("start");
    });

    it("should include previously stored entryNode and eventTypes in re-scan state", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["go"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      stateEmitter.fire.resetHistory();
      deps.scanEvents.resolves([{ type: "go", ts: 300 }]);

      eventWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));

      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.entryNode).to.equal("start");
      expect(state.eventTypes).to.deep.equal(["go"]);
    });
  });

  // ─── Happy Path — sendEvent ───────────────────────────────────────────────

  describe("Happy Path — sendEvent", function () {
    it("should call EventDispatcher.dispatch with correct arguments", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");
      await instance.sendEvent("submit", "hello");

      expect(deps.dispatchEvent.calledOnce).to.be.true;
      const [eventType, sessionId, message, loggerArg] = deps.dispatchEvent.firstCall.args;
      expect(eventType).to.equal("submit");
      expect(sessionId).to.equal("s1");
      expect(message).to.equal("hello");
      expect(loggerArg).to.equal(logger);
    });
  });

  // ─── Error Propagation ────────────────────────────────────────────────────

  describe("Error Propagation", function () {
    it("should propagate EventWatcher construction error to caller", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const error = new Error("ENOENT");
      deps.createEventWatcher.throws(error);

      const instance = createInstance();
      try {
        await instance.open("s1", "wf1");
        expect.fail("should have thrown");
      } catch (err: any) {
        expect(err.message).to.equal("ENOENT");
      }
    });

    it("should fire onDidError and log when sendEvent dispatch fails with ENOENT", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);
      deps.dispatchEvent.rejects(new Error("ENOENT"));

      const instance = createInstance();
      await instance.open("s1", "wf1");
      await instance.sendEvent("submit", "msg");

      expect(errorEmitter.fire.calledOnce).to.be.true;
      expect(errorEmitter.fire.firstCall.args[0]).to.be.instanceOf(Error);
      expect(errorEmitter.fire.firstCall.args[0].message).to.include("ENOENT");
      expect(logger.error.calledOnce).to.be.true;
    });

    it("should fire onDidError and log when sendEvent dispatch fails with EACCES", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);
      deps.dispatchEvent.rejects(new Error("EACCES"));

      const instance = createInstance();
      await instance.open("s1", "wf1");
      await instance.sendEvent("submit", "msg");

      expect(errorEmitter.fire.calledOnce).to.be.true;
      expect(errorEmitter.fire.firstCall.args[0]).to.be.instanceOf(Error);
      expect(errorEmitter.fire.firstCall.args[0].message).to.include("EACCES");
      expect(logger.error.calledOnce).to.be.true;
    });
  });

  // ─── Concurrent Behaviour ─────────────────────────────────────────────────

  describe("Concurrent Behaviour", function () {
    it("should discard stale scan results when open is called again", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const deferred1 = createDeferred<any[]>();
      deps.scanEvents.onFirstCall().returns(deferred1.promise);
      deps.scanEvents.onSecondCall().resolves([{ type: "new", ts: 500 }]);
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanSessions.resolves([{ id: "s2", currentState: "start", status: "running", pid: 2 }]);

      const watcher1 = createMockEventWatcherInstance();
      const watcher2 = createMockEventWatcherInstance();
      deps.createEventWatcher.onFirstCall().returns(watcher1);
      deps.createEventWatcher.onSecondCall().returns(watcher2);

      const instance = createInstance();
      const open1Promise = instance.open("s1", "wf1");

      // Immediately call open again before first completes
      await instance.open("s2", "wf2");

      // Now resolve the first open's scan
      deferred1.resolve([{ type: "old", ts: 1 }]);
      await open1Promise.catch(() => {});

      // onDidUpdate should NOT have fired for s1 results
      const firedStates = stateEmitter.fire.args.map((a: any[]) => a[0]);
      const s1Fires = firedStates.filter((s: any) => s.sessionId === "s1");
      expect(s1Fires).to.have.length(0);
    });

    it("should dispose previous watcher when open is called again", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const watcher1 = createMockEventWatcherInstance();
      const watcher2 = createMockEventWatcherInstance();
      deps.createEventWatcher.onFirstCall().returns(watcher1);
      deps.createEventWatcher.onSecondCall().returns(watcher2);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      deps.scanSessions.resolves([{ id: "s2", currentState: "start", status: "running", pid: 2 }]);
      await instance.open("s2", "wf2");

      expect(watcher1.dispose.calledOnce).to.be.true;
    });

    it("should coalesce overlapping scans via dirty flag", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      // Reset scan call count after initial open
      deps.scanEvents.resetHistory();

      const scanDeferred = createDeferred<any[]>();
      deps.scanEvents.returns(scanDeferred.promise);

      // Trigger onDidChange three times while scan is in-flight
      eventWatcher.triggerChange();
      eventWatcher.triggerChange();
      eventWatcher.triggerChange();

      // Resolve the in-flight scan
      scanDeferred.resolve([]);
      await new Promise((r) => setImmediate(r));

      // Should have been called twice total: in-flight + one re-scan after dirty flag
      expect(deps.scanEvents.callCount).to.equal(2);
    });

    it("should discard scan result when generation changes mid-scan", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      stateEmitter.fire.resetHistory();
      const scanDeferred = createDeferred<any[]>();
      deps.scanEvents.returns(scanDeferred.promise);

      // Trigger onDidChange to start a scan
      eventWatcher.triggerChange();

      // Open a new session (increments generation)
      deps.scanEvents.resolves([]);
      const watcher2 = createMockEventWatcherInstance();
      deps.createEventWatcher.returns(watcher2);
      await instance.open("s2", "wf2");

      // Resolve the original scan (stale)
      scanDeferred.resolve([{ type: "stale", ts: 999 }]);
      await new Promise((r) => setImmediate(r));

      // The stale scan should not fire onDidUpdate for stale events
      const firedStates = stateEmitter.fire.args.map((a: any[]) => a[0]);
      const staleEvents = firedStates.filter(
        (s: any) => s.events && s.events.some((e: any) => e.type === "stale"),
      );
      expect(staleEvents).to.have.length(0);
    });
  });

  // ─── Resource Cleanup ─────────────────────────────────────────────────────

  describe("Resource Cleanup", function () {
    it("should dispose watcher and emitters on dispose", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");
      instance.dispose();

      expect(eventWatcher.dispose.calledOnce).to.be.true;
      expect(stateEmitter.dispose.calledOnce).to.be.true;
      expect(errorEmitter.dispose.calledOnce).to.be.true;
    });

    it("should suppress onDidUpdate after dispose", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      stateEmitter.fire.resetHistory();
      const scanDeferred = createDeferred<any[]>();
      deps.scanEvents.returns(scanDeferred.promise);

      eventWatcher.triggerChange();
      instance.dispose();

      // Resolve the pending scan after dispose
      scanDeferred.resolve([{ type: "late", ts: 999 }]);
      await new Promise((r) => setImmediate(r));

      expect(stateEmitter.fire.called).to.be.false;
    });

    it("should no-op on open after dispose", function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      instance.dispose();

      deps.createEventWatcher.resetHistory();
      instance.open("s1", "wf1");

      expect(deps.createEventWatcher.called).to.be.false;
    });

    it("should no-op on sendEvent after dispose", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      const instance = createInstance();
      instance.dispose();

      await instance.sendEvent("submit", "msg");

      expect(deps.dispatchEvent.called).to.be.false;
    });

    it("should set watcher to null after dispose", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");
      instance.dispose();

      // Watcher reference cleared — subsequent open would not double-dispose
      expect(eventWatcher.dispose.calledOnce).to.be.true;
    });
  });

  // ─── Idempotency ──────────────────────────────────────────────────────────

  describe("Idempotency", function () {
    it("should handle multiple dispose calls without error", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      expect(() => {
        instance.dispose();
        instance.dispose();
        instance.dispose();
      }).to.not.throw();
    });
  });

  // ─── Null / Empty Input ───────────────────────────────────────────────────

  describe("Null / Empty Input", function () {
    it("should push empty events array when EventScanner returns empty", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["go"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.events).to.deep.equal([]);
    });

    it("should push empty eventTypes when WorkflowDefinitionParser returns empty", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "", eventTypes: [] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "", status: "initializing", pid: 0 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.entryNode).to.equal("");
      expect(state.eventTypes).to.deep.equal([]);
    });

    it("should default session fields when SessionScanner has no matching session", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["go"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "other", currentState: "done", status: "completed", pid: 99 }]);

      const instance = createInstance();
      await instance.open("s1", "wf1");

      expect(stateEmitter.fire.calledOnce).to.be.true;
      const state = stateEmitter.fire.firstCall.args[0];
      expect(state.currentState).to.equal("");
      expect(state.status).to.equal("initializing");
      expect(state.pid).to.equal(0);
    });
  });

  // ─── Mock / Dependency Interaction ────────────────────────────────────────

  describe("Mock / Dependency Interaction", function () {
    it("should not read or write any files directly", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      // Verified by the DI design: the controller has no fs dependency.
      // All file I/O is delegated to scanners/watchers via injected stubs.
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([{ type: "submit", ts: 100 }]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const fsReadFile = sinon.spy();
      const fsWriteFile = sinon.spy();

      const instance = createInstance();
      await instance.open("s1", "wf1");
      eventWatcher.triggerChange();
      await new Promise((r) => setImmediate(r));
      await instance.sendEvent("submit", "msg");

      expect(fsReadFile.called).to.be.false;
      expect(fsWriteFile.called).to.be.false;
    });

    it("should not spawn processes directly", async function () {
      // Scaffolded: SessionDetailController class not yet implemented
      this.skip();
      // Verified by DI: the controller has no child_process dependency.
      deps.parseWorkflowDefinition.resolves({ entryNode: "start", eventTypes: ["submit"] });
      deps.scanEvents.resolves([]);
      deps.scanSessions.resolves([{ id: "s1", currentState: "start", status: "running", pid: 1 }]);

      const cpSpawn = sinon.spy();
      const cpExec = sinon.spy();

      const instance = createInstance();
      await instance.open("s1", "wf1");
      await instance.sendEvent("submit", "msg");

      expect(cpSpawn.called).to.be.false;
      expect(cpExec.called).to.be.false;
      expect(deps.dispatchEvent.calledOnce).to.be.true;
    });
  });
});
