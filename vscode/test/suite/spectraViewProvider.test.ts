/**
 * Unit tests for SpectraViewProvider.
 *
 * Test spec: spec/test/vscode/src/views/spectraViewProvider.md
 * Source under test: vscode/src/views/spectraViewProvider.ts
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createStubWebviewView,
  createStubExtensionUri,
  createStubWebviewViewResolveContext,
  createStubCancellationToken,
  createMockPanelLogger,
  type StubWebviewView,
  type StubUri,
  type StubWebviewViewResolveContext,
  type StubCancellationToken,
  type MockPanelLogger,
} from "./helpers/webviewStubs";

import { SpectraViewProvider } from "../../src/views/spectraViewProvider";

describe("SpectraViewProvider", function () {
  let sandbox: sinon.SinonSandbox;
  let mockExtensionUri: StubUri;
  let mockLogger: MockPanelLogger;
  let mockWebviewView: StubWebviewView;
  let mockContext: StubWebviewViewResolveContext;
  let mockToken: StubCancellationToken;
  let getWebviewContentStub: sinon.SinonStub;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    mockExtensionUri = createStubExtensionUri();
    mockLogger = createMockPanelLogger();
    mockWebviewView = createStubWebviewView();
    mockContext = createStubWebviewViewResolveContext();
    mockToken = createStubCancellationToken();
    getWebviewContentStub = sandbox.stub().returns("<html></html>");
  });

  afterEach(function () {
    sandbox.restore();
  });

  // ─── Helper: create instance ────────────────────────────────────────────────

  /**
   * Creates a SpectraViewProvider instance.
   * Expected signature: new SpectraViewProvider(extensionUri, logger, deps?)
   */
  function createInstance(): SpectraViewProvider {
    return new SpectraViewProvider(mockExtensionUri as any, mockLogger as any, {
      getWebviewContent: getWebviewContentStub,
    });
  }

  /**
   * Calls resolveWebviewView on the instance.
   */
  function resolveView(instance: SpectraViewProvider, view?: StubWebviewView): void {
    const v = view || mockWebviewView;
    instance.resolveWebviewView(v as any, mockContext as any, mockToken as any);
  }

  // ─── Happy Path — Construction ─────────────────────────────────────────────

  describe("Happy Path — Construction", function () {
    it("should store extensionUri and logger", function () {
      const instance = createInstance();
      // Verify via subsequent method calls that extensionUri and logger are stored
      resolveView(instance);
      expect(getWebviewContentStub.calledOnce).to.be.true;
      const [, uriArg] = getWebviewContentStub.firstCall.args;
      expect(uriArg).to.equal(mockExtensionUri);
    });

    it("should expose onDidReceiveMessage event", function () {
      const instance = createInstance();
      expect(instance.onDidReceiveMessage).to.be.a("function");
    });

    it("should initialize view as null", function () {
      const instance = createInstance();
      // Calling showSessionList stores as pending (no postMessage call)
      instance.showSessionList({ sessions: [], workflows: [] });
      // — no error thrown, no postMessage (no view yet)
    });

    it("should initialize pendingMessage as null", function () {
      const instance = createInstance();
      resolveView(instance);
      expect(mockWebviewView.webview.postMessage.called).to.be.false;
    });
  });

  // ─── Happy Path — resolveWebviewView ───────────────────────────────────────

  describe("Happy Path — resolveWebviewView", function () {
    it("should configure webview options with enableScripts and localResourceRoots", function () {
      const instance = createInstance();
      resolveView(instance);
      expect(mockWebviewView.webview.options).to.deep.equal({
        enableScripts: true,
        localResourceRoots: [mockExtensionUri],
      });
    });

    it("should assign HTML from getWebviewContent to webview", function () {
      getWebviewContentStub.returns("<html>test</html>");
      const instance = createInstance();
      resolveView(instance);
      expect(mockWebviewView.webview.html).to.equal("<html>test</html>");
    });

    it("should deliver pendingMessage after HTML assignment", function () {
      const instance = createInstance();
      instance.showNotInitialized(); // stores as pending
      resolveView(instance);
      expect(mockWebviewView.webview.postMessage.calledOnce).to.be.true;
      expect(mockWebviewView.webview.postMessage.firstCall.args[0]).to.deep.equal(
        { type: "showNotInitialized" }
      );
    });

    it("should clear pendingMessage after delivery", function () {
      const instance = createInstance();
      instance.showSessionList({ sessions: [], workflows: [] });
      resolveView(instance);
      // Trigger dispose → null the view
      mockWebviewView.triggerDispose();
      // Resolve with a new view
      const secondView = createStubWebviewView();
      resolveView(instance, secondView);
      expect(secondView.webview.postMessage.called).to.be.false;
    });

    it("should subscribe to webview onDidReceiveMessage", function () {
      const instance = createInstance();
      resolveView(instance);
      const spy = sinon.spy();
      instance.onDidReceiveMessage(spy);
      mockWebviewView.triggerMessage({ command: "navigateToList" });
      expect(spy.calledOnce).to.be.true;
      expect(spy.firstCall.args[0]).to.deep.equal({ command: "navigateToList" });
    });

    it("should set view to null on webviewView dispose", function () {
      const instance = createInstance();
      resolveView(instance);
      mockWebviewView.triggerDispose();
      // After dispose, showSessionList stores as pending
      instance.showSessionList({ sessions: [], workflows: [] });
      expect(mockWebviewView.webview.postMessage.called).to.be.false;
      expect(mockLogger.info.called).to.be.true;
    });

    it("should log view resolution", function () {
      const instance = createInstance();
      resolveView(instance);
      expect(mockLogger.info.called).to.be.true;
    });
  });

  // ─── Happy Path — showSessionList ──────────────────────────────────────────

  describe("Happy Path — showSessionList", function () {
    it("should post showSessions message to webview", function () {
      const instance = createInstance();
      resolveView(instance);
      instance.showSessionList({ sessions: [], workflows: ["wf1"] });
      expect(mockWebviewView.webview.postMessage.calledOnce).to.be.true;
      expect(mockWebviewView.webview.postMessage.firstCall.args[0]).to.deep.equal({
        type: "showSessions",
        state: { sessions: [], workflows: ["wf1"] },
      });
    });

    it("should store as pendingMessage when view is null", function () {
      const instance = createInstance();
      // Do NOT call resolveWebviewView
      expect(() => {
        instance.showSessionList({ sessions: [], workflows: [] });
      }).to.not.throw();
      // Verify delivery on subsequent resolveWebviewView
    });
  });

  // ─── Happy Path — showSessionDetail ────────────────────────────────────────

  describe("Happy Path — showSessionDetail", function () {
    it("should post showDetail message to webview", function () {
      const instance = createInstance();
      resolveView(instance);
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
      expect(mockWebviewView.webview.postMessage.calledOnce).to.be.true;
      expect(mockWebviewView.webview.postMessage.firstCall.args[0]).to.deep.equal({
        type: "showDetail",
        state,
      });
    });

    it("should store as pendingMessage when view is null", function () {
      const instance = createInstance();
      expect(() => {
        instance.showSessionDetail({
          sessionId: "s1", workflowName: "wf1",
          entryNode: "start", currentState: "start",
          status: "running", pid: 42, eventTypes: [], events: [],
        });
      }).to.not.throw();
    });
  });

  // ─── Happy Path — showNotInitialized ───────────────────────────────────────

  describe("Happy Path — showNotInitialized", function () {
    it("should post showNotInitialized message to webview", function () {
      const instance = createInstance();
      resolveView(instance);
      instance.showNotInitialized();
      expect(mockWebviewView.webview.postMessage.calledOnce).to.be.true;
      expect(mockWebviewView.webview.postMessage.firstCall.args[0]).to.deep.equal(
        { type: "showNotInitialized" }
      );
    });

    it("should store as pendingMessage when view is null", function () {
      const instance = createInstance();
      expect(() => {
        instance.showNotInitialized();
      }).to.not.throw();
    });
  });

  // ─── Idempotency ──────────────────────────────────────────────────────────

  describe("Idempotency", function () {
    it("should overwrite previous pendingMessage with latest call", function () {
      const instance = createInstance();
      instance.showNotInitialized();
      const state = { sessions: [], workflows: [] };
      instance.showSessionList(state);
      resolveView(instance);
      expect(mockWebviewView.webview.postMessage.calledOnce).to.be.true;
      expect(mockWebviewView.webview.postMessage.firstCall.args[0]).to.deep.equal({
        type: "showSessions",
        state,
      });
    });

    it("should handle resolveWebviewView called again after view disposal", function () {
      const instance = createInstance();
      resolveView(instance);
      mockWebviewView.triggerDispose();
      getWebviewContentStub.returns("<html>new</html>");
      const secondView = createStubWebviewView();
      resolveView(instance, secondView);
      expect(secondView.webview.html).to.equal("<html>new</html>");
      expect(secondView.webview.onDidReceiveMessage.called).to.be.true;
    });
  });

  // ─── Mock / Dependency Interaction ─────────────────────────────────────────

  describe("Mock / Dependency Interaction", function () {
    it("should forward webview messages to onDidReceiveMessage subscribers", function () {
      const instance = createInstance();
      resolveView(instance);
      const spy = sinon.spy();
      instance.onDidReceiveMessage(spy);
      mockWebviewView.triggerMessage({
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
      const instance = createInstance();
      resolveView(instance);
      const spy = sinon.spy();
      instance.onDidReceiveMessage(spy);
      mockWebviewView.triggerMessage({ command: "unknownCommand", data: 123 });
      expect(spy.calledOnce).to.be.true;
      expect(spy.firstCall.args[0]).to.deep.equal({ command: "unknownCommand", data: 123 });
    });

    it("should call getWebviewContent with webview and extensionUri", function () {
      const instance = createInstance();
      resolveView(instance);
      expect(getWebviewContentStub.calledOnce).to.be.true;
      const [webviewArg, uriArg] = getWebviewContentStub.firstCall.args;
      expect(webviewArg).to.equal(mockWebviewView.webview);
      expect(uriArg).to.equal(mockExtensionUri);
    });
  });

  // ─── Resource Cleanup ─────────────────────────────────────────────────────

  describe("Resource Cleanup", function () {
    it("should dispose EventEmitter on dispose", function () {
      const instance = createInstance();
      resolveView(instance);
      const spy = sinon.spy();
      instance.onDidReceiveMessage(spy);
      instance.dispose();
      mockWebviewView.triggerMessage({ command: "late" });
      expect(spy.called).to.be.false;
    });

    it("should set view to null on dispose", function () {
      const instance = createInstance();
      resolveView(instance);
      instance.dispose();
      // Subsequent showSessionList stores as pending (no postMessage on disposed view)
      instance.showSessionList({ sessions: [], workflows: [] });
      expect(mockWebviewView.webview.postMessage.called).to.be.false;
    });

    it("should set pendingMessage to null on dispose", function () {
      const instance = createInstance();
      instance.showNotInitialized();
      instance.dispose();
      const newView = createStubWebviewView();
      resolveView(instance, newView);
      expect(newView.webview.postMessage.called).to.be.false;
    });

    it("should not fire onDidReceiveMessage after dispose", function () {
      const instance = createInstance();
      resolveView(instance);
      const spy = sinon.spy();
      instance.onDidReceiveMessage(spy);
      instance.dispose();
      mockWebviewView.triggerMessage({ command: "afterDispose" });
      expect(spy.called).to.be.false;
    });
  });

  // ─── Asynchronous Flow ─────────────────────────────────────────────────────

  describe("Asynchronous Flow", function () {
    it("should transition from notInitialized to sessions when showSessionList called later", function () {
      const instance = createInstance();
      resolveView(instance);
      instance.showNotInitialized();
      const state = { sessions: [], workflows: [] };
      instance.showSessionList(state);
      expect(mockWebviewView.webview.postMessage.calledTwice).to.be.true;
      expect(mockWebviewView.webview.postMessage.firstCall.args[0]).to.deep.equal(
        { type: "showNotInitialized" }
      );
      expect(mockWebviewView.webview.postMessage.secondCall.args[0]).to.deep.equal(
        { type: "showSessions", state }
      );
    });
  });
});
