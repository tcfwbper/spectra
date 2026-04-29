package race_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// TestTransitionEvaluator_ConcurrentCalls verifies multiple goroutines can call
// TransitionEvaluator concurrently with no data races and no corruption.
func TestTransitionEvaluator_ConcurrentCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	transitions := make([]storage.Transition, 10)
	for i := 0; i < 10; i++ {
		transitions[i] = storage.Transition{
			FromNode:  fmt.Sprintf("Node%d", i),
			EventType: fmt.Sprintf("event%d", i),
			ToNode:    fmt.Sprintf("Node%d", i+1),
		}
	}
	wf := &storage.WorkflowDefinition{
		Name:        "TestWorkflow",
		EntryNode:   "Entry",
		Nodes:       []storage.Node{{Name: "Entry", Type: "human"}},
		Transitions: transitions,
	}

	// Snapshot workflow for comparison after concurrent calls
	snapshotTransitions := make([]storage.Transition, len(wf.Transitions))
	copy(snapshotTransitions, wf.Transitions)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			nodeIdx := idx % 10
			state := fmt.Sprintf("Node%d", nodeIdx)
			eventType := fmt.Sprintf("event%d", nodeIdx)

			transition, _, err := runtime.EvaluateTransition(wf, state, eventType)
			require.NoError(t, err)
			require.NotNil(t, transition)
			assert.Equal(t, fmt.Sprintf("Node%d", nodeIdx+1), transition.ToNode)
		}(i)
	}
	wg.Wait()

	// Workflow definition should be unchanged
	assert.Equal(t, snapshotTransitions, wf.Transitions, "WorkflowDefinition unchanged after concurrent calls")
}

// TestTransitionEvaluator_ConcurrentReadsSameTransition verifies multiple goroutines
// reading the same transition concurrently return identical results with no races.
func TestTransitionEvaluator_ConcurrentReadsSameTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes:     []storage.Node{{Name: "Entry", Type: "human"}},
		Transitions: []storage.Transition{
			{FromNode: "A", EventType: "go", ToNode: "B"},
		},
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	type result struct {
		transition *storage.Transition
		isExit     bool
	}
	results := make([]result, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			transition, isExit, err := runtime.EvaluateTransition(wf, "A", "go")
			require.NoError(t, err)
			results[idx] = result{transition: transition, isExit: isExit}
		}(i)
	}
	wg.Wait()

	// All results should be identical
	for i := 1; i < goroutines; i++ {
		assert.Equal(t, results[0].transition, results[i].transition, "all goroutines should return same transition pointer")
		assert.Equal(t, results[0].isExit, results[i].isExit, "all goroutines should return same isExit value")
	}
}
