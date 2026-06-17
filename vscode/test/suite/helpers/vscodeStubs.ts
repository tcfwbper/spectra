/**
 * Shared test helpers for stubbing VS Code workspace APIs.
 *
 * These helpers isolate the vscode module stubbing logic so that
 * individual test files remain focused on assertions.
 *
 * Scaffolded: depends on `vscode` module being available in the test
 * environment (via @vscode/test-electron or a mock resolver).
 */
import * as sinon from "sinon";

/**
 * Represents a minimal VS Code workspace folder shape used in tests.
 */
export interface FakeWorkspaceFolder {
  uri: { fsPath: string };
}

/**
 * Options for configuring workspace stubs in a test.
 */
export interface WorkspaceStubOptions {
  /** Value to return for vscode.workspace.workspaceFolders */
  workspaceFolders?: FakeWorkspaceFolder[] | undefined;
  /** Value to return for getConfiguration('spectra').get('projectRoot') */
  projectRootConfig?: string | null | undefined;
}

/**
 * Creates a fake configuration object matching the shape returned by
 * vscode.workspace.getConfiguration().
 */
export function createFakeConfiguration(
  projectRoot: string | null | undefined,
): {
  get: sinon.SinonStub;
} {
  const getStub = sinon.stub();
  getStub.withArgs("projectRoot").returns(projectRoot);
  return { get: getStub };
}

/**
 * Sets up stubs for vscode.workspace used by ProjectRootResolver.
 *
 * Returns the stubs/spies so tests can make interaction assertions.
 *
 * NOTE: The actual stubbing mechanism depends on how the vscode module
 * is exposed in the test environment. This helper assumes the vscode
 * module can be stubbed via sinon (e.g., using proxyquire or a mock
 * resolver). The exact wiring will be finalized when the extension
 * build tooling is in place.
 */
export function setupWorkspaceStubs(
  vscode: any,
  opts: WorkspaceStubOptions,
): {
  getConfigurationStub: sinon.SinonStub;
  configGetStub: sinon.SinonStub;
} {
  // Stub workspaceFolders property
  sinon.stub(vscode.workspace, "workspaceFolders").value(opts.workspaceFolders);

  // Stub getConfiguration
  const fakeConfig = createFakeConfiguration(opts.projectRootConfig);
  const getConfigurationStub = sinon
    .stub(vscode.workspace, "getConfiguration")
    .returns(fakeConfig);

  return {
    getConfigurationStub,
    configGetStub: fakeConfig.get,
  };
}

/**
 * Builds a FakeWorkspaceFolder array from simple path strings.
 */
export function buildWorkspaceFolders(
  ...paths: string[]
): FakeWorkspaceFolder[] {
  return paths.map((fsPath) => ({ uri: { fsPath } }));
}
