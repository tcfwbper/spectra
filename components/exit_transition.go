package components

import "fmt"

// ExitTransition identifies a specific transition triple that triggers workflow
// completion when traversed. It is a pure immutable value object.
type ExitTransition struct {
	fromNode  string
	eventType string
	toNode    string
}

// NewExitTransition constructs and validates an ExitTransition. Returns an error
// if any constraint is violated.
func NewExitTransition(fromNode string, eventType string, toNode string) (*ExitTransition, error) {
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

	return &ExitTransition{
		fromNode:  fromNode,
		eventType: eventType,
		toNode:    toNode,
	}, nil
}

// FromNode returns the source node of the exit transition.
func (et *ExitTransition) FromNode() string { return et.fromNode }

// EventType returns the event type that triggers workflow completion.
func (et *ExitTransition) EventType() string { return et.eventType }

// ToNode returns the destination node (workflow terminates here).
func (et *ExitTransition) ToNode() string { return et.toNode }
