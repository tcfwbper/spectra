package storage_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfwbper/spectra/storage"
)

// TestStorageLayout_SpectraDirConstant verifies SpectraDir constant value
func TestStorageLayout_SpectraDirConstant(t *testing.T) {
	assert.Equal(t, ".spectra", storage.SpectraDir)
}

// TestStorageLayout_SessionsDirConstant verifies SessionsDir constant value
func TestStorageLayout_SessionsDirConstant(t *testing.T) {
	assert.Equal(t, ".spectra/sessions", storage.SessionsDir)
}

// TestStorageLayout_WorkflowsDirConstant verifies WorkflowsDir constant value
func TestStorageLayout_WorkflowsDirConstant(t *testing.T) {
	assert.Equal(t, ".spectra/workflows", storage.WorkflowsDir)
}

// TestStorageLayout_AgentsDirConstant verifies AgentsDir constant value
func TestStorageLayout_AgentsDirConstant(t *testing.T) {
	assert.Equal(t, ".spectra/agents", storage.AgentsDir)
}

// TestStorageLayout_SessionMetadataFileConstant verifies SessionMetadataFile constant value
func TestStorageLayout_SessionMetadataFileConstant(t *testing.T) {
	assert.Equal(t, "session.json", storage.SessionMetadataFile)
}

// TestStorageLayout_EventHistoryFileConstant verifies EventHistoryFile constant value
func TestStorageLayout_EventHistoryFileConstant(t *testing.T) {
	assert.Equal(t, "events.jsonl", storage.EventHistoryFile)
}

// TestStorageLayout_RuntimeSocketFileConstant verifies RuntimeSocketFile constant value
func TestStorageLayout_RuntimeSocketFileConstant(t *testing.T) {
	assert.Equal(t, "runtime.sock", storage.RuntimeSocketFile)
}

// TestGetSpectraDir_AbsolutePath returns absolute path to .spectra directory
func TestGetSpectraDir_AbsolutePath(t *testing.T) {
	projectRoot := "/home/user/project"
	expected := filepath.Join(projectRoot, ".spectra")
	result := storage.GetSpectraDir(projectRoot)
	assert.Equal(t, expected, result)
}

// TestGetSpectraDir_TrailingSlash handles project root with trailing slash correctly
func TestGetSpectraDir_TrailingSlash(t *testing.T) {
	projectRoot := "/home/user/project/"
	result := storage.GetSpectraDir(projectRoot)
	assert.NotContains(t, result, "//")
	assert.True(t, filepath.IsAbs(result) || filepath.VolumeName(result) != "")
}

// TestGetSessionsDir_AbsolutePath returns absolute path to sessions directory
func TestGetSessionsDir_AbsolutePath(t *testing.T) {
	projectRoot := "/home/user/project"
	expected := filepath.Join(projectRoot, ".spectra", "sessions")
	result := storage.GetSessionsDir(projectRoot)
	assert.Equal(t, expected, result)
}

// TestGetWorkflowsDir_AbsolutePath returns absolute path to workflows directory
func TestGetWorkflowsDir_AbsolutePath(t *testing.T) {
	projectRoot := "/home/user/project"
	expected := filepath.Join(projectRoot, ".spectra", "workflows")
	result := storage.GetWorkflowsDir(projectRoot)
	assert.Equal(t, expected, result)
}

// TestGetAgentsDir_AbsolutePath returns absolute path to agents directory
func TestGetAgentsDir_AbsolutePath(t *testing.T) {
	projectRoot := "/home/user/project"
	expected := filepath.Join(projectRoot, ".spectra", "agents")
	result := storage.GetAgentsDir(projectRoot)
	assert.Equal(t, expected, result)
}

// TestGetSessionDir_ValidUUID returns absolute path to session directory with valid UUID
func TestGetSessionDir_ValidUUID(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	expected := filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID)
	result := storage.GetSessionDir(projectRoot, sessionUUID)
	assert.Equal(t, expected, result)
}

// TestGetSessionDir_UUIDWithUppercase preserves UUID case as provided
func TestGetSessionDir_UUIDWithUppercase(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "123E4567-E89B-12D3-A456-426614174000"
	result := storage.GetSessionDir(projectRoot, sessionUUID)
	assert.Contains(t, result, sessionUUID)
}

// TestGetSessionMetadataPath_ValidUUID returns absolute path to session.json file
func TestGetSessionMetadataPath_ValidUUID(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	expected := filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID, "session.json")
	result := storage.GetSessionMetadataPath(projectRoot, sessionUUID)
	assert.Equal(t, expected, result)
}

// TestGetEventHistoryPath_ValidUUID returns absolute path to events.jsonl file
func TestGetEventHistoryPath_ValidUUID(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	expected := filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID, "events.jsonl")
	result := storage.GetEventHistoryPath(projectRoot, sessionUUID)
	assert.Equal(t, expected, result)
}

