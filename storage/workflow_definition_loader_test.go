package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAgentDefinitionLoader is a mock implementation for testing.
// It records all Load calls and allows customizing the return behavior via loadFn.
type MockAgentDefinitionLoader struct {
	mu     sync.Mutex
	loadFn func(agentRole string) (*AgentDefinition, error)
	calls  []string
}

// NewMockAgentDefinitionLoader creates a new mock loader
func NewMockAgentDefinitionLoader() *MockAgentDefinitionLoader {
	return &MockAgentDefinitionLoader{
		calls: []string{},
	}
}

// Load calls the mock function and records the call
func (m *MockAgentDefinitionLoader) Load(agentRole string) (*AgentDefinition, error) {
	m.mu.Lock()
	m.calls = append(m.calls, agentRole)
	m.mu.Unlock()

	if m.loadFn != nil {
		return m.loadFn(agentRole)
	}
	return &AgentDefinition{Role: agentRole}, nil
}

// GetCalls returns the list of calls made
func (m *MockAgentDefinitionLoader) GetCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.calls))
	copy(result, m.calls)
	return result
}

// Test helper functions

func setupWorkflowTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, WorkflowsDir)
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))
	return tmpDir
}

func writeWorkflowYAML(t *testing.T, projectRoot, workflowName, content string) {
	t.Helper()
	workflowPath := GetWorkflowPath(projectRoot, workflowName)
	require.NoError(t, os.WriteFile(workflowPath, []byte(content), 0644))
}

