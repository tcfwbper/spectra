package race_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

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

	loader := storage.NewAgentDefinitionLoader(tmpDir)

	done := make(chan *storage.AgentDefinition, 10)
	for range 10 {
		go func() {
			def, err := loader.Load("Architect")
			require.NoError(t, err)
			done <- def
		}()
	}

	results := make([]*storage.AgentDefinition, 10)
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

	loader := storage.NewAgentDefinitionLoader(tmpDir)

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
