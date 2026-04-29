package runtime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/storage"
)

// --- Helper functions for TransitionEvaluator tests ---

func buildTransitionEvalWorkflow(transitions []storage.Transition, exitTransitions []storage.ExitTransition) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:            "TestWorkflow",
		EntryNode:       "Entry",
		Nodes:           []storage.Node{{Name: "Entry", Type: "human"}},
		Transitions:     transitions,
		ExitTransitions: exitTransitions,
	}
}

// --- Happy Path — Regular Transition ---

func TestTransitionEvaluator_RegularTransition_Found(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "A", transition.FromNode)
	assert.Equal(t, "success", transition.EventType)
	assert.Equal(t, "B", transition.ToNode)
	assert.False(t, isExit)
}

func TestTransitionEvaluator_RegularTransition_MultipleInDefinition(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
			{FromNode: "A", EventType: "failure", ToNode: "C"},
			{FromNode: "B", EventType: "done", ToNode: "D"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "failure")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "A", transition.FromNode)
	assert.Equal(t, "failure", transition.EventType)
	assert.Equal(t, "C", transition.ToNode)
	assert.False(t, isExit)
}

// --- Happy Path — Exit Transition ---

func TestTransitionEvaluator_ExitTransition_Found(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "Final", EventType: "complete", ToNode: "End"},
		},
		[]storage.ExitTransition{
			{FromNode: "Final", EventType: "complete", ToNode: "End"},
		},
	)

	transition, isExit, err := EvaluateTransition(wf, "Final", "complete")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "Final", transition.FromNode)
	assert.Equal(t, "complete", transition.EventType)
	assert.Equal(t, "End", transition.ToNode)
	assert.True(t, isExit)
}

func TestTransitionEvaluator_ExitTransition_ExactMatch(t *testing.T) {
	// Transition A->B (success) in Transitions; ExitTransitions has A->C (success)
	// The exit entry does NOT match the found transition (different to_node)
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		[]storage.ExitTransition{
			{FromNode: "A", EventType: "success", ToNode: "C"},
		},
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "B", transition.ToNode)
	assert.False(t, isExit, "ExitTransitions entry doesn't match all fields — should not be marked as exit")
}

// --- Happy Path — No Transition Found ---

func TestTransitionEvaluator_NoMatch_ReturnsNil(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
			{FromNode: "B", EventType: "next", ToNode: "C"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "failure")

	require.NoError(t, err)
	assert.Nil(t, transition)
	assert.False(t, isExit)
}

func TestTransitionEvaluator_NoMatch_EmptyTransitions(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "Any", "any")

	require.NoError(t, err)
	assert.Nil(t, transition)
	assert.False(t, isExit)
}

// --- Happy Path — Transition Lookup ---

func TestTransitionEvaluator_MatchesFromNodeAndEventType(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "Start", EventType: "init", ToNode: "Process"},
		},
		nil,
	)

	transition, _, err := EvaluateTransition(wf, "Start", "init")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "Process", transition.ToNode, "matches based on from_node and event_type regardless of to_node")
}

func TestTransitionEvaluator_FirstMatchReturned(t *testing.T) {
	// Duplicate transitions violate validation but documents current behavior
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
			{FromNode: "A", EventType: "go", ToNode: "C"},
		},
		nil,
	)

	// Should not panic; returns first matched transition
	assert.NotPanics(t, func() {
		transition, _, err := EvaluateTransition(wf, "A", "go")
		require.NoError(t, err)
		require.NotNil(t, transition)
		assert.Equal(t, "B", transition.ToNode, "returns first matching transition (undefined behavior per spec)")
	})
}

// --- Idempotency ---

func TestTransitionEvaluator_RepeatedCalls_IdenticalResults(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		nil,
	)

	t1, isExit1, err1 := EvaluateTransition(wf, "A", "success")
	t2, isExit2, err2 := EvaluateTransition(wf, "A", "success")
	t3, isExit3, err3 := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	assert.Equal(t, t1, t2)
	assert.Equal(t, t2, t3)
	assert.Equal(t, isExit1, isExit2)
	assert.Equal(t, isExit2, isExit3)
	assert.False(t, isExit1)
}

func TestTransitionEvaluator_RepeatedCalls_NoStateChange(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		nil,
	)

	// Snapshot before evaluation
	snapshotTransitions := make([]storage.Transition, len(wf.Transitions))
	copy(snapshotTransitions, wf.Transitions)

	for i := 0; i < 10; i++ {
		EvaluateTransition(wf, "A", "success")
	}

	assert.Equal(t, snapshotTransitions, wf.Transitions, "WorkflowDefinition should remain unchanged after repeated calls")
}

// --- State Transitions ---

