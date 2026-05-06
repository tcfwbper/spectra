package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — GetSpectraDir ---

func TestGetSpectraDir_AbsolutePath(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSpectraDir("/home/user/project")
	assert.Equal(t, "/home/user/project/.spectra", result)
}

func TestGetSpectraDir_TrailingSlash(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSpectraDir("/home/user/project/")
	// filepath.Join normalizes trailing slash
	expected := filepath.Join("/home/user/project", ".spectra")
	assert.Equal(t, expected, result)
}

// --- Happy Path — GetSessionsDir ---

func TestGetSessionsDir_AbsolutePath(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSessionsDir("/home/user/project")
	assert.Equal(t, "/home/user/project/.spectra/sessions", result)
}

// --- Happy Path — GetWorkflowsDir ---

func TestGetWorkflowsDir_AbsolutePath(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetWorkflowsDir("/home/user/project")
	assert.Equal(t, "/home/user/project/.spectra/workflows", result)
}

// --- Happy Path — GetAgentsDir ---

func TestGetAgentsDir_AbsolutePath(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetAgentsDir("/home/user/project")
	assert.Equal(t, "/home/user/project/.spectra/agents", result)
}

// --- Happy Path — GetSessionDir ---

func TestGetSessionDir_ValidUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSessionDir("/home/user/project", "550e8400-e29b-41d4-a716-446655440000")
	assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000", result)
}

// --- Happy Path — GetSessionMetadataPath ---

func TestGetSessionMetadataPath_ValidUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSessionMetadataPath("/home/user/project", "550e8400-e29b-41d4-a716-446655440000")
	assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/session.json", result)
}

// --- Happy Path — GetEventHistoryPath ---

func TestGetEventHistoryPath_ValidUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetEventHistoryPath("/home/user/project", "550e8400-e29b-41d4-a716-446655440000")
	assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/events.jsonl", result)
}

// --- Happy Path — GetRuntimeSocketPath ---

func TestGetRuntimeSocketPath_ValidUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetRuntimeSocketPath("/home/user/project", "550e8400-e29b-41d4-a716-446655440000")
	assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/runtime.sock", result)
}

// --- Happy Path — GetWorkflowPath ---

func TestGetWorkflowPath_ValidName(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetWorkflowPath("/home/user/project", "CodeReview")
	assert.Equal(t, "/home/user/project/.spectra/workflows/CodeReview.yaml", result)
}

// --- Happy Path — GetAgentPath ---

func TestGetAgentPath_ValidRole(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetAgentPath("/home/user/project", "Architect")
	assert.Equal(t, "/home/user/project/.spectra/agents/Architect.yaml", result)
}

// --- Null / Empty Input ---

func TestGetSessionDir_EmptyUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSessionDir("/home/user/project", "")
	// Returns malformed path ending with `.spectra/sessions/`; no panic
	assert.Contains(t, result, ".spectra/sessions")
}

func TestGetSpectraDir_EmptyProjectRoot(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSpectraDir("")
	assert.Equal(t, ".spectra", result)
}

func TestGetWorkflowPath_EmptyName(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetWorkflowPath("/home/user/project", "")
	// Returns path ending with `.yaml`; no panic
	assert.Contains(t, result, ".yaml")
}

func TestGetAgentPath_EmptyRole(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetAgentPath("/home/user/project", "")
	// Returns path ending with `.yaml`; no panic
	assert.Contains(t, result, ".yaml")
}

// --- Boundary Values — projectRoot ---

func TestGetSpectraDir_RelativeProjectRoot(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSpectraDir("./project")
	// filepath.Join normalizes `./`
	assert.Equal(t, "project/.spectra", result)
}

func TestGetSessionDir_PathSeparatorInUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/storage_layout.go")

	result := GetSessionDir("/home/user/project", "../malicious")
	// filepath.Join resolves the path; no error or panic
	expected := filepath.Join("/home/user/project", ".spectra", "sessions", "../malicious")
	assert.Equal(t, expected, result)
}