func createMinimalValidWorkflowYAML(name string) string {
	return `name: "` + name + `"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
    description: "Start node"
  - name: "End"
    type: "human"
    description: "End node"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
}

// Happy Path — Construction

func TestWorkflowDefinitionLoader_New(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)

	assert.NotNil(t, loader)
}

// Happy Path — Load

func TestWorkflowDefinitionLoader_Load_MinimalValidWorkflow(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	require.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, "Simple", def.Name)
	assert.Equal(t, "Start", def.EntryNode)
	assert.Len(t, def.Nodes, 2)
	assert.Len(t, def.Transitions, 1)
	assert.Len(t, def.ExitTransitions, 1)
}

func TestWorkflowDefinitionLoader_Load_MultiNodeWorkflow(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Complex"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "Agent1"
    type: "agent"
    agent_role: "Architect"
  - name: "Agent2"
    type: "agent"
    agent_role: "Reviewer"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Agent1"
  - from_node: "Agent1"
    event_type: "Done"
    to_node: "Agent2"
  - from_node: "Agent2"
    event_type: "Complete"
    to_node: "End"
exit_transitions:
  - from_node: "Agent2"
    event_type: "Complete"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Complex", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Complex")

	require.NoError(t, err)
	assert.Len(t, def.Nodes, 4)
	assert.Len(t, def.Transitions, 3)
}

func TestWorkflowDefinitionLoader_Load_NameWithDigits(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "V2Workflow", createMinimalValidWorkflowYAML("V2Workflow"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("V2Workflow")

	require.NoError(t, err)
	assert.Equal(t, "V2Workflow", def.Name)
}

func TestWorkflowDefinitionLoader_Load_NameWithConsecutiveUppercase(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "DefaultLOGICSPEC", createMinimalValidWorkflowYAML("DefaultLOGICSPEC"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("DefaultLOGICSPEC")

	require.NoError(t, err)
	assert.Equal(t, "DefaultLOGICSPEC", def.Name)
}

func TestWorkflowDefinitionLoader_Load_SingleUppercaseLetter(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "A", createMinimalValidWorkflowYAML("A"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("A")

	require.NoError(t, err)
	assert.Equal(t, "A", def.Name)
}

func TestWorkflowDefinitionLoader_Load_WithOptionalDescription(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Described"
description: "Test workflow"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Described", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Described")

	require.NoError(t, err)
	assert.Equal(t, "Test workflow", def.Description)
}

func TestWorkflowDefinitionLoader_Load_WithoutDescription(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	require.NoError(t, err)
	assert.Empty(t, def.Description)
}

func TestWorkflowDefinitionLoader_Load_UnreachableNode(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Unreachable"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "Isolated"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Unreachable", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Unreachable")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Unreachable' validation failed: node 'Isolated' is unreachable (no incoming transitions)")
}

func TestWorkflowDefinitionLoader_Load_ExitTargetWithOutgoingTransitions(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "ExitWithOut"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "Middle"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Middle"
  - from_node: "Middle"
    event_type: "Next"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Middle"
  - from_node: "Middle"
    event_type: "Next"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "ExitWithOut", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("ExitWithOut")

	require.NoError(t, err)
	assert.NotNil(t, def)
}

func TestWorkflowDefinitionLoader_Load_UnknownFieldsIgnored(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Extended"
custom_metadata: "extra"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Extended", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Extended")

	require.NoError(t, err)
	assert.NotNil(t, def)
}

// Happy Path — Agent Reference Validation

func TestWorkflowDefinitionLoader_Load_AgentNodeReferencesValidAgent(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "WithAgent"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "AgentNode"
    type: "agent"
    agent_role: "Architect"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentNode"
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "WithAgent", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("WithAgent")

	require.NoError(t, err)
	assert.NotNil(t, def)
	calls := mockLoader.GetCalls()
	assert.Contains(t, calls, "Architect")
}

func TestWorkflowDefinitionLoader_Load_MultipleAgentNodesValidated(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "MultiAgent"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "Agent1"
    type: "agent"
    agent_role: "Architect"
  - name: "Agent2"
    type: "agent"
    agent_role: "Reviewer"
  - name: "Agent3"
    type: "agent"
    agent_role: "Coder"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Agent1"
  - from_node: "Agent1"
    event_type: "Next1"
    to_node: "Agent2"
  - from_node: "Agent2"
    event_type: "Next2"
    to_node: "Agent3"
  - from_node: "Agent3"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "Agent3"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "MultiAgent", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("MultiAgent")

	require.NoError(t, err)
	assert.NotNil(t, def)
	calls := mockLoader.GetCalls()
	assert.Contains(t, calls, "Architect")
	assert.Contains(t, calls, "Reviewer")
	assert.Contains(t, calls, "Coder")
}

func TestWorkflowDefinitionLoader_Load_DuplicateAgentRoleValidatedOnce(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "DupeRole"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "Agent1"
    type: "agent"
    agent_role: "Coder"
  - name: "Agent2"
    type: "agent"
    agent_role: "Coder"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Agent1"
  - from_node: "Agent1"
    event_type: "Next"
    to_node: "Agent2"
  - from_node: "Agent2"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "Agent2"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "DupeRole", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("DupeRole")

	require.NoError(t, err)
	assert.NotNil(t, def)
	calls := mockLoader.GetCalls()
	coderCount := 0
	for _, call := range calls {
		if call == "Coder" {
			coderCount++
		}
	}
	assert.Equal(t, 1, coderCount, "Coder should be loaded exactly once")
}

// Validation Failures — File Not Found

func TestWorkflowDefinitionLoader_Load_FileNotFound(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition not found: Simple")
}

func TestWorkflowDefinitionLoader_Load_EmptyWorkflowName(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition not found:")
}

// Validation Failures — File Read Errors

func TestWorkflowDefinitionLoader_Load_PermissionDenied(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))
	workflowPath := GetWorkflowPath(tmpDir, "Simple")
	require.NoError(t, os.Chmod(workflowPath, 0000))
	defer os.Chmod(workflowPath, 0644)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read workflow definition 'Simple'")
	assert.Contains(t, err.Error(), "permission denied")
}

// Validation Failures — YAML Parsing

func TestWorkflowDefinitionLoader_Load_EmptyFile(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", "")

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow definition 'Simple'")
	assert.Contains(t, err.Error(), "EOF")
}

func TestWorkflowDefinitionLoader_Load_InvalidYAMLSyntax(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	invalidYAML := "name:\n  - invalid:\nbroken"
	writeWorkflowYAML(t, tmpDir, "Simple", invalidYAML)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow definition 'Simple'")
	assert.Contains(t, err.Error(), "yaml: line")
}

// Validation Failures — Missing Required Fields

func TestWorkflowDefinitionLoader_Load_MissingName(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Simple", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Simple' validation failed: missing required field 'name'")
}

func TestWorkflowDefinitionLoader_Load_MissingEntryNode(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Simple"
nodes:
  - name: "Start"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Simple", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Simple' validation failed: missing required field 'entry_node'")
}

func TestWorkflowDefinitionLoader_Load_MissingNodes(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Simple"
entry_node: "Start"
nodes: []
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Simple", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Simple' validation failed: missing required field 'nodes'")
}

func TestWorkflowDefinitionLoader_Load_MissingTransitions(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Simple"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
transitions: []
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Simple", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Simple' validation failed: missing required field 'transitions'")
}

func TestWorkflowDefinitionLoader_Load_MissingExitTransitions(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Simple"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions: []
`
	writeWorkflowYAML(t, tmpDir, "Simple", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Simple' validation failed: missing required field 'exit_transitions'")
}

// Validation Failures — Name Format

func TestWorkflowDefinitionLoader_Load_NameWithSpaces(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Default LogicSpec"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Default LogicSpec", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Default LogicSpec")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Default LogicSpec' validation failed: name must be PascalCase with no spaces or special characters")
}

