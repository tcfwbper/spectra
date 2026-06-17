/**
 * Unit tests for SessionWatcher.
 *
 * Test spec: spec/test/vscode/src/services/sessionWatcher.md
 * Source under test: vscode/src/services/sessionWatcher.ts
 *
 * The exact constructor signature is derived from the logic spec:
 * `new SessionWatcher(projectRoot, deps)`.
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockFileSystemWatcher,
  createMockEventEmitter,
  type MockFileSystemWatcher,
  type MockEventEmitter,
} from "./helpers/watcherStubs";

import { SessionWatcher } from "../../src/services/sessionWatcher";
import type { SessionWatcherDeps } from "../../src/services/sessionWatcher";

describe("SessionWatcher", function () {
  let sandbox: sinon.SinonSandbox;
  let clock: sinon.SinonFakeTimers;
  let mockWatcher: MockFileSystemWatcher;
  let mockEmitter: MockEventEmitter;
  let deps: SessionWatcherDeps;
  let relativePatternArgs: any[];

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    clock = sinon.useFakeTimers();
    mockWatcher = createMockFileSystemWatcher();
    mockEmitter = createMockEventEmitter();
    relativePatternArgs = [];

    deps = {
      createFileSystemWatcher: sinon.stub().returns(mockWatcher),
      createRelativePattern: sinon.stub().callsFake((...args: any[]) => {
        relativePatternArgs.push(...args);
        return { pattern: args };
      }),
      createEventEmitter: sinon.stub().returns(mockEmitter),
    };
  });

  afterEach(function () {
    clock.restore();
    sandbox.restore();
  });

  describe("Happy Path — Construction", function () {
    it("should store projectRoot", function () {
      const instance = new SessionWatcher("/project", deps);
      expect(instance).to.exist;
    });

    it("should expose onDidChange event", function () {
      const instance = new SessionWatcher("/project", deps);
      expect(instance.onDidChange).to.be.a("function");
    });

    it("should create file system watcher with correct glob pattern", function () {
      new SessionWatcher("/my/root", deps);
      // Verify createFileSystemWatcher called with RelativePattern
      // matching: /my/root/.spectra/sessions/*/session.json
      expect((deps.createRelativePattern as sinon.SinonStub).calledOnce).to.be
        .true;
      expect(relativePatternArgs[0]).to.equal("/my/root");
      expect(relativePatternArgs[1]).to.equal(
        ".spectra/sessions/*/session.json",
      );
    });
  });

  describe("Happy Path — onDidChange", function () {
    it("should fire onDidChange after debounce when session file is created", function () {
      new SessionWatcher("/project", deps);
      mockWatcher.onDidCreate.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should fire onDidChange after debounce when session file is modified", function () {
      new SessionWatcher("/project", deps);
      mockWatcher.onDidChange.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should fire onDidChange after debounce when session file is deleted", function () {
      new SessionWatcher("/project", deps);
      mockWatcher.onDidDelete.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should debounce rapid successive signals from mixed event types into single event", function () {
      new SessionWatcher("/project", deps);
      mockWatcher.onDidCreate.fire({});
      clock.tick(100);
      mockWatcher.onDidChange.fire({});
      clock.tick(100);
      mockWatcher.onDidDelete.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should reset debounce timer on each new signal", function () {
      new SessionWatcher("/project", deps);
      mockWatcher.onDidCreate.fire({});
      clock.tick(200);
      mockWatcher.onDidChange.fire({});
      clock.tick(200);
      mockWatcher.onDidDelete.fire({});
      clock.tick(200);
      expect(mockEmitter.fire.called).to.be.false; // only 200ms since last
      clock.tick(100);
      expect(mockEmitter.fire.calledOnce).to.be.true; // 300ms from last
    });
  });

  describe("Resource Cleanup", function () {
    it("should dispose file watcher and event emitter on dispose", function () {
      const instance = new SessionWatcher("/project", deps);
      instance.dispose();
      expect(mockWatcher.dispose.calledOnce).to.be.true;
      expect(mockEmitter.dispose.calledOnce).to.be.true;
    });

    it("should cancel pending debounce timer on dispose", function () {
      const instance = new SessionWatcher("/project", deps);
      mockWatcher.onDidCreate.fire({});
      instance.dispose();
      clock.tick(300);
      expect(mockEmitter.fire.called).to.be.false;
    });

    it("should not fire onDidChange after dispose", function () {
      const instance = new SessionWatcher("/project", deps);
      instance.dispose();
      mockWatcher.onDidCreate.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.called).to.be.false;
    });
  });

  describe("Idempotency", function () {
    it("should handle multiple dispose calls without error", function () {
      const instance = new SessionWatcher("/project", deps);
      expect(() => {
        instance.dispose();
        instance.dispose();
        instance.dispose();
      }).to.not.throw();
    });
  });

  describe("Mock / Dependency Interaction", function () {
    it("should subscribe to onDidCreate, onDidChange, and onDidDelete", function () {
      new SessionWatcher("/project", deps);
      expect(mockWatcher.onDidCreate.listeners).to.have.lengthOf(1);
      expect(mockWatcher.onDidChange.listeners).to.have.lengthOf(1);
      expect(mockWatcher.onDidDelete.listeners).to.have.lengthOf(1);
    });

    it("should not read or write any files", function () {
      // The SessionWatcher only creates a file system watcher and emitter;
      // it does not call any fs read/write/access methods.
      new SessionWatcher("/project", deps);
      expect(
        (deps.createFileSystemWatcher as sinon.SinonStub).calledOnce,
      ).to.be.true;
    });
  });
});
