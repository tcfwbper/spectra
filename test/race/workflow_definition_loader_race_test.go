package race_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

// Node represents a node in a workflow
type Node struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	AgentRole   string `yaml:"agent_role,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Transition represents a state transition in a workflow
type Transition struct {
	FromNode  string `yaml:"from_node"`
	EventType string `yaml:"event_type"`
	ToNode    string `yaml:"to_node"`
}

// ExitTransition represents a workflow exit transition
type ExitTransition struct {
	FromNode  string `yaml:"from_node"`
	EventType string `yaml:"event_type"`
	ToNode    string `yaml:"to_node"`
}

// WorkflowDefinition represents a workflow configuration loaded from YAML
// This is a stub awaiting implementation.
type WorkflowDefinition struct {
	Name            string           `yaml:"name"`
	Description     string           `yaml:"description,omitempty"`
	EntryNode       string           `yaml:"entry_node"`
	ExitTransitions []ExitTransition `yaml:"exit_transitions"`
	Nodes           []Node           `yaml:"nodes"`
	Transitions     []Transition     `yaml:"transitions"`
}

// MockAgentDefinitionLoader is a mock implementation for testing
type MockAgentDefinitionLoader struct {
	mu    sync.Mutex
	calls []string
}

// NewMockAgentDefinitionLoader creates a new mock loader
func NewMockAgentDefinitionLoader() *MockAgentDefinitionLoader {
	return &MockAgentDefinitionLoader{
		calls: []string{},
	}
}

// Load calls the mock function
func (m *MockAgentDefinitionLoader) Load(agentRole string) (*AgentDefinition, error) {
	m.mu.Lock()
	m.calls = append(m.calls, agentRole)
	m.mu.Unlock()
	return &AgentDefinition{Role: agentRole}, nil
}

// AgentLoader is an interface for loading agent definitions
type AgentLoader interface {
	Load(agentRole string) (*AgentDefinition, error)
}

// WorkflowDefinitionLoader loads workflow definitions from .spectra/workflows/
// This is a stub awaiting implementation.
type WorkflowDefinitionLoader struct {
	projectRoot string
	agentLoader AgentLoader
}

// NewWorkflowDefinitionLoader creates a new WorkflowDefinitionLoader
func NewWorkflowDefinitionLoader(projectRoot string, agentLoader AgentLoader) *WorkflowDefinitionLoader {
	return &WorkflowDefinitionLoader{
		projectRoot: projectRoot,
		agentLoader: agentLoader,
	}
}

// Load loads a workflow definition from disk
// Stub implementation - will be provided by the implementation phase
func (l *WorkflowDefinitionLoader) Load(workflowName string) (*WorkflowDefinition, error) {
	return nil, nil
}

// Test helper functions

func setupWorkflowTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, storage.WorkflowsDir)
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))
	return tmpDir
}

func writeWorkflowYAML(t *testing.T, projectRoot, workflowName, content string) {
	t.Helper()
	workflowPath := storage.GetWorkflowPath(projectRoot, workflowName)
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

// TestWorkflowDefinitionLoader_Load_ConcurrentSameWorkflow tests multiple goroutines loading the same workflow concurrently
func TestWorkflowDefinitionLoader_Load_ConcurrentSameWorkflow(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	writeWorkflowYAML(t, tmpDir, "Simple", createMinimalValidWorkflowYAML("Simple"))

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)

	done := make(chan *WorkflowDefinition, 10)
	for range 10 {
		go func() {
			def, err := loader.Load("Simple")
			require.NoError(t, err)
			done <- def
		}()
	}

	results := make([]*WorkflowDefinition, 10)
	for i := range 10 {
		results[i] = <-done
	}

	for i := 1; i < 10; i++ {
		assert.Equal(t, results[0].Name, results[i].Name)
	}
}

// TestWorkflowDefinitionLoader_Load_ConcurrentDifferentWorkflows tests multiple goroutines loading different workflows concurrently
func TestWorkflowDefinitionLoader_Load_ConcurrentDifferentWorkflows(t *testing.T) {
	tmpDir := setupWorkflowTestDir(t)
	mockLoader := NewMockAgentDefinitionLoader()
	workflows := []string{"Workflow1", "Workflow2", "Workflow3", "Workflow4", "Workflow5"}
	for _, wf := range workflows {
		writeWorkflowYAML(t, tmpDir, wf, createMinimalValidWorkflowYAML(wf))
	}

	loader := NewWorkflowDefinitionLoader(tmpDir, mockLoader)

	done := make(chan string, 10)
	for i := range 10 {
		wf := workflows[i%len(workflows)]
		go func(w string) {
			def, err := loader.Load(w)
			require.NoError(t, err)
			done <- def.Name
		}(wf)
	}

	for range 10 {
		result := <-done
		assert.Contains(t, workflows, result)
	}
}
