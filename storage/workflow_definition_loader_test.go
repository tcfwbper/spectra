package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectra-ai/spectra/components"
)

// --- Mock AgentLoader ---

// mockAgentLoader implements the AgentLoader interface for testing.
type mockAgentLoader struct {
	// loadFunc defines custom Load behavior. If nil, returns a default valid AgentDefinition.
	loadFunc func(agentRole string) (*components.AgentDefinition, error)
	// calls records all Load invocations in order.
	calls []string
	mu    sync.Mutex
}

func newMockAgentLoader() *mockAgentLoader {
	return &mockAgentLoader{}
}

// withLoadFunc sets a custom Load function.
func (m *mockAgentLoader) withLoadFunc(fn func(agentRole string) (*components.AgentDefinition, error)) *mockAgentLoader {
	m.loadFunc = fn
	return m
}

// withSuccess configures the mock to return success for all calls.
func (m *mockAgentLoader) withSuccess() *mockAgentLoader {
	m.loadFunc = func(agentRole string) (*components.AgentDefinition, error) {
		return components.NewAgentDefinition(agentRole, "claude-sonnet-4-20250514", "high", "Mock prompt.", ".", nil, nil)
	}
	return m
}

// withError configures the mock to return an error for a specific role.
func (m *mockAgentLoader) withErrorForRole(role string, err error) *mockAgentLoader {
	prev := m.loadFunc
	m.loadFunc = func(agentRole string) (*components.AgentDefinition, error) {
		if agentRole == role {
			return nil, err
		}
		if prev != nil {
			return prev(agentRole)
		}
		return components.NewAgentDefinition(agentRole, "claude-sonnet-4-20250514", "high", "Mock prompt.", ".", nil, nil)
	}
	return m
}

// withFailAlways configures the mock to always fail.
func (m *mockAgentLoader) withFailAlways(err error) *mockAgentLoader {
	m.loadFunc = func(agentRole string) (*components.AgentDefinition, error) {
		return nil, err
	}
	return m
}

func (m *mockAgentLoader) Load(agentRole string) (*components.AgentDefinition, error) {
	m.mu.Lock()
	m.calls = append(m.calls, agentRole)
	m.mu.Unlock()

	if m.loadFunc != nil {
		return m.loadFunc(agentRole)
	}
	// Default: return a valid AgentDefinition.
	return components.NewAgentDefinition(agentRole, "claude-sonnet-4-20250514", "high", "Default mock.", ".", nil, nil)
}

// getCalls returns a copy of all recorded calls.
func (m *mockAgentLoader) getCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.calls))
	copy(out, m.calls)
	return out
}

// --- YAML Fixture Builders for WorkflowDefinitionLoader ---

// validWorkflowYAML returns a well-formed workflow YAML with standard structure.
// It defines: a human entry node, an agent node, transitions between them, and an exit transition.
func validWorkflowYAML(description string) string {
	return fmt.Sprintf(`description: "%s"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User provides input"
  - name: "AgentWork"
    type: "agent"
    agentRole: "Worker"
    description: "Agent performs work"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentWork"
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
`, description)
}

// validWorkflowYAMLMultipleAgents returns a workflow YAML with multiple agent-type nodes.
func validWorkflowYAMLMultipleAgents() string {
	return `description: "Multi-agent workflow"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User provides input"
  - name: "AgentOne"
    type: "agent"
    agentRole: "Architect"
    description: "First agent"
  - name: "AgentTwo"
    type: "agent"
    agentRole: "Developer"
    description: "Second agent"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentOne"
  - fromNode: "AgentOne"
    eventType: "Review"
    toNode: "AgentTwo"
  - fromNode: "AgentTwo"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "AgentTwo"
    eventType: "Done"
    toNode: "HumanInput"
`
}

// validWorkflowYAMLHumanOnly returns a workflow YAML with only human-type nodes.
func validWorkflowYAMLHumanOnly() string {
	return `description: "Human-only workflow"
entryNode: "StepOne"
nodes:
  - name: "StepOne"
    type: "human"
    description: "First human step"
  - name: "StepTwo"
    type: "human"
    description: "Second human step"
transitions:
  - fromNode: "StepOne"
    eventType: "Next"
    toNode: "StepTwo"
  - fromNode: "StepTwo"
    eventType: "Back"
    toNode: "StepOne"
exitTransitions:
  - fromNode: "StepTwo"
    eventType: "Back"
    toNode: "StepOne"
`
}

