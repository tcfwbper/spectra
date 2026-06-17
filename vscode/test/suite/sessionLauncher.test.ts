/**
 * Unit tests for SessionLauncher.
 *
 * Test spec: spec/test/vscode/src/services/sessionLauncher.md
 * Source under test: vscode/src/services/sessionLauncher.ts
 *
 * Scaffolded: The production module (sessionLauncher.ts) does not yet exist.
 * All tests are structured with full setup/assertions but import is deferred.
 * Once the production surface is created with a dependency-injectable static method,
 * remove the t.skip markers and wire the import.
 *
 * Expected production surface:
 *   SessionLauncher.launch(workflowName, logger, deps?)
 *   where deps = { getConfiguration, spawn, randomUUID }
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockChildProcess,
  createMockServiceLogger,
  createSessionLauncherDeps,
  type MockChildProcess,
  type MockLogger,
  type SessionLauncherDeps,
} from "./helpers/processStubs";

// Scaffolded import — uncomment when production module exists:
// import { SessionLauncher } from '../../src/services/sessionLauncher';

describe("SessionLauncher", function () {
  let sandbox: sinon.SinonSandbox;
  let mockChild: MockChildProcess;
  let logger: MockLogger;
  let deps: SessionLauncherDeps;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    mockChild = createMockChildProcess();
    logger = createMockServiceLogger();
    deps = createSessionLauncherDeps("spectra", mockChild, "test-uuid-1234");
  });

  afterEach(function () {
    sandbox.restore();
  });

  describe("Happy Path — launch", function () {
    it("should spawn detached process with correct arguments", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // await SessionLauncher.launch('myWorkflow', logger, deps);
      // expect(deps.spawn.calledOnce).to.be.true;
      // const [binary, args, options] = deps.spawn.firstCall.args;
      // expect(binary).to.equal('spectra');
      // expect(args).to.deep.equal([
      //   'run', '--workflow', 'myWorkflow', '--session-id', 'test-uuid-1234'
      // ]);
      // expect(options).to.deep.equal({ detached: true, stdio: 'ignore' });
    });

    it("should call unref on the child process", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // await SessionLauncher.launch('wf', logger, deps);
      // expect(mockChild.unref.calledOnce).to.be.true;
    });

    it("should log info message with workflow name and session id", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const uuidDeps = createSessionLauncherDeps('spectra', mockChild, 'uuid-abc');
      // await SessionLauncher.launch('deploy', logger, uuidDeps);
      // expect(logger.info.calledOnce).to.be.true;
      // const infoMsg = logger.info.firstCall.args[0];
      // expect(infoMsg).to.include('deploy');
      // expect(infoMsg).to.include('uuid-abc');
    });

    it("should resolve with void on successful spawn", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const result = await SessionLauncher.launch('wf', logger, deps);
      // expect(result).to.be.undefined;
    });
  });

  describe("Happy Path — configuration default", function () {
    it("should default to spectra when config is undefined", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const undefinedDeps = createSessionLauncherDeps(undefined, mockChild, 'uuid');
      // await SessionLauncher.launch('wf', logger, undefinedDeps);
      // const [binary] = undefinedDeps.spawn.firstCall.args;
      // expect(binary).to.equal('spectra');
    });

    it("should default to spectra when config is empty string", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const emptyDeps = createSessionLauncherDeps('' as any, mockChild, 'uuid');
      // await SessionLauncher.launch('wf', logger, emptyDeps);
      // const [binary] = emptyDeps.spawn.firstCall.args;
      // expect(binary).to.equal('spectra');
    });

    it("should use custom binary path from configuration", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const customDeps = createSessionLauncherDeps('/usr/local/bin/spectra', mockChild, 'uuid');
      // await SessionLauncher.launch('wf', logger, customDeps);
      // const [binary] = customDeps.spawn.firstCall.args;
      // expect(binary).to.equal('/usr/local/bin/spectra');
    });
  });

  describe("Error Propagation", function () {
    it("should throw when spawn fails with ENOENT", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const errorChild = createMockChildProcess();
      // const enoentDeps = createSessionLauncherDeps('/missing/spectra', errorChild, 'uuid');
      // enoentDeps.spawn.callsFake(() => {
      //   const cp = createMockChildProcess();
      //   process.nextTick(() => {
      //     const err: any = new Error('spawn ENOENT');
      //     err.code = 'ENOENT';
      //     cp.emit('error', err);
      //   });
      //   return cp;
      // });
      // try {
      //   await SessionLauncher.launch('wf', logger, enoentDeps);
      //   expect.fail('should have thrown');
      // } catch (err: any) {
      //   expect(err.message).to.include('/missing/spectra');
      // }
    });

    it("should throw when spawn fails with EACCES", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // const eaccesDeps = createSessionLauncherDeps('/no-exec/spectra', createMockChildProcess(), 'uuid');
      // eaccesDeps.spawn.callsFake(() => {
      //   const cp = createMockChildProcess();
      //   process.nextTick(() => {
      //     const err: any = new Error('spawn EACCES');
      //     err.code = 'EACCES';
      //     cp.emit('error', err);
      //   });
      //   return cp;
      // });
      // await expect(SessionLauncher.launch('wf', logger, eaccesDeps))
      //   .to.be.rejected;
    });
  });

  describe("Mock / Dependency Interaction", function () {
    it("should generate a fresh UUID for every invocation", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // deps.randomUUID.onFirstCall().returns('uuid-1');
      // deps.randomUUID.onSecondCall().returns('uuid-2');
      // await SessionLauncher.launch('wf', logger, deps);
      // await SessionLauncher.launch('wf', logger, deps);
      // const firstArgs = deps.spawn.firstCall.args[1];
      // const secondArgs = deps.spawn.secondCall.args[1];
      // expect(firstArgs).to.include('uuid-1');
      // expect(secondArgs).to.include('uuid-2');
    });

    it("should spawn with detached true", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // await SessionLauncher.launch('wf', logger, deps);
      // const options = deps.spawn.firstCall.args[2];
      // expect(options.detached).to.be.true;
    });

    it("should spawn with stdio ignore", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // await SessionLauncher.launch('wf', logger, deps);
      // const options = deps.spawn.firstCall.args[2];
      // expect(options.stdio).to.equal('ignore');
    });

    it("should read configuration on every invocation", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // await SessionLauncher.launch('wf', logger, deps);
      // await SessionLauncher.launch('wf', logger, deps);
      // expect(deps.getConfiguration.callCount).to.be.at.least(2);
    });

    it("should pass workflowName with special characters as single argv element", function () {
      // Scaffolded: awaiting SessionLauncher production surface
      this.skip(); // Missing: SessionLauncher.launch static method in sessionLauncher.ts

      // await SessionLauncher.launch('my workflow (v2)', logger, deps);
      // const args = deps.spawn.firstCall.args[1];
      // const workflowIdx = args.indexOf('--workflow') + 1;
      // expect(args[workflowIdx]).to.equal('my workflow (v2)');
    });
  });
});
