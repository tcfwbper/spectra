package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

func setupAgentTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, AgentsDir)
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	return tmpDir
}

func writeAgentYAML(t *testing.T, projectRoot, agentRole, content string) {
	t.Helper()
	agentPath := GetAgentPath(projectRoot, agentRole)
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))
}

func createValidAgentYAML(role, agentRoot string) string {
	return `role: "` + role + `"
model: "sonnet"
effort: "high"
system_prompt: "You are a test agent."
agent_root: "` + agentRoot + `"
allowed_tools: []
disallowed_tools: []
`
}

// Happy Path — Construction

func TestAgentDefinitionLoader_New(t *testing.T) {
	tmpDir := setupAgentTestDir(t)

	loader := NewAgentDefinitionLoader(tmpDir)

	assert.NotNil(t, loader)
}

// Happy Path — Load

func TestAgentDefinitionLoader_Load_ValidDefinition(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", createValidAgentYAML("Architect", "."))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	require.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, "Architect", def.Role)
	assert.Equal(t, "sonnet", def.Model)
	assert.Equal(t, "high", def.Effort)
	assert.Equal(t, "You are a test agent.", def.SystemPrompt)
	assert.Equal(t, ".", def.AgentRoot)
	assert.Empty(t, def.AllowedTools)
	assert.Empty(t, def.DisallowedTools)
}

func TestAgentDefinitionLoader_Load_WithOptionalTools(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Coder"
model: "sonnet"
effort: "medium"
system_prompt: "You are a coding agent."
agent_root: "."
allowed_tools:
  - "Read"
  - "Write"
disallowed_tools:
  - "Bash"
`
	writeAgentYAML(t, tmpDir, "Coder", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Coder")

	require.NoError(t, err)
	assert.Equal(t, []string{"Read", "Write"}, def.AllowedTools)
	assert.Equal(t, []string{"Bash"}, def.DisallowedTools)
}

func TestAgentDefinitionLoader_Load_WithoutOptionalTools(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Reviewer"
model: "sonnet"
effort: "low"
system_prompt: "You are a reviewer."
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "Reviewer", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Reviewer")

	require.NoError(t, err)
	assert.Empty(t, def.AllowedTools)
	assert.Empty(t, def.DisallowedTools)
}

func TestAgentDefinitionLoader_Load_RoleWithDigits(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "V2Architect", createValidAgentYAML("V2Architect", "."))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("V2Architect")

	require.NoError(t, err)
	assert.Equal(t, "V2Architect", def.Role)
}

func TestAgentDefinitionLoader_Load_RoleWithConsecutiveUppercase(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "QAReviewer", createValidAgentYAML("QAReviewer", "."))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("QAReviewer")

	require.NoError(t, err)
	assert.Equal(t, "QAReviewer", def.Role)
}

func TestAgentDefinitionLoader_Load_SingleUppercaseLetter(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "A", createValidAgentYAML("A", "."))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("A")

	require.NoError(t, err)
	assert.Equal(t, "A", def.Role)
}

func TestAgentDefinitionLoader_Load_AgentRootDot(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "ProjectAgent", createValidAgentYAML("ProjectAgent", "."))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("ProjectAgent")

	require.NoError(t, err)
	assert.Equal(t, ".", def.AgentRoot)
}

func TestAgentDefinitionLoader_Load_AgentRootNestedPath(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	nestedDir := filepath.Join(tmpDir, "agents", "subdir")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))
	writeAgentYAML(t, tmpDir, "NestedAgent", createValidAgentYAML("NestedAgent", "agents/subdir"))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("NestedAgent")

	require.NoError(t, err)
	assert.Equal(t, "agents/subdir", def.AgentRoot)
}

func TestAgentDefinitionLoader_Load_SystemPromptWithYAMLFrontMatter(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "PromptAgent"
model: "sonnet"
effort: "medium"
system_prompt: "---\ntitle: Test\n---\nYou are..."
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "PromptAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("PromptAgent")

	require.NoError(t, err)
	assert.Contains(t, def.SystemPrompt, "---")
}

func TestAgentDefinitionLoader_Load_InvalidModelValue(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "invalid-model-123"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	require.NoError(t, err)
	assert.Equal(t, "invalid-model-123", def.Model)
}

func TestAgentDefinitionLoader_Load_InvalidEffortValue(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "ultra-mega-high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	require.NoError(t, err)
	assert.Equal(t, "ultra-mega-high", def.Effort)
}

func TestAgentDefinitionLoader_Load_InvalidToolIdentifiers(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "medium"
system_prompt: "Test prompt"
agent_root: "."
allowed_tools:
  - "Invalid(**)"
  - "Bad@Tool"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	require.NoError(t, err)
	assert.Equal(t, []string{"Invalid(**)", "Bad@Tool"}, def.AllowedTools)
}

func TestAgentDefinitionLoader_Load_ConflictingToolLists(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "medium"
system_prompt: "Test prompt"
agent_root: "."
allowed_tools:
  - "Read"
disallowed_tools:
  - "Read"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	require.NoError(t, err)
	assert.Contains(t, def.AllowedTools, "Read")
	assert.Contains(t, def.DisallowedTools, "Read")
}

// Validation Failures — File Not Found

func TestAgentDefinitionLoader_Load_FileNotFound(t *testing.T) {
	tmpDir := setupAgentTestDir(t)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition not found: Architect")
}

func TestAgentDefinitionLoader_Load_EmptyAgentRole(t *testing.T) {
	tmpDir := setupAgentTestDir(t)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition not found:")
}

// Validation Failures — File Read Errors

func TestAgentDefinitionLoader_Load_PermissionDenied(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", createValidAgentYAML("Architect", "."))
	agentPath := GetAgentPath(tmpDir, "Architect")
	require.NoError(t, os.Chmod(agentPath, 0000))
	defer os.Chmod(agentPath, 0644)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read agent definition 'Architect'")
	assert.Contains(t, err.Error(), "permission denied")
}

// Validation Failures — YAML Parsing

func TestAgentDefinitionLoader_Load_EmptyFile(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", "")

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent definition 'Architect'")
	assert.Contains(t, err.Error(), "EOF")
}

func TestAgentDefinitionLoader_Load_InvalidYAMLSyntax(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	invalidYAML := "role:\n  - invalid:\nbroken"
	writeAgentYAML(t, tmpDir, "Architect", invalidYAML)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent definition 'Architect'")
	assert.Contains(t, err.Error(), "yaml: line")
}

func TestAgentDefinitionLoader_Load_UnknownFieldsIgnored(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Architect"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
custom_metadata: "extra"
`
	writeAgentYAML(t, tmpDir, "Architect", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	require.NoError(t, err)
	assert.NotNil(t, def)
}

// Validation Failures — Missing Required Fields

func TestAgentDefinitionLoader_Load_MissingRole(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "Architect", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Architect' validation failed: missing required field 'role'")
}

func TestAgentDefinitionLoader_Load_MissingModel(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Architect"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "Architect", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Architect' validation failed: missing required field 'model'")
}

