package storage

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- YAML Fixture Builders for AgentDefinitionLoader ---

// validAgentYAML returns a well-formed agent YAML with all fields using camelCase keys.
func validAgentYAML(model, effort, systemPrompt, agentRoot string) string {
	return "model: \"" + model + "\"\n" +
		"effort: \"" + effort + "\"\n" +
		"systemPrompt: \"" + systemPrompt + "\"\n" +
		"agentRoot: \"" + agentRoot + "\"\n"
}

// validAgentYAMLWithTools returns agent YAML with allowedTools and disallowedTools arrays.
func validAgentYAMLWithTools(model, effort, systemPrompt, agentRoot string, allowed, disallowed []string) string {
	yaml := validAgentYAML(model, effort, systemPrompt, agentRoot)
	if len(allowed) > 0 {
		yaml += "allowedTools:\n"
		for _, tool := range allowed {
			yaml += "  - \"" + tool + "\"\n"
		}
	}
	if len(disallowed) > 0 {
		yaml += "disallowedTools:\n"
		for _, tool := range disallowed {
			yaml += "  - \"" + tool + "\"\n"
		}
	}
	return yaml
}

// --- Filesystem Fixture Builders for AgentDefinitionLoader ---

// makeTempDirWithAgents creates a temp dir containing `.spectra/agents/` directory.
func makeTempDirWithAgents(t *testing.T) string {
	t.Helper()
	dir := makeTempDirWithSpectra(t)
	agentsDir := filepath.Join(dir, ".spectra", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("makeTempDirWithAgents: failed to create agents dir: %v", err)
	}
	return dir
}

// writeAgentYAML writes YAML content to .spectra/agents/<role>.yaml.
func writeAgentYAML(t *testing.T, projectRoot, role, content string) {
	t.Helper()
	filePath := GetAgentPath(projectRoot, role)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}

// makeAgentRootDir creates a directory at projectRoot/<relPath> to serve as agentRoot.
func makeAgentRootDir(t *testing.T, projectRoot, relPath string) {
	t.Helper()
	absPath := filepath.Join(projectRoot, relPath)
	err := os.MkdirAll(absPath, 0755)
	require.NoError(t, err)
}

// --- Happy Path Tests ---

func TestAgentDefinitionLoader_Load_ValidDefinition(t *testing.T) {
	projectRoot := makeTempDirWithAgents(t)
	agentRoot := "src"
	makeAgentRootDir(t, projectRoot, agentRoot)
	writeAgentYAML(t, projectRoot, "MyAgent", validAgentYAML("claude-sonnet-4-20250514", "high", "You are an agent.", agentRoot))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("MyAgent")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "MyAgent", def.Role())
	assert.Equal(t, "claude-sonnet-4-20250514", def.Model())
	assert.Equal(t, "high", def.Effort())
	assert.Equal(t, "You are an agent.", def.SystemPrompt())
	assert.Equal(t, agentRoot, def.AgentRoot())
}

func TestAgentDefinitionLoader_Load_RoleDerivedFromFilename(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	agentRoot := "arch"
	makeAgentRootDir(t, projectRoot, agentRoot)
	writeAgentYAML(t, projectRoot, "Architect", validAgentYAML("claude-sonnet-4-20250514", "high", "You are an architect.", agentRoot))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Architect")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "Architect", def.Role())
}

func TestAgentDefinitionLoader_Load_AllowedToolsAndDisallowedTools(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	agentRoot := "worker"
	makeAgentRootDir(t, projectRoot, agentRoot)

	allowed := []string{"Read", "Write"}
	disallowed := []string{"Bash"}
	yaml := validAgentYAMLWithTools("claude-sonnet-4-20250514", "high", "Worker prompt.", agentRoot, allowed, disallowed)
	writeAgentYAML(t, projectRoot, "Worker", yaml)

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Worker")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, allowed, def.AllowedTools())
	assert.Equal(t, disallowed, def.DisallowedTools())
}

func TestAgentDefinitionLoader_Load_MissingToolsFields(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	agentRoot := "simple"
	makeAgentRootDir(t, projectRoot, agentRoot)
	writeAgentYAML(t, projectRoot, "Simple", validAgentYAML("claude-sonnet-4-20250514", "high", "Simple prompt.", agentRoot))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Simple")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Empty(t, def.AllowedTools())
	assert.Empty(t, def.DisallowedTools())
}

func TestAgentDefinitionLoader_Load_AgentRootDot(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	// agentRoot "." resolves to projectRoot itself, which already exists.
	writeAgentYAML(t, projectRoot, "RootAgent", validAgentYAML("claude-sonnet-4-20250514", "high", "Root agent.", "."))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("RootAgent")

	require.NoError(t, err)
	require.NotNil(t, def)
}

func TestAgentDefinitionLoader_Load_AgentRootSymlinkToDir(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)

	// Create target directory.
	targetDir := filepath.Join(projectRoot, "real_dir")
	require.NoError(t, os.Mkdir(targetDir, 0755))

	// Create symlink pointing to target directory.
	symlinkPath := filepath.Join(projectRoot, "linked_dir")
	require.NoError(t, os.Symlink(targetDir, symlinkPath))

	writeAgentYAML(t, projectRoot, "Linked", validAgentYAML("claude-sonnet-4-20250514", "high", "Linked agent.", "linked_dir"))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Linked")

	require.NoError(t, err)
	require.NotNil(t, def)
}

// --- Error Propagation Tests ---

func TestAgentDefinitionLoader_Load_FileNotFound(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Missing")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition not found: Missing")
}