// --- Filesystem Fixture Builders for WorkflowDefinitionLoader ---

// makeTempDirWithWorkflows creates a temp dir containing `.spectra/workflows/` directory.
func makeTempDirWithWorkflows(t *testing.T) string {
	t.Helper()
	dir := makeTempDirWithSpectra(t)
	workflowsDir := filepath.Join(dir, ".spectra", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("makeTempDirWithWorkflows: failed to create workflows dir: %v", err)
	}
	return dir
}

// writeWorkflowYAML writes YAML content to .spectra/workflows/<name>.yaml.
func writeWorkflowYAML(t *testing.T, projectRoot, name, content string) {
	t.Helper()
	filePath := GetWorkflowPath(projectRoot, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}

// --- Happy Path Tests ---

func TestWorkflowDefinitionLoader_Load_ValidDefinition(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "MyWorkflow", validWorkflowYAML("A test workflow"))

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("MyWorkflow")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "MyWorkflow", def.Name())
	assert.Equal(t, "A test workflow", def.Description())
	assert.Equal(t, "HumanInput", def.EntryNode())
	assert.Len(t, def.Nodes(), 2)
	assert.Len(t, def.Transitions(), 2)
	assert.Len(t, def.ExitTransitions(), 1)
}

func TestWorkflowDefinitionLoader_Load_NameDerivedFromFilename(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "CodeReview", validWorkflowYAML("Code review workflow"))

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("CodeReview")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "CodeReview", def.Name())
}

func TestWorkflowDefinitionLoader_Load_EmptyDescription(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// YAML without description field.
	yaml := `entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User provides input"
  - name: "AgentWork"
    type: "agent"
    agentRole: "Worker"
    description: "Agent performs work"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentWork"
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "Minimal", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Minimal")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "", def.Description())
}

func TestWorkflowDefinitionLoader_Load_MultipleAgentNodes(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "Multi", validWorkflowYAMLMultipleAgents())

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Multi")

	require.NoError(t, err)
	require.NotNil(t, def)

	// Verify AgentLoader was called for each unique agent role.
	calls := mock.getCalls()
	assert.Contains(t, calls, "Architect")
	assert.Contains(t, calls, "Developer")
}

// --- Error Propagation Tests ---

func TestWorkflowDefinitionLoader_Load_FileNotFound(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Missing")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition not found: Missing")
}

func TestWorkflowDefinitionLoader_Load_ReadPermissionDenied(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "Locked", validWorkflowYAML("Locked workflow"))

	// Remove read permissions.
	filePath := GetWorkflowPath(projectRoot, "Locked")
	require.NoError(t, os.Chmod(filePath, 0000))
	t.Cleanup(func() { os.Chmod(filePath, 0644) })

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Locked")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read workflow definition 'Locked':")
	assert.Contains(t, err.Error(), "permission")
}

func TestWorkflowDefinitionLoader_Load_YamlSyntaxError(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "Bad", "entryNode: \"unclosed\n  bad: [indent")

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Bad")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow definition 'Bad':")
}

func TestWorkflowDefinitionLoader_Load_YamlUnknownField(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	yaml := validWorkflowYAML("Extra field") + "customField: value\n"
	writeWorkflowYAML(t, projectRoot, "Extra", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Extra")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow definition 'Extra':")
}

func TestWorkflowDefinitionLoader_Load_YamlSnakeCaseField(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// Use snake_case entry_node instead of camelCase entryNode.
	yaml := `entry_node: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "Snake", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Snake")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow definition 'Snake':")
}

