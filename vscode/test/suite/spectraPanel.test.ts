/**
 * Unit tests for SpectraPanel.
 *
 * Test spec: spec/test/vscode/src/views/spectraPanel.md
 * Source under test: vscode/src/views/spectraPanel.ts
 *
 * Scaffolded: The source file does not yet exist. These tests are structured
 * to compile and provide coverage once the production surface is created.
 *
 * Missing production surface:
 *   - vscode/src/views/spectraPanel.ts
 *   - SpectraPanel class (static createOrReveal, showSessionList, showSessionDetail, dispose)
 *   - vscode/src/views/getWebviewContent.ts (imported by SpectraPanel)
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createStubWebviewPanel,
  createStubExtensionUri,
  createStubExtensionContext,
  createMockPanelLogger,
  type StubWebviewPanel,
  type StubExtensionContext,
  type StubUri,
  type MockPanelLogger,
} from "./helpers/webviewStubs";

import { SpectraPanel } from "../../src/views/spectraPanel";

// Stub for getWebviewContent — will be wired via sinon/proxyquire when
// the production module exists.
// The exact stubbing mechanism depends on the module loader. For now,
// we define the expected interaction shape and use sinon stubs.

describe("SpectraPanel", function () {
  let sandbox: sinon.SinonSandbox;
  let mockPanel: StubWebviewPanel;
  let mockContext: StubExtensionContext;
  let mockExtensionUri: StubUri;
  let mockLogger: MockPanelLogger;
  let createWebviewPanelStub: sinon.SinonStub;
  let getWebviewContentStub: sinon.SinonStub;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    mockPanel = createStubWebviewPanel();
    mockContext = createStubExtensionContext();
    mockExtensionUri = createStubExtensionUri();
    mockLogger = createMockPanelLogger();

    // Stub vscode.window.createWebviewPanel
    createWebviewPanelStub = sandbox.stub().returns(mockPanel);

    // Stub getWebviewContent
    getWebviewContentStub = sandbox.stub().returns("<html></html>");

    // Reset singleton between tests.
    // The exact mechanism depends on the production surface exposing a
    // test-only reset or using module reload. We rely on dispose clearing
    // the static instance (per logic spec).
    resetSingleton();
  });

  afterEach(function () {
    sandbox.restore();
  });

  // ─── Helper: reset singleton state ─────────────────────────────────────────
  /**
   * Ensures the SpectraPanel singleton is cleared between tests.
   * Production surface must clear the static instance on dispose (per spec).
   * This helper triggers dispose if an instance exists, or accesses a
   * test-only reset path.
   *
   * Scaffolded: exact mechanism will depend on the production singleton
   * implementation. May need module reload or a static `_resetForTest()`.
   */
  function resetSingleton(): void {
    // Attempt to clear via the public API by triggering dispose on any
    // existing mock panel. When the production surface exists, the
    // onDidDispose callback clears the static instance.
    // For now this is a no-op placeholder.
  }

  // ─── Helper: create panel via static factory ────────────────────────────────
  /**
   * Calls SpectraPanel.createOrReveal with the test mocks.
   *
   * NOTE: The exact wiring of createWebviewPanelStub and getWebviewContentStub
   * into the production module depends on the module loader / DI approach.
   * When the production surface exists, this may use proxyquire or a DI seam.
   */
  function createOrReveal(): any {
    return SpectraPanel.createOrReveal(
      mockContext as any,
      mockExtensionUri as any,
      mockLogger as any,
    );
  }

  // ─── Happy Path — Construction ─────────────────────────────────────────────

  describe("Happy Path — Construction", function () {
    it("should create a new WebviewPanel with correct options", function () {
      const instance = createOrReveal();

      expect(createWebviewPanelStub.calledOnce).to.be.true;
      const [viewType, title, column, options] =
        createWebviewPanelStub.firstCall.args;

      expect(viewType).to.equal("spectra");
      expect(title).to.equal("Spectra");
      // vscode.ViewColumn.One === 1
      expect(column).to.equal(1);
      expect(options).to.deep.include({
        enableScripts: true,
        retainContextWhenHidden: true,
      });
      expect(options.localResourceRoots).to.deep.equal([mockExtensionUri]);
    });

    it("should assign HTML from getWebviewContent to webview", function () {
      getWebviewContentStub.returns("<html>test</html>");

      createOrReveal();

      expect(mockPanel.webview.html).to.equal("<html>test</html>");
    });

    it("should push instance into context.subscriptions", function () {
      const lengthBefore = mockContext.subscriptions.length;

      createOrReveal();

      expect(mockContext.subscriptions.length).to.equal(lengthBefore + 1);
    });

    it("should log panel creation", function () {
      createOrReveal();

      expect(mockLogger.info.called).to.be.true;
    });

    it("should expose onDidReceiveMessage event", function () {
      const instance = createOrReveal();

      expect(instance.onDidReceiveMessage).to.be.a("function");
    });

    it("should expose onDidDispose event", function () {
      const instance = createOrReveal();

      expect(instance.onDidDispose).to.be.a("function");
    });
  });

  // ─── Idempotency ──────────────────────────────────────────────────────────

  describe("Idempotency", function () {
    it("should reveal existing panel when called again", function () {
      const instance1 = createOrReveal();

      // Second call
      const instance2 = createOrReveal();

      // createWebviewPanel called only once
      expect(createWebviewPanelStub.calledOnce).to.be.true;
      // reveal called on the existing panel
      expect(mockPanel.reveal.calledOnce).to.be.true;
      // Same instance returned
      expect(instance2).to.equal(instance1);
    });

    it("should create new panel after previous was disposed", function () {
      createOrReveal();

      // Trigger dispose
      mockPanel.triggerDispose();

      // Create a fresh mock panel for the second creation
      const secondPanel = createStubWebviewPanel();
      createWebviewPanelStub.returns(secondPanel);

      const instance2 = createOrReveal();

      // createWebviewPanel should be called a second time
      expect(createWebviewPanelStub.calledTwice).to.be.true;
      expect(instance2).to.exist;
    });
  });

  // ─── Happy Path — showSessionList ──────────────────────────────────────────

  describe("Happy Path — showSessionList", function () {
    it("should post showSessions message to webview", function () {
      const instance = createOrReveal();
      const state = { sessions: [], workflows: ["wf1"] };

      instance.showSessionList(state);

      expect(mockPanel.webview.postMessage.calledOnce).to.be.true;
      expect(mockPanel.webview.postMessage.firstCall.args[0]).to.deep.equal({
        type: "showSessions",
        state: { sessions: [], workflows: ["wf1"] },
      });
    });

    it("should update currentPage to sessions", function () {
      const instance = createOrReveal();

      // First go to detail
      instance.showSessionDetail({
        sessionId: "s1",
        workflowName: "wf1",
        entryNode: "start",
        currentState: "start",
        status: "running",
        pid: 42,
        eventTypes: [],
        events: [],
      });

      mockPanel.webview.postMessage.resetHistory();

      // Then back to sessions
      instance.showSessionList({ sessions: [], workflows: [] });

      expect(mockPanel.webview.postMessage.calledOnce).to.be.true;
      expect(mockPanel.webview.postMessage.firstCall.args[0].type).to.equal(
        "showSessions",
      );
    });
  });

  // ─── Happy Path — showSessionDetail ────────────────────────────────────────

  describe("Happy Path — showSessionDetail", function () {
    it("should post showDetail message to webview", function () {
      const instance = createOrReveal();
      const state = {
        sessionId: "s1",
        workflowName: "wf1",
        entryNode: "start",
        currentState: "start",
        status: "running",
        pid: 42,
        eventTypes: ["submit"],
        events: [],
      };

      instance.showSessionDetail(state);

      expect(mockPanel.webview.postMessage.calledOnce).to.be.true;
      expect(mockPanel.webview.postMessage.firstCall.args[0]).to.deep.equal({
        type: "showDetail",
        state: {
          sessionId: "s1",
          workflowName: "wf1",
          entryNode: "start",
          currentState: "start",
          status: "running",
          pid: 42,
          eventTypes: ["submit"],
          events: [],
        },
      });
    });

    it("should update currentPage to detail", function () {
      const instance = createOrReveal();
      const state = {
        sessionId: "s1",
        workflowName: "wf1",
        entryNode: "start",
        currentState: "start",
        status: "running",
        pid: 42,
        eventTypes: [],
        events: [],
      };

      instance.showSessionDetail(state);

      expect(mockPanel.webview.postMessage.calledOnce).to.be.true;
      expect(mockPanel.webview.postMessage.firstCall.args[0].type).to.equal(
        "showDetail",
      );
    });
  });

  // ─── Mock / Dependency Interaction ─────────────────────────────────────────

  describe("Mock / Dependency Interaction", function () {
    it("should forward webview messages to onDidReceiveMessage subscribers", function () {
      const instance = createOrReveal();
      const spy = sinon.spy();

      instance.onDidReceiveMessage(spy);

      // Simulate webview sending a message
      mockPanel.triggerMessage({
        command: "navigateToDetail",
        sessionId: "s1",
        workflowName: "wf1",
      });

      expect(spy.calledOnce).to.be.true;
      expect(spy.firstCall.args[0]).to.deep.equal({
        command: "navigateToDetail",
        sessionId: "s1",
        workflowName: "wf1",
      });
    });

    it("should forward unrecognized commands without filtering", function () {
      const instance = createOrReveal();
      const spy = sinon.spy();

      instance.onDidReceiveMessage(spy);

      mockPanel.triggerMessage({ command: "unknownCommand", data: 123 });

      expect(spy.calledOnce).to.be.true;
      expect(spy.firstCall.args[0]).to.deep.equal({
        command: "unknownCommand",
        data: 123,
      });
    });

    it("should fire onDidDispose when panel is closed", function () {
      const instance = createOrReveal();
      const spy = sinon.spy();

      instance.onDidDispose(spy);

      mockPanel.triggerDispose();

      expect(spy.calledOnce).to.be.true;
    });

    it("should call getWebviewContent with webview and extensionUri", function () {
      createOrReveal();

      expect(getWebviewContentStub.calledOnce).to.be.true;
      const [webviewArg, uriArg] = getWebviewContentStub.firstCall.args;
      expect(webviewArg).to.equal(mockPanel.webview);
      expect(uriArg).to.equal(mockExtensionUri);
    });
  });

  // ─── Resource Cleanup ─────────────────────────────────────────────────────

  describe("Resource Cleanup", function () {
    it("should dispose underlying panel when dispose is called", function () {
      const instance = createOrReveal();

      instance.dispose();

      expect(mockPanel.dispose.calledOnce).to.be.true;
    });

    it("should set static instance to null on disposal", function () {
      createOrReveal();

      // Trigger dispose to clear static instance
      mockPanel.triggerDispose();

      // Create a fresh panel for next call
      const secondPanel = createStubWebviewPanel();
      createWebviewPanelStub.returns(secondPanel);

      // A new panel should be created (proving instance was null)
      createOrReveal();
      expect(createWebviewPanelStub.calledTwice).to.be.true;
    });

    it("should log on panel disposal", function () {
      createOrReveal();

      mockLogger.info.resetHistory();
      mockPanel.triggerDispose();

      expect(mockLogger.info.called).to.be.true;
    });

    it("should not fire onDidReceiveMessage after disposal", function () {
      const instance = createOrReveal();
      const spy = sinon.spy();

      instance.onDidReceiveMessage(spy);

      // Dispose the panel
      mockPanel.triggerDispose();

      // Simulate message after disposal
      mockPanel.triggerMessage({ command: "late", data: "ignored" });

      // Spy should not have been called after disposal
      // (The spy may have been called 0 times if the listener is properly
      // deregistered, or it may be called only before disposal.)
      expect(spy.calledWith(sinon.match({ command: "late" }))).to.be.false;
    });

    it("should handle multiple dispose calls gracefully", function () {
      const instance = createOrReveal();

      expect(() => {
        instance.dispose();
        instance.dispose();
      }).to.not.throw();
    });

    it("should not throw when showSessionList is called after disposal", function () {
      const instance = createOrReveal();
      mockPanel.triggerDispose();

      expect(() => {
        instance.showSessionList({ sessions: [], workflows: [] });
      }).to.not.throw();
    });

    it("should not throw when showSessionDetail is called after disposal", function () {
      const instance = createOrReveal();
      mockPanel.triggerDispose();

      expect(() => {
        instance.showSessionDetail({
          sessionId: "s1",
          workflowName: "wf1",
          entryNode: "start",
          currentState: "start",
          status: "running",
          pid: 42,
          eventTypes: [],
          events: [],
        });
      }).to.not.throw();
    });
  });

  // ─── Asynchronous Flow ─────────────────────────────────────────────────────

  describe("Asynchronous Flow", function () {
    it("should not lose messages posted before webview JS initializes", function () {
      const instance = createOrReveal();
      const state = {
        sessionId: "s1",
        workflowName: "wf1",
        entryNode: "start",
        currentState: "start",
        status: "running",
        pid: 42,
        eventTypes: [],
        events: [],
      };

      // Call immediately after creation — should not throw
      expect(() => {
        instance.showSessionDetail(state);
      }).to.not.throw();

      // postMessage should have been called (VS Code handles queuing internally)
      expect(mockPanel.webview.postMessage.calledOnce).to.be.true;
    });
  });
});
