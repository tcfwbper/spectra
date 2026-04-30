package spectra_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

// TestInit_UsesBuiltinResourceCopier verifies init command uses BuiltinResourceCopier for copying files.
func TestInit_UsesBuiltinResourceCopier(t *testing.T) {
	tmpDir := t.TempDir()

	mock := spectra.NewMockBuiltinResourceCopier()
	handler := spectra.NewInitHandler(tmpDir, mock)

	exitCode := handler.Execute()

	assert.Equal(t, 0, exitCode)
	assert.True(t, mock.CopyWorkflowsCalled(), "CopyWorkflows should be called")
	assert.True(t, mock.CopyAgentsCalled(), "CopyAgents should be called")
	assert.True(t, mock.CopySpecFilesCalled(), "CopySpecFiles should be called")
	assert.Equal(t, tmpDir, mock.CopyWorkflowsProjectRoot(), "CopyWorkflows called with correct project root")
	assert.Equal(t, tmpDir, mock.CopyAgentsProjectRoot(), "CopyAgents called with correct project root")
	assert.Equal(t, tmpDir, mock.CopySpecFilesProjectRoot(), "CopySpecFiles called with correct project root")
}

// TestInit_UsesEmbedFS verifies init command uses embed.FS for built-in files.
func TestInit_UsesEmbedFS(t *testing.T) {
	// Verify embedded filesystems builtinWorkflows, builtinAgents, and builtinSpecFiles
	// are populated from //go:embed directives
	workflows := spectra.BuiltinWorkflowsFS()
	agents := spectra.BuiltinAgentsFS()
	specFiles := spectra.BuiltinSpecFilesFS()

	assert.NotNil(t, workflows, "builtinWorkflows embed.FS should be populated")
	assert.NotNil(t, agents, "builtinAgents embed.FS should be populated")
	assert.NotNil(t, specFiles, "builtinSpecFiles embed.FS should be populated")
}
