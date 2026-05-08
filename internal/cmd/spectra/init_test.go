package spectra

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Init ---

func TestInit_AllPhasesSucceed(t *testing.T) {
	deps := newFakeInitDeps()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stdout contains success message; exit code 0
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Spectra project initialized successfully")
	assert.Empty(t, stderr.String())
}

func TestInit_Phase2_WarningsPrinted(t *testing.T) {
	deps := newFakeInitDeps()
	deps.copier.copyWorkflowsWarnings = []string{"workflow X already exists, skipping"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stdout contains warning and success message
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "workflow X already exists, skipping")
	assert.Contains(t, stdout.String(), "Spectra project initialized successfully")
	assert.Empty(t, stderr.String())
}

// --- Ordering — Phase Sequencing ---

func TestInit_PhasesExecuteInOrder(t *testing.T) {
	deps := newFakeInitDeps()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runInit(deps, stdout, stderr)

	// Expected: Phases execute in order
	expected := []string{"gitignore", "directories", "workflows", "agents", "specfiles"}
	assert.Equal(t, expected, deps.callOrder)
}

// --- Error Propagation ---

func TestInit_GetwdFails(t *testing.T) {
	deps := newFakeInitDeps()
	deps.getwdErr = errFakeGetwdFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stderr contains error; exit code 1
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "Error: failed to determine working directory:")
}

func TestInit_Phase0_GitignoreFails(t *testing.T) {
	deps := newFakeInitDeps()
	deps.gitignoreEnsurer.err = errFakePhaseFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stderr contains error; exit code 1; DirectoryCreator.CreateAll() not called
	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr.String())
	assert.False(t, deps.directoryCreator.called, "DirectoryCreator.CreateAll() should not have been called")
}

func TestInit_Phase1_DirectoryCreatorFails(t *testing.T) {
	deps := newFakeInitDeps()
	deps.directoryCreator.err = errFakePhaseFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stderr contains error; exit code 1; BuiltinResourceCopier methods not called
	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr.String())
	assert.False(t, deps.copier.copyWorkflowsCalled, "CopyWorkflows should not have been called")
	assert.False(t, deps.copier.copyAgentsCalled, "CopyAgents should not have been called")
	assert.False(t, deps.copier.copySpecFilesCalled, "CopySpecFiles should not have been called")
}

func TestInit_Phase2a_CopyWorkflowsFails(t *testing.T) {
	deps := newFakeInitDeps()
	deps.copier.copyWorkflowsErr = errFakePhaseFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stderr contains error; CopyAgents and CopySpecFiles not called
	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr.String())
	assert.False(t, deps.copier.copyAgentsCalled, "CopyAgents should not have been called")
	assert.False(t, deps.copier.copySpecFilesCalled, "CopySpecFiles should not have been called")
}

func TestInit_Phase2b_CopyAgentsFails(t *testing.T) {
	deps := newFakeInitDeps()
	deps.copier.copyAgentsErr = errFakePhaseFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stderr contains error; CopySpecFiles not called
	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr.String())
	assert.False(t, deps.copier.copySpecFilesCalled, "CopySpecFiles should not have been called")
}

func TestInit_Phase2c_CopySpecFilesFails(t *testing.T) {
	deps := newFakeInitDeps()
	deps.copier.copySpecFilesErr = errFakePhaseFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stderr contains error; exit code 1
	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr.String())
}

// --- Mock / Dependency Interaction ---

func TestInit_PassesProjectRootToAllPhases(t *testing.T) {
	deps := newFakeInitDeps()
	deps.cwd = "/fake/project"
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runInit(deps, stdout, stderr)

	// Expected: All phases receive /fake/project as projectRoot
	assert.Equal(t, "/fake/project", deps.gitignoreEnsurer.receivedProjectRoot)
	assert.Equal(t, "/fake/project", deps.directoryCreator.receivedProjectRoot)
	assert.Equal(t, "/fake/project", deps.copier.receivedProjectRoot)
}

func TestInit_FailFast_Phase0_SkipsSubsequent(t *testing.T) {
	deps := newFakeInitDeps()
	deps.gitignoreEnsurer.err = errFakePhaseFailure
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runInit(deps, stdout, stderr)

	// Expected: No subsequent phases called
	assert.False(t, deps.directoryCreator.called)
	assert.False(t, deps.copier.copyWorkflowsCalled)
	assert.False(t, deps.copier.copyAgentsCalled)
	assert.False(t, deps.copier.copySpecFilesCalled)
}

// --- Idempotency ---

func TestInit_ReInitialization_WarningsOnly(t *testing.T) {
	deps := newFakeInitDeps()
	deps.copier.copyWorkflowsWarnings = []string{"workflow A exists"}
	deps.copier.copyAgentsWarnings = []string{"agent B exists"}
	deps.copier.copySpecFilesWarnings = []string{"spec C exists"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := runInit(deps, stdout, stderr)

	// Expected: stdout contains all warnings and success message
	assert.Equal(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "workflow A exists")
	assert.Contains(t, output, "agent B exists")
	assert.Contains(t, output, "spec C exists")
	assert.Contains(t, output, "Spectra project initialized successfully")
	assert.Empty(t, stderr.String())
}

// --- Utility: suppress unused import warnings ---

var (
	_ = require.NoError
)
