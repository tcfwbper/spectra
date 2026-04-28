package race_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

// AgentDefinition represents an agent configuration loaded from YAML
type AgentDefinition struct {
	Role            string   `yaml:"role"`
	Model           string   `yaml:"model"`
	Effort          string   `yaml:"effort"`
	SystemPrompt    string   `yaml:"system_prompt"`
	AgentRoot       string   `yaml:"agent_root"`
	AllowedTools    []string `yaml:"allowed_tools"`
	DisallowedTools []string `yaml:"disallowed_tools"`
}

// AgentDefinitionLoader loads agent definitions from .spectra/agents/
// This is a stub awaiting implementation.
type AgentDefinitionLoader struct {
	projectRoot string
}

// NewAgentDefinitionLoader creates a new AgentDefinitionLoader
func NewAgentDefinitionLoader(projectRoot string) *AgentDefinitionLoader {
	return &AgentDefinitionLoader{projectRoot: projectRoot}
}

// Load loads an agent definition from disk
// Stub implementation - will be provided by the implementation phase
func (l *AgentDefinitionLoader) Load(agentRole string) (*AgentDefinition, error) {
	return nil, nil
}

// Test helper functions

func setupAgentTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, storage.AgentsDir)
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	return tmpDir
}

func writeAgentYAML(t *testing.T, projectRoot, agentRole, content string) {
	t.Helper()
	agentPath := storage.GetAgentPath(projectRoot, agentRole)
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

// TestAgentDefinitionLoader_Load_ConcurrentSameRole tests multiple goroutines loading the same agent role concurrently
func TestAgentDefinitionLoader_Load_ConcurrentSameRole(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	writeAgentYAML(t, tmpDir, "Architect", createValidAgentYAML("Architect", "."))

	loader := NewAgentDefinitionLoader(tmpDir)

	done := make(chan *AgentDefinition, 10)
	for range 10 {
		go func() {
			def, err := loader.Load("Architect")
			require.NoError(t, err)
			done <- def
		}()
	}

	results := make([]*AgentDefinition, 10)
	for i := range 10 {
		results[i] = <-done
	}

	for i := 1; i < 10; i++ {
		assert.Equal(t, results[0].Role, results[i].Role)
	}
}

// TestAgentDefinitionLoader_Load_ConcurrentDifferentRoles tests multiple goroutines loading different agent roles concurrently
func TestAgentDefinitionLoader_Load_ConcurrentDifferentRoles(t *testing.T) {
	tmpDir := setupAgentTestDir(t)
	roles := []string{"Agent1", "Agent2", "Agent3", "Agent4", "Agent5"}
	for _, role := range roles {
		writeAgentYAML(t, tmpDir, role, createValidAgentYAML(role, "."))
	}

	loader := NewAgentDefinitionLoader(tmpDir)

	done := make(chan string, 10)
	for i := range 10 {
		role := roles[i%len(roles)]
		go func(r string) {
			def, err := loader.Load(r)
			require.NoError(t, err)
			done <- def.Role
		}(role)
	}

	for range 10 {
		result := <-done
		assert.Contains(t, roles, result)
	}
}
