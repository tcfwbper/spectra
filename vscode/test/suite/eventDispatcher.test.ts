/**
 * Unit tests for EventDispatcher.
 *
 * Test spec: spec/test/vscode/src/services/eventDispatcher.md
 * Source under test: vscode/src/services/eventDispatcher.ts
 *
 * Scaffolded: The production module (eventDispatcher.ts) does not yet exist.
 * All tests are structured with full setup/assertions but import is deferred.
 * Once the production surface is created with a dependency-injectable static method,
 * remove the t.skip markers and wire the import.
 *
 * Expected production surface:
 *   EventDispatcher.dispatch(eventType, sessionId, message, logger, deps?)
 *   where deps = { getConfiguration, execFile }
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockChildProcess,
  createMockServiceLogger,
  createEventDispatcherDeps,
  type MockChildProcess,
  type MockLogger,
  type EventDispatcherDeps,
} from "./helpers/processStubs";

// Scaffolded import — uncomment when production module exists:
// import { EventDispatcher } from '../../src/services/eventDispatcher';

describe("EventDispatcher", function () {
  let sandbox: sinon.SinonSandbox;
  let clock: sinon.SinonFakeTimers;
  let mockChild: MockChildProcess;
  let logger: MockLogger;
  let deps: EventDispatcherDeps;

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    clock = sinon.useFakeTimers();
    mockChild = createMockChildProcess();
    logger = createMockServiceLogger();
    deps = createEventDispatcherDeps("spectra-agent", mockChild);
  });

  afterEach(function () {
    clock.restore();
    sandbox.restore();
  });

  describe("Happy Path — dispatch", function () {
    it("should spawn spectra-agent with correct arguments", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      // Expected: execFile called with 'spectra-agent' and
      //   ['event', 'emit', 'ReviewNeeded', '--session-id', 'abc-123', '--message', 'hello world']
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // await EventDispatcher.dispatch('ReviewNeeded', 'abc-123', 'hello world', logger, deps);
      // expect(deps.execFile.calledOnce).to.be.true;
      // const [binary, args] = deps.execFile.firstCall.args;
      // expect(binary).to.equal('spectra-agent');
      // expect(args).to.deep.equal([
      //   'event', 'emit', 'ReviewNeeded', '--session-id', 'abc-123', '--message', 'hello world'
      // ]);
    });

    it("should log info message with event type and session id", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // await EventDispatcher.dispatch('SessionStarted', 'uuid-1', 'started', logger, deps);
      // expect(logger.info.calledOnce).to.be.true;
      // const infoMsg = logger.info.firstCall.args[0];
      // expect(infoMsg).to.include('SessionStarted');
      // expect(infoMsg).to.include('uuid-1');
    });

    it("should resolve without waiting for child process exit", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      // Uses fake timers to prove promise resolves immediately
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // The mock child process never emits 'exit' — promise must still resolve
      // const promise = EventDispatcher.dispatch('Ping', 's1', 'm', logger, deps);
      // await promise; // Should resolve without advancing timers
    });
  });

  describe("Happy Path — configuration default", function () {
    it("should default to spectra-agent when config is undefined", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const undefinedDeps = createEventDispatcherDeps(undefined, mockChild);
      // await EventDispatcher.dispatch('E', 's', 'm', logger, undefinedDeps);
      // const [binary] = undefinedDeps.execFile.firstCall.args;
      // expect(binary).to.equal('spectra-agent');
    });

    it("should default to spectra-agent when config is empty string", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const emptyDeps = createEventDispatcherDeps('' as any, mockChild);
      // await EventDispatcher.dispatch('E', 's', 'm', logger, emptyDeps);
      // const [binary] = emptyDeps.execFile.firstCall.args;
      // expect(binary).to.equal('spectra-agent');
    });

    it("should use custom binary path from configuration", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const customDeps = createEventDispatcherDeps('/opt/bin/spectra-agent', mockChild);
      // await EventDispatcher.dispatch('E', 's', 'm', logger, customDeps);
      // const [binary] = customDeps.execFile.firstCall.args;
      // expect(binary).to.equal('/opt/bin/spectra-agent');
    });
  });

  describe("Error Propagation", function () {
    it("should throw when spawn fails with ENOENT", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const errorChild = createMockChildProcess();
      // const enoentDeps = createEventDispatcherDeps('/missing/spectra-agent', errorChild);
      // // Simulate synchronous error event after spawn
      // enoentDeps.execFile.callsFake(() => {
      //   const cp = createMockChildProcess();
      //   process.nextTick(() => {
      //     const err: any = new Error('spawn ENOENT');
      //     err.code = 'ENOENT';
      //     cp.emit('error', err);
      //   });
      //   return cp;
      // });
      // try {
      //   await EventDispatcher.dispatch('E', 's', 'm', logger, enoentDeps);
      //   expect.fail('should have thrown');
      // } catch (err: any) {
      //   expect(err.message).to.include('/missing/spectra-agent');
      // }
    });

    it("should throw when spawn fails with EACCES", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const eaccesDeps = createEventDispatcherDeps('/no-exec/spectra-agent', createMockChildProcess());
      // eaccesDeps.execFile.callsFake(() => {
      //   const cp = createMockChildProcess();
      //   process.nextTick(() => {
      //     const err: any = new Error('spawn EACCES');
      //     err.code = 'EACCES';
      //     cp.emit('error', err);
      //   });
      //   return cp;
      // });
      // await expect(EventDispatcher.dispatch('E', 's', 'm', logger, eaccesDeps))
      //   .to.be.rejected;
    });
  });

  describe("Mock / Dependency Interaction", function () {
    it("should not use shell for spawn", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // await EventDispatcher.dispatch('E', 's', 'm', logger, deps);
      // const callArgs = deps.execFile.firstCall.args;
      // // execFile should be called without shell: true in options
      // if (callArgs.length > 2 && typeof callArgs[2] === 'object') {
      //   expect(callArgs[2]).to.not.have.property('shell', true);
      // }
    });

    it("should log warning on non-zero exit code", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // await EventDispatcher.dispatch('E', 's', 'm', logger, deps);
      // // Simulate non-zero exit after promise resolved
      // mockChild.emit('exit', 1);
      // expect(logger.warn.calledOnce).to.be.true;
      // expect(logger.warn.firstCall.args[0]).to.include('1');
    });

    it("should not throw on non-zero exit code", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const result = await EventDispatcher.dispatch('E', 's', 'm', logger, deps);
      // expect(result).to.be.undefined;
      // // Emit exit code 2 after resolution — should not cause unhandled rejection
      // mockChild.emit('exit', 2);
    });

    it("should pass special characters in message without shell interpretation", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // const specialMsg = 'hello "world" \n $PATH';
      // await EventDispatcher.dispatch('E', 's', specialMsg, logger, deps);
      // const [, args] = deps.execFile.firstCall.args;
      // expect(args[args.length - 1]).to.equal(specialMsg);
    });

    it("should read configuration on every invocation", function () {
      // Scaffolded: awaiting EventDispatcher production surface
      this.skip(); // Missing: EventDispatcher.dispatch static method in eventDispatcher.ts

      // await EventDispatcher.dispatch('E', 's', 'm', logger, deps);
      // await EventDispatcher.dispatch('E', 's', 'm', logger, deps);
      // expect(deps.getConfiguration.callCount).to.be.at.least(2);
    });
  });
});
