package components

import "fmt"

// Transition defines an event-driven edge between two nodes in a workflow.
// It is a pure immutable value object.
type Transition struct {
	fromNode  string
	eventType string
	toNode    string
}

// NewTransition constructs and validates a Transition. Returns an error if any
// constraint is violated.
func NewTransition(fromNode string, eventType string, toNode string) (*Transition, error) {
	if fromNode == "" {
		return nil, fmt.Errorf("from_node cannot be empty")
	}
	if !isPascalCase(fromNode) {
		return nil, fmt.Errorf("from_node must be PascalCase (start with uppercase, alphanumeric only)")
	}

	if eventType == "" {
		return nil, fmt.Errorf("event_type cannot be empty")
	}
	if !isPascalCase(eventType) {
		return nil, fmt.Errorf("event_type must be PascalCase (start with uppercase, alphanumeric only)")
	}

	if toNode == "" {
		return nil, fmt.Errorf("to_node cannot be empty")
	}
	if !isPascalCase(toNode) {
		return nil, fmt.Errorf("to_node must be PascalCase (start with uppercase, alphanumeric only)")
	}

	if fromNode == toNode {
		return nil, fmt.Errorf("from_node and to_node must be different")
	}

	return &Transition{
		fromNode:  fromNode,
		eventType: eventType,
		toNode:    toNode,
	}, nil
}

// FromNode returns the source node of the transition.
func (t *Transition) FromNode() string { return t.fromNode }

// EventType returns the event type that triggers this transition.
func (t *Transition) EventType() string { return t.eventType }

// ToNode returns the destination node of the transition.
func (t *Transition) ToNode() string { return t.toNode }
