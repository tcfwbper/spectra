package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
)

// =============================================================================
// Test Helpers — TransitionEvaluator
// =============================================================================

// mockWorkflowDef provides a minimal mock for the WorkflowDefinition interface
// used by EvaluateTransition. This decouples tests from the complex validation
// in components.NewWorkflowDefinition.
//
// The production EvaluateTransition is expected to accept an interface with:
//   - Transitions() []*components.Transition
//   - ExitTransitions() []*components.ExitTransition
type mockWorkflowDef struct {
	transitions     []*components.Transition
	exitTransitions []*components.ExitTransition
}

func (m *mockWorkflowDef) Transitions() []*components.Transition {
	return m.transitions
}

func (m *mockWorkflowDef) ExitTransitions() []*components.ExitTransition {
	return m.exitTransitions
}

// mustNewTransition creates a Transition or fails the test; for test setup only.
func mustNewTransition(t *testing.T, from, eventType, to string) *components.Transition {
	t.Helper()
	tr, err := components.NewTransition(from, eventType, to)
	require.NoError(t, err, "mustNewTransition(%q, %q, %q)", from, eventType, to)
	return tr
}

// mustNewExitTransition creates an ExitTransition or fails the test; for test setup only.
func mustNewExitTransition(t *testing.T, from, eventType, to string) *components.ExitTransition {
	t.Helper()
	et, err := components.NewExitTransition(from, eventType, to)
	require.NoError(t, err, "mustNewExitTransition(%q, %q, %q)", from, eventType, to)
	return et
}

// =============================================================================
// Happy Path — EvaluateTransition
// =============================================================================

func TestEvaluateTransition_RegularTransition(t *testing.T) {
	// Setup: one regular transition A->Done->B, no exit transitions.
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "A", "Done")

	// Assert
	require.NotNil(t, tr)
	assert.Equal(t, "B", tr.ToNode())
	assert.False(t, isExit)
}

func TestEvaluateTransition_ExitTransition(t *testing.T) {
	// Setup: transition B->Complete->End is also an exit transition.
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "B", "Complete", "End")},
		exitTransitions: []*components.ExitTransition{mustNewExitTransition(t, "B", "Complete", "End")},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "B", "Complete")

	// Assert
	require.NotNil(t, tr)
	assert.Equal(t, "End", tr.ToNode())
	assert.True(t, isExit)
}

func TestEvaluateTransition_NoMatch(t *testing.T) {
	// Setup: one transition A->Done->B
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: query with non-matching event type
	tr, isExit := EvaluateTransition(wfDef, "A", "Error")

	// Assert
	assert.Nil(t, tr)
	assert.False(t, isExit)
}

func TestEvaluateTransition_MultipleTransitions_CorrectMatch(t *testing.T) {
	// Setup: multiple transitions
	wfDef := &mockWorkflowDef{
		transitions: []*components.Transition{
			mustNewTransition(t, "A", "Done", "B"),
			mustNewTransition(t, "A", "Error", "C"),
			mustNewTransition(t, "B", "Done", "D"),
		},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: query A + Error
	tr, isExit := EvaluateTransition(wfDef, "A", "Error")

	// Assert: selects correct transition
	require.NotNil(t, tr)
	assert.Equal(t, "C", tr.ToNode())
	assert.False(t, isExit)
}

func TestEvaluateTransition_MultipleExitTransitions(t *testing.T) {
	// Setup
	wfDef := &mockWorkflowDef{
		transitions: []*components.Transition{
			mustNewTransition(t, "X", "Done", "Y"),
			mustNewTransition(t, "Y", "Finish", "End"),
		},
		exitTransitions: []*components.ExitTransition{
			mustNewExitTransition(t, "Y", "Finish", "End"),
		},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "Y", "Finish")

	// Assert
	require.NotNil(t, tr)
	assert.Equal(t, "End", tr.ToNode())
	assert.True(t, isExit)
}

// =============================================================================
// Null / Empty Input — EvaluateTransition
// =============================================================================

func TestEvaluateTransition_EmptyCurrentState(t *testing.T) {
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: empty currentState
	tr, isExit := EvaluateTransition(wfDef, "", "Done")

	// Assert
	assert.Nil(t, tr)
	assert.False(t, isExit)
}

func TestEvaluateTransition_EmptyEventType(t *testing.T) {
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: empty eventType
	tr, isExit := EvaluateTransition(wfDef, "A", "")

	// Assert
	assert.Nil(t, tr)
	assert.False(t, isExit)
}

func TestEvaluateTransition_EmptyTransitionsList(t *testing.T) {
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "A", "Done")

	// Assert
	assert.Nil(t, tr)
	assert.False(t, isExit)
}