func TestWorkflowDefinitionLoader_Load_NameWithUnderscore(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Default_LogicSpec"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Default_LogicSpec", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Default_LogicSpec")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Default_LogicSpec' validation failed: name must be PascalCase with no spaces or special characters")
}

func TestWorkflowDefinitionLoader_Load_NameWithHyphen(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Default-LogicSpec"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Default-LogicSpec", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Default-LogicSpec")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Default-LogicSpec' validation failed: name must be PascalCase with no spaces or special characters")
}

func TestWorkflowDefinitionLoader_Load_NameWithDot(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Default.LogicSpec"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Default.LogicSpec", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Default.LogicSpec")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Default.LogicSpec' validation failed: name must be PascalCase with no spaces or special characters")
}

func TestWorkflowDefinitionLoader_Load_NameStartsLowercase(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "defaultLogicSpec"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "defaultLogicSpec", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("defaultLogicSpec")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'defaultLogicSpec' validation failed: name must be PascalCase with no spaces or special characters")
}

// Validation Failures — EntryNode

func TestWorkflowDefinitionLoader_Load_EntryNodeNotFound(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "BadEntry"
entry_node: "NonExistent"
nodes:
  - name: "Start"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "BadEntry", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("BadEntry")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadEntry' validation failed: entry_node 'NonExistent' references non-existent node")
}

func TestWorkflowDefinitionLoader_Load_EntryNodeNotHumanType(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "AgentEntry"
entry_node: "AgentNode"
nodes:
  - name: "AgentNode"
    type: "agent"
    agent_role: "Architect"
  - name: "End"
    type: "human"
transitions:
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "AgentEntry", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("AgentEntry")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'AgentEntry' validation failed: entry_node 'AgentNode' must have type 'human', but has type 'agent'")
}

// Validation Failures — Node Integrity

