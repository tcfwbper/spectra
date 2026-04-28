package entities

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Event represents a typed signal that drives state transitions in the workflow state machine.
type Event struct {
	ID        uuid.UUID
	Type      string
	Message   string
	Payload   json.RawMessage
	EmittedBy string
	EmittedAt int64
	SessionID uuid.UUID
}

// NewEvent creates a new Event with validation.
// It validates that:
// - Type is non-empty and follows PascalCase convention
// - Payload is a valid JSON object (not primitive or array) if provided
// - SessionID references an existing session (placeholder validation for now)
func NewEvent(
	eventType string,
	message string,
	payload json.RawMessage,
	sessionID uuid.UUID,
) (*Event, error) {
	// Validate Type
	if eventType == "" {
		return nil, fmt.Errorf("type must be non-empty")
	}

	// Check if Type follows PascalCase convention (starts with uppercase letter)
	if !isValidPascalCase(eventType) {
		return nil, fmt.Errorf("type must be in PascalCase format")
	}

	// Default empty payload to empty JSON object
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	// Validate Payload JSON
	var payloadObj any
	if err := json.Unmarshal(payload, &payloadObj); err != nil {
		return nil, fmt.Errorf("invalid JSON in payload: failed to parse: %w", err)
	}

	// Ensure payload is a JSON object, not primitive or array
	switch payloadObj.(type) {
	case map[string]any:
		// Valid JSON object
	case nil:
		return nil, fmt.Errorf("payload must not be null")
	default:
		return nil, fmt.Errorf("payload must be a JSON object")
	}

	// TODO: Validate SessionID references an existing session
	// TODO: Set EmittedBy from session's CurrentState
	// TODO: Validate event type is defined in workflow
	// For now, we use placeholder values

	return &Event{
		ID:        uuid.New(),
		Type:      eventType,
		Message:   message,
		Payload:   payload,
		EmittedBy: "", // Should be set by runtime from session's CurrentState
		EmittedAt: 0,  // Should be set by runtime to current timestamp
		SessionID: sessionID,
	}, nil
}

// isValidPascalCase checks if a string follows PascalCase convention
func isValidPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Check if first character is uppercase
	firstChar := rune(s[0])
	if firstChar < 'A' || firstChar > 'Z' {
		return false
	}
	// Check for invalid characters (like underscores or hyphens)
	for _, c := range s {
		if c == '_' || c == '-' || c == ' ' {
			return false
		}
	}
	return true
}
