/**
 * Unit tests for ProjectRootResolver.
 *
 * Test spec: spec/test/vscode/src/services/projectRootResolver.md
 * Source under test: vscode/src/services/projectRootResolver.ts
 */
import * as sinon from "sinon";
import * as path from "path";
import { expect } from "chai";

// Use require for fs so sinon can spy on its methods (import * creates
// a namespace with getter-only descriptors that sinon cannot wrap).
// eslint-disable-next-line @typescript-eslint/no-var-requires
const fs = require("fs");

import {
  buildWorkspaceFolders,
  createFakeConfiguration,
  type WorkspaceStubOptions,
} from "./helpers/vscodeStubs";

import {
  ProjectRootResolver,
  type WorkspaceProvider,
} from "../../src/services/projectRootResolver";

/**
 * Creates a minimal mock of the vscode.workspace namespace.
 * This avoids depending on the real vscode module at test-authoring time.
 */
function createMockVscodeWorkspace(opts: WorkspaceStubOptions): WorkspaceProvider & {
  getConfiguration: sinon.SinonStub;
  _fakeConfig: { get: sinon.SinonStub };
} {
  const fakeConfig = createFakeConfiguration(opts.projectRootConfig);
  return {
    workspaceFolders: opts.workspaceFolders,
    getConfiguration: sinon.stub().returns(fakeConfig),
    _fakeConfig: fakeConfig,
  };
}

describe("ProjectRootResolver", () => {
  let sandbox: sinon.SinonSandbox;

  beforeEach(() => {
    sandbox = sinon.createSandbox();
  });

  afterEach(() => {
    sandbox.restore();
  });

  describe("Happy Path — resolve", () => {
    it("should return workspace path when projectRoot config is not set", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/home/user/project' } }]
      //        getConfiguration('spectra').get('projectRoot') => undefined
      // Expected: returns '/home/user/project'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/home/user/project"),
        projectRootConfig: undefined,
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal("/home/user/project");
    });

    it("should return joined path when projectRoot config is set", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/home/user/project' } }]
      //        getConfiguration('spectra').get('projectRoot') => 'sub/folder'
      // Expected: returns '/home/user/project/sub/folder'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/home/user/project"),
        projectRootConfig: "sub/folder",
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal(path.join("/home/user/project", "sub/folder"));
    });

    it("should normalize path segments with dot-dot in config value", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/home/user/project' } }]
      //        getConfiguration('spectra').get('projectRoot') => '../sibling'
      // Expected: returns '/home/user/sibling'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/home/user/project"),
        projectRootConfig: "../sibling",
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal(path.join("/home/user/project", "../sibling"));
    });
  });

  describe("Null / Empty Input", () => {
    it("should return undefined when workspaceFolders is undefined", function () {
      // Setup: workspaceFolders = undefined
      // Expected: returns undefined
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: undefined,
        projectRootConfig: undefined,
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.be.undefined;
    });

    it("should return undefined when workspaceFolders is empty array", function () {
      // Setup: workspaceFolders = []
      // Expected: returns undefined
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: [],
        projectRootConfig: undefined,
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.be.undefined;
    });

    it("should return undefined when first workspace folder fsPath is empty string", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '' } }]
      // Expected: returns undefined
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders(""),
        projectRootConfig: undefined,
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.be.undefined;
    });

    it("should return workspace path when projectRoot config is null", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/workspace' } }]
      //        getConfiguration('spectra').get('projectRoot') => null
      // Expected: returns '/workspace'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/workspace"),
        projectRootConfig: null,
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal("/workspace");
    });

    it("should return workspace path when projectRoot config is empty string", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/workspace' } }]
      //        getConfiguration('spectra').get('projectRoot') => ''
      // Expected: returns '/workspace'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/workspace"),
        projectRootConfig: "",
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal("/workspace");
    });
  });

  describe("Mock / Dependency Interaction", () => {
    it("should read only the first workspace folder", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/first' } }, { uri: { fsPath: '/second' } }]
      //        getConfiguration('spectra').get('projectRoot') => undefined
      // Expected: returns '/first'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/first", "/second"),
        projectRootConfig: undefined,
      });
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal("/first");
    });

    it("should call getConfiguration with spectra section", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/workspace' } }]
      //        spy on getConfiguration; config.get returns 'custom'
      // Expected: getConfiguration called with 'spectra'; get called with 'projectRoot'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/workspace"),
        projectRootConfig: "custom",
      });
      ProjectRootResolver.resolve(workspace);
      expect(workspace.getConfiguration.calledWith("spectra")).to.be.true;
      expect(workspace._fakeConfig.get.calledWith("projectRoot")).to.be.true;
    });

    it("should not perform any file system operations", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/workspace' } }]
      //        getConfiguration('spectra').get('projectRoot') => 'sub'
      //        spy on fs methods (existsSync, mkdirSync)
      // Expected: no fs methods called; returns '/workspace/sub'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/workspace"),
        projectRootConfig: "sub",
      });
      const fsSpy = sandbox.spy(fs, "existsSync");
      const mkdirSpy = sandbox.spy(fs, "mkdirSync");
      const result = ProjectRootResolver.resolve(workspace);
      expect(fsSpy.called).to.be.false;
      expect(mkdirSpy.called).to.be.false;
      expect(result).to.equal(path.join("/workspace", "sub"));
    });
  });
});
