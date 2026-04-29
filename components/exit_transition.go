package components

import (
	"fmt"
)

// ExitTransition identifies a specific transition that triggers workflow completion when traversed.
type ExitTransition struct {
	fromNode  string
	eventType string
	toNode    string
}

// NewExitTransition creates a new ExitTransition with the given parameters.
func NewExitTransition(fromNode, eventType, toNode string) (*ExitTransition, error) {
	// Validate fromNode
	if fromNode == "" {
		return nil, fmt.Errorf("exit transition from_node must be non-empty")
	}

	// Validate toNode
	if toNode == "" {
		return nil, fmt.Errorf("exit transition to_node must be non-empty")
	}

	// Validate event_type
	if eventType == "" {
		return nil, fmt.Errorf("exit transition event_type must be non-empty")
	}
	if !pascalCasePattern.MatchString(eventType) {
		if hasSpaces(eventType) {
			return nil, fmt.Errorf("exit transition event_type must be PascalCase with no spaces")
		}
		return nil, fmt.Errorf("exit transition event_type must be PascalCase")
	}

	return &ExitTransition{
		fromNode:  fromNode,
		eventType: eventType,
		toNode:    toNode,
	}, nil
}

// GetFromNode returns the source node of the exit transition.
func (e *ExitTransition) GetFromNode() string {
	return e.fromNode
}

// GetEventType returns the event type that triggers the exit transition.
func (e *ExitTransition) GetEventType() string {
	return e.eventType
}

// GetToNode returns the destination node of the exit transition.
func (e *ExitTransition) GetToNode() string {
	return e.toNode
}