// TestGetRuntimeSocketPath_ValidUUID returns absolute path to runtime.sock file
func TestGetRuntimeSocketPath_ValidUUID(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	expected := filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID, "runtime.sock")
	result := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	assert.Equal(t, expected, result)
}

// TestGetWorkflowPath_PascalCaseName returns absolute path to workflow YAML file
func TestGetWorkflowPath_PascalCaseName(t *testing.T) {
	projectRoot := "/home/user/project"
	workflowName := "CodeReview"
	expected := filepath.Join(projectRoot, ".spectra", "workflows", workflowName+".yaml")
	result := storage.GetWorkflowPath(projectRoot, workflowName)
	assert.Equal(t, expected, result)
}

// TestGetAgentPath_PascalCaseRole returns absolute path to agent YAML file
func TestGetAgentPath_PascalCaseRole(t *testing.T) {
	projectRoot := "/home/user/project"
	agentRole := "Architect"
	expected := filepath.Join(projectRoot, ".spectra", "agents", agentRole+".yaml")
	result := storage.GetAgentPath(projectRoot, agentRole)
	assert.Equal(t, expected, result)
}

// TestGetSessionDir_EmptyUUID returns malformed path with empty UUID
func TestGetSessionDir_EmptyUUID(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := ""
	result := storage.GetSessionDir(projectRoot, sessionUUID)
	expected := filepath.Join(projectRoot, ".spectra", "sessions", "")
	assert.Equal(t, expected, result)
}

// TestGetWorkflowPath_EmptyName returns malformed path with empty workflow name
func TestGetWorkflowPath_EmptyName(t *testing.T) {
	projectRoot := "/home/user/project"
	workflowName := ""
	result := storage.GetWorkflowPath(projectRoot, workflowName)
	assert.Contains(t, result, ".yaml")
}

// TestGetAgentPath_EmptyRole returns malformed path with empty agent role
func TestGetAgentPath_EmptyRole(t *testing.T) {
	projectRoot := "/home/user/project"
	agentRole := ""
	result := storage.GetAgentPath(projectRoot, agentRole)
	assert.Contains(t, result, ".yaml")
}

// TestGetSpectraDir_RelativePath returns relative path when project root is relative
func TestGetSpectraDir_RelativePath(t *testing.T) {
	projectRoot := "./project"
	result := storage.GetSpectraDir(projectRoot)
	expected := filepath.Join(".", "project", ".spectra")
	assert.Equal(t, expected, result)
}

// TestGetSessionDir_RelativePathProjectRoot returns relative path when project root is relative
func TestGetSessionDir_RelativePathProjectRoot(t *testing.T) {
	projectRoot := "./project"
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	result := storage.GetSessionDir(projectRoot, sessionUUID)
	expected := filepath.Join(".", "project", ".spectra", "sessions", sessionUUID)
	assert.Equal(t, expected, result)
}

// TestGetSessionDir_UUIDWithPathSeparator does not validate UUID format; passes through as-is
func TestGetSessionDir_UUIDWithPathSeparator(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "../malicious"
	result := storage.GetSessionDir(projectRoot, sessionUUID)
	assert.Contains(t, result, "..")
	assert.Contains(t, result, "malicious")
}

// TestGetWorkflowPath_NameWithPathSeparator does not validate workflow name; passes through as-is
func TestGetWorkflowPath_NameWithPathSeparator(t *testing.T) {
	projectRoot := "/home/user/project"
	workflowName := "../malicious/workflow"
	result := storage.GetWorkflowPath(projectRoot, workflowName)
	assert.Contains(t, result, "..")
	assert.Contains(t, result, "malicious")
}

// TestStorageLayout_IdempotentComposition tests that multiple calls with same inputs return identical paths
func TestStorageLayout_IdempotentComposition(t *testing.T) {
	projectRoot := "/home/user/project"
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	result1 := storage.GetSessionDir(projectRoot, sessionUUID)
	result2 := storage.GetSessionDir(projectRoot, sessionUUID)
	result3 := storage.GetSessionDir(projectRoot, sessionUUID)

	assert.Equal(t, result1, result2)
	assert.Equal(t, result2, result3)
}

// TestStorageLayout_PlatformSpecificSeparators uses platform-appropriate path separators
func TestStorageLayout_PlatformSpecificSeparators(t *testing.T) {
	projectRoot := "/home/user/project"
	if runtime.GOOS == "windows" {
		projectRoot = "C:\\Users\\user\\project"
	}

	result := storage.GetSessionDir(projectRoot, "test-uuid")

	if runtime.GOOS == "windows" {
		assert.Contains(t, result, "\\")
	} else {
		assert.Contains(t, result, "/")
	}
}
