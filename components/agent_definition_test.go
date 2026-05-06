package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewAgentDefinition_AllFieldsValid(t *testing.T) {
	allowedTools := []string{"Bash", "Read"}
	disallowedTools := []string{"Write"}

	ad, err := NewAgentDefinition(
		"DefaultArchitect",
		"claude-sonnet-4-20250514",
		"high",
		"You are an architect.",
		"spec",
		allowedTools,
		disallowedTools,
	)

	require.NoError(t, err)
	assert.Equal(t, "DefaultArchitect", ad.Role())
	assert.Equal(t, "claude-sonnet-4-20250514", ad.Model())
	assert.Equal(t, "high", ad.Effort())
	assert.Equal(t, "You are an architect.", ad.SystemPrompt())
	assert.Equal(t, "spec", ad.AgentRoot())
	assert.Equal(t, []string{"Bash", "Read"}, ad.AllowedTools())
	assert.Equal(t, []string{"Write"}, ad.DisallowedTools())
}

func TestNewAgentDefinition_EmptyAllowedTools(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Writer",
		"claude-sonnet-4-20250514",
		"low",
		"Write docs.",
		"docs",
		[]string{},
		[]string{},
	)

	require.NoError(t, err)
	assert.Equal(t, []string{}, ad.AllowedTools())
}

func TestNewAgentDefinition_NilAllowedTools(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Writer",
		"claude-sonnet-4-20250514",
		"low",
		"Write docs.",
		"docs",
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, []string{}, ad.AllowedTools())
	assert.Equal(t, []string{}, ad.DisallowedTools())
}

func TestNewAgentDefinition_AgentRootDot(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Root",
		"claude-sonnet-4-20250514",
		"medium",
		"Prompt.",
		".",
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, ".", ad.AgentRoot())
}

func TestNewAgentDefinition_AgentRootNestedPath(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Nested",
		"claude-sonnet-4-20250514",
		"medium",
		"Prompt.",
		"src/internal/pkg",
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "src/internal/pkg", ad.AgentRoot())
}

func TestNewAgentDefinition_OverlappingTools(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Tester",
		"claude-sonnet-4-20250514",
		"high",
		"Test.",
		"test",
		[]string{"Bash"},
		[]string{"Bash"},
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"Bash"}, ad.AllowedTools())
	assert.Equal(t, []string{"Bash"}, ad.DisallowedTools())
}

func TestNewAgentDefinition_UnrecognizedModelAccepted(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Agent",
		"invalid-model",
		"high",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "invalid-model", ad.Model())
}

// --- Validation Failures — Role ---

func TestNewAgentDefinition_EmptyRole(t *testing.T) {
	_, err := NewAgentDefinition(
		"",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "role cannot be empty", err.Error())
}

func TestNewAgentDefinition_RoleStartsLowercase(t *testing.T) {
	_, err := NewAgentDefinition(
		"defaultArchitect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "role must be PascalCase (start with uppercase, alphanumeric only)", err.Error())
}

func TestNewAgentDefinition_RoleContainsHyphen(t *testing.T) {
	_, err := NewAgentDefinition(
		"Default-Architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "role must be PascalCase (start with uppercase, alphanumeric only)", err.Error())
}

func TestNewAgentDefinition_RoleContainsUnderscore(t *testing.T) {
	_, err := NewAgentDefinition(
		"default_architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "role must be PascalCase (start with uppercase, alphanumeric only)", err.Error())
}

// --- Validation Failures — Model ---

func TestNewAgentDefinition_EmptyModel(t *testing.T) {
	_, err := NewAgentDefinition(
		"Architect",
		"",
		"high",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "model cannot be empty", err.Error())
}

// --- Validation Failures — Effort ---

func TestNewAgentDefinition_EmptyEffort(t *testing.T) {
	_, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"",
		"Prompt.",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "effort cannot be empty", err.Error())
}

// --- Validation Failures — SystemPrompt ---

func TestNewAgentDefinition_EmptySystemPrompt(t *testing.T) {
	_, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"",
		"src",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "system_prompt cannot be empty", err.Error())
}

// --- Validation Failures — AgentRoot ---

func TestNewAgentDefinition_EmptyAgentRoot(t *testing.T) {
	_, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "agent_root cannot be empty", err.Error())
}

func TestNewAgentDefinition_AgentRootAbsoluteUnix(t *testing.T) {
	_, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"/usr/local",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "agent_root must be a relative path", err.Error())
}

func TestNewAgentDefinition_AgentRootDriveLetter(t *testing.T) {
	_, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"C:\\projects",
		nil,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "agent_root must be a relative path", err.Error())
}

// --- Immutability ---

func TestAgentDefinition_Immutability(t *testing.T) {
	ad, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"You are an architect.",
		"spec",
		[]string{"Bash"},
		[]string{"Write"},
	)
	require.NoError(t, err)

	// First access.
	assert.Equal(t, "Architect", ad.Role())
	assert.Equal(t, "claude-sonnet-4-20250514", ad.Model())
	assert.Equal(t, "high", ad.Effort())
	assert.Equal(t, "You are an architect.", ad.SystemPrompt())
	assert.Equal(t, "spec", ad.AgentRoot())
	assert.Equal(t, []string{"Bash"}, ad.AllowedTools())
	assert.Equal(t, []string{"Write"}, ad.DisallowedTools())

	// Second access — ensure no mutation between calls.
	assert.Equal(t, "Architect", ad.Role())
	assert.Equal(t, "claude-sonnet-4-20250514", ad.Model())
	assert.Equal(t, "high", ad.Effort())
	assert.Equal(t, "You are an architect.", ad.SystemPrompt())
	assert.Equal(t, "spec", ad.AgentRoot())
	assert.Equal(t, []string{"Bash"}, ad.AllowedTools())
	assert.Equal(t, []string{"Write"}, ad.DisallowedTools())
}

// --- Data Independence (Copy Semantics) ---

func TestAgentDefinition_AllowedToolsCopySemantics(t *testing.T) {
	allowedTools := []string{"Bash", "Read"}

	ad, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"src",
		allowedTools,
		nil,
	)
	require.NoError(t, err)

	// Mutate the original slice after construction.
	allowedTools[0] = "Mutated"

	// Getter must still return the original values.
	assert.Equal(t, []string{"Bash", "Read"}, ad.AllowedTools())
}

func TestAgentDefinition_DisallowedToolsCopySemantics(t *testing.T) {
	disallowedTools := []string{"Write"}

	ad, err := NewAgentDefinition(
		"Architect",
		"claude-sonnet-4-20250514",
		"high",
		"Prompt.",
		"src",
		nil,
		disallowedTools,
	)
	require.NoError(t, err)

	// Mutate the original slice after construction.
	disallowedTools[0] = "Mutated"

	// Getter must still return the original values.
	assert.Equal(t, []string{"Write"}, ad.DisallowedTools())
}
