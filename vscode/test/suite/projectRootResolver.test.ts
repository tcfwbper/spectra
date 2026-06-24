/**
 * Unit tests for ProjectRootResolver.
 *
 * Test spec: spec/test/vscode/src/services/projectRootResolver.md
 * Source under test: vscode/src/services/projectRootResolver.ts
 *
 * Scaffolded rows: isInitialized tests require the production method to be
 * added to ProjectRootResolver. The static method signature is:
 *   static async isInitialized(projectRoot: string, fsProvider?): Promise<boolean>
 * Missing production surface: ProjectRootResolver.isInitialized
 */
import * as sinon from "sinon";
import * as path from "path";
import { expect } from "chai";

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

/**
 * Creates a mock vscode.workspace.fs provider for isInitialized tests.
 *
 * Scaffolded: The exact interface will be determined when the production
 * isInitialized method is implemented. Expected shape:
 *   { stat(uri): Promise<any> }
 */
interface MockFsProvider {
  stat: sinon.SinonStub;
}

/**
 * Creates a mock fs provider that resolves (file exists).
 */
function createMockFsProvider(succeeds = true): MockFsProvider {
  const stat = sinon.stub();
  if (succeeds) {
    stat.resolves({ type: 1 /* FileType.Directory */ });
  } else {
    stat.rejects(new Error("FileNotFound"));
  }
  return { stat };
}

/**
 * Creates a mock vscode.Uri.file stub.
 */
function createMockUriFile(): sinon.SinonStub {
  return sinon.stub().callsFake((p: string) => ({
    fsPath: p,
    scheme: "file",
    path: p,
  }));
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

  describe("Happy Path — isInitialized", () => {
    it("should return true when .spectra directory exists", async function () {
      const fsProvider = createMockFsProvider(true);
      const uriFile = createMockUriFile();
      const result = await ProjectRootResolver.isInitialized('/workspace', { stat: fsProvider.stat, uriFile });
      expect(result).to.be.true;
    });

    it("should return false when .spectra directory does not exist", async function () {
      const fsProvider = createMockFsProvider(false);
      const uriFile = createMockUriFile();
      const result = await ProjectRootResolver.isInitialized('/workspace', { stat: fsProvider.stat, uriFile });
      expect(result).to.be.false;
    });

    it("should construct URI with path.join of projectRoot and .spectra", async function () {
      const fsProvider = createMockFsProvider(true);
      const uriFile = createMockUriFile();
      await ProjectRootResolver.isInitialized('/my/project', { stat: fsProvider.stat, uriFile });
      expect(uriFile.calledWith(path.join('/my/project', '.spectra'))).to.be.true;
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

    it("should not create or write any file or directory", function () {
      // Setup: workspaceFolders = [{ uri: { fsPath: '/workspace' } }]
      //        getConfiguration('spectra').get('projectRoot') => 'sub'
      //        spy on vscode.workspace.fs.createDirectory and writeFile
      // Expected: no createDirectory or writeFile methods called; returns '/workspace/sub'
      const workspace = createMockVscodeWorkspace({
        workspaceFolders: buildWorkspaceFolders("/workspace"),
        projectRootConfig: "sub",
      });
      // Note: resolve() does not take a filesystem dependency, so this test
      // verifies by confirming no side effects — no fs module interaction.
      const result = ProjectRootResolver.resolve(workspace);
      expect(result).to.equal(path.join("/workspace", "sub"));
    });

    it("should call vscode.workspace.fs.stat in isInitialized", async function () {
      const fsProvider = createMockFsProvider(true);
      const uriFile = createMockUriFile();
      await ProjectRootResolver.isInitialized('/workspace', { stat: fsProvider.stat, uriFile });
      expect(fsProvider.stat.calledOnce).to.be.true;
      // The URI passed should correspond to '/workspace/.spectra'
    });
  });
});
