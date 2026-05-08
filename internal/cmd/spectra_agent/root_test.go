package spectraagent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errNotInitialized is a test-side sentinel matching storage.ErrNotInitialized.
// The production code imports the real sentinel; tests use this to configure fakes.
var errNotInitialized = errors.New("spectra not initialized: .spectra directory not found")

// --- Happy Path — Execute ---

func TestExecute_NoSubcommandPrintsUsage(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc123"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "spectra-agent [command]")
}

func TestExecute_HelpFlag(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--help"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "spectra-agent")
}

// --- Validation Failures — session-id ---

func TestExecute_MissingSessionID(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"error", "msg"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--session-id flag is required")
}

func TestExecute_EmptySessionID(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "", "error", "msg"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--session-id flag is required")
}

func TestExecute_InvalidUUIDSessionIDAccepted(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "not-a-uuid", "--help"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)
	assert.NotContains(t, result.stderr, "session-id")
}

// --- Validation Failures ---

func TestExecute_ProjectRootNotFound(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{err: errNotInitialized}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "msg"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, ".spectra directory not found. Are you in a Spectra project?")
}

func TestExecute_UnknownSubcommand(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "unknown"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "unknown")
}

// --- Error Propagation ---

func TestExecute_PropagatesSubcommandExitCode2(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 2, stderr: "Error: socket not found"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "some error"}, finder, sender)

	assert.Equal(t, 2, result.exitCode)
}

func TestExecute_PropagatesSubcommandExitCode3(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 3, stderr: "Error: runtime error"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "some error"}, finder, sender)

	assert.Equal(t, 3, result.exitCode)
}

// --- Mock / Dependency Interaction ---

func TestExecute_CallsSpectraFinderWithEmptyString(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	executeCommand(t, []string{"--session-id", "abc", "--help"}, finder, sender)

	calls := finder.calls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "", calls[0])
}

func TestExecute_PropagatesSessionIDToSubcommand(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	executeCommand(t, []string{"--session-id", "my-sess-123", "error", "msg"}, finder, sender)

	calls := sender.calls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "my-sess-123", calls[0].sessionID)
}

func TestExecute_PropagatesProjectRootToSubcommand(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/my-project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	executeCommand(t, []string{"--session-id", "abc", "error", "msg"}, finder, sender)

	calls := sender.calls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "/tmp/my-project", calls[0].projectRoot)
}
