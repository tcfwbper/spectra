package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewExitTransition_ValidInputs(t *testing.T) {
	et, err := NewExitTransition("Reviewer", "Approved", "HumanApproval")

	require.NoError(t, err)
	assert.Equal(t, "Reviewer", et.FromNode())
	assert.Equal(t, "Approved", et.EventType())
	assert.Equal(t, "HumanApproval", et.ToNode())
}

func TestNewExitTransition_SameFromAndToNode(t *testing.T) {
	et, err := NewExitTransition("HumanApproval", "Completed", "HumanApproval")

	require.NoError(t, err)
	assert.Equal(t, "HumanApproval", et.FromNode())
	assert.Equal(t, "Completed", et.EventType())
	assert.Equal(t, "HumanApproval", et.ToNode())
}

// --- Validation Failures — FromNode ---

func TestNewExitTransition_EmptyFromNode(t *testing.T) {
	_, err := NewExitTransition("", "Approved", "HumanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "from_node cannot be empty")
}

func TestNewExitTransition_FromNodeStartsLowercase(t *testing.T) {
	_, err := NewExitTransition("reviewer", "Approved", "HumanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewExitTransition_FromNodeContainsSpecialChar(t *testing.T) {
	_, err := NewExitTransition("Review-Node", "Approved", "HumanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Validation Failures — EventType ---

func TestNewExitTransition_EmptyEventType(t *testing.T) {
	_, err := NewExitTransition("Reviewer", "", "HumanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event_type cannot be empty")
}

func TestNewExitTransition_EventTypeStartsLowercase(t *testing.T) {
	_, err := NewExitTransition("Reviewer", "approved", "HumanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

func TestNewExitTransition_EventTypeContainsHyphen(t *testing.T) {
	_, err := NewExitTransition("Reviewer", "Requirement-Approved", "HumanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Validation Failures — ToNode ---

func TestNewExitTransition_EmptyToNode(t *testing.T) {
	_, err := NewExitTransition("Reviewer", "Approved", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_node cannot be empty")
}

func TestNewExitTransition_ToNodeStartsLowercase(t *testing.T) {
	_, err := NewExitTransition("Reviewer", "Approved", "humanApproval")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PascalCase")
}

// --- Immutability ---

func TestExitTransition_Immutability(t *testing.T) {
	et, err := NewExitTransition("Reviewer", "Approved", "HumanApproval")
	require.NoError(t, err)

	// Access getters multiple times; values must remain identical to construction inputs.
	assert.Equal(t, "Reviewer", et.FromNode())
	assert.Equal(t, "Approved", et.EventType())
	assert.Equal(t, "HumanApproval", et.ToNode())

	// Second access — ensure no mutation between calls.
	assert.Equal(t, "Reviewer", et.FromNode())
	assert.Equal(t, "Approved", et.EventType())
	assert.Equal(t, "HumanApproval", et.ToNode())
}