func TestAgentDefinitionLoader_Load_ReadPermissionDenied(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	writeAgentYAML(t, projectRoot, "Locked", validAgentYAML("claude-sonnet-4-20250514", "high", "Locked.", "src"))

	// Remove read permissions.
	filePath := GetAgentPath(projectRoot, "Locked")
	require.NoError(t, os.Chmod(filePath, 0000))
	t.Cleanup(func() { os.Chmod(filePath, 0644) })

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Locked")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read agent definition 'Locked':")
	assert.Contains(t, err.Error(), "permission")
}

func TestAgentDefinitionLoader_Load_YamlSyntaxError(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	writeAgentYAML(t, projectRoot, "Bad", "model: \"unclosed\n  bad: [indent")

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Bad")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent definition 'Bad':")
}

func TestAgentDefinitionLoader_Load_YamlUnknownField(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	yaml := validAgentYAML("claude-sonnet-4-20250514", "high", "Agent.", "src") + "customField: value\n"
	writeAgentYAML(t, projectRoot, "Extra", yaml)

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Extra")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent definition 'Extra':")
}

func TestAgentDefinitionLoader_Load_YamlSnakeCaseField(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	// Use snake_case system_prompt instead of camelCase systemPrompt.
	yaml := "model: \"claude-sonnet-4-20250514\"\n" +
		"effort: \"high\"\n" +
		"system_prompt: \"Snake case prompt.\"\n" +
		"agentRoot: \"src\"\n"
	writeAgentYAML(t, projectRoot, "Snake", yaml)

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Snake")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent definition 'Snake':")
}

func TestAgentDefinitionLoader_Load_ConstructorValidationFails(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	// Empty model triggers constructor error.
	yaml := "model: \"\"\n" +
		"effort: \"high\"\n" +
		"systemPrompt: \"No model.\"\n" +
		"agentRoot: \"src\"\n"
	writeAgentYAML(t, projectRoot, "NoModel", yaml)

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("NoModel")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'NoModel' validation failed:")
}

func TestAgentDefinitionLoader_Load_AgentRootNotExists(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	// Reference agentRoot that does not exist.
	writeAgentYAML(t, projectRoot, "Orphan", validAgentYAML("claude-sonnet-4-20250514", "high", "Orphan.", "missing_dir"))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Orphan")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Orphan' validation failed: agent_root directory not found:")
}

func TestAgentDefinitionLoader_Load_AgentRootIsFile(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	// Create a regular file where a directory is expected.
	filePath := filepath.Join(projectRoot, "somefile")
	require.NoError(t, os.WriteFile(filePath, []byte("not a dir"), 0644))

	writeAgentYAML(t, projectRoot, "FileRoot", validAgentYAML("claude-sonnet-4-20250514", "high", "FileRoot.", "somefile"))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("FileRoot")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'FileRoot' validation failed: agent_root is not a directory:")
}

func TestAgentDefinitionLoader_Load_AgentRootSymlinkBroken(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	// Create a dangling symlink.
	symlinkPath := filepath.Join(projectRoot, "broken_link")
	require.NoError(t, os.Symlink("/nonexistent/target", symlinkPath))

	writeAgentYAML(t, projectRoot, "Broken", validAgentYAML("claude-sonnet-4-20250514", "high", "Broken.", "broken_link"))

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Broken")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition 'Broken' validation failed: agent_root directory not found:")
}

// --- Null / Empty Input Tests ---

func TestAgentDefinitionLoader_Load_EmptyAgentRole(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent definition not found: ")
}

func TestAgentDefinitionLoader_Load_EmptyYamlFile(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	writeAgentYAML(t, projectRoot, "Empty", "")

	loader := NewAgentDefinitionLoader(projectRoot)
	def, err := loader.Load("Empty")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent definition 'Empty':")
}

// --- Boundary Values Tests ---

func TestAgentDefinitionLoader_Load_PathTraversal(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)

	loader := NewAgentDefinitionLoader(projectRoot)
	_, err := loader.Load("../malicious/agent")

	require.Error(t, err)
}

// --- Idempotency Tests ---

func TestAgentDefinitionLoader_Load_NoCaching(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	agentRoot := "src"
	makeAgentRootDir(t, projectRoot, agentRoot)

	// Write initial YAML with first model.
	writeAgentYAML(t, projectRoot, "Mutable", validAgentYAML("model-v1", "high", "Mutable agent.", agentRoot))

	loader := NewAgentDefinitionLoader(projectRoot)

	// First load returns initial model.
	def1, err := loader.Load("Mutable")
	require.NoError(t, err)
	assert.Equal(t, "model-v1", def1.Model())

	// Overwrite YAML with different model.
	writeAgentYAML(t, projectRoot, "Mutable", validAgentYAML("model-v2", "high", "Mutable agent.", agentRoot))

	// Second load reflects updated content.
	def2, err := loader.Load("Mutable")
	require.NoError(t, err)
	assert.Equal(t, "model-v2", def2.Model())
}

// --- Concurrent Behaviour Tests ---

func TestAgentDefinitionLoader_Load_ConcurrentAccess(t *testing.T) {

	projectRoot := makeTempDirWithAgents(t)
	agentRoot := "shared"
	makeAgentRootDir(t, projectRoot, agentRoot)
	writeAgentYAML(t, projectRoot, "Shared", validAgentYAML("claude-sonnet-4-20250514", "high", "Shared agent.", agentRoot))

	loader := NewAgentDefinitionLoader(projectRoot)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errs := make([]error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			def, err := loader.Load("Shared")
			errs[idx] = err
			if err == nil {
				assert.Equal(t, "Shared", def.Role())
			}
		}(i)
	}

	wg.Wait()
	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d returned error", i)
	}
}

// Suppress unused import warnings for packages used only in skipped test bodies.
var (
	_ = strings.Contains
)
