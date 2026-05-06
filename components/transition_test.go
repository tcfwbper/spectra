package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewTransition_ValidInputs(t *testing.T) {
	tr, err := NewTransition("Architect", "DraftCompleted", "Reviewer")

	require.NoError(t, err)
	assert.Equal(t, "Architect", tr.FromNode())
	assert.Equal(t, "DraftCompleted", tr.EventType())
	assert.Equal(t, "Reviewer", tr.ToNode())
}

// --- Validation Failures — FromNode ---

func TestNewTransition_EmptyFromNode(t *testing.T) {
	_, err := NewTransition("", "DraftCompleted", "Reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "from_node cannot be empty")
}

func TestNewTransition_FromNodeStartsLowercase(t *testing.T) {
	_, err := NewTransition("architect", "DraftCompleted", "Reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewTransition_FromNodeContainsHyphen(t *testing.T) {
	_, err := NewTransition("Archi-Tect", "DraftCompleted", "Reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Validation Failures — EventType ---

func TestNewTransition_EmptyEventType(t *testing.T) {
	_, err := NewTransition("Architect", "", "Reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event_type cannot be empty")
}

func TestNewTransition_EventTypeStartsLowercase(t *testing.T) {
	_, err := NewTransition("Architect", "draftCompleted", "Reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewTransition_EventTypeContainsUnderscore(t *testing.T) {
	_, err := NewTransition("Architect", "Draft_Completed", "Reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Validation Failures — ToNode ---

func TestNewTransition_EmptyToNode(t *testing.T) {
	_, err := NewTransition("Architect", "DraftCompleted", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_node cannot be empty")
}

func TestNewTransition_ToNodeStartsLowercase(t *testing.T) {
	_, err := NewTransition("Architect", "DraftCompleted", "reviewer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Validation Failures — Self-Loop ---

func TestNewTransition_SelfLoop(t *testing.T) {
	_, err := NewTransition("Architect", "DraftCompleted", "Architect")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "from_node and to_node must be different")
}

// --- Immutability ---

func TestTransition_Immutability(t *testing.T) {
	tr, err := NewTransition("Architect", "DraftCompleted", "Reviewer")
	require.NoError(t, err)

	// Access getters multiple times; values must remain identical to construction inputs.
	assert.Equal(t, "Architect", tr.FromNode())
	assert.Equal(t, "DraftCompleted", tr.EventType())
	assert.Equal(t, "Reviewer", tr.ToNode())

	// Second access — ensure no mutation between calls.
	assert.Equal(t, "Architect", tr.FromNode())
	assert.Equal(t, "DraftCompleted", tr.EventType())
	assert.Equal(t, "Reviewer", tr.ToNode())
}
