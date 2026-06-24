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
  let mockFileWatcher: MockFileSystemWatcher;
  let mockDirWatcher: MockFileSystemWatcher;
  let mockEmitter: MockEventEmitter;
  let deps: SessionWatcherDeps;
  let relativePatternArgs: any[][];

  beforeEach(function () {
    sandbox = sinon.createSandbox();
    clock = sinon.useFakeTimers();
    mockFileWatcher = createMockFileSystemWatcher();
    mockDirWatcher = createMockFileSystemWatcher();
    mockEmitter = createMockEventEmitter();
    relativePatternArgs = [];

    const createFileSystemWatcherStub = sinon.stub();
    createFileSystemWatcherStub.onFirstCall().returns(mockFileWatcher);
    createFileSystemWatcherStub.onSecondCall().returns(mockDirWatcher);

    deps = {
      createFileSystemWatcher: createFileSystemWatcherStub,
      createRelativePattern: sinon.stub().callsFake((...args: any[]) => {
        relativePatternArgs.push(args);
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

    it("should create file watcher with correct glob pattern for session.json files", function () {
      new SessionWatcher("/my/root", deps);
      // First call to createRelativePattern should be for session.json files
      expect(relativePatternArgs.length).to.be.at.least(1);
      expect(relativePatternArgs[0][0]).to.equal("/my/root");
      expect(relativePatternArgs[0][1]).to.equal(
        ".spectra/sessions/*/session.json",
      );
      // First createFileSystemWatcher call for the file pattern
      expect((deps.createFileSystemWatcher as sinon.SinonStub).calledTwice).to
        .be.true;
    });

    it("should create directory watcher with correct glob pattern for session directories", function () {
      new SessionWatcher("/my/root", deps);
      // Second call to createRelativePattern should be for session directories
      expect(relativePatternArgs.length).to.be.at.least(2);
      expect(relativePatternArgs[1][0]).to.equal("/my/root");
      expect(relativePatternArgs[1][1]).to.equal(".spectra/sessions/*");
      // Second createFileSystemWatcher call for the directory pattern
      expect((deps.createFileSystemWatcher as sinon.SinonStub).calledTwice).to
        .be.true;
    });
  });

  describe("Happy Path — onDidChange", function () {
    it("should fire onDidChange after debounce when session file is created", function () {
      new SessionWatcher("/project", deps);
      mockFileWatcher.onDidCreate.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should fire onDidChange after debounce when session file is modified", function () {
      new SessionWatcher("/project", deps);
      mockFileWatcher.onDidChange.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should fire onDidChange after debounce when session file is deleted", function () {
      new SessionWatcher("/project", deps);
      mockFileWatcher.onDidDelete.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should fire onDidChange after debounce when session directory is created", function () {
      new SessionWatcher("/project", deps);
      mockDirWatcher.onDidCreate.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should fire onDidChange after debounce when session directory is deleted", function () {
      new SessionWatcher("/project", deps);
      mockDirWatcher.onDidDelete.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should debounce rapid successive signals from both watchers into single event", function () {
      new SessionWatcher("/project", deps);
      mockFileWatcher.onDidCreate.fire({});
      clock.tick(100);
      mockDirWatcher.onDidCreate.fire({});
      clock.tick(100);
      mockFileWatcher.onDidChange.fire({});
      clock.tick(300);
      expect(mockEmitter.fire.calledOnce).to.be.true;
    });

    it("should reset debounce timer on each new signal from either watcher", function () {
      new SessionWatcher("/project", deps);
      mockFileWatcher.onDidCreate.fire({});
      clock.tick(200);
      mockDirWatcher.onDidCreate.fire({});
      clock.tick(200);
      mockFileWatcher.onDidDelete.fire({});
      clock.tick(200);
      expect(mockEmitter.fire.called).to.be.false; // only 200ms since last
      clock.tick(100);
      expect(mockEmitter.fire.calledOnce).to.be.true; // 300ms from last
    });
  });

  describe("Resource Cleanup", function () {
    it("should dispose both watchers and event emitter on dispose", function () {
      const instance = new SessionWatcher("/project", deps);
      instance.dispose();
      expect(mockFileWatcher.dispose.calledOnce).to.be.true;
      expect(mockDirWatcher.dispose.calledOnce).to.be.true;
      expect(mockEmitter.dispose.calledOnce).to.be.true;
    });

    it("should cancel pending debounce timer on dispose", function () {
      const instance = new SessionWatcher("/project", deps);
      mockFileWatcher.onDidCreate.fire({});
      instance.dispose();
      clock.tick(300);
      expect(mockEmitter.fire.called).to.be.false;
    });

    it("should not fire onDidChange after dispose", function () {
      const instance = new SessionWatcher("/project", deps);
      instance.dispose();
      mockFileWatcher.onDidCreate.fire({});
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
    it("should subscribe to onDidCreate, onDidChange, and onDidDelete on file watcher", function () {
      new SessionWatcher("/project", deps);
      expect(mockFileWatcher.onDidCreate.listeners).to.have.lengthOf(1);
      expect(mockFileWatcher.onDidChange.listeners).to.have.lengthOf(1);
      expect(mockFileWatcher.onDidDelete.listeners).to.have.lengthOf(1);
    });

    it("should subscribe to onDidCreate and onDidDelete on directory watcher", function () {
      new SessionWatcher("/project", deps);
      expect(mockDirWatcher.onDidCreate.listeners).to.have.lengthOf(1);
      expect(mockDirWatcher.onDidDelete.listeners).to.have.lengthOf(1);
    });

    it("should not read or write any files", function () {
      // The SessionWatcher only creates file system watchers and emitter;
      // it does not call any fs read/write/access methods.
      new SessionWatcher("/project", deps);
      expect((deps.createFileSystemWatcher as sinon.SinonStub).calledTwice).to
        .be.true;
    });
  });
});