func TestWorkflowDefinitionLoader_Load_DuplicateNodeName(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "DuplicateNode"
entry_node: "Review"
nodes:
  - name: "Review"
    type: "human"
  - name: "Review"
    type: "human"
transitions:
  - from_node: "Review"
    event_type: "Done"
    to_node: "Review"
exit_transitions:
  - from_node: "Review"
    event_type: "Done"
    to_node: "Review"
`
	writeWorkflowYAML(t, tmpDir, "DuplicateNode", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("DuplicateNode")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'DuplicateNode' validation failed: duplicate node name 'Review'")
}

// Validation Failures — Transition Integrity

func TestWorkflowDefinitionLoader_Load_TransitionFromNodeNotFound(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "BadFrom"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Ghost"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Ghost"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "BadFrom", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("BadFrom")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadFrom' validation failed: transition references non-existent node 'Ghost'")
}

func TestWorkflowDefinitionLoader_Load_TransitionToNodeNotFound(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "BadTo"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Ghost"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Ghost"
`
	writeWorkflowYAML(t, tmpDir, "BadTo", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("BadTo")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadTo' validation failed: transition references non-existent node 'Ghost'")
}

func TestWorkflowDefinitionLoader_Load_TransitionSelfLoop(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "SelfLoop"
entry_node: "Review"
nodes:
  - name: "Review"
    type: "human"
transitions:
  - from_node: "Review"
    event_type: "Retry"
    to_node: "Review"
exit_transitions:
  - from_node: "Review"
    event_type: "Retry"
    to_node: "Review"
`
	writeWorkflowYAML(t, tmpDir, "SelfLoop", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("SelfLoop")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'SelfLoop' validation failed: transition from_node and to_node must be different (node 'Review', event")
}

func TestWorkflowDefinitionLoader_Load_DuplicateTransitionKey(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "DupeTrans"
entry_node: "Review"
nodes:
  - name: "Review"
    type: "human"
  - name: "End1"
    type: "human"
  - name: "End2"
    type: "human"
transitions:
  - from_node: "Review"
    event_type: "approve"
    to_node: "End1"
  - from_node: "Review"
    event_type: "approve"
    to_node: "End2"
exit_transitions:
  - from_node: "Review"
    event_type: "approve"
    to_node: "End1"
`
	writeWorkflowYAML(t, tmpDir, "DupeTrans", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("DupeTrans")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'DupeTrans' validation failed: duplicate transition for event 'approve' from node 'Review'")
}

func TestWorkflowDefinitionLoader_Load_NodeWithoutOutgoingTransitions(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Isolated"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "Isolated"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "Isolated"
  - from_node: "Start"
    event_type: "Skip"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Skip"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Isolated", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Isolated")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Isolated' validation failed: node 'Isolated' has no outgoing transitions and is not an exit target")
}

// Validation Failures — ExitTransition Integrity

func TestWorkflowDefinitionLoader_Load_DuplicateExitTransition(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "DupeExit"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "DupeExit", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("DupeExit")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'DupeExit' validation failed: duplicate exit transition")
}

func TestWorkflowDefinitionLoader_Load_ExitTransitionNoCorrespondingTransition(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "Orphan"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Other"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Orphan", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Orphan")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'Orphan' validation failed: exit transition")
	assert.Contains(t, err.Error(), "has no corresponding transition definition")
}

func TestWorkflowDefinitionLoader_Load_ExitTransitionTargetsAgentNode(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "AgentExit"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "AgentNode"
    type: "agent"
    agent_role: "Architect"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentNode"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentNode"
`
	writeWorkflowYAML(t, tmpDir, "AgentExit", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("AgentExit")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'AgentExit' validation failed: exit transition")
	assert.Contains(t, err.Error(), "must target a human node")
}

// Validation Failures — Agent Reference Integrity

func TestWorkflowDefinitionLoader_Load_AgentNodeReferencesNonExistentAgent(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	mockLoader.loadFn = func(agentRole string) (*AgentDefinition, error) {
		if agentRole == "Ghost" {
			return nil, &os.PathError{Op: "open", Path: "Ghost", Err: os.ErrNotExist}
		}
		return &AgentDefinition{Role: agentRole}, nil
	}
	yamlContent := `name: "BadAgent"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "AgentNode"
    type: "agent"
    agent_role: "Ghost"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentNode"
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "BadAgent", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("BadAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadAgent' validation failed: node")
	assert.Contains(t, err.Error(), "references invalid agent_role 'Ghost'")
}

