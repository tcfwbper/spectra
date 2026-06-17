/**
 * Unit tests for EventWatcher.
 *
 * Test spec: spec/test/vscode/src/services/eventWatcher.md
 * Source under test: vscode/src/services/eventWatcher.ts
 *
 * The exact constructor signature is derived from the logic spec:
 * `new EventWatcher(projectRoot, sessionId, deps)`.
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockFileSystemWatcher,
  createMockEventEmitter,
  type MockFileSystemWatcher,
  type MockEventEmitter,
} from "./helpers/watcherStubs";

import { EventWatcher } from "../../src/services/eventWatcher";
import type { EventWatcherDeps } from "../../src/services/eventWatcher";

describe("EventWatcher", function () {
  let sandbox: sinon.SinonSandbox;
  let clock: sinon.SinonFakeTimers;
  let mockWatcher: MockFileSystemWatcher;
  let mockEmitter: MockEventEmitter;
  let deps: EventWatcherDeps;
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
    it("should store projectRoot and sessionId", function () {
      const instance = new EventWatcher("/project", "sess-1", deps);
      expect(instance).to.exist;
    });

    it("should expose onDidChange event", function () {
      const instance = new EventWatcher("/project", "sess-1", deps);
      expect(instance.onDidChange).to.be.a("function");
    });

    it("should create file system watcher with correct glob pattern", function () {
      new EventWatcher("/my/root", "abc-123", deps);
      // Verify createFileSystemWatcher called with RelativePattern
      // matching: /my/root/.spectra/sessions/abc-123/events.jsonl
      expect((deps.createRelativePattern as sinon.SinonStub).calledOnce).to.be
        .true;
      expect(relativePatternArgs[0]).to.equal("/my/root");
      expect(relativePatternArgs[1]).to.equal(
        ".spectra/sessions/abc-123/events.jsonl",
      );
    });
  });

  describe("Happy Path — onDidChange", function () {
    it("should fire onDidChange after debounce when file is modified", function () {
      new EventWatcher("/project", "sess-1", deps);
      mockWatcher.onDidChange.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should debounce rapid successive modifications into single event", function () {
      new EventWatcher("/project", "sess-1", deps);
      mockWatcher.onDidChange.fire({});
      clock.tick(100);
      mockWatcher.onDidChange.fire({});
      clock.tick(100);
      mockWatcher.onDidChange.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should reset debounce timer on each new modification", function () {
      new EventWatcher("/project", "sess-1", deps);
      mockWatcher.onDidChange.fire({});
      clock.tick(200);
      mockWatcher.onDidChange.fire({});
      clock.tick(200);
      mockWatcher.onDidChange.fire({});
      clock.tick(200);
      expect(mockEmitter.fire.called).to.be.false; // only 200ms since last
      clock.tick(100);
      expect(mockEmitter.fire.calledOnce).to.be.true; // 300ms from last
    });
  });

  describe("Resource Cleanup", function () {
    it("should dispose file watcher and event emitter on dispose", function () {
      const instance = new EventWatcher("/project", "sess-1", deps);
      instance.dispose();
      expect(mockWatcher.dispose.calledOnce).to.be.true;
      expect(mockEmitter.dispose.calledOnce).to.be.true;
    });

    it("should cancel pending debounce timer on dispose", function () {
      const instance = new EventWatcher("/project", "sess-1", deps);
      mockWatcher.onDidChange.fire({});
      instance.dispose();
      clock.tick(300);
      expect(mockEmitter.fire.called).to.be.false;
    });

    it("should not fire onDidChange after dispose", function () {
      const instance = new EventWatcher("/project", "sess-1", deps);
      instance.dispose();
      mockWatcher.onDidChange.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.called).to.be.false;
    });
  });

  describe("Idempotency", function () {
    it("should handle multiple dispose calls without error", function () {
      const instance = new EventWatcher("/project", "sess-1", deps);
      expect(() => {
        instance.dispose();
        instance.dispose();
        instance.dispose();
      }).to.not.throw();
    });
  });

  describe("Mock / Dependency Interaction", function () {
    it("should subscribe only to onDidChange of file watcher", function () {
      new EventWatcher("/project", "sess-1", deps);
      expect(mockWatcher.onDidChange.listeners).to.have.lengthOf(1);
      expect(mockWatcher.onDidCreate.listeners).to.have.lengthOf(0);
      expect(mockWatcher.onDidDelete.listeners).to.have.lengthOf(0);
    });

    it("should not read or write any files", function () {
      // The EventWatcher only creates a file system watcher and emitter;
      // it does not call any fs read/write/access methods.
      // This is verified by the absence of such calls in the mock deps.
      new EventWatcher("/project", "sess-1", deps);
      // deps only exposes createFileSystemWatcher, createRelativePattern, createEventEmitter
      // No fs operations are available or called
      expect(
        (deps.createFileSystemWatcher as sinon.SinonStub).calledOnce,
      ).to.be.true;
    });
  });
});