// =============================================================================
// Boundary Values — ExitTransition Classification
// =============================================================================

func TestEvaluateTransition_PartialExitMatch_DifferentToNode(t *testing.T) {
	// Setup: exit transition has same FromNode and EventType but different ToNode.
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{mustNewExitTransition(t, "A", "Done", "C")},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "A", "Done")

	// Assert: NOT classified as exit (ToNode mismatch)
	require.NotNil(t, tr)
	assert.Equal(t, "B", tr.ToNode())
	assert.False(t, isExit)
}

func TestEvaluateTransition_PartialExitMatch_DifferentEventType(t *testing.T) {
	// Setup: exit transition has same FromNode and ToNode but different EventType.
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{mustNewExitTransition(t, "A", "Error", "B")},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "A", "Done")

	// Assert: NOT classified as exit (EventType mismatch)
	require.NotNil(t, tr)
	assert.Equal(t, "B", tr.ToNode())
	assert.False(t, isExit)
}

func TestEvaluateTransition_PartialExitMatch_DifferentFromNode(t *testing.T) {
	// Setup: exit transition has same EventType and ToNode but different FromNode.
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{mustNewExitTransition(t, "X", "Done", "B")},
	}

	// Act
	tr, isExit := EvaluateTransition(wfDef, "A", "Done")

	// Assert: NOT classified as exit (FromNode mismatch)
	require.NotNil(t, tr)
	assert.Equal(t, "B", tr.ToNode())
	assert.False(t, isExit)
}

// =============================================================================
// Idempotency — EvaluateTransition
// =============================================================================

func TestEvaluateTransition_RepeatedCalls_SameResult(t *testing.T) {
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: call twice
	tr1, isExit1 := EvaluateTransition(wfDef, "A", "Done")
	tr2, isExit2 := EvaluateTransition(wfDef, "A", "Done")

	// Assert: identical results
	require.NotNil(t, tr1)
	require.NotNil(t, tr2)
	assert.Equal(t, tr1.ToNode(), tr2.ToNode())
	assert.Equal(t, tr1.FromNode(), tr2.FromNode())
	assert.Equal(t, tr1.EventType(), tr2.EventType())
	assert.Equal(t, isExit1, isExit2)
}

// =============================================================================
// Mock / Dependency Interaction — EvaluateTransition
// =============================================================================

func TestEvaluateTransition_DoesNotModifyWorkflowDefinition(t *testing.T) {
	// Setup: record original slices
	transitions := []*components.Transition{mustNewTransition(t, "A", "Done", "B")}
	exitTransitions := []*components.ExitTransition{}

	wfDef := &mockWorkflowDef{
		transitions:     transitions,
		exitTransitions: exitTransitions,
	}

	origTransLen := len(wfDef.Transitions())
	origExitLen := len(wfDef.ExitTransitions())

	// Act
	_, _ = EvaluateTransition(wfDef, "A", "Done")

	// Assert: no mutations
	assert.Equal(t, origTransLen, len(wfDef.Transitions()))
	assert.Equal(t, origExitLen, len(wfDef.ExitTransitions()))
}

func TestEvaluateTransition_InvalidCurrentState_NoError(t *testing.T) {
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: non-existent node name — should not panic
	assert.NotPanics(t, func() {
		tr, isExit := EvaluateTransition(wfDef, "NonExistentNode", "Done")
		assert.Nil(t, tr)
		assert.False(t, isExit)
	})
}

func TestEvaluateTransition_InvalidEventType_NoError(t *testing.T) {
	wfDef := &mockWorkflowDef{
		transitions:     []*components.Transition{mustNewTransition(t, "A", "Done", "B")},
		exitTransitions: []*components.ExitTransition{},
	}

	// Act: undefined event type — should not panic
	assert.NotPanics(t, func() {
		tr, isExit := EvaluateTransition(wfDef, "A", "UndefinedEvent")
		assert.Nil(t, tr)
		assert.False(t, isExit)
	})
}