func TestAgentDefinitionLoader_Load_MissingEffort(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Architect"
model: "sonnet"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "Architect", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Architect' validation failed: missing required field 'effort'")
}

func TestAgentDefinitionLoader_Load_MissingSystemPrompt(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Architect"
model: "sonnet"
effort: "high"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "Architect", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Architect' validation failed: missing required field 'system_prompt'")
}

func TestAgentDefinitionLoader_Load_MissingAgentRoot(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "Architect"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
`
	writeAgentYAML(t, tmpDir, "Architect", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Architect' validation failed: missing required field 'agent_root'")
}

// Validation Failures — Role Format

func TestAgentDefinitionLoader_Load_RoleWithSpaces(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "QA Reviewer"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "QAReviewer", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("QAReviewer")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'QAReviewer' validation failed: role must be PascalCase with no spaces or special characters")
}

func TestAgentDefinitionLoader_Load_RoleWithUnderscore(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "QA_Reviewer"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "QA_Reviewer", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("QA_Reviewer")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'QA_Reviewer' validation failed: role must be PascalCase with no spaces or special characters")
}

func TestAgentDefinitionLoader_Load_RoleWithHyphen(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "QA-Reviewer"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "QA-Reviewer", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("QA-Reviewer")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'QA-Reviewer' validation failed: role must be PascalCase with no spaces or special characters")
}

func TestAgentDefinitionLoader_Load_RoleWithDot(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "QA.Reviewer"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "QA.Reviewer", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("QA.Reviewer")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'QA.Reviewer' validation failed: role must be PascalCase with no spaces or special characters")
}

func TestAgentDefinitionLoader_Load_RoleStartsLowercase(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "qaReviewer"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "qaReviewer", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("qaReviewer")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'qaReviewer' validation failed: role must be PascalCase with no spaces or special characters")
}

// Validation Failures — AgentRoot Path

func TestAgentDefinitionLoader_Load_AgentRootAbsolutePath(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "/usr/local/bin"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'TestAgent' validation failed: agent_root must be a relative path")
}

func TestAgentDefinitionLoader_Load_AgentRootWithDriveLetter(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "C:\\Users\\test"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'TestAgent' validation failed: agent_root must be a relative path")
}

func TestAgentDefinitionLoader_Load_AgentRootDirectoryNotFound(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "nonexistent/dir"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'TestAgent' validation failed: agent_root directory not found:")
}

func TestAgentDefinitionLoader_Load_AgentRootIsFile(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	filePath := filepath.Join(tmpDir, "somefile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "somefile.txt"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'TestAgent' validation failed: agent_root is not a directory:")
}

func TestAgentDefinitionLoader_Load_AgentRootSymlinkToDirectory(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	targetDir := filepath.Join(tmpDir, "target")
	linkPath := filepath.Join(tmpDir, "link")
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	require.NoError(t, os.Symlink(targetDir, linkPath))
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "link"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	require.NoError(t, err)
	assert.NotNil(t, def)
}

func TestAgentDefinitionLoader_Load_AgentRootSymlinkToNonexistent(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	linkPath := filepath.Join(tmpDir, "broken_link")
	require.NoError(t, os.Symlink(filepath.Join(tmpDir, "nonexistent"), linkPath))
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "broken_link"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'TestAgent' validation failed: agent_root directory not found:")
}

func TestAgentDefinitionLoader_Load_AgentRootUnreadableDirectory(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	restrictedDir := filepath.Join(tmpDir, "restricted")
	require.NoError(t, os.MkdirAll(restrictedDir, 0000))
	defer os.Chmod(restrictedDir, 0755)
	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "high"
system_prompt: "Test prompt"
agent_root: "restricted"
`
	writeAgentYAML(t, tmpDir, "TestAgent", yamlContent)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("TestAgent")

	require.NoError(t, err)
	assert.NotNil(t, def)
}