func TestTransitionEvaluator_SequentialTransitions(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "next", ToNode: "B"},
			{FromNode: "B", EventType: "next", ToNode: "C"},
			{FromNode: "C", EventType: "done", ToNode: "End"},
		},
		[]storage.ExitTransition{
			{FromNode: "C", EventType: "done", ToNode: "End"},
		},
	)

	// First transition: A -> B
	t1, isExit1, err1 := EvaluateTransition(wf, "A", "next")
	require.NoError(t, err1)
	require.NotNil(t, t1)
	assert.Equal(t, "B", t1.ToNode)
	assert.False(t, isExit1)

	// Second transition: B -> C
	t2, isExit2, err2 := EvaluateTransition(wf, "B", "next")
	require.NoError(t, err2)
	require.NotNil(t, t2)
	assert.Equal(t, "C", t2.ToNode)
	assert.False(t, isExit2)

	// Third transition: C -> End (exit)
	t3, isExit3, err3 := EvaluateTransition(wf, "C", "done")
	require.NoError(t, err3)
	require.NotNil(t, t3)
	assert.Equal(t, "End", t3.ToNode)
	assert.True(t, isExit3)
}

// --- Validation Failures — Invalid State ---

func TestTransitionEvaluator_InvalidCurrentState_NoMatch(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
			{FromNode: "B", EventType: "next", ToNode: "C"},
			{FromNode: "C", EventType: "done", ToNode: "D"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "InvalidNode", "success")

	require.NoError(t, err)
	assert.Nil(t, transition)
	assert.False(t, isExit)
}

func TestTransitionEvaluator_EmptyCurrentState_NoMatch(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "", "go")

	require.NoError(t, err)
	assert.Nil(t, transition)
	assert.False(t, isExit)
}

// --- Validation Failures — Invalid Event Type ---

func TestTransitionEvaluator_InvalidEventType_NoMatch(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
			{FromNode: "A", EventType: "failure", ToNode: "C"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "unknown_event")

	require.NoError(t, err)
	assert.Nil(t, transition)
	assert.False(t, isExit)
}

func TestTransitionEvaluator_EmptyEventType_NoMatch(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "")

	require.NoError(t, err)
	assert.Nil(t, transition)
	assert.False(t, isExit)
}

// --- Boundary Values — Transitions Array ---

func TestTransitionEvaluator_SingleTransition(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "go")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "B", transition.ToNode)
	assert.False(t, isExit)
}

func TestTransitionEvaluator_LargeTransitionsArray(t *testing.T) {
	transitions := make([]storage.Transition, 1000)
	for i := 0; i < 1000; i++ {
		transitions[i] = storage.Transition{
			FromNode:  fmt.Sprintf("State%d", i),
			EventType: fmt.Sprintf("event%d", i),
			ToNode:    fmt.Sprintf("State%d", i+1),
		}
	}
	wf := buildTransitionEvalWorkflow(transitions, nil)

	transition, isExit, err := EvaluateTransition(wf, "State500", "event500")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "State500", transition.FromNode)
	assert.Equal(t, "event500", transition.EventType)
	assert.Equal(t, "State501", transition.ToNode)
	assert.False(t, isExit)
}

// --- Boundary Values — Exit Transitions ---

func TestTransitionEvaluator_EmptyExitTransitions(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		[]storage.ExitTransition{},
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.False(t, isExit, "all transitions are regular when ExitTransitions is empty")
}

func TestTransitionEvaluator_AllTransitionsAreExit(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
			{FromNode: "B", EventType: "done", ToNode: "C"},
		},
		[]storage.ExitTransition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
			{FromNode: "B", EventType: "done", ToNode: "C"},
		},
	)

	t1, isExit1, err1 := EvaluateTransition(wf, "A", "go")
	require.NoError(t, err1)
	require.NotNil(t, t1)
	assert.True(t, isExit1)

	t2, isExit2, err2 := EvaluateTransition(wf, "B", "done")
	require.NoError(t, err2)
	require.NotNil(t, t2)
	assert.True(t, isExit2)
}

// --- Boundary Values — Edge Case States ---

func TestTransitionEvaluator_CaseSensitiveNodeNames(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "Start", EventType: "go", ToNode: "B"},
			{FromNode: "start", EventType: "go", ToNode: "C"},
		},
		nil,
	)

	transition, _, err := EvaluateTransition(wf, "Start", "go")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "B", transition.ToNode, "should match case-sensitive 'Start', not 'start'")
}

func TestTransitionEvaluator_CaseSensitiveEventTypes(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "Success", ToNode: "B"},
			{FromNode: "A", EventType: "success", ToNode: "C"},
		},
		nil,
	)

	transition, _, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "C", transition.ToNode, "should match case-sensitive 'success', not 'Success'")
}

// --- Immutability ---

func TestTransitionEvaluator_WorkflowDefinitionUnmodified(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
			{FromNode: "B", EventType: "done", ToNode: "C"},
		},
		[]storage.ExitTransition{
			{FromNode: "B", EventType: "done", ToNode: "C"},
		},
	)

	// Deep copy for comparison
	snapshotTransitions := make([]storage.Transition, len(wf.Transitions))
	copy(snapshotTransitions, wf.Transitions)
	snapshotExitTransitions := make([]storage.ExitTransition, len(wf.ExitTransitions))
	copy(snapshotExitTransitions, wf.ExitTransitions)
	snapshotName := wf.Name
	snapshotEntryNode := wf.EntryNode

	EvaluateTransition(wf, "A", "go")

	assert.Equal(t, snapshotName, wf.Name)
	assert.Equal(t, snapshotEntryNode, wf.EntryNode)
	assert.Equal(t, snapshotTransitions, wf.Transitions)
	assert.Equal(t, snapshotExitTransitions, wf.ExitTransitions)
}

