package runtime

import "github.com/tcfwbper/spectra/components"

// WorkflowDef defines the read-only interface for workflow definitions consumed
// by EvaluateTransition. It provides access to transitions and exit transitions.
type WorkflowDef interface {
	Transitions() []*components.Transition
	ExitTransitions() []*components.ExitTransition
}

// EvaluateTransition is a stateless pure function that determines the next state
// transition in a workflow state machine. Given a WorkflowDefinition, the current
// state (node name), and an event type, it finds the matching transition and
// determines whether it is an exit transition.
//
// Returns (nil, false) if no matching transition is found.
// Returns (transition, true) if the transition is an exit transition.
// Returns (transition, false) if the transition is a regular transition.
func EvaluateTransition(wfDef WorkflowDef, currentState, eventType string) (*components.Transition, bool) {
	// Find matching transition by (currentState, eventType).
	var matched *components.Transition
	for _, tr := range wfDef.Transitions() {
		if tr.FromNode() == currentState && tr.EventType() == eventType {
			matched = tr
			break
		}
	}
	if matched == nil {
		return nil, false
	}

	// Check whether the matched transition is an exit transition.
	for _, et := range wfDef.ExitTransitions() {
		if et.FromNode() == matched.FromNode() &&
			et.EventType() == matched.EventType() &&
			et.ToNode() == matched.ToNode() {
			return matched, true
		}
	}

	return matched, false
}
