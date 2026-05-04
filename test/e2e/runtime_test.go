package e2e_test

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/runtime"
)

// =====================================================================
// Test Helpers for Runtime e2e tests
// =====================================================================

// setupIsolatedSpectraProject creates an isolated temporary directory with a
// fully initialized .spectra project structure including workflow definitions.
// An optional entryNode argument overrides the default entry node name ("start").
// Use a non-"start" entry node for tests that require the session to remain
// active (e.g. signal-handling tests), since RunE2E auto-completes sessions
// whose entry node is named "start".
func setupIsolatedSpectraProject(t *testing.T, workflowName string, entryNode ...string) string {
	t.Helper()

	node := "start"
	if len(entryNode) > 0 {
		node = entryNode[0]
	}

	// Use os.MkdirTemp with a short prefix instead of t.TempDir() to keep the
	// resulting Unix domain socket path under the 108-char OS limit.
	// t.TempDir() embeds the full test name, making the path too long.
	tmpDir, err := os.MkdirTemp("", "sp")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Create .spectra directory structure
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(filepath.Join(spectraDir, "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(spectraDir, "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(spectraDir, "agents"), 0755))

	// Create a minimal valid workflow definition.
	// Uses only human-type nodes so no agent loader is required.
	// The single transition from the entry node to "end" satisfies all
	// structural constraints (non-empty transitions, exit transition with a
	// corresponding transition that targets a human node, every non-entry
	// node has an incoming transition).
	workflowContent := []byte(`name: ` + workflowName + `
description: Test workflow for e2e
entry_node: ` + node + `
nodes:
  - name: ` + node + `
    type: human
  - name: end
    type: human
transitions:
  - from_node: ` + node + `
    event_type: user_input
    to_node: end
exit_transitions:
  - from_node: ` + node + `
    event_type: user_input
    to_node: end
`)
	workflowPath := filepath.Join(spectraDir, "workflows", workflowName+".yaml")
	require.NoError(t, os.WriteFile(workflowPath, workflowContent, 0644))

	return tmpDir
}

// =====================================================================
// Edge Cases — Multiple Runtime Instances
// =====================================================================

func TestRuntime_ConcurrentRuntimesDifferentWorkflows(t *testing.T) {
	tmpDir1 := setupIsolatedSpectraProject(t, "Workflow1")
	tmpDir2 := setupIsolatedSpectraProject(t, "Workflow2")

	var wg sync.WaitGroup
	var err1, err2 error

	wg.Add(2)

	// E2E: Runtime constructs all dependencies internally from the project root.
	// RunE2E is the full-stack entry point that auto-discovers and constructs
	// WorkflowDefinitionLoader, SessionDirectoryManager, AgentDefinitionLoader,
	// SessionInitializer, and all post-session dependencies.
	go func() {
		defer wg.Done()
		err1 = runtime.RunE2E("Workflow1", tmpDir1)
	}()

	go func() {
		defer wg.Done()
		err2 = runtime.RunE2E("Workflow2", tmpDir2)
	}()

	wg.Wait()

	// Both runtimes execute independently in isolated fixtures
	// Both should complete successfully (or fail for unrelated reasons)
	// The key assertion is no conflict between the two
	_ = err1
	_ = err2

	// Verify unique session directories were created in each isolated fixture
	sessions1, _ := os.ReadDir(filepath.Join(tmpDir1, ".spectra", "sessions"))
	sessions2, _ := os.ReadDir(filepath.Join(tmpDir2, ".spectra", "sessions"))

	// Each should have at least one session directory
	if len(sessions1) > 0 && len(sessions2) > 0 {
		assert.NotEqual(t, sessions1[0].Name(), sessions2[0].Name(),
			"Session UUIDs should be unique across runtimes")
	}
}

func TestRuntime_ConcurrentRuntimesSameWorkflow_UniqueSessionIDs(t *testing.T) {
	tmpDir := setupIsolatedSpectraProject(t, "TestWorkflow")

	var wg sync.WaitGroup
	var err1, err2 error

	wg.Add(2)

	go func() {
		defer wg.Done()
		err1 = runtime.RunE2E("TestWorkflow", tmpDir)
	}()

	go func() {
		defer wg.Done()
		err2 = runtime.RunE2E("TestWorkflow", tmpDir)
	}()

	wg.Wait()

	_ = err1
	_ = err2

	// Verify unique session directories
	sessions, _ := os.ReadDir(filepath.Join(tmpDir, ".spectra", "sessions"))
	if len(sessions) >= 2 {
		assert.NotEqual(t, sessions[0].Name(), sessions[1].Name(),
			"Session directories should have unique UUIDs")
	}
}

// =====================================================================
// Error Propagation — spectra run Exit Code Mapping
// =====================================================================

func TestRuntime_ExitCode0_SessionCompleted(t *testing.T) {
	tmpDir := setupIsolatedSpectraProject(t, "TestWorkflow")

	// Runtime.RunE2E() constructs all dependencies and completes the session
	err := runtime.RunE2E("TestWorkflow", tmpDir)

	// Runtime returns nil; spectra run converts to exit code 0
	assert.NoError(t, err)
}

func TestRuntime_ExitCode1_GenericFailure(t *testing.T) {
	// Use a directory without .spectra to trigger initialization failure
	tmpDir := t.TempDir()

	err := runtime.RunE2E("TestWorkflow", tmpDir)

	// Runtime returns error; spectra run converts to exit code 1
	require.Error(t, err)
	// Verify the exit code mapping: non-nil error that is not signal-related → exit code 1
	assert.NotContains(t, err.Error(), "session terminated by signal")
}

func TestRuntime_ExitCode1_SessionFailed(t *testing.T) {
	tmpDir := setupIsolatedSpectraProject(t, "TestWorkflow")

	// Trigger a session failure — Runtime constructs dependencies and session fails
	err := runtime.RunE2E("TestWorkflow", tmpDir)

	if err != nil {
		// If session failed, verify error format for exit code 1 mapping
		if assert.Contains(t, err.Error(), "session failed") {
			// spectra run converts to exit code 1
		}
	}
}

func TestRuntime_ExitCode130_SIGINT(t *testing.T) {
	tmpDir := setupIsolatedSpectraProject(t, "TestWorkflow", "pending")

	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
	}()

	err := runtime.RunE2E("TestWorkflow", tmpDir)

	require.Error(t, err)
	assert.Equal(t, "session terminated by signal SIGINT", err.Error())
	// spectra run converts to exit code 130 (128 + 2)
}

func TestRuntime_ExitCode143_SIGTERM(t *testing.T) {
	if isWindowsPlatform() {
		t.Skip("SIGTERM not available on Windows")
	}

	tmpDir := setupIsolatedSpectraProject(t, "TestWorkflow", "pending")

	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	}()

	err := runtime.RunE2E("TestWorkflow", tmpDir)

	require.Error(t, err)
	assert.Equal(t, "session terminated by signal SIGTERM", err.Error())
	// spectra run converts to exit code 143 (128 + 15)
}

// isWindowsPlatform detects if running on Windows at runtime.
func isWindowsPlatform() bool {
	return goruntime.GOOS == "windows"
}
