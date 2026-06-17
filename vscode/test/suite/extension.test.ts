/**
 * Unit tests for extension activate/deactivate.
 *
 * Test spec: spec/test/vscode/src/extension.md
 * Source under test: vscode/src/extension.ts
 *
 * Scaffolded: The production extension.ts file does not yet exist.
 * All tests are marked with t.skip() naming the missing production surface.
 * Once extension.ts is implemented and exports `activate` and `deactivate`,
 * these tests can be unwired from skip and run against the real module.
 *
 * The test structure, mocks, fixtures, and assertion intent are all in place
 * so that the engineer implementing extension.ts can simply remove the skips
 * and adjust the import/wiring mechanism (e.g., proxyquire or DI parameter)
 * to inject the mocked dependencies.
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createExtensionTestFixture,
  createMockOutputChannel,
  type ExtensionTestFixture,
} from "./helpers/extensionStubs";

/**
 * NOTE: When extension.ts exists, import activate/deactivate here:
 *
 *   import { activate, deactivate } from '../../src/extension';
 *
 * The exact injection mechanism (proxyquire, DI params, or module-level
 * stubs) will be determined by the extension.ts implementation. Tests
 * below use the fixture pattern and assume a dependency-injection approach
 * or a module-stubbing approach (like proxyquire) to wire mocks.
 */

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
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: fixture.vscode.window.createOutputChannel called with 'Spectra'
      expect(
        fixture.vscode.window.createOutputChannel.calledOnceWith("Spectra"),
      ).to.be.true;
    });

    it("test_activate_logsActivationStart: logs activation start before resolving project root", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: fixture.projectRootResolveStub called exactly once
      expect(fixture.projectRootResolveStub.calledOnce).to.be.true;
    });

    it("test_activate_createsSessionListController: constructs SessionListController with projectRoot and logger", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: SessionListController constructor called with '/workspace' and logger
      expect(fixture.sessionListControllerConstructorStub.calledOnce).to.be
        .true;
      const [projectRoot] =
        fixture.sessionListControllerConstructorStub.firstCall.args;
      expect(projectRoot).to.equal("/workspace");
    });

    it("test_activate_createsSessionDetailController: constructs SessionDetailController with projectRoot and logger", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: SessionDetailController constructor called with '/workspace' and logger
      expect(fixture.sessionDetailControllerConstructorStub.calledOnce).to.be
        .true;
      const [projectRoot] =
        fixture.sessionDetailControllerConstructorStub.firstCall.args;
      expect(projectRoot).to.equal("/workspace");
    });

    it("test_activate_callsCreateOrReveal: calls SpectraPanel.createOrReveal with context, extensionUri, and logger", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: SpectraPanel.createOrReveal called with context, context.extensionUri, logger
      expect(fixture.spectraPanelCreateOrRevealStub.calledOnce).to.be.true;
      const [ctx, uri] = fixture.spectraPanelCreateOrRevealStub.firstCall.args;
      expect(ctx).to.equal(fixture.context);
      expect(uri).to.equal(fixture.context.extensionUri);
    });

    it("test_activate_registersOpenPanelCommand: registers the spectra.openPanel command", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: vscode.commands.registerCommand called with 'spectra.openPanel' and a handler
      expect(fixture.vscode.commands.registerCommand.called).to.be.true;
      const registerCall = fixture.vscode.commands.registerCommand.args.find(
        (args: any[]) => args[0] === "spectra.openPanel",
      );
      expect(registerCall).to.exist;
      expect(registerCall![1]).to.be.a("function");
    });

    it("test_activate_pushesAllDisposablesToSubscriptions: pushes all disposables to context.subscriptions", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // ASSERT: context.subscriptions contains at least:
      //   OutputChannel, sessionListController, sessionDetailController, panel, command disposable
      expect(fixture.context.subscriptions.length).to.be.at.least(5);
    });

    it("test_activate_logsSuccessWithProjectRoot: logs successful activation including the resolved projectRoot", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      const fixtureCustom = createExtensionTestFixture("/my/project");

      // INTENT: Call activate(fixtureCustom.context)
      // ASSERT: Logger info called with message containing '/my/project'
      const projectRootLog = fixtureCustom.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("/my/project"),
      );
      expect(projectRootLog).to.exist;
    });
  });

  // ─── Happy Path — onDidUpdate subscriptions ────────────────────────────────

  describe("activate — onDidUpdate subscriptions", function () {
    it("test_activate_sessionListOnDidUpdate_cachesStateAndShowsList: caches state and calls panel.showSessionList", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Then trigger sessionListController.onDidUpdate with a fake state
      const fakeState = { sessions: [{ id: "s1" }], workflows: ["wf1"] };
      fixture.sessionListController.triggerUpdate(fakeState);

      // ASSERT: panel.showSessionList called with fakeState
      expect(fixture.panel.showSessionList.calledOnce).to.be.true;
      expect(fixture.panel.showSessionList.firstCall.args[0]).to.deep.equal(
        fakeState,
      );
    });

    it("test_activate_sessionDetailOnDidUpdate_showsDetail: calls panel.showSessionDetail on controller update", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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
      expect(fixture.panel.showSessionDetail.calledOnce).to.be.true;
      expect(fixture.panel.showSessionDetail.firstCall.args[0]).to.deep.equal(
        fakeDetailState,
      );
    });
  });

  // ─── Error Propagation ─────────────────────────────────────────────────────

  describe("activate — Error Propagation", function () {
    it("test_activate_sessionListOnDidError_showsErrorMessage: shows error when sessionListController fires onDidError", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Then trigger sessionListController.onDidError
      fixture.sessionListController.triggerError({ message: "scan failed" });

      // ASSERT: vscode.window.showErrorMessage called with 'scan failed'
      expect(
        fixture.vscode.window.showErrorMessage.calledOnceWith("scan failed"),
      ).to.be.true;
    });

    it("test_activate_sessionDetailOnDidError_showsErrorMessage: shows error when sessionDetailController fires onDidError", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Then trigger sessionDetailController.onDidError
      fixture.sessionDetailController.triggerError({
        message: "detail error",
      });

      // ASSERT: vscode.window.showErrorMessage called with 'detail error'
      expect(
        fixture.vscode.window.showErrorMessage.calledOnceWith("detail error"),
      ).to.be.true;
    });

    it("test_activate_createOrRevealThrows_propagatesError: error propagates when SpectraPanel.createOrReveal throws", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // Setup: SpectraPanel.createOrReveal throws
      fixture.spectraPanelCreateOrRevealStub.throws(
        new Error("internal error"),
      );

      // INTENT: Call activate(fixture.context)
      // ASSERT: activate throws/rejects with 'internal error'
      // (Exact mechanism — sync throw or async reject — depends on implementation)
    });
  });

  // ─── Happy Path — onDidReceiveMessage routing ──────────────────────────────

  describe("activate — onDidReceiveMessage routing", function () {
    it("test_activate_messageRouting_navigateToDetail: routes navigateToDetail to sessionDetailController.open", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Then trigger panel.onDidReceiveMessage with navigateToDetail
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

    it("test_activate_messageRouting_navigateToList_withCache: routes navigateToList to panel.showSessionList with cached state", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // First trigger onDidUpdate to populate cache
      const cachedState = { sessions: [{ id: "s1" }], workflows: ["wf1"] };
      fixture.sessionListController.triggerUpdate(cachedState);
      fixture.panel.showSessionList.resetHistory();

      // Then trigger navigateToList
      fixture.panel.triggerMessage({ command: "navigateToList" });

      // ASSERT: panel.showSessionList called with cachedState
      expect(fixture.panel.showSessionList.calledOnce).to.be.true;
      expect(fixture.panel.showSessionList.firstCall.args[0]).to.deep.equal(
        cachedState,
      );
    });

    it("test_activate_messageRouting_navigateToList_noCacheNoOp: no-op when navigateToList received before first onDidUpdate", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Do NOT trigger sessionListController.onDidUpdate
      // Trigger navigateToList
      fixture.panel.triggerMessage({ command: "navigateToList" });

      // ASSERT: panel.showSessionList is NOT called
      expect(fixture.panel.showSessionList.called).to.be.false;
    });

    it("test_activate_messageRouting_launchSession: routes launchSession to sessionListController.launch", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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

  // ─── Happy Path — onDidDispose ─────────────────────────────────────────────

  describe("activate — onDidDispose", function () {
    it("test_activate_panelOnDidDispose_disposesBothControllers: disposes both controllers when panel is disposed", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Trigger panel.onDidDispose
      fixture.panel.triggerDispose();

      // ASSERT: both controllers disposed
      expect(fixture.sessionListController.dispose.calledOnce).to.be.true;
      expect(fixture.sessionDetailController.dispose.calledOnce).to.be.true;
    });
  });

  // ─── Happy Path — spectra.openPanel command ────────────────────────────────

  describe("activate — spectra.openPanel command", function () {
    it("test_activate_openPanelCommand_callsCreateOrReveal: command handler calls SpectraPanel.createOrReveal", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
      // Capture the handler registered with registerCommand for 'spectra.openPanel'
      const registerCall = fixture.vscode.commands.registerCommand.args.find(
        (args: any[]) => args[0] === "spectra.openPanel",
      );
      expect(registerCall).to.exist;

      const handler = registerCall![1];

      // Reset createOrReveal call history
      fixture.spectraPanelCreateOrRevealStub.resetHistory();

      // Invoke the command handler
      handler();

      // ASSERT: SpectraPanel.createOrReveal called with context, extensionUri, logger
      expect(fixture.spectraPanelCreateOrRevealStub.calledOnce).to.be.true;
      const [ctx, uri] = fixture.spectraPanelCreateOrRevealStub.firstCall.args;
      expect(ctx).to.equal(fixture.context);
      expect(uri).to.equal(fixture.context.extensionUri);
    });
  });

  // ─── Null / Empty Input ────────────────────────────────────────────────────

  describe("activate — Null / Empty Input", function () {
    it("test_activate_projectRootUndefined_showsErrorAndReturnsEarly: shows error and returns early when projectRoot is undefined", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // Setup: ProjectRootResolver.resolve() returns undefined
      const undefinedFixture = createExtensionTestFixture(undefined);

      // INTENT: Call activate(undefinedFixture.context)

      // ASSERT: vscode.window.showErrorMessage called with a descriptive message
      expect(undefinedFixture.vscode.window.showErrorMessage.calledOnce).to.be
        .true;

      // ASSERT: no commands registered
      expect(undefinedFixture.vscode.commands.registerCommand.called).to.be
        .false;

      // ASSERT: no controllers created
      expect(undefinedFixture.sessionListControllerConstructorStub.called).to.be
        .false;
      expect(undefinedFixture.sessionDetailControllerConstructorStub.called).to
        .be.false;

      // ASSERT: only OutputChannel pushed to subscriptions
      expect(undefinedFixture.context.subscriptions.length).to.equal(1);
    });
  });

  // ─── Mock / Dependency Interaction ─────────────────────────────────────────

  describe("activate — Mock / Dependency Interaction", function () {
    it("test_activate_loggerWrapsOutputChannel: logger delegates info/warn/error to outputChannel.appendLine with severity tags", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // Setup: activate returns early (projectRoot undefined) but we can
      // observe the logger calling outputChannel.appendLine with [INFO] tag
      const undefinedFixture = createExtensionTestFixture(undefined);

      // INTENT: Call activate(undefinedFixture.context)

      // ASSERT: outputChannel.appendLine called with strings containing severity prefix
      const infoCall = undefinedFixture.outputChannel.appendLine.args.find(
        (args: any[]) =>
          typeof args[0] === "string" && args[0].includes("[INFO]"),
      );
      expect(infoCall).to.exist;
    });

    it("test_activate_terminateFromDetailPage_routesToSessionListController: terminateSession from detail page routes to sessionListController.terminate", function () {
      // Scaffolded: depends on extension.ts exporting activate()
      this.skip(); // Missing: vscode/src/extension.ts — activate() export

      // INTENT: Call activate(fixture.context)
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
  });

  // ─── deactivate ────────────────────────────────────────────────────────────

  describe("deactivate", function () {
    it("test_deactivate_isEmptyFunction: deactivate does nothing — returns undefined without errors", function () {
      // Scaffolded: depends on extension.ts exporting deactivate()
      this.skip(); // Missing: vscode/src/extension.ts — deactivate() export

      // INTENT: Call deactivate()
      // ASSERT: returns undefined, no errors thrown
      // const result = deactivate();
      // expect(result).to.be.undefined;
    });
  });
});
