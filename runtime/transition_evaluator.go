package runtime

import (
	"github.com/tcfwbper/spectra/storage"
)

// EvaluateTransition is a stateless function that determines the next state transition
// in a workflow state machine. It returns the matching transition, whether it's an exit
// transition, and always returns nil for error (no-match is not an error).
func EvaluateTransition(
	workflowDef *storage.WorkflowDefinition,
	currentState string,
	eventType string,
) (*storage.Transition, bool, error) {
	// Step 3: Search for matching transition
	var matchedTransition *storage.Transition
	for i := range workflowDef.Transitions {
		t := &workflowDef.Transitions[i]
		if t.FromNode == currentState && t.EventType == eventType {
			matchedTransition = t
			break
		}
	}

	// Step 4: If no matching transition found, return nil
	if matchedTransition == nil {
		return nil, false, nil
	}

	// Step 5: Check if it's an exit transition
	isExitTransition := false
	for i := range workflowDef.ExitTransitions {
		et := &workflowDef.ExitTransitions[i]
		if et.FromNode == matchedTransition.FromNode &&
			et.EventType == matchedTransition.EventType &&
			et.ToNode == matchedTransition.ToNode {
			isExitTransition = true
			break
		}
	}

	// Step 6-7: Return the result
	return matchedTransition, isExitTransition, nil
}
