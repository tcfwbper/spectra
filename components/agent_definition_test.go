package components_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/components"
)

// TestAgentDefinition_ValidAgentAllFields creates AgentDefinition with all fields provided
func TestAgentDefinition_ValidAgentAllFields(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; spec/ directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)
	specDir := filepath.Join(tmpDir, "spec")
	err = os.MkdirAll(specDir, 0755)
	require.NoError(t, err)

	// Input: Role="QaReviewer", Model="sonnet", Effort="high", SystemPrompt="You are a QA reviewer", AgentRoot="spec", AllowedTools=["Read(*)"], DisallowedTools=["Bash(spectra *)"]
	agent, err := components.NewAgentDefinition(
		"QaReviewer",
		"sonnet",
		"high",
		"You are a QA reviewer",
		"spec",
		[]string{"Read(*)"},
		[]string{"Bash(spectra *)"},
	)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; all fields match input
	require.Equal(t, "QaReviewer", agent.GetRole())
	require.Equal(t, "sonnet", agent.GetModel())
	require.Equal(t, "high", agent.GetEffort())
	require.Equal(t, "You are a QA reviewer", agent.GetSystemPrompt())
	require.Equal(t, "spec", agent.GetAgentRoot())
	require.Equal(t, []string{"Read(*)"}, agent.GetAllowedTools())
	require.Equal(t, []string{"Bash(spectra *)"}, agent.GetDisallowedTools())

	// Expected: YAML file created at <test-dir>/.spectra/agents/QaReviewer.yaml
	yamlPath := filepath.Join(agentsDir, "QaReviewer.yaml")
	err = agent.SaveToFile(yamlPath)
	require.NoError(t, err)
	require.FileExists(t, yamlPath)
}

// TestAgentDefinition_EmptyToolLists creates AgentDefinition with empty AllowedTools and DisallowedTools
func TestAgentDefinition_EmptyToolLists(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="SimpleAgent", Model="sonnet", Effort="medium", SystemPrompt="Simple agent", AgentRoot=".", AllowedTools=[], DisallowedTools=[]
	agent, err := components.NewAgentDefinition(
		"SimpleAgent",
		"sonnet",
		"medium",
		"Simple agent",
		".",
		[]string{},
		[]string{},
	)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; AllowedTools=[], DisallowedTools=[]
	require.Equal(t, []string{}, agent.GetAllowedTools())
	require.Equal(t, []string{}, agent.GetDisallowedTools())
}

// TestAgentDefinition_AgentRootCurrentDir creates AgentDefinition with AgentRoot set to current directory
func TestAgentDefinition_AgentRootCurrentDir(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="RootAgent", Model="opus", Effort="low", SystemPrompt="Root agent", AgentRoot=".", AllowedTools=[], DisallowedTools=[]
	agent, err := components.NewAgentDefinition(
		"RootAgent",
		"opus",
		"low",
		"Root agent",
		".",
		[]string{},
		[]string{},
	)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; AgentRoot="."
	require.Equal(t, ".", agent.GetAgentRoot())
}

// TestAgentDefinition_MultiLineSystemPrompt creates AgentDefinition with multi-line SystemPrompt
func TestAgentDefinition_MultiLineSystemPrompt(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="MultiLineAgent", Model="sonnet", Effort="medium", SystemPrompt="Line 1\nLine 2\nLine 3", AgentRoot=".", AllowedTools=[], DisallowedTools=[]
	agent, err := components.NewAgentDefinition(
		"MultiLineAgent",
		"sonnet",
		"medium",
		"Line 1\nLine 2\nLine 3",
		".",
		[]string{},
		[]string{},
	)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; SystemPrompt preserves newlines
	require.Equal(t, "Line 1\nLine 2\nLine 3", agent.GetSystemPrompt())
}