func TestTransitionEvaluator_NoGlobalState(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
			{FromNode: "B", EventType: "stop", ToNode: "C"},
		},
		nil,
	)

	// First call
	t1, _, err1 := EvaluateTransition(wf, "A", "go")
	require.NoError(t, err1)
	require.NotNil(t, t1)
	assert.Equal(t, "B", t1.ToNode)

	// Second call — should not reference or depend on first call
	t2, _, err2 := EvaluateTransition(wf, "B", "stop")
	require.NoError(t, err2)
	require.NotNil(t, t2)
	assert.Equal(t, "C", t2.ToNode)
}

// --- Type Hierarchy ---

func TestTransitionEvaluator_ReturnsNilNotError(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
		},
		nil,
	)

	transition, isExit, err := EvaluateTransition(wf, "X", "nonexistent")

	assert.Nil(t, transition, "returns nil transition when no match")
	assert.False(t, isExit)
	assert.NoError(t, err, "error field is nil — no match is not an error condition")
}

// --- Mock / Dependency Interaction ---

func TestTransitionEvaluator_AccessesTransitionsArray(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		nil,
	)

	transition, _, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition, "function must access WorkflowDefinition.Transitions for lookup")
	assert.Equal(t, "B", transition.ToNode)
}

func TestTransitionEvaluator_AccessesExitTransitionsArray(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		[]storage.ExitTransition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
	)

	_, isExit, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	assert.True(t, isExit, "function must access WorkflowDefinition.ExitTransitions to determine exit status")
}

func TestTransitionEvaluator_NoSessionAccess(t *testing.T) {
	// TransitionEvaluator is a stateless function that does not accept or
	// access a Session object. It takes only WorkflowDefinition, currentState,
	// and eventType as parameters.
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		nil,
	)

	// Function works correctly without any Session dependency
	transition, _, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "B", transition.ToNode)
}

// --- Error Propagation ---

func TestTransitionEvaluator_NeverReturnsError(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
		},
		nil,
	)

	testCases := []struct {
		name         string
		currentState string
		eventType    string
	}{
		{"valid match", "A", "go"},
		{"no match - wrong state", "Z", "go"},
		{"no match - wrong event", "A", "unknown"},
		{"empty state", "", "go"},
		{"empty event", "A", ""},
		{"both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := EvaluateTransition(wf, tc.currentState, tc.eventType)
			assert.NoError(t, err, "function never returns error")
		})
	}
}

// --- Atomic Replacement ---

func TestTransitionEvaluator_ReturnedTransitionIsReference(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
			{FromNode: "B", EventType: "done", ToNode: "C"},
		},
		nil,
	)

	transition, _, err := EvaluateTransition(wf, "A", "go")

	require.NoError(t, err)
	require.NotNil(t, transition)

	// Returned transition should point to the same memory as the original in Transitions array
	assert.Same(t, &wf.Transitions[0], transition, "returned transition pointer should reference the original in Transitions array")
}

// --- Happy Path — Exit Transition Matching ---

func TestTransitionEvaluator_ExitTransition_AllFieldsMatch(t *testing.T) {
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "End", EventType: "complete", ToNode: "Exit"},
		},
		[]storage.ExitTransition{
			{FromNode: "End", EventType: "complete", ToNode: "Exit"},
		},
	)

	transition, isExit, err := EvaluateTransition(wf, "End", "complete")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "End", transition.FromNode)
	assert.Equal(t, "complete", transition.EventType)
	assert.Equal(t, "Exit", transition.ToNode)
	assert.True(t, isExit, "all three fields match ExitTransitions entry")
}

func TestTransitionEvaluator_ExitTransition_PartialMatch_NotExit(t *testing.T) {
	// Transition A->B (success) in Transitions; ExitTransitions has A->B (failure) — different event_type
	wf := buildTransitionEvalWorkflow(
		[]storage.Transition{
			{FromNode: "A", EventType: "success", ToNode: "B"},
		},
		[]storage.ExitTransition{
			{FromNode: "A", EventType: "failure", ToNode: "B"},
		},
	)

	transition, isExit, err := EvaluateTransition(wf, "A", "success")

	require.NoError(t, err)
	require.NotNil(t, transition)
	assert.Equal(t, "B", transition.ToNode)
	assert.False(t, isExit, "partial match in ExitTransitions should not mark as exit")
}

// --- Boundary Values — Nil Inputs ---

func TestTransitionEvaluator_WorkflowDefinitionNil_Undefined(t *testing.T) {
	// Behavior is undefined when WorkflowDefinition is nil.
	// This test documents current behavior — may panic or return nil.
	// Caller must pass a valid WorkflowDefinition.
	defer func() {
		// If it panics, that is acceptable undefined behavior
		_ = recover()
	}()

	transition, isExit, err := EvaluateTransition(nil, "A", "go")

	// If we reach here (no panic), the result is implementation-defined
	_ = transition
	_ = isExit
	_ = err
}