func TestWorkflowDefinitionLoader_Load_NodeConstructorFails_WithName(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// Node with valid name but invalid type "bot".
	yaml := `description: "Bad node"
entryNode: "HumanInput"
nodes:
  - name: "NodeName"
    type: "bot"
    description: "Invalid type"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "NodeName"
exitTransitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "NodeName"
`
	writeWorkflowYAML(t, projectRoot, "BadNode", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("BadNode")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadNode' validation failed: node 'NodeName':")
}

func TestWorkflowDefinitionLoader_Load_NodeConstructorFails_EmptyName(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// Node with empty name field.
	yaml := `description: "No name"
entryNode: "HumanInput"
nodes:
  - name: ""
    type: "human"
    description: "No name node"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "NoName", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("NoName")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'NoName' validation failed: node[0]:")
}

func TestWorkflowDefinitionLoader_Load_TransitionConstructorFails(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// Transition where fromNode == toNode.
	yaml := `description: "Bad transition"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
  - name: "AgentWork"
    type: "agent"
    agentRole: "Worker"
    description: "Agent"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "BadTrans", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("BadTrans")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadTrans' validation failed: transition (from")
}

func TestWorkflowDefinitionLoader_Load_ExitTransitionConstructorFails(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// ExitTransition with empty fields that would cause constructor failure.
	yaml := `description: "Bad exit"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
  - name: "AgentWork"
    type: "agent"
    agentRole: "Worker"
    description: "Agent"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentWork"
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: ""
    eventType: ""
    toNode: ""
`
	writeWorkflowYAML(t, projectRoot, "BadExit", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("BadExit")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadExit' validation failed: exit_transition (from")
}

func TestWorkflowDefinitionLoader_Load_WorkflowDefinitionConstructorFails(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// Duplicate (fromNode, eventType) pair triggers WorkflowDefinition constructor error.
	yaml := `description: "Bad graph"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
  - name: "AgentOne"
    type: "agent"
    agentRole: "Worker"
    description: "First agent"
  - name: "AgentTwo"
    type: "agent"
    agentRole: "Worker"
    description: "Second agent"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentOne"
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentTwo"
  - fromNode: "AgentOne"
    eventType: "Done"
    toNode: "HumanInput"
  - fromNode: "AgentTwo"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "AgentOne"
    eventType: "Done"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "BadGraph", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("BadGraph")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadGraph' validation failed:")
}

func TestWorkflowDefinitionLoader_Load_AgentRoleNotFound(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// Workflow with agent node referencing "NonExistent" role.
	yaml := `description: "Bad agent ref"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
  - name: "NodeName"
    type: "agent"
    agentRole: "NonExistent"
    description: "Agent with bad role"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "NodeName"
  - fromNode: "NodeName"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions:
  - fromNode: "NodeName"
    eventType: "Done"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "BadRef", yaml)

	mock := newMockAgentLoader().withErrorForRole("NonExistent", fmt.Errorf("agent definition not found: NonExistent"))
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("BadRef")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'BadRef' validation failed: node 'NodeName' references invalid agent_role 'NonExistent':")
}

func TestWorkflowDefinitionLoader_Load_AgentRoleValidationFails(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "InvalidAgent", validWorkflowYAML("Agent validation fails"))

	mock := newMockAgentLoader().withErrorForRole("Worker", fmt.Errorf("agent definition 'Worker' validation failed: agent_root directory not found: /some/path"))
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("InvalidAgent")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'InvalidAgent' validation failed: node")
	assert.Contains(t, err.Error(), "agent_root directory not found")
}

// --- Null / Empty Input Tests ---

func TestWorkflowDefinitionLoader_Load_EmptyWorkflowName(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition not found: ")
}

func TestWorkflowDefinitionLoader_Load_EmptyYamlFile(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "Empty", "")

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("Empty")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow definition 'Empty':")
}

