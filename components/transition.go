package components

import (
	"fmt"
)

// Transition defines an event-driven state change in a workflow.
type Transition struct {
	fromNode  string
	eventType string
	toNode    string
}

// NewTransition creates a new Transition with the given parameters.
func NewTransition(fromNode, eventType, toNode string) (*Transition, error) {
	// Validate fromNode
	if fromNode == "" {
		return nil, fmt.Errorf("transition from_node must be non-empty")
	}

	// Validate toNode
	if toNode == "" {
		return nil, fmt.Errorf("transition to_node must be non-empty")
	}

	// Validate no self-loop
	if fromNode == toNode {
		return nil, fmt.Errorf("transition from_node and to_node must be different")
	}

	// Validate event_type
	if eventType == "" {
		return nil, fmt.Errorf("transition event_type must be non-empty")
	}
	if !pascalCasePattern.MatchString(eventType) {
		if hasSpaces(eventType) {
			return nil, fmt.Errorf("transition event_type must be PascalCase with no spaces")
		}
		return nil, fmt.Errorf("transition event_type must be PascalCase")
	}

	return &Transition{
		fromNode:  fromNode,
		eventType: eventType,
		toNode:    toNode,
	}, nil
}

// GetFromNode returns the source node of the transition.
func (t *Transition) GetFromNode() string {
	return t.fromNode
}

// GetEventType returns the event type that triggers the transition.
func (t *Transition) GetEventType() string {
	return t.eventType
}

// GetToNode returns the destination node of the transition.
func (t *Transition) GetToNode() string {
	return t.toNode
}