// TestAgentDefinition_LoadValidYAML loads AgentDefinition from valid YAML file
func TestAgentDefinition_LoadValidYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with all required fields; spec/ directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)
	specDir := filepath.Join(tmpDir, "spec")
	err = os.MkdirAll(specDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Architect"
model: "opus"
effort: "high"
system_prompt: "You are an architect"
agent_root: "spec"
allowed_tools:
  - "Read(*)"
  - "Write(*)"
disallowed_tools:
  - "Bash(*)"
`
	yamlPath := filepath.Join(agentsDir, "Architect.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load agent with Role="Architect"
	agent, err := components.LoadAgentDefinition(yamlPath)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; all fields match YAML content
	require.Equal(t, "Architect", agent.GetRole())
	require.Equal(t, "opus", agent.GetModel())
	require.Equal(t, "high", agent.GetEffort())
	require.Equal(t, "You are an architect", agent.GetSystemPrompt())
	require.Equal(t, "spec", agent.GetAgentRoot())
	require.Equal(t, []string{"Read(*)", "Write(*)"}, agent.GetAllowedTools())
	require.Equal(t, []string{"Bash(*)"}, agent.GetDisallowedTools())
}

// TestAgentDefinition_LoadWithEmptyToolLists loads AgentDefinition with empty tool lists from YAML
func TestAgentDefinition_LoadWithEmptyToolLists(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with allowed_tools: [], disallowed_tools: []; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Agent"
model: "sonnet"
effort: "medium"
system_prompt: "Test agent"
agent_root: "."
allowed_tools: []
disallowed_tools: []
`
	yamlPath := filepath.Join(agentsDir, "Agent.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load agent with Role="Agent"
	agent, err := components.LoadAgentDefinition(yamlPath)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; AllowedTools=[], DisallowedTools=[]
	require.Equal(t, []string{}, agent.GetAllowedTools())
	require.Equal(t, []string{}, agent.GetDisallowedTools())
}

// TestAgentDefinition_LoadWithMultiLinePrompt loads AgentDefinition with multi-line SystemPrompt from YAML
func TestAgentDefinition_LoadWithMultiLinePrompt(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with multi-line system_prompt using YAML |- syntax; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Prompter"
model: "sonnet"
effort: "medium"
system_prompt: |-
  Line 1
  Line 2
  Line 3
agent_root: "."
allowed_tools: []
disallowed_tools: []
`
	yamlPath := filepath.Join(agentsDir, "Prompter.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load agent with Role="Prompter"
	agent, err := components.LoadAgentDefinition(yamlPath)
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; SystemPrompt preserves newlines and formatting
	require.Equal(t, "Line 1\nLine 2\nLine 3", agent.GetSystemPrompt())
}

// TestAgentDefinition_EmptyRole rejects AgentDefinition with empty Role
func TestAgentDefinition_EmptyRole(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("", "sonnet", "medium", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /role.*non-empty/i
	assertErrorMatches(t, err, `(?i)role.*non-empty`)
}

// TestAgentDefinition_RoleWithSpaces rejects AgentDefinition with Role containing spaces
func TestAgentDefinition_RoleWithSpaces(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="Qa Reviewer", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("Qa Reviewer", "sonnet", "medium", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /role.*PascalCase.*spaces.*special.*characters/i
	assertErrorMatches(t, err, `(?i)role.*PascalCase.*spaces.*special.*characters`)
}

// TestAgentDefinition_RoleWithUnderscores rejects AgentDefinition with Role containing underscores
func TestAgentDefinition_RoleWithUnderscores(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="Qa_Reviewer", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("Qa_Reviewer", "sonnet", "medium", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /role.*PascalCase.*spaces.*special.*characters/i
	assertErrorMatches(t, err, `(?i)role.*PascalCase.*spaces.*special.*characters`)
}

// TestAgentDefinition_RoleWithHyphens rejects AgentDefinition with Role containing hyphens
func TestAgentDefinition_RoleWithHyphens(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="Qa-Reviewer", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("Qa-Reviewer", "sonnet", "medium", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /role.*PascalCase.*spaces.*special.*characters/i
	assertErrorMatches(t, err, `(?i)role.*PascalCase.*spaces.*special.*characters`)
}

// TestAgentDefinition_RoleNotPascalCase rejects AgentDefinition with Role not in PascalCase
func TestAgentDefinition_RoleNotPascalCase(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="qaReviewer", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("qaReviewer", "sonnet", "medium", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /role.*PascalCase/i
	assertErrorMatches(t, err, `(?i)role.*PascalCase`)
}

// TestAgentDefinition_EmptyModel rejects AgentDefinition with empty Model
func TestAgentDefinition_EmptyModel(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="", Effort="medium", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("TestAgent", "", "medium", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /model.*non-empty/i
	assertErrorMatches(t, err, `(?i)model.*non-empty`)
}

// TestAgentDefinition_EmptyEffort rejects AgentDefinition with empty Effort
func TestAgentDefinition_EmptyEffort(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="", SystemPrompt="test", AgentRoot="."
	_, err = components.NewAgentDefinition("TestAgent", "sonnet", "", "test", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /effort.*non-empty/i
	assertErrorMatches(t, err, `(?i)effort.*non-empty`)
}

// TestAgentDefinition_EmptySystemPrompt rejects AgentDefinition with empty SystemPrompt
func TestAgentDefinition_EmptySystemPrompt(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="medium", SystemPrompt="", AgentRoot="."
	_, err = components.NewAgentDefinition("TestAgent", "sonnet", "medium", "", ".", []string{}, []string{})

	// Expected: Returns error; error message matches /system_prompt.*non-empty/i
	assertErrorMatches(t, err, `(?i)system_prompt.*non-empty`)
}

// TestAgentDefinition_EmptyAgentRoot rejects AgentDefinition with empty AgentRoot
func TestAgentDefinition_EmptyAgentRoot(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot=""
	_, err = components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", "", []string{}, []string{})

	// Expected: Returns error; error message matches /agent_root.*non-empty/i
	assertErrorMatches(t, err, `(?i)agent_root.*non-empty`)
}

// TestAgentDefinition_AbsolutePathAgentRoot rejects AgentDefinition with absolute path AgentRoot (Unix)
func TestAgentDefinition_AbsolutePathAgentRoot(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="/usr/local"
	_, err = components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", "/usr/local", []string{}, []string{})

	// Expected: Returns error; error message matches /agent_root.*relative.*path/i
	assertErrorMatches(t, err, `(?i)agent_root.*relative.*path`)
}

// TestAgentDefinition_AbsolutePathAgentRootWindows rejects AgentDefinition with absolute path AgentRoot (Windows drive letter)
func TestAgentDefinition_AbsolutePathAgentRootWindows(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot="C:\\spectra"
	_, err = components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", "C:\\spectra", []string{}, []string{})

	// Expected: Returns error; error message matches /agent_root.*relative.*path/i
	assertErrorMatches(t, err, `(?i)agent_root.*relative.*path`)
}

// TestAgentDefinition_NonExistentAgentRoot rejects AgentDefinition when AgentRoot directory does not exist
func TestAgentDefinition_NonExistentAgentRoot(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; nonexistent/ directory does NOT exist
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "TestAgent"
model: "sonnet"
effort: "medium"
system_prompt: "test"
agent_root: "nonexistent"
allowed_tools: []
disallowed_tools: []
`
	yamlPath := filepath.Join(agentsDir, "TestAgent.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load agent with AgentRoot="nonexistent"
	_, err = components.LoadAgentDefinition(yamlPath)

	// Expected: Returns error during agent definition load; error message matches /agent_root.*directory.*not found/i
	assertErrorMatches(t, err, `(?i)agent_root.*directory.*not found`)
}

// TestAgentDefinition_DuplicateRole rejects loading multiple AgentDefinitions with same Role
func TestAgentDefinition_DuplicateRole(t *testing.T) {
	// Setup: Temporary test directory created; YAML file at Architect.yaml; second YAML with same role
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)
	specDir := filepath.Join(tmpDir, "spec")
	err = os.MkdirAll(specDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Architect"
model: "opus"
effort: "high"
system_prompt: "You are an architect"
agent_root: "spec"
allowed_tools: []
disallowed_tools: []
`
	yamlPath1 := filepath.Join(agentsDir, "Architect.yaml")
	err = os.WriteFile(yamlPath1, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load first agent
	registry := components.NewAgentRegistry()
	agent1, err := components.LoadAgentDefinition(yamlPath1)
	require.NoError(t, err)
	err = registry.Register(agent1)
	require.NoError(t, err)

	// Create second agent with same role
	yamlPath2 := filepath.Join(agentsDir, "Architect2.yaml")
	err = os.WriteFile(yamlPath2, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load both agents with Role="Architect"
	agent2, err := components.LoadAgentDefinition(yamlPath2)
	require.NoError(t, err)
	err = registry.Register(agent2)

	// Expected: Second agent load returns error; error message matches /agent.*Architect.*already exists/i
	assertErrorMatches(t, err, `(?i)agent.*Architect.*already exists`)
}

// TestAgentDefinition_FileDoesNotExist returns error when agent YAML file does not exist
func TestAgentDefinition_FileDoesNotExist(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created but empty
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Load agent with Role="NonExistent"
	yamlPath := filepath.Join(agentsDir, "NonExistent.yaml")
	_, err = components.LoadAgentDefinition(yamlPath)

	// Expected: Returns error; error message matches /agent.*not found/i
	assertErrorMatches(t, err, `(?i)agent.*not found`)
}

// TestAgentDefinition_MalformedYAML rejects AgentDefinition with malformed YAML syntax
func TestAgentDefinition_MalformedYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with invalid YAML syntax (unclosed quote)
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Broken
model: "sonnet"
`
	yamlPath := filepath.Join(agentsDir, "Broken.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load agent with Role="Broken"
	_, err = components.LoadAgentDefinition(yamlPath)

	// Expected: Returns parse error; error message indicates YAML syntax issue
	require.Error(t, err)
}

// TestAgentDefinition_MissingRequiredField rejects AgentDefinition with missing required field (model)
func TestAgentDefinition_MissingRequiredField(t *testing.T) {
	// Setup: Temporary test directory created; YAML file missing model field
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Incomplete"
effort: "medium"
system_prompt: "test"
agent_root: "."
allowed_tools: []
disallowed_tools: []
`
	yamlPath := filepath.Join(agentsDir, "Incomplete.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load agent with Role="Incomplete"
	_, err = components.LoadAgentDefinition(yamlPath)

	// Expected: Returns error; error message matches /model.*required/i
	assertErrorMatches(t, err, `(?i)model.*required`)
}

// TestAgentDefinition_InvalidModelPassthrough AgentDefinition with invalid Model is loaded without validation; validation deferred to Claude CLI
func TestAgentDefinition_InvalidModelPassthrough(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="invalid-model", Effort="medium", SystemPrompt="test", AgentRoot="."
	agent, err := components.NewAgentDefinition("TestAgent", "invalid-model", "medium", "test", ".", []string{}, []string{})
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; Model="invalid-model" stored without validation
	require.Equal(t, "invalid-model", agent.GetModel())
}

// TestAgentDefinition_InvalidEffortPassthrough AgentDefinition with invalid Effort is loaded without validation; validation deferred to Claude CLI
func TestAgentDefinition_InvalidEffortPassthrough(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="super-high", SystemPrompt="test", AgentRoot="."
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "super-high", "test", ".", []string{}, []string{})
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; Effort="super-high" stored without validation
	require.Equal(t, "super-high", agent.GetEffort())
}

// TestAgentDefinition_InvalidToolsPassthrough AgentDefinition with invalid tool identifiers is loaded without validation; validation deferred to Claude CLI
func TestAgentDefinition_InvalidToolsPassthrough(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot=".", AllowedTools=["InvalidTool"]
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", ".", []string{"InvalidTool"}, []string{})
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; AllowedTools=["InvalidTool"] stored without validation
	require.Equal(t, []string{"InvalidTool"}, agent.GetAllowedTools())
}

// TestAgentDefinition_ToolConflictPassthrough AgentDefinition with same tool in both AllowedTools and DisallowedTools is loaded without validation; conflict resolution deferred to Claude CLI
func TestAgentDefinition_ToolConflictPassthrough(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/agents/ directory created; . directory exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: Role="TestAgent", Model="sonnet", Effort="medium", SystemPrompt="test", AgentRoot=".", AllowedTools=["Read(*)"], DisallowedTools=["Read(*)"]
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", ".", []string{"Read(*)"}, []string{"Read(*)"})
	require.NoError(t, err)

	// Expected: Returns valid AgentDefinition; both tool lists contain "Read(*)"; no validation error
	require.Equal(t, []string{"Read(*)"}, agent.GetAllowedTools())
	require.Equal(t, []string{"Read(*)"}, agent.GetDisallowedTools())
}

// TestAgentDefinition_ToYAML AgentDefinition serializes to YAML correctly
func TestAgentDefinition_ToYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: AgentDefinition with all fields populated
	agent, err := components.NewAgentDefinition(
		"TestAgent",
		"sonnet",
		"high",
		"Test prompt",
		"spec",
		[]string{"Read(*)", "Write(*)"},
		[]string{"Bash(*)"},
	)
	require.NoError(t, err)

	yamlPath := filepath.Join(agentsDir, "TestAgent.yaml")
	err = agent.SaveToFile(yamlPath)
	require.NoError(t, err)

	// Expected: YAML contains role, model, effort, system_prompt, agent_root, allowed_tools, disallowed_tools with correct values
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	yamlStr := string(content)
	require.Contains(t, yamlStr, "role: TestAgent")
	require.Contains(t, yamlStr, "model: sonnet")
	require.Contains(t, yamlStr, "effort: high")
	require.Contains(t, yamlStr, "system_prompt:")
	require.Contains(t, yamlStr, "agent_root: spec")
	require.Contains(t, yamlStr, "allowed_tools:")
	require.Contains(t, yamlStr, "disallowed_tools:")
}

// TestAgentDefinition_ToYAMLEmptyToolLists AgentDefinition with empty tool lists serializes to YAML correctly
func TestAgentDefinition_ToYAMLEmptyToolLists(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: AgentDefinition with AllowedTools=[], DisallowedTools=[]
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", ".", []string{}, []string{})
	require.NoError(t, err)

	yamlPath := filepath.Join(agentsDir, "TestAgent.yaml")
	err = agent.SaveToFile(yamlPath)
	require.NoError(t, err)

	// Expected: YAML contains allowed_tools: [], disallowed_tools: []
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	yamlStr := string(content)
	require.Contains(t, yamlStr, "allowed_tools: []")
	require.Contains(t, yamlStr, "disallowed_tools: []")
}

// TestAgentDefinition_ToYAMLMultiLinePrompt AgentDefinition with multi-line SystemPrompt serializes to YAML correctly
func TestAgentDefinition_ToYAMLMultiLinePrompt(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Input: AgentDefinition with SystemPrompt="Line 1\nLine 2\nLine 3"
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "medium", "Line 1\nLine 2\nLine 3", ".", []string{}, []string{})
	require.NoError(t, err)

	yamlPath := filepath.Join(agentsDir, "TestAgent.yaml")
	err = agent.SaveToFile(yamlPath)
	require.NoError(t, err)

	// Expected: YAML uses YAML multi-line syntax (|-) and preserves newlines
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	yamlStr := string(content)
	require.Contains(t, yamlStr, "system_prompt: |-")
}

// TestAgentDefinition_FieldsImmutable AgentDefinition fields cannot be modified after creation
func TestAgentDefinition_FieldsImmutable(t *testing.T) {
	// Setup: AgentDefinition instance created
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", ".", []string{"Read(*)"}, []string{"Bash(*)"})
	require.NoError(t, err)

	// Expected: Field modification attempt fails or has no effect; original values remain
	// In Go, we enforce immutability through unexported fields and getter methods only
	require.Equal(t, "TestAgent", agent.GetRole())
	require.Equal(t, "sonnet", agent.GetModel())
	require.Equal(t, "medium", agent.GetEffort())
	require.Equal(t, "test", agent.GetSystemPrompt())
	require.Equal(t, ".", agent.GetAgentRoot())
	require.Equal(t, []string{"Read(*)"}, agent.GetAllowedTools())
	require.Equal(t, []string{"Bash(*)"}, agent.GetDisallowedTools())
}

// TestAgentDefinition_ImplementsInterface AgentDefinition type implements expected interface
func TestAgentDefinition_ImplementsInterface(t *testing.T) {
	// Setup: AgentDefinition instance created
	agent, err := components.NewAgentDefinition("TestAgent", "sonnet", "medium", "test", ".", []string{}, []string{})
	require.NoError(t, err)

	// Expected: AgentDefinition satisfies AgentDefinition interface contract
	// Verify all getter methods are available
	_ = agent.GetRole()
	_ = agent.GetModel()
	_ = agent.GetEffort()
	_ = agent.GetSystemPrompt()
	_ = agent.GetAgentRoot()
	_ = agent.GetAllowedTools()
	_ = agent.GetDisallowedTools()
}

// TestAgentDefinition_ReferencedByNode verifies AgentDefinition successfully referenced by workflow Node
func TestAgentDefinition_ReferencedByNode(t *testing.T) {
	// Setup: Temporary test directory created; AgentDefinition file at <test-dir>/.spectra/agents/Architect.yaml;
	//        workflow definition with agent Node referencing AgentRole="Architect"
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)
	specDir := filepath.Join(tmpDir, "spec")
	err = os.MkdirAll(specDir, 0755)
	require.NoError(t, err)

	yamlContent := `role: "Architect"
model: "opus"
effort: "high"
system_prompt: "You are an architect"
agent_root: "spec"
allowed_tools:
  - "Read(*)"
disallowed_tools: []
`
	yamlPath := filepath.Join(agentsDir, "Architect.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load agent and verify it exists
	agent, err := components.LoadAgentDefinition(yamlPath)
	require.NoError(t, err)
	require.Equal(t, "Architect", agent.GetRole())

	// Input: Load workflow and validate nodes; Node with AgentRole="Architect" references valid AgentDefinition
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "ArchitectNode", "agent", "Architect", "AI architect drafts spec"),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "ArchitectNode"),
		createTransition(t, "ArchitectNode", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "ArchitectNode", "Done", "Start"),
	}

	// Expected: Node with AgentRole="Architect" references valid AgentDefinition; workflow validation succeeds
	workflow, err := components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	archNode := workflow.GetNodes()[1]
	require.Equal(t, "Architect", archNode.GetAgentRole())
	require.Equal(t, agent.GetRole(), archNode.GetAgentRole())
}