func TestWorkflowDefinitionLoader_Load_EmptyNodesArray(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	yaml := `description: "No nodes"
entryNode: "HumanInput"
nodes: []
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentWork"
exitTransitions:
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "NoNodes", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("NoNodes")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'NoNodes' validation failed:")
}

func TestWorkflowDefinitionLoader_Load_EmptyTransitionsArray(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	yaml := `description: "No transitions"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
  - name: "AgentWork"
    type: "agent"
    agentRole: "Worker"
    description: "Agent"
transitions: []
exitTransitions:
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
`
	writeWorkflowYAML(t, projectRoot, "NoTrans", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("NoTrans")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'NoTrans' validation failed:")
}

func TestWorkflowDefinitionLoader_Load_EmptyExitTransitionsArray(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	yaml := `description: "No exit transitions"
entryNode: "HumanInput"
nodes:
  - name: "HumanInput"
    type: "human"
    description: "User"
  - name: "AgentWork"
    type: "agent"
    agentRole: "Worker"
    description: "Agent"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "AgentWork"
  - fromNode: "AgentWork"
    eventType: "Done"
    toNode: "HumanInput"
exitTransitions: []
`
	writeWorkflowYAML(t, projectRoot, "NoExit", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("NoExit")

	assert.Nil(t, def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow definition 'NoExit' validation failed:")
}

// --- Boundary Values Tests ---

func TestWorkflowDefinitionLoader_Load_PathTraversal(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	_, err := loader.Load("../malicious/workflow")

	require.Error(t, err)
}

// --- Mock / Dependency Interaction Tests ---

func TestWorkflowDefinitionLoader_Load_AgentLoaderNotCalledForHumanNodes(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "HumanOnly", validWorkflowYAMLHumanOnly())

	// Mock that panics if called.
	mock := newMockAgentLoader().withLoadFunc(func(agentRole string) (*components.AgentDefinition, error) {
		t.Fatalf("AgentLoader should not be called for human-only workflow, but called with %q", agentRole)
		return nil, nil
	})

	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	def, err := loader.Load("HumanOnly")

	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Empty(t, mock.getCalls())
}

func TestWorkflowDefinitionLoader_Load_AgentLoaderFailFast(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "TwoAgents", validWorkflowYAMLMultipleAgents())

	// Fail on the first agent role encountered.
	callCount := 0
	mock := newMockAgentLoader().withLoadFunc(func(agentRole string) (*components.AgentDefinition, error) {
		callCount++
		return nil, fmt.Errorf("agent definition not found: %s", agentRole)
	})

	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	_, err := loader.Load("TwoAgents")

	require.Error(t, err)
	assert.Equal(t, 1, callCount, "AgentLoader should be called at most once (fail-fast)")
}

func TestWorkflowDefinitionLoader_Load_NodeConstructionFailFast(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	// First node is invalid (bad type), second is valid.
	yaml := `description: "Fail fast"
entryNode: "HumanInput"
nodes:
  - name: "BadNode"
    type: "bot"
    description: "Invalid"
  - name: "HumanInput"
    type: "human"
    description: "Valid"
transitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "BadNode"
exitTransitions:
  - fromNode: "HumanInput"
    eventType: "Submit"
    toNode: "BadNode"
`
	writeWorkflowYAML(t, projectRoot, "MultiNode", yaml)

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)
	_, err := loader.Load("MultiNode")

	require.Error(t, err)
	// Error should reference first node only.
	assert.Contains(t, err.Error(), "node 'BadNode':")
	assert.NotContains(t, err.Error(), "HumanInput")
}

// --- Idempotency Tests ---

func TestWorkflowDefinitionLoader_Load_NoCaching(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "Mutable", validWorkflowYAML("version one"))

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)

	// First load returns original description.
	def1, err := loader.Load("Mutable")
	require.NoError(t, err)
	assert.Equal(t, "version one", def1.Description())

	// Overwrite YAML with different description.
	writeWorkflowYAML(t, projectRoot, "Mutable", validWorkflowYAML("version two"))

	// Second load reflects updated content.
	def2, err := loader.Load("Mutable")
	require.NoError(t, err)
	assert.Equal(t, "version two", def2.Description())
}

// --- Concurrent Behaviour Tests ---

func TestWorkflowDefinitionLoader_Load_ConcurrentAccess(t *testing.T) {
	t.Skip("scaffolded: requires storage.WorkflowDefinitionLoader and storage.NewWorkflowDefinitionLoader (production type not yet implemented)")

	projectRoot := makeTempDirWithWorkflows(t)
	writeWorkflowYAML(t, projectRoot, "Shared", validWorkflowYAML("Shared workflow"))

	mock := newMockAgentLoader().withSuccess()
	loader := NewWorkflowDefinitionLoader(projectRoot, mock)

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
				assert.Equal(t, "Shared", def.Name())
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
	_ = (*components.AgentDefinition)(nil)
)
