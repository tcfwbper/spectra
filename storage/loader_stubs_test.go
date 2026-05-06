package storage

// TEMPORARY STUBS — Delete this file once agent_definition_loader.go and
// workflow_definition_loader.go are implemented in production code.
//
// These stubs exist solely to allow test scaffolding to compile before
// the production surface is available. They do NOT implement production
// behavior and always return errors indicating the stub is active.

import (
	"fmt"

	"github.com/spectra-ai/spectra/components"
)

// --- AgentLoader interface (expected in production) ---

// AgentLoader is the interface that WorkflowDefinitionLoader depends on for
// agent-role referential integrity. Defined here temporarily until the
// production file declares it.
type AgentLoader interface {
	Load(agentRole string) (*components.AgentDefinition, error)
}

// --- AgentDefinitionLoader stub ---

// AgentDefinitionLoader is a temporary stub for the production type.
// Missing production symbol: storage.AgentDefinitionLoader (agent_definition_loader.go)
type AgentDefinitionLoader struct {
	projectRoot string
}

// NewAgentDefinitionLoader is a temporary stub constructor.
// Missing production symbol: storage.NewAgentDefinitionLoader (agent_definition_loader.go)
func NewAgentDefinitionLoader(projectRoot string) *AgentDefinitionLoader {
	return &AgentDefinitionLoader{projectRoot: projectRoot}
}

// Load is a stub that always returns an error indicating production code is missing.
// Missing production method: (*AgentDefinitionLoader).Load (agent_definition_loader.go)
func (l *AgentDefinitionLoader) Load(agentRole string) (*components.AgentDefinition, error) {
	return nil, fmt.Errorf("STUB: AgentDefinitionLoader.Load not yet implemented")
}

// --- WorkflowDefinitionLoader stub ---

// WorkflowDefinitionLoader is a temporary stub for the production type.
// Missing production symbol: storage.WorkflowDefinitionLoader (workflow_definition_loader.go)
type WorkflowDefinitionLoader struct {
	projectRoot string
	agentLoader AgentLoader
}

// NewWorkflowDefinitionLoader is a temporary stub constructor.
// Missing production symbol: storage.NewWorkflowDefinitionLoader (workflow_definition_loader.go)
func NewWorkflowDefinitionLoader(projectRoot string, agentLoader AgentLoader) *WorkflowDefinitionLoader {
	return &WorkflowDefinitionLoader{projectRoot: projectRoot, agentLoader: agentLoader}
}

// Load is a stub that always returns an error indicating production code is missing.
// Missing production method: (*WorkflowDefinitionLoader).Load (workflow_definition_loader.go)
func (l *WorkflowDefinitionLoader) Load(workflowName string) (*components.WorkflowDefinition, error) {
	return nil, fmt.Errorf("STUB: WorkflowDefinitionLoader.Load not yet implemented")
}