func TestWorkflowDefinitionLoader_Load_AgentNodeReferencesInvalidAgent(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	mockLoader.loadFn = func(agentRole string) (*AgentDefinition, error) {
		if agentRole == "InvalidAgent" {
			return nil, assert.AnError
		}
		return &AgentDefinition{Role: agentRole}, nil
	}
	yamlContent := `name: "InvalidAgent"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "AgentNode"
    type: "agent"
    agent_role: "InvalidAgent"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentNode"
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "InvalidAgent", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("InvalidAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'InvalidAgent' validation failed: node")
	assert.Contains(t, err.Error(), "references invalid agent_role 'InvalidAgent'")
}

// Validation Failures — Path Injection

func TestWorkflowDefinitionLoader_Load_WorkflowNameWithPathTraversal(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("../malicious/workflow")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition not found:")
}

// Idempotency

func TestWorkflowDefinitionLoader_Load_RepeatedCalls(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)

	def1, err1 := loader.Load("Simple")
	require.NoError(t, err1)

	def2, err2 := loader.Load("Simple")
	require.NoError(t, err2)

	def3, err3 := loader.Load("Simple")
	require.NoError(t, err3)

	assert.Equal(t, def1.Name, def2.Name)
	assert.Equal(t, def2.Name, def3.Name)
}

func TestWorkflowDefinitionLoader_Load_FileModifiedBetweenCalls(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)

	def1, err1 := loader.Load("Simple")
	require.NoError(t, err1)
	assert.Empty(t, def1.Description)

	// Modify file
	modifiedYAML := `name: "Simple"
description: "Modified description"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
exit_transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "Simple", modifiedYAML)

	def2, err2 := loader.Load("Simple")
	require.NoError(t, err2)
	assert.Equal(t, "Modified description", def2.Description)
}

// Concurrent Behaviour tests are in test/race/workflow_definition_loader_race_test.go

// Dependency Interaction

func TestWorkflowDefinitionLoader_Load_UsesStorageLayout(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	require.NoError(t, err)
	assert.NotNil(t, def)
	expectedPath := GetWorkflowPath(tmpDir, "Simple")
	assert.FileExists(t, expectedPath)
}

func TestWorkflowDefinitionLoader_Load_ReadsFromCorrectPath(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	require.NoError(t, err)
	assert.NotNil(t, def)
	// Verify that the file at the expected path computed by GetWorkflowPath exists and was read
	expectedPath := GetWorkflowPath(tmpDir, "Simple")
	assert.FileExists(t, expectedPath)
}

func TestWorkflowDefinitionLoader_Load_CallsAgentDefinitionLoaderForEachAgentNode(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	yamlContent := `name: "TwoAgents"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "AgentA"
    type: "agent"
    agent_role: "A"
  - name: "AgentB"
    type: "agent"
    agent_role: "B"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentA"
  - from_node: "AgentA"
    event_type: "NextA"
    to_node: "AgentB"
  - from_node: "AgentB"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "AgentB"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "TwoAgents", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	_, err := loader.Load("TwoAgents")

	require.NoError(t, err)
	calls := mockLoader.GetCalls()
	assert.Contains(t, calls, "A")
	assert.Contains(t, calls, "B")
}

// Boundary Values — ProjectRoot

func TestWorkflowDefinitionLoader_New_RelativeProjectRoot(t *testing.T) {
	mockLoader := NewMockAgentDefinitionLoader()

	loader := NewWorkflowDefinitionLoader("./project", mockLoader)

	assert.NotNil(t, loader)
}

func TestWorkflowDefinitionLoader_Load_ProjectRootWithoutSpectraDir(t *testing.T) {
	tmpDir := t.TempDir()
	mockLoader := NewMockAgentDefinitionLoader()

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("Simple")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition not found: Simple")
}

// Error Propagation

func TestWorkflowDefinitionLoader_Load_AgentDefinitionLoaderErrorPropagated(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	mockLoader.loadFn = func(agentRole string) (*AgentDefinition, error) {
		return nil, assert.AnError
	}
	yamlContent := `name: "PropError"
entry_node: "Start"
nodes:
  - name: "Start"
    type: "human"
  - name: "AgentNode"
    type: "agent"
    agent_role: "TestAgent"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Begin"
    to_node: "AgentNode"
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
exit_transitions:
  - from_node: "AgentNode"
    event_type: "Done"
    to_node: "End"
`
	writeWorkflowYAML(t, tmpDir, "PropError", yamlContent)

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)
	def, err := loader.Load("PropError")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'PropError' validation failed: node")
	assert.Contains(t, err.Error(), "references invalid agent_role 'TestAgent'")
}