// Validation Failures — Path Injection

func TestAgentDefinitionLoader_Load_AgentRoleWithPathTraversal(t *testing.T) {
	tmpDir := setupAgentTestDir(t)

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("../malicious/agent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition not found:")
}

// Idempotency

func TestAgentDefinitionLoader_Load_RepeatedCalls(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", createValidAgentYAML("Architect", "."))

	loader := NewAgentDefinitionLoader(tmpDir)

	def1, err1 := loader.Load("Architect")
	require.NoError(t, err1)

	def2, err2 := loader.Load("Architect")
	require.NoError(t, err2)

	def3, err3 := loader.Load("Architect")
	require.NoError(t, err3)

	assert.Equal(t, def1, def2)
	assert.Equal(t, def2, def3)
}

func TestAgentDefinitionLoader_Load_FileModifiedBetweenCalls(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", createValidAgentYAML("Architect", "."))

	loader := NewAgentDefinitionLoader(tmpDir)

	def1, err1 := loader.Load("Architect")
	require.NoError(t, err1)
	assert.Equal(t, "high", def1.Effort)

	// Modify file
	modifiedYAML := `role: "Architect"
model: "sonnet"
effort: "low"
system_prompt: "Modified prompt"
agent_root: "."
`
	writeAgentYAML(t, tmpDir, "Architect", modifiedYAML)

	def2, err2 := loader.Load("Architect")
	require.NoError(t, err2)
	assert.Equal(t, "low", def2.Effort)
	assert.Equal(t, "Modified prompt", def2.SystemPrompt)
}

// Concurrent Behaviour tests are in test/race/agent_definition_loader_race_test.go

// Dependency Interaction

func TestAgentDefinitionLoader_Load_ReadsFromCorrectPath(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", createValidAgentYAML("Architect", "."))

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	require.NoError(t, err)
	assert.NotNil(t, def)
	// Verify that the file at the expected path computed by GetAgentPath exists and was read
	expectedPath := GetAgentPath(tmpDir, "Architect")
	assert.FileExists(t, expectedPath)
}

// Boundary Values — ProjectRoot

func TestAgentDefinitionLoader_New_RelativeProjectRoot(t *testing.T) {
	loader := NewAgentDefinitionLoader("./project")

	assert.NotNil(t, loader)
}

func TestAgentDefinitionLoader_Load_ProjectRootWithoutSpectraDir(t *testing.T) {
	tmpDir := t.TempDir()

	loader := NewAgentDefinitionLoader(tmpDir)
	def, err := loader.Load("Architect")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition not found: Architect")
}
