package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewNode_AgentTypeValid(t *testing.T) {
	n, err := NewNode("ReviewStep", "agent", "Architect", "Reviews code")

	require.NoError(t, err)
	assert.Equal(t, "ReviewStep", n.Name())
	assert.Equal(t, "agent", n.Type())
	assert.Equal(t, "Architect", n.AgentRole())
	assert.Equal(t, "Reviews code", n.Description())
}

func TestNewNode_HumanTypeValid(t *testing.T) {
	n, err := NewNode("HumanApproval", "human", "", "Waits for approval")

	require.NoError(t, err)
	assert.Equal(t, "HumanApproval", n.Name())
	assert.Equal(t, "human", n.Type())
	assert.Equal(t, "", n.AgentRole())
	assert.Equal(t, "Waits for approval", n.Description())
}

func TestNewNode_EmptyDescription(t *testing.T) {
	n, err := NewNode("Draft", "agent", "Writer", "")

	require.NoError(t, err)
	assert.Equal(t, "", n.Description())
}

// --- Validation Failures — Name ---

func TestNewNode_EmptyName(t *testing.T) {
	_, err := NewNode("", "agent", "Writer", "desc")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "node name cannot be empty")
}

func TestNewNode_NameStartsLowercase(t *testing.T) {
	_, err := NewNode("reviewStep", "agent", "Writer", "desc")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewNode_NameContainsHyphen(t *testing.T) {
	_, err := NewNode("Review-Step", "agent", "Writer", "desc")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewNode_NameContainsUnderscore(t *testing.T) {
	_, err := NewNode("review_step", "agent", "Writer", "desc")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Validation Failures — Type ---

func TestNewNode_EmptyType(t *testing.T) {
	_, err := NewNode("Review", "", "Writer", "desc")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "node type must be 'agent' or 'human'")
}

func TestNewNode_InvalidType(t *testing.T) {
	_, err := NewNode("Review", "bot", "Writer", "desc")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "node type must be 'agent' or 'human'")
}

// --- Validation Failures — AgentRole ---

func TestNewNode_AgentTypeEmptyRole(t *testing.T) {
	_, err := NewNode("Review", "agent", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent_role is required when type is 'agent'")
}

func TestNewNode_AgentTypeRoleStartsLowercase(t *testing.T) {
	_, err := NewNode("Review", "agent", "architect", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewNode_HumanTypeNonEmptyRole(t *testing.T) {
	_, err := NewNode("Approval", "human", "Architect", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent_role must be empty when type is 'human'")
}

// --- Immutability ---

func TestNode_Immutability(t *testing.T) {
	n, err := NewNode("Draft", "agent", "Writer", "Drafts content")
	require.NoError(t, err)

	// Access getters multiple times; values must remain identical to construction inputs.
	assert.Equal(t, "Draft", n.Name())
	assert.Equal(t, "agent", n.Type())
	assert.Equal(t, "Writer", n.AgentRole())
	assert.Equal(t, "Drafts content", n.Description())

	// Second access — ensure no mutation between calls.
	assert.Equal(t, "Draft", n.Name())
	assert.Equal(t, "agent", n.Type())
	assert.Equal(t, "Writer", n.AgentRole())
	assert.Equal(t, "Drafts content", n.Description())
}
