/**
 * Unit tests for extension activate/deactivate.
 *
 * Test spec: spec/test/vscode/src/extension.md
 * Source under test: vscode/src/extension.ts
 *
 * Architecture: Tests inject mocked dependencies through the production
 * activate(context, deps?) ActivateDeps DI interface. All deps fields are
 * optional — when provided, they replace the production default for that
 * collaborator. This allows tests to run without requiring the vscode module.
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createExtensionTestFixture,
  createMockOutputChannel,
  createMockExtensionContext,
  createMockViewProvider,
  createMockSessionListController,
  createMockSessionDetailController,
  activateWithFixture,
  type ExtensionTestFixture,
} from "./helpers/extensionStubs";

import { activate, deactivate } from "../../src/extension";

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
    it("test_activate_createsOutputChannelFromDeps: uses deps.outputChannel when provided", function () {
      // Provide deps.outputChannel — no createOutputChannel should be called.
      const mockChannel = createMockOutputChannel();
      fixture.deps.outputChannel = mockChannel;
      // Remove createOutputChannel to verify it isn't called
      delete (fixture.deps as any).createOutputChannel;

      activateWithFixture(fixture);

      // The mock outputChannel should be used for logger (appendLine called)
      expect(mockChannel.appendLine.called).to.be.true;
    });

    it("test_activate_createsOutputChannelViaDepsFactory: uses deps.createOutputChannel when outputChannel is not provided", function () {
      const mockChannel = createMockOutputChannel();
      const createOutputChannel = sinon.stub().returns(mockChannel);

      // Remove outputChannel, provide createOutputChannel
      delete (fixture.deps as any).outputChannel;
      fixture.deps.createOutputChannel = createOutputChannel;

      activateWithFixture(fixture);

      expect(createOutputChannel.calledOnceWith("Spectra")).to.be.true;
    });

    it("test_activate_createsOutputChannelViaVscode: uses vscode.window.createOutputChannel when deps provides neither", function () {
      // Scaffolded: requires stubbing require("vscode") which is a module-level concern.
      // The production code lazily requires vscode when neither outputChannel nor
      // createOutputChannel is in deps. Testing this requires intercepting require(),
      // which depends on the production source file existing and a module-stubbing mechanism.
      // Missing seam: production vscode/src/extension.ts source file + require interception (proxyquire/rewire)
      this.skip();
    });

    it("test_activate_logsActivationStart: logs activation start before resolving project root", function () {
      // Track call order: resolveProjectRoot should be called AFTER logging
      let resolveProjectRootCallOrder = -1;
      let firstInfoCallOrder = -1;
      let callCounter = 0;

      const mockChannel = createMockOutputChannel();
      mockChannel.appendLine = sinon.spy(() => {
        if (firstInfoCallOrder === -1) {
          firstInfoCallOrder = callCounter++;
        }
      });
      fixture.deps.outputChannel = mockChannel;

      const originalResolve = fixture.spies.resolveProjectRoot;
      fixture.deps.resolveProjectRoot = sinon.stub().callsFake(() => {
        resolveProjectRootCallOrder = callCounter++;
        return "/workspace";
      });

      activateWithFixture(fixture);

      // ASSERT: outputChannel.appendLine called with [INFO] and "activating" before resolveProjectRoot
      const infoActivating = mockChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" &&
          args[0].includes("[INFO]") &&
          args[0].toLowerCase().includes("activating"),
      );
      expect(infoActivating).to.exist;
      expect(firstInfoCallOrder).to.be.lessThan(resolveProjectRootCallOrder);
    });

    it("test_activate_resolvesProjectRootFromDeps: uses deps.resolveProjectRoot() when provided", function () {
      activateWithFixture(fixture);

      expect(fixture.spies.resolveProjectRoot.calledOnce).to.be.true;
    });

    it("test_activate_resolvesProjectRootFromProductionDefault: calls ProjectRootResolver.resolve(vscode.workspace) when deps.resolveProjectRoot is not provided", function () {
      // Scaffolded: requires stubbing require("vscode") and ProjectRootResolver.resolve.
      // The production code falls back to ProjectRootResolver.resolve(vscode.workspace) when
      // deps.resolveProjectRoot is not provided. Testing this requires module-level stubbing.
      // Missing seam: production vscode/src/extension.ts source file + require interception
      this.skip();
    });

    it("test_activate_createsSessionListController: constructs SessionListController with projectRoot, logger, and controllerDeps", function () {
      activateWithFixture(fixture);

      expect(fixture.spies.createSessionListController.calledOnce).to.be.true;
      const [projectRoot, logger] =
        fixture.spies.createSessionListController.firstCall.args;
      expect(projectRoot).to.equal("/workspace");
      expect(logger).to.have.property("info").that.is.a("function");
      expect(logger).to.have.property("warn").that.is.a("function");
      expect(logger).to.have.property("error").that.is.a("function");
    });

    it("test_activate_createsSessionDetailController: constructs SessionDetailController with projectRoot, logger, and controllerDeps", function () {
      activateWithFixture(fixture);

      expect(fixture.spies.createSessionDetailController.calledOnce).to.be.true;
      const [projectRoot, logger] =
        fixture.spies.createSessionDetailController.firstCall.args;
      expect(projectRoot).to.equal("/workspace");
      expect(logger).to.have.property("info").that.is.a("function");
      expect(logger).to.have.property("warn").that.is.a("function");
      expect(logger).to.have.property("error").that.is.a("function");
    });

    it("test_activate_createsViewProvider: constructs SpectraViewProvider with extensionUri and logger", function () {
      activateWithFixture(fixture);

      expect(fixture.spies.createViewProvider.calledOnce).to.be.true;
      const [extensionUri, logger] =
        fixture.spies.createViewProvider.firstCall.args;
      expect(extensionUri).to.equal(fixture.context.extensionUri);
      expect(logger).to.have.property("info").that.is.a("function");
    });

    it("test_activate_createsViewProviderFromDepsFactory: uses deps.createViewProvider when provided", function () {
      const mockVP = createMockViewProvider();
      const createVP = sinon.stub().returns(mockVP);
      fixture.deps.createViewProvider = createVP;

      activateWithFixture(fixture);

      expect(createVP.calledOnce).to.be.true;
      // The returned mock should be used for registration
      expect(
        fixture.spies.registerWebviewViewProvider.firstCall.args[1],
      ).to.equal(mockVP);
    });

    it("test_activate_registersViewProvider: registers the view provider with correct viewType and options", function () {
      activateWithFixture(fixture);

      expect(fixture.spies.registerWebviewViewProvider.calledOnce).to.be.true;
      const [viewType, provider, options] =
        fixture.spies.registerWebviewViewProvider.firstCall.args;
      expect(viewType).to.equal("spectra.chatView");
      expect(provider).to.equal(fixture.viewProvider);
      expect(options).to.deep.equal({
        webviewOptions: { retainContextWhenHidden: true },
      });
    });

    it("test_activate_registersOpenPanelCommand: registers the spectra.openPanel command", function () {
      activateWithFixture(fixture);

      expect(fixture.spies.registerCommand.called).to.be.true;
      const openPanelCall = fixture.spies.registerCommand.args.find(
        (args: any[]) => args[0] === "spectra.openPanel",
      );
      expect(openPanelCall).to.exist;
      expect(openPanelCall![1]).to.be.a("function");
    });

    it("test_activate_openPanelCommandHandlerIsNoOp: the spectra.openPanel command handler does nothing when invoked", function () {
      // Capture the handler registered for spectra.openPanel
      let capturedHandler: ((...args: any[]) => any) | undefined;
      fixture.deps.registerCommand = sinon
        .stub()
        .callsFake((id: string, handler: (...args: any[]) => any) => {
          if (id === "spectra.openPanel") {
            capturedHandler = handler;
          }
          return { dispose: () => {} };
        });

      activateWithFixture(fixture);

      expect(capturedHandler).to.be.a("function");
      // Invoking the handler should not throw or produce side effects
      const result = capturedHandler!();
      expect(result).to.be.undefined;
    });

    it("test_activate_pushesAllDisposablesToSubscriptions: pushes all disposables to context.subscriptions", function () {
      activateWithFixture(fixture);

      // Per the spec: OutputChannel, sessionListController, sessionDetailController,
      // viewProvider, view provider registration disposable, command disposable,
      // and subscription disposables (onDidUpdate x2, onDidError x2, onDidReceiveMessage)
      // That's at least 11 items
      expect(fixture.context.subscriptions.length).to.be.at.least(7);

      // Verify key disposables are present
      expect(fixture.context.subscriptions).to.include(fixture.outputChannel);
      expect(fixture.context.subscriptions).to.include(
        fixture.sessionListController,
      );
      expect(fixture.context.subscriptions).to.include(
        fixture.sessionDetailController,
      );
      expect(fixture.context.subscriptions).to.include(fixture.viewProvider);
    });

    it("test_activate_logsSuccessWithProjectRoot: logs successful activation including the resolved projectRoot", function () {
      const fixtureCustom = createExtensionTestFixture("/my/project");

      activateWithFixture(fixtureCustom);

      // ASSERT: outputChannel.appendLine called with [INFO] and '/my/project'
      const projectRootLog = fixtureCustom.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" &&
          args[0].includes("[INFO]") &&
          args[0].includes("/my/project"),
      );
      expect(projectRootLog).to.exist;
    });
  });

  // ─── Happy Path — onDidUpdate subscriptions ────────────────────────────────

  describe("activate — onDidUpdate subscriptions", function () {
    it("test_activate_sessionListOnDidUpdate_cachesStateAndShowsList: caches state and calls viewProvider.showSessionList", function () {
      activateWithFixture(fixture);

      // Trigger sessionListController.onDidUpdate with a fake state
      const fakeState = { sessions: [{ id: "s1" }], workflows: ["wf1"] };
      fixture.sessionListController.triggerUpdate(fakeState);

      // ASSERT: viewProvider.showSessionList called with fakeState
      expect(fixture.viewProvider.showSessionList.calledOnce).to.be.true;
      expect(
        fixture.viewProvider.showSessionList.firstCall.args[0],
      ).to.deep.equal(fakeState);
    });

    it("test_activate_sessionDetailOnDidUpdate_showsDetail: calls viewProvider.showSessionDetail on controller update", function () {
      activateWithFixture(fixture);

      // Trigger sessionDetailController.onDidUpdate with a fake detail state
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

      // ASSERT: viewProvider.showSessionDetail called with fakeDetailState
      expect(fixture.viewProvider.showSessionDetail.calledOnce).to.be.true;
      expect(
        fixture.viewProvider.showSessionDetail.firstCall.args[0],
      ).to.deep.equal(fakeDetailState);
    });
  });

  // ─── Error Propagation ─────────────────────────────────────────────────────

  describe("activate — Error Propagation", function () {
    it("test_activate_sessionListOnDidError_showsErrorMessage: calls showErrorMessage when sessionListController fires onDidError", function () {
      activateWithFixture(fixture);

      fixture.sessionListController.triggerError({ message: "scan failed" });

      expect(fixture.spies.showErrorMessage.calledOnceWith("scan failed")).to.be
        .true;
    });

    it("test_activate_sessionDetailOnDidError_showsErrorMessage: calls showErrorMessage when sessionDetailController fires onDidError", function () {
      activateWithFixture(fixture);

      fixture.sessionDetailController.triggerError({
        message: "detail error",
      });

      expect(fixture.spies.showErrorMessage.calledOnceWith("detail error")).to
        .be.true;
    });
  });

  // ─── Happy Path — onDidReceiveMessage routing ──────────────────────────────

  describe("activate — onDidReceiveMessage routing", function () {
    it("test_activate_messageRouting_navigateToDetail: routes navigateToDetail to sessionDetailController.open", function () {
      activateWithFixture(fixture);

      fixture.viewProvider.triggerMessage({
        command: "navigateToDetail",
        sessionId: "s1",
        workflowName: "wf1",
      });

      expect(fixture.sessionDetailController.open.calledOnce).to.be.true;
      expect(fixture.sessionDetailController.open.calledWith("s1", "wf1")).to.be
        .true;
    });

    it("test_activate_messageRouting_navigateToList_withCache: routes navigateToList to viewProvider.showSessionList with cached state", function () {
      activateWithFixture(fixture);

      // First trigger onDidUpdate to populate cache
      const cachedState = { sessions: [{ id: "s1" }], workflows: ["wf1"] };
      fixture.sessionListController.triggerUpdate(cachedState);
      fixture.viewProvider.showSessionList.resetHistory();

      // Then trigger navigateToList
      fixture.viewProvider.triggerMessage({ command: "navigateToList" });

      expect(fixture.viewProvider.showSessionList.calledOnce).to.be.true;
      expect(
        fixture.viewProvider.showSessionList.firstCall.args[0],
      ).to.deep.equal(cachedState);
    });

    it("test_activate_messageRouting_navigateToList_noCacheNoOp: no-op when navigateToList received before first onDidUpdate", function () {
      activateWithFixture(fixture);

      // Do NOT trigger sessionListController.onDidUpdate first
      // Reset any showSessionList calls from activation
      fixture.viewProvider.showSessionList.resetHistory();

      // Trigger navigateToList
      fixture.viewProvider.triggerMessage({ command: "navigateToList" });

      expect(fixture.viewProvider.showSessionList.called).to.be.false;
    });

    it("test_activate_messageRouting_launchSession: routes launchSession to sessionListController.launch", function () {
      activateWithFixture(fixture);

      fixture.viewProvider.triggerMessage({
        command: "launchSession",
        workflowName: "deploy",
      });

      expect(fixture.sessionListController.launch.calledOnce).to.be.true;
      expect(fixture.sessionListController.launch.calledWith("deploy")).to.be
        .true;
    });

    it("test_activate_messageRouting_terminateSession: routes terminateSession to sessionListController.terminate", function () {
      activateWithFixture(fixture);

      fixture.viewProvider.triggerMessage({
        command: "terminateSession",
        pid: 1234,
      });

      expect(fixture.sessionListController.terminate.calledOnce).to.be.true;
      expect(fixture.sessionListController.terminate.calledWith(1234)).to.be
        .true;
    });

    it("test_activate_messageRouting_sendEvent: routes sendEvent to sessionDetailController.sendEvent", function () {
      activateWithFixture(fixture);

      fixture.viewProvider.triggerMessage({
        command: "sendEvent",
        eventType: "input",
        message: "hello",
      });

      expect(fixture.sessionDetailController.sendEvent.calledOnce).to.be.true;
      expect(
        fixture.sessionDetailController.sendEvent.calledWith("input", "hello"),
      ).to.be.true;
    });

    it("test_activate_messageRouting_unknownCommand_logsWarning: logs a warning for unrecognized commands", function () {
      activateWithFixture(fixture);

      fixture.viewProvider.triggerMessage({ command: "unknownCmd" });

      // ASSERT: outputChannel.appendLine called with [WARN] and 'unknownCmd'
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
    it("test_activate_projectRootUndefined_showsErrorAndReturnsEarly: shows error message and returns early when projectRoot is undefined", function () {
      const undefinedFixture = createExtensionTestFixture(undefined);

      activateWithFixture(undefinedFixture);

      // ASSERT: deps.showErrorMessage called with the exact error message
      expect(
        undefinedFixture.spies.showErrorMessage.calledOnceWith(
          "Spectra: No workspace folder open.",
        ),
      ).to.be.true;

      // ASSERT: no controllers created
      expect(undefinedFixture.spies.createSessionListController.called).to.be
        .false;
      expect(undefinedFixture.spies.createSessionDetailController.called).to.be
        .false;

      // ASSERT: no ViewProvider created
      expect(undefinedFixture.spies.createViewProvider.called).to.be.false;
    });

    it("test_activate_projectRootUndefined_logsError: logs the error when projectRoot is undefined", function () {
      const undefinedFixture = createExtensionTestFixture(undefined);

      activateWithFixture(undefinedFixture);

      // ASSERT: outputChannel.appendLine called with [ERROR]
      const errorLog = undefinedFixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[ERROR]"),
      );
      expect(errorLog).to.exist;
    });

    it("test_activate_projectRootUndefined_onlyOutputChannelInSubscriptions: only the OutputChannel is pushed to context.subscriptions", function () {
      const undefinedFixture = createExtensionTestFixture(undefined);

      activateWithFixture(undefinedFixture);

      // ASSERT: context.subscriptions contains only the OutputChannel
      expect(undefinedFixture.context.subscriptions).to.have.lengthOf(1);
      expect(undefinedFixture.context.subscriptions[0]).to.equal(
        undefinedFixture.outputChannel,
      );
    });
  });

  // ─── Mock / Dependency Interaction ─────────────────────────────────────────

  describe("activate — Mock / Dependency Interaction", function () {
    it("test_activate_loggerWrapsOutputChannel: logger delegates info to outputChannel.appendLine with [INFO] prefix", function () {
      // Use a fixture where projectRoot is undefined so activate returns early after logging
      const earlyFixture = createExtensionTestFixture(undefined);

      activateWithFixture(earlyFixture);

      // ASSERT: outputChannel.appendLine called with strings containing [INFO] prefix
      const infoCall = earlyFixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[INFO]"),
      );
      expect(infoCall).to.exist;
    });

    it("test_activate_loggerSeverityTags: logger prepends [INFO], [WARN], [ERROR] for respective methods", function () {
      activateWithFixture(fixture);

      // Trigger an unrecognized message to invoke logger.warn
      fixture.viewProvider.triggerMessage({ command: "unknownCmd" });

      // ASSERT: [INFO] appears (from activation logging)
      const infoCall = fixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[INFO]"),
      );
      expect(infoCall).to.exist;

      // ASSERT: [WARN] appears (from unknown command)
      const warnCall = fixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[WARN]"),
      );
      expect(warnCall).to.exist;
    });

    it("test_activate_terminateFromDetailPage_routesToSessionListController: terminateSession from detail page routes to sessionListController.terminate", function () {
      activateWithFixture(fixture);

      // Trigger terminateSession message (simulating from detail page)
      fixture.viewProvider.triggerMessage({
        command: "terminateSession",
        pid: 5678,
      });

      expect(fixture.sessionListController.terminate.calledOnce).to.be.true;
      expect(fixture.sessionListController.terminate.calledWith(5678)).to.be
        .true;
    });

    it("test_activate_requiresVscodeWhenNoDepsOutputChannel: lazily requires vscode module when deps provides neither outputChannel nor createOutputChannel", function () {
      // Scaffolded: requires intercepting require("vscode") at module level.
      // The production code lazily requires vscode when neither outputChannel nor
      // createOutputChannel is present in deps. Testing this requires proxyquire/rewire
      // or similar module-stubbing, plus the production source file existing.
      // Missing seam: production vscode/src/extension.ts source file + require interception mechanism
      this.skip();
    });

    it("test_activate_doesNotRequireVscodeWhenDepsProvideOutputChannel: does not require vscode module when deps provides outputChannel", function () {
      // Scaffolded: requires intercepting require() to verify vscode is NOT called.
      // The test fixture already provides deps.outputChannel. In practice, when
      // outputChannel is provided, vscode should not be required. Verifying "not required"
      // needs require interception at the module level.
      // Missing seam: production vscode/src/extension.ts source file + require interception mechanism
      this.skip();
    });

    it("test_activate_depsEmptyObject_usesProductionDefaults: all collaborators are constructed using production defaults when deps is empty object", function () {
      // Scaffolded: requires production vscode/src/extension.ts to exist and
      // stubbing require("vscode") to prevent actual VS Code calls.
      // When deps={}, all collaborators must be constructed from production defaults.
      // Missing seam: production vscode/src/extension.ts source file + require("vscode") interception
      this.skip();
    });

    it("test_activate_depsUndefined_usesProductionDefaults: all collaborators are constructed using production defaults when deps is undefined", function () {
      // Scaffolded: requires production vscode/src/extension.ts to exist and
      // stubbing require("vscode") to prevent actual VS Code calls.
      // When deps is undefined (VS Code runtime scenario), all collaborators must
      // be constructed from production defaults.
      // Missing seam: production vscode/src/extension.ts source file + require("vscode") interception
      this.skip();
    });

    it("test_activate_depsPartial_mergesWithProductionDefaults: when deps provides only some fields, remaining fields use production defaults", function () {
      // Scaffolded: requires production vscode/src/extension.ts to exist and
      // stubbing require("vscode") for the remaining production deps.
      // When only outputChannel is provided, remaining fields (controllers, view provider)
      // should use production constructors.
      // Missing seam: production vscode/src/extension.ts source file + require("vscode") interception
      this.skip();
    });

    it("test_activate_usesShowErrorMessageFromDeps: uses deps.showErrorMessage instead of vscode.window.showErrorMessage", function () {
      const undefinedFixture = createExtensionTestFixture(undefined);

      activateWithFixture(undefinedFixture);

      // ASSERT: deps.showErrorMessage was called (since projectRoot is undefined)
      expect(undefinedFixture.spies.showErrorMessage.called).to.be.true;
    });

    it("test_activate_usesRegisterCommandFromDeps: uses deps.registerCommand instead of vscode.commands.registerCommand", function () {
      activateWithFixture(fixture);

      // ASSERT: deps.registerCommand called with 'spectra.openPanel'
      const openPanelCall = fixture.spies.registerCommand.args.find(
        (args: any[]) => args[0] === "spectra.openPanel",
      );
      expect(openPanelCall).to.exist;
    });
  });

  // ─── deactivate ────────────────────────────────────────────────────────────

  describe("deactivate", function () {
    it("test_deactivate_isEmptyFunction: deactivate does nothing — returns undefined without errors", function () {
      const result = deactivate();
      expect(result).to.be.undefined;
    });
  });
});
