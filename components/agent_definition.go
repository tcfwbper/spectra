package components

import (
	"fmt"
	"strings"
	"unicode"
)

// AgentDefinition describes the metadata for an AI agent role. It is a pure
// immutable value object that validates its own field formats at construction
// time.
type AgentDefinition struct {
	role            string
	model           string
	effort          string
	systemPrompt    string
	agentRoot       string
	allowedTools    []string
	disallowedTools []string
}

// NewAgentDefinition constructs and validates an AgentDefinition. Returns an
// error if any constraint is violated.
func NewAgentDefinition(
	role string,
	model string,
	effort string,
	systemPrompt string,
	agentRoot string,
	allowedTools []string,
	disallowedTools []string,
) (*AgentDefinition, error) {
	if role == "" {
		return nil, fmt.Errorf("role cannot be empty")
	}
	if !isPascalCase(role) {
		return nil, fmt.Errorf("role must be PascalCase (start with uppercase, alphanumeric only)")
	}

	if model == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}

	if effort == "" {
		return nil, fmt.Errorf("effort cannot be empty")
	}

	if systemPrompt == "" {
		return nil, fmt.Errorf("system_prompt cannot be empty")
	}

	if agentRoot == "" {
		return nil, fmt.Errorf("agent_root cannot be empty")
	}
	if strings.HasPrefix(agentRoot, "/") {
		return nil, fmt.Errorf("agent_root must be a relative path")
	}
	if hasDriveLetterPrefix(agentRoot) {
		return nil, fmt.Errorf("agent_root must be a relative path")
	}

	// Normalize nil slices to empty slices.
	at := make([]string, len(allowedTools))
	copy(at, allowedTools)

	dt := make([]string, len(disallowedTools))
	copy(dt, disallowedTools)

	return &AgentDefinition{
		role:            role,
		model:           model,
		effort:          effort,
		systemPrompt:    systemPrompt,
		agentRoot:       agentRoot,
		allowedTools:    at,
		disallowedTools: dt,
	}, nil
}

// hasDriveLetterPrefix checks if s starts with a drive letter prefix (e.g., "C:").
func hasDriveLetterPrefix(s string) bool {
	if len(s) < 2 {
		return false
	}
	return unicode.IsLetter(rune(s[0])) && s[1] == ':'
}

// Role returns the agent role identifier.
func (ad *AgentDefinition) Role() string { return ad.role }

// Model returns the Claude model identifier.
func (ad *AgentDefinition) Model() string { return ad.model }

// Effort returns the Claude effort level.
func (ad *AgentDefinition) Effort() string { return ad.effort }

// SystemPrompt returns the system prompt for the agent.
func (ad *AgentDefinition) SystemPrompt() string { return ad.systemPrompt }

// AgentRoot returns the working directory for agent execution (relative path).
func (ad *AgentDefinition) AgentRoot() string { return ad.agentRoot }

// AllowedTools returns a copy of the allowed tools slice.
func (ad *AgentDefinition) AllowedTools() []string {
	out := make([]string, len(ad.allowedTools))
	copy(out, ad.allowedTools)
	return out
}

// DisallowedTools returns a copy of the disallowed tools slice.
func (ad *AgentDefinition) DisallowedTools() []string {
	out := make([]string, len(ad.disallowedTools))
	copy(out, ad.disallowedTools)
	return out
}
