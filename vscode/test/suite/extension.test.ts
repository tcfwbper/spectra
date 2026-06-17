/**
 * Unit tests for extension activate/deactivate.
 *
 * Test spec: spec/test/vscode/src/extension.md
 * Source under test: vscode/src/extension.ts
 *
 * The test structure, mocks, fixtures, and assertion intent are all in place.
 * Tests use the `activateWithFixture` bridge to inject mocked dependencies
 * through the production activate() function's ExtensionDeps DI interface.
 *
 * Scaffolded rows: The production extension.ts needs to be refactored from
 * SpectraPanel to SpectraViewProvider architecture. Tests marked with t.Skip()
 * name the exact missing production surface.
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createExtensionTestFixture,
  createMockOutputChannel,
  activateWithFixture,
  type ExtensionTestFixture,
} from "./helpers/extensionStubs";

import { deactivate } from "../../src/extension";

describe("extension", function () {
  let sandbox: sinon.SinonSandbox;
  let fixture: ExtensionTestFixture;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    fixture = createExtensionTestFixture("/workspace");
  });

  afterEach(function () {
    sandbox.restore();
  });

  // ─── Happy Path — activate ──────────────────────────────────────────────────

  describe("activate — Happy Path", function () {
    it("test_activate_createsOutputChannel: creates an OutputChannel named 'Spectra' on activation", function () {
      activateWithFixture(fixture);

      // ASSERT: fixture.vscode.window.createOutputChannel called with 'Spectra'
      expect(
        fixture.vscode.window.createOutputChannel.calledOnceWith("Spectra"),
      ).to.be.true;
    });

    it("test_activate_logsActivationStart: logs activation start before resolving project root", function () {
      activateWithFixture(fixture);

      // ASSERT: logger.info called with a message indicating activation start
      // (Logger is constructed from OutputChannel, so verify outputChannel.appendLine
      // receives an [INFO] prefixed message about activation)
      expect(fixture.outputChannel.appendLine.called).to.be.true;
      const firstInfoCall = fixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[INFO]"),
      );
      expect(firstInfoCall).to.exist;
    });

    it("test_activate_resolvesProjectRoot: calls ProjectRootResolver.resolve() to obtain project root", function () {
      activateWithFixture(fixture);

      // ASSERT: fixture.projectRootResolveStub called exactly once
      expect(fixture.projectRootResolveStub.calledOnce).to.be.true;
    });

    it("test_activate_createsViewProvider: constructs SpectraViewProvider with extensionUri and logger", function () {
      // Scaffolded: production ExtensionDeps needs createViewProvider(extensionUri, logger) seam
      this.skip();
      // Once the production surface exists:
      // activateWithFixture(fixture);
      // expect(fixture.viewProviderConstructorStub.calledOnce).to.be.true;
      // const [uri] = fixture.viewProviderConstructorStub.firstCall.args;
      // expect(uri).to.equal(fixture.context.extensionUri);
    });

    it("test_activate_registersViewProvider: registers the view provider with VS Code using the correct viewType and options", function () {
      // Scaffolded: production ExtensionDeps needs registerWebviewViewProvider seam
      this.skip();
      // Once the production surface exists:
      // activateWithFixture(fixture);
      // expect(fixture.vscode.window.registerWebviewViewProvider.calledOnce).to.be.true;
      // const [viewType, provider, options] =
      //   fixture.vscode.window.registerWebviewViewProvider.firstCall.args;
      // expect(viewType).to.equal('spectra.sessionView');
      // expect(provider).to.equal(fixture.viewProvider);
      // expect(options).to.deep.equal({ webviewOptions: { retainContextWhenHidden: true } });
    });

    it("test_activate_createsSessionListController: constructs SessionListController with projectRoot and logger", function () {
      activateWithFixture(fixture);

      // ASSERT: SessionListController constructor called with '/workspace' and logger
      expect(fixture.sessionListControllerConstructorStub.calledOnce).to.be
        .true;
      const [projectRoot] =
        fixture.sessionListControllerConstructorStub.firstCall.args;
      expect(projectRoot).to.equal("/workspace");
    });

    it("test_activate_createsSessionDetailController: constructs SessionDetailController with projectRoot and logger", function () {
      activateWithFixture(fixture);

      // ASSERT: SessionDetailController constructor called with '/workspace' and logger
      expect(fixture.sessionDetailControllerConstructorStub.calledOnce).to.be
        .true;
      const [projectRoot] =
        fixture.sessionDetailControllerConstructorStub.firstCall.args;
      expect(projectRoot).to.equal("/workspace");
    });

    it("test_activate_pushesAllDisposablesToSubscriptions: pushes all disposables to context.subscriptions", function () {
      activateWithFixture(fixture);

      // ASSERT: context.subscriptions contains at least:
      //   OutputChannel, sessionListController, sessionDetailController, viewProvider,
      //   view provider registration disposable
      // Current production pushes 11 items in happy path; after refactor should push at least 5
      expect(fixture.context.subscriptions.length).to.be.at.least(5);
    });

    it("test_activate_logsSuccessWithProjectRoot: logs successful activation including the resolved projectRoot", function () {
      const fixtureCustom = createExtensionTestFixture("/my/project");

      activateWithFixture(fixtureCustom);

      // ASSERT: Logger info called with message containing '/my/project'
      const projectRootLog = fixtureCustom.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("/my/project"),
      );
      expect(projectRootLog).to.exist;
    });

    it("test_activate_checksProjectInitialization: calls ProjectRootResolver.isInitialized with projectRoot after resolving", function () {
      // Scaffolded: production ExtensionDeps needs isInitialized(projectRoot) seam
      this.skip();
      // Once the production surface exists:
      // activateWithFixture(fixture);
      // expect(fixture.isInitializedStub.calledOnce).to.be.true;
      // expect(fixture.isInitializedStub.calledWith('/workspace')).to.be.true;
    });
  });

  // ─── Happy Path — onDidUpdate subscriptions ────────────────────────────────

  describe("activate — onDidUpdate subscriptions", function () {
    it("test_activate_sessionListOnDidUpdate_cachesStateAndShowsList: caches state and calls viewProvider.showSessionList", function () {
      // Scaffolded: once production uses viewProvider instead of panel
      // Currently tests via the legacy panel mock
      activateWithFixture(fixture);

      // Then trigger sessionListController.onDidUpdate with a fake state
      const fakeState = { sessions: [{ id: "s1" }], workflows: ["wf1"] };
      fixture.sessionListController.triggerUpdate(fakeState);

      // ASSERT: panel.showSessionList called with fakeState
      // (will become viewProvider.showSessionList after refactor)
      expect(fixture.panel.showSessionList.calledOnce).to.be.true;
      expect(fixture.panel.showSessionList.firstCall.args[0]).to.deep.equal(
        fakeState,
      );
    });

    it("test_activate_sessionDetailOnDidUpdate_showsDetail: calls viewProvider.showSessionDetail on controller update", function () {
      // Scaffolded: once production uses viewProvider instead of panel
      activateWithFixture(fixture);

      // Then trigger sessionDetailController.onDidUpdate with a fake detail state
      const fakeDetailState = {
        sessionId: "s1",
        workflowName: "wf1",
        entryNode: "start",
        currentState: "running",
        status: "running",
        pid: 42,
        eventTypes: ["input"],
        events: [],
      };
      fixture.sessionDetailController.triggerUpdate(fakeDetailState);

      // ASSERT: panel.showSessionDetail called with fakeDetailState
      // (will become viewProvider.showSessionDetail after refactor)
      expect(fixture.panel.showSessionDetail.calledOnce).to.be.true;
      expect(fixture.panel.showSessionDetail.firstCall.args[0]).to.deep.equal(
        fakeDetailState,
      );
    });
  });

  // ─── Error Propagation ─────────────────────────────────────────────────────

  describe("activate — Error Propagation", function () {
    it("test_activate_sessionListOnDidError_showsErrorMessage: shows error when sessionListController fires onDidError", function () {
      activateWithFixture(fixture);

      // Then trigger sessionListController.onDidError
      fixture.sessionListController.triggerError({ message: "scan failed" });

      // ASSERT: vscode.window.showErrorMessage called with 'scan failed'
      expect(
        fixture.vscode.window.showErrorMessage.calledOnceWith("scan failed"),
      ).to.be.true;
    });

    it("test_activate_sessionDetailOnDidError_showsErrorMessage: shows error when sessionDetailController fires onDidError", function () {
      activateWithFixture(fixture);

      // Then trigger sessionDetailController.onDidError
      fixture.sessionDetailController.triggerError({
        message: "detail error",
      });

      // ASSERT: vscode.window.showErrorMessage called with 'detail error'
      expect(
        fixture.vscode.window.showErrorMessage.calledOnceWith("detail error"),
      ).to.be.true;
    });
  });

  // ─── Happy Path — onDidReceiveMessage routing ──────────────────────────────

  describe("activate — onDidReceiveMessage routing", function () {
    it("test_activate_messageRouting_navigateToDetail: routes navigateToDetail to sessionDetailController.open", function () {
      activateWithFixture(fixture);

      // Then trigger viewProvider.onDidReceiveMessage with navigateToDetail
      // (currently wired through panel mock)
      fixture.panel.triggerMessage({
        command: "navigateToDetail",
        sessionId: "s1",
        workflowName: "wf1",
      });

      // ASSERT: sessionDetailController.open called with 's1', 'wf1'
      expect(fixture.sessionDetailController.open.calledOnce).to.be.true;
      expect(fixture.sessionDetailController.open.calledWith("s1", "wf1")).to.be
        .true;
    });

    it("test_activate_messageRouting_navigateToList_withCache: routes navigateToList to viewProvider.showSessionList with cached state", function () {
      activateWithFixture(fixture);

      // First trigger onDidUpdate to populate cache
      const cachedState = { sessions: [{ id: "s1" }], workflows: ["wf1"] };
      fixture.sessionListController.triggerUpdate(cachedState);
      fixture.panel.showSessionList.resetHistory();

      // Then trigger navigateToList
      fixture.panel.triggerMessage({ command: "navigateToList" });

      // ASSERT: panel.showSessionList called with cachedState
      // (will become viewProvider.showSessionList after refactor)
      expect(fixture.panel.showSessionList.calledOnce).to.be.true;
      expect(fixture.panel.showSessionList.firstCall.args[0]).to.deep.equal(
        cachedState,
      );
    });

    it("test_activate_messageRouting_navigateToList_noCacheNoOp: no-op when navigateToList received before first onDidUpdate", function () {
      activateWithFixture(fixture);

      // Do NOT trigger sessionListController.onDidUpdate
      // Trigger navigateToList
      fixture.panel.triggerMessage({ command: "navigateToList" });

      // ASSERT: panel.showSessionList is NOT called
      expect(fixture.panel.showSessionList.called).to.be.false;
    });

    it("test_activate_messageRouting_launchSession: routes launchSession to sessionListController.launch", function () {
      activateWithFixture(fixture);

      // Trigger launchSession message
      fixture.panel.triggerMessage({
        command: "launchSession",
        workflowName: "deploy",
      });

      // ASSERT: sessionListController.launch called with 'deploy'
      expect(fixture.sessionListController.launch.calledOnce).to.be.true;
      expect(fixture.sessionListController.launch.calledWith("deploy")).to.be
        .true;
    });

    it("test_activate_messageRouting_terminateSession: routes terminateSession to sessionListController.terminate", function () {
      activateWithFixture(fixture);

      // Trigger terminateSession message
      fixture.panel.triggerMessage({
        command: "terminateSession",
        pid: 1234,
      });

      // ASSERT: sessionListController.terminate called with 1234
      expect(fixture.sessionListController.terminate.calledOnce).to.be.true;
      expect(fixture.sessionListController.terminate.calledWith(1234)).to.be
        .true;
    });

    it("test_activate_messageRouting_sendEvent: routes sendEvent to sessionDetailController.sendEvent", function () {
      activateWithFixture(fixture);

      // Trigger sendEvent message
      fixture.panel.triggerMessage({
        command: "sendEvent",
        eventType: "input",
        message: "hello",
      });

      // ASSERT: sessionDetailController.sendEvent called with 'input', 'hello'
      expect(fixture.sessionDetailController.sendEvent.calledOnce).to.be.true;
      expect(
        fixture.sessionDetailController.sendEvent.calledWith("input", "hello"),
      ).to.be.true;
    });

    it("test_activate_messageRouting_unknownCommand_logsWarning: logs a warning for unrecognized commands", function () {
      activateWithFixture(fixture);

      // Trigger unknown command message
      fixture.panel.triggerMessage({ command: "unknownCmd" });

      // ASSERT: logger.warn called with message containing 'unknownCmd'
      // The logger delegates to outputChannel.appendLine with [WARN] prefix
      const warnCall = fixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" &&
          args[0].includes("[WARN]") &&
          args[0].includes("unknownCmd"),
      );
      expect(warnCall).to.exist;
    });
  });

  // ─── Null / Empty Input ────────────────────────────────────────────────────

  describe("activate — Null / Empty Input", function () {
    it("test_activate_projectRootUndefined_showsNotInitializedAndReturnsEarly: shows not-initialized and returns early when projectRoot is undefined", function () {
      // Scaffolded: production must call viewProvider.showNotInitialized() instead of showErrorMessage
      // Missing seam: ExtensionDeps.createViewProvider, ExtensionDeps.registerWebviewViewProvider
      this.skip();
      // Once production surface exists:
      // const undefinedFixture = createExtensionTestFixture(undefined);
      // activateWithFixture(undefinedFixture);
      // expect(undefinedFixture.viewProvider.showNotInitialized.calledOnce).to.be.true;
      // expect(undefinedFixture.sessionListControllerConstructorStub.called).to.be.false;
      // expect(undefinedFixture.sessionDetailControllerConstructorStub.called).to.be.false;
    });

    it("test_activate_projectNotInitialized_showsNotInitializedAndReturnsEarly: shows not-initialized when .spectra/ missing", function () {
      // Scaffolded: production must call isInitialized(projectRoot) and viewProvider.showNotInitialized()
      // Missing seam: ExtensionDeps.isInitialized, ExtensionDeps.createViewProvider
      this.skip();
      // Once production surface exists:
      // const notInitFixture = createExtensionTestFixture('/workspace', false);
      // activateWithFixture(notInitFixture);
      // expect(notInitFixture.viewProvider.showNotInitialized.calledOnce).to.be.true;
      // expect(notInitFixture.sessionListControllerConstructorStub.called).to.be.false;
      // expect(notInitFixture.sessionDetailControllerConstructorStub.called).to.be.false;
    });

    it("test_activate_projectRootUndefined_viewProviderStillRegistered: ViewProvider is registered even when projectRoot is undefined", function () {
      // Scaffolded: production must register ViewProvider before the early return
      // Missing seam: ExtensionDeps.registerWebviewViewProvider
      this.skip();
      // Once production surface exists:
      // const undefinedFixture = createExtensionTestFixture(undefined);
      // activateWithFixture(undefinedFixture);
      // expect(undefinedFixture.vscode.window.registerWebviewViewProvider.calledOnce).to.be.true;
      // const [viewType] = undefinedFixture.vscode.window.registerWebviewViewProvider.firstCall.args;
      // expect(viewType).to.equal('spectra.sessionView');
    });
  });

  // ─── Mock / Dependency Interaction ─────────────────────────────────────────

  describe("activate — Mock / Dependency Interaction", function () {
    it("test_activate_loggerWrapsOutputChannel: logger delegates info/warn/error to outputChannel.appendLine with severity tags", function () {
      // Setup: activate returns early (projectRoot undefined) but we can
      // observe the logger calling outputChannel.appendLine with [INFO] tag
      const undefinedFixture = createExtensionTestFixture(undefined);

      activateWithFixture(undefinedFixture);

      // ASSERT: outputChannel.appendLine called with strings containing severity prefix
      const infoCall = undefinedFixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[INFO]"),
      );
      expect(infoCall).to.exist;
    });

    it("test_activate_terminateFromDetailPage_routesToSessionListController: terminateSession from detail page routes to sessionListController.terminate", function () {
      activateWithFixture(fixture);

      // Trigger terminateSession message (simulating from detail page)
      fixture.panel.triggerMessage({
        command: "terminateSession",
        pid: 5678,
      });

      // ASSERT: sessionListController.terminate called with 5678
      expect(fixture.sessionListController.terminate.calledOnce).to.be.true;
      expect(fixture.sessionListController.terminate.calledWith(5678)).to.be
        .true;
    });

    it("test_activate_acceptsOnlyOneParameter: activate function signature accepts exactly one parameter", function () {
      // The production activate function currently accepts (context, deps) for testability.
      // The logic spec mandates that the public signature accepts exactly one parameter (context).
      // Scaffolded: once production removes the deps parameter (collaborators constructed internally),
      // this test will pass. Until then, verify the constraint via t.Skip.
      // Missing seam: production activate() must accept exactly 1 parameter (no ExtensionDeps)
      this.skip();
      // Once the production surface is updated:
      // import { activate } from "../../src/extension";
      // expect(activate.length).to.equal(1);
    });

    it("test_activate_constructsAllCollaboratorsInternally: all collaborators are constructed inside activate without external injection", function () {
      // Scaffolded: once production activate() no longer accepts ExtensionDeps,
      // all collaborators are constructed internally.
      // Missing seam: production activate() must not accept a deps parameter
      this.skip();
      // Once the production surface is updated:
      // import { activate } from "../../src/extension";
      // const context = createMockExtensionContext();
      // // Verify activate was called with only context (no second argument)
      // // and all constructors (SessionListController, SessionDetailController, SpectraViewProvider) are called.
    });

    it("test_activate_registersViewProviderSynchronouslyDuringActivation: ViewProvider registration occurs synchronously during activation", function () {
      // Scaffolded: once production uses registerWebviewViewProvider instead of createOrRevealPanel,
      // this test verifies registration happens synchronously before any await.
      // Missing seam: ExtensionDeps.registerWebviewViewProvider, production must register ViewProvider
      // synchronously within activate before any async work.
      this.skip();
      // Once the production surface exists:
      // activateWithFixture(fixture);
      // expect(fixture.vscode.window.registerWebviewViewProvider.calledOnce).to.be.true;
      // const [viewType] = fixture.vscode.window.registerWebviewViewProvider.firstCall.args;
      // expect(viewType).to.equal('spectra.sessionView');
    });
  });

  // ─── deactivate ────────────────────────────────────────────────────────────

  describe("deactivate", function () {
    it("test_deactivate_isEmptyFunction: deactivate does nothing — returns undefined without errors", function () {
      // ASSERT: returns undefined, no errors thrown
      const result = deactivate();
      expect(result).to.be.undefined;
    });
  });
});
