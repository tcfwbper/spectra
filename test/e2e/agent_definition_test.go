package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAgentDefinition_BuiltinCopiedDuringInit verifies built-in agents are copied to .spectra/agents/ during spectra init
func TestAgentDefinition_BuiltinCopiedDuringInit(t *testing.T) {
	// Category: e2e
	// Setup: Temporary test directory created; no .spectra/agents/ directory exists; all file operations occur within test fixtures
	tmpDir := t.TempDir()
	_ = tmpDir

	// Input: Execute `spectra init`
	// Expected: Built-in agent files copied to <test-dir>/.spectra/agents/; files readable and valid
	t.Skip("Requires CLI infrastructure")

	// Test implementation placeholder:
	// 1. Execute spectra init in tmpDir
	// 2. Verify .spectra/agents/ directory created
	// 3. Verify built-in agent YAML files exist (e.g., QaReviewer.yaml, Architect.yaml)
	// 4. Read and validate YAML content of copied files
	// 5. Ensure all required fields are present and valid
}

// TestAgentDefinition_ExistingAgentNotOverwritten verifies existing agent file is not overwritten during spectra init
func TestAgentDefinition_ExistingAgentNotOverwritten(t *testing.T) {
	// Category: e2e
	// Setup: Temporary test directory created; .spectra/agents/QaReviewer.yaml exists with custom content; all file operations occur within test fixtures
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	customContent := `role: "QaReviewer"
model: "custom-model"
effort: "ultra-high"
system_prompt: "Custom QA reviewer prompt"
agent_root: "custom"
allowed_tools: ["CustomTool"]
disallowed_tools: []
`
	customAgentPath := filepath.Join(agentsDir, "QaReviewer.yaml")
	err = os.WriteFile(customAgentPath, []byte(customContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write custom agent: %v", err)
	}

	// Input: Execute `spectra init`
	// Expected: QaReviewer.yaml content unchanged; other built-in agents copied; no error returned
	t.Skip("Requires CLI infrastructure")

	// Test implementation placeholder:
	// 1. Execute spectra init in tmpDir
	// 2. Read QaReviewer.yaml content
	// 3. Verify content matches customContent exactly
	// 4. Verify other built-in agents were copied
	// 5. Ensure no error was returned
}

// TestAgentDefinition_ListAgents verifies CLI lists all available agents
func TestAgentDefinition_ListAgents(t *testing.T) {
	// Category: e2e
	// Setup: Temporary test directory created; multiple agent definition files in <test-dir>/.spectra/agents/; all file operations occur within test fixtures
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create test agent files
	agents := map[string]string{
		"Architect": `role: "Architect"
model: "opus"
effort: "high"
system_prompt: "You are an architect"
agent_root: "spec"
allowed_tools: []
disallowed_tools: []
`,
		"QaReviewer": `role: "QaReviewer"
model: "sonnet"
effort: "high"
system_prompt: "You are a QA reviewer"
agent_root: "."
allowed_tools: []
disallowed_tools: []
`,
		"Implementer": `role: "Implementer"
model: "sonnet"
effort: "medium"
system_prompt: "You are an implementer"
agent_root: "src"
allowed_tools: []
disallowed_tools: []
`,
	}

	for role, content := range agents {
		agentPath := filepath.Join(agentsDir, role+".yaml")
		err := os.WriteFile(agentPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write agent %s: %v", role, err)
		}
	}

	// Input: Execute `spectra agent list`
	// Expected: Command succeeds; output lists all agents with role names and agent_root paths
	t.Skip("Requires CLI infrastructure")

	// Test implementation placeholder:
	// 1. Execute spectra agent list in tmpDir
	// 2. Parse output
	// 3. Verify all 3 agents are listed
	// 4. Verify each agent shows role name and agent_root
	// 5. Verify exit code is 0
}

// TestAgentDefinition_ShowAgentDetails verifies CLI shows details for specific agent
func TestAgentDefinition_ShowAgentDetails(t *testing.T) {
	// Category: e2e
	// Setup: Temporary test directory created; agent definition file at <test-dir>/.spectra/agents/Architect.yaml; all file operations occur within test fixtures
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	agentContent := `role: "Architect"
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
	agentPath := filepath.Join(agentsDir, "Architect.yaml")
	err = os.WriteFile(agentPath, []byte(agentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write agent: %v", err)
	}

	// Input: Execute `spectra agent show --role Architect`
	// Expected: Command succeeds; output displays all fields: role, model, effort, system_prompt, agent_root, allowed_tools, disallowed_tools
	t.Skip("Requires CLI infrastructure")

	// Test implementation placeholder:
	// 1. Execute spectra agent show --role Architect in tmpDir
	// 2. Parse output
	// 3. Verify all fields are displayed:
	//    - role: "Architect"
	//    - model: "opus"
	//    - effort: "high"
	//    - system_prompt: "You are an architect"
	//    - agent_root: "spec"
	//    - allowed_tools: ["Read(*)", "Write(*)"]
	//    - disallowed_tools: ["Bash(*)"]
	// 4. Verify exit code is 0
}

// TestAgentDefinition_ValidateAgent verifies CLI validates agent definition file
func TestAgentDefinition_ValidateAgent(t *testing.T) {
	// Category: e2e
	// Setup: Temporary test directory created; valid agent definition file at <test-dir>/.spectra/agents/TestAgent.yaml; <test-dir>/spec/ directory exists; all file operations occur within test fixtures
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}
	specDir := filepath.Join(tmpDir, "spec")
	err = os.MkdirAll(specDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create spec directory: %v", err)
	}

	agentContent := `role: "TestAgent"
model: "sonnet"
effort: "medium"
system_prompt: "You are a test agent"
agent_root: "spec"
allowed_tools: []
disallowed_tools: []
`
	agentPath := filepath.Join(agentsDir, "TestAgent.yaml")
	err = os.WriteFile(agentPath, []byte(agentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write agent: %v", err)
	}

	// Input: Execute `spectra agent validate --role TestAgent`
	// Expected: Command succeeds; no errors reported
	t.Skip("Requires CLI infrastructure")

	// Test implementation placeholder:
	// 1. Execute spectra agent validate --role TestAgent in tmpDir
	// 2. Verify exit code is 0
	// 3. Verify no validation errors in output
	// 4. Verify success message is displayed
}

// TestAgentDefinition_ValidateAgentInvalidRoot verifies CLI validation fails for agent with non-existent AgentRoot
func TestAgentDefinition_ValidateAgentInvalidRoot(t *testing.T) {
	// Category: e2e
	// Setup: Temporary test directory created; agent definition file at <test-dir>/.spectra/agents/BadAgent.yaml with agent_root: "nonexistent"; directory does NOT exist; all file operations occur within test fixtures
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	agentContent := `role: "BadAgent"
model: "sonnet"
effort: "medium"
system_prompt: "You are a bad agent"
agent_root: "nonexistent"
allowed_tools: []
disallowed_tools: []
`
	agentPath := filepath.Join(agentsDir, "BadAgent.yaml")
	err = os.WriteFile(agentPath, []byte(agentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write agent: %v", err)
	}

	// Input: Execute `spectra agent validate --role BadAgent`
	// Expected: Command fails; error message matches /agent_root.*directory.*not found/i
	t.Skip("Requires CLI infrastructure")

	// Test implementation placeholder:
	// 1. Execute spectra agent validate --role BadAgent in tmpDir
	// 2. Verify exit code is non-zero
	// 3. Verify error message contains "agent_root" and "directory" and "not found"
	// 4. Verify validation failure is reported
}
