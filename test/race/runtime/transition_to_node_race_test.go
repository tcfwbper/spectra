package runtime_race

import (
	"testing"
)

// =============================================================================
// Concurrent Behaviour — TransitionToNode
// =============================================================================
//
// Production surface required:
//   - runtime.NewTransitionToNode(ps, wfDef, loader, invoker, opts...)
//   - runtime.TransitionToNode.Execute(targetNodeName, message string) error
//   - runtime.WithOutput(w io.Writer) TransitionToNodeOption
//   - runtime.TransitionWorkflowDef interface
//   - runtime.TransitionAgentDefLoader interface
//   - runtime.TransitionAgentInvoker interface
//
// This test verifies concurrent calls to Execute for the same session
// serialize state updates without data races.
// =============================================================================

func TestTransitionToNode_Execute_ConcurrentCalls(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/transition_to_node.go — NewTransitionToNode, TransitionToNode.Execute, WithOutput; also requires exported interfaces or test-accessible constructors for race testing from external package")

	// Setup:
	// - Mock WorkflowDefinition.Nodes() returns nodes "NodeA" (human) and "NodeB" (human).
	// - Capture stdout via thread-safe buffer.
	// - Mock PersistentSession.UpdateCurrentStateSafe serializes via internal lock and records calls.
	//
	// Act:
	// - Launch two goroutines: one calls Execute("NodeA", "a"), another calls Execute("NodeB", "b") concurrently.
	//
	// Assert:
	// - Both calls return nil.
	// - PersistentSession.UpdateCurrentStateSafe called exactly twice (once with "NodeA", once with "NodeB").
	// - No data race detected by race detector (-race flag).
}
