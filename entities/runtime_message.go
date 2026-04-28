package entities

import (
	"encoding/json"
	"fmt"
)

// RuntimeMessage represents a message sent from spectra-agent to RuntimeSocketManager.
// It is used for communication over the runtime socket and carries either an event
// emission or an error report.
type RuntimeMessage struct {
	Type            string          `json:"type"`
	ClaudeSessionID string          `json:"claudeSessionID,omitempty"`
	Payload         json.RawMessage `json:"payload"`
}

// EventPayload represents the payload structure for RuntimeMessage with type="event".
type EventPayload struct {
	EventType string          `json:"eventType"`
	Message   string          `json:"message,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// ErrorPayload represents the payload structure for RuntimeMessage with type="error".
type ErrorPayload struct {
	Message string          `json:"message"`
	Detail  json.RawMessage `json:"detail,omitempty"`
}

// NewRuntimeMessage creates a new RuntimeMessage with validation.
// It validates that:
// - Type is non-empty and is one of "event" or "error"
// - Payload is a valid JSON object
// - Payload structure matches the type
func NewRuntimeMessage(msgType string, claudeSessionID string, payload json.RawMessage) (*RuntimeMessage, error) {
	// Validate type
	if msgType == "" {
		return nil, fmt.Errorf("type field must not be empty")
	}

	if msgType != "event" && msgType != "error" {
		return nil, fmt.Errorf("invalid message type '%s'", msgType)
	}

	// Validate payload is present and is a JSON object
	if len(payload) == 0 {
		return nil, fmt.Errorf("missing required field 'payload'")
	}

	var payloadObj any
	if err := json.Unmarshal(payload, &payloadObj); err != nil {
		return nil, fmt.Errorf("invalid JSON in payload: %w", err)
	}

	// Ensure payload is a JSON object
	switch payloadObj.(type) {
	case map[string]any:
		// Valid JSON object
	case nil:
		return nil, fmt.Errorf("payload must be a JSON object")
	default:
		return nil, fmt.Errorf("payload must be a JSON object")
	}

	// Validate payload structure based on type
	switch msgType {
	case "event":
		var eventPayload EventPayload
		if err := json.Unmarshal(payload, &eventPayload); err != nil {
			return nil, fmt.Errorf("invalid event payload: %w", err)
		}
		if eventPayload.EventType == "" {
			return nil, fmt.Errorf("eventType must not be empty")
		}
	case "error":
		var errorPayload ErrorPayload
		if err := json.Unmarshal(payload, &errorPayload); err != nil {
			return nil, fmt.Errorf("invalid error payload: %w", err)
		}
		if errorPayload.Message == "" {
			return nil, fmt.Errorf("error payload missing required field 'message'")
		}
	}

	return &RuntimeMessage{
		Type:            msgType,
		ClaudeSessionID: claudeSessionID,
		Payload:         payload,
	}, nil
}

// Validate validates a RuntimeMessage after deserialization from JSON.
// This is used by RuntimeSocketManager to validate incoming messages.
func (rm *RuntimeMessage) Validate() error {
	// Validate type
	if rm.Type == "" {
		return fmt.Errorf("type field must not be empty")
	}

	if rm.Type != "event" && rm.Type != "error" {
		return fmt.Errorf("invalid message type '%s'", rm.Type)
	}

	// Validate payload is present and is a JSON object
	if len(rm.Payload) == 0 {
		return fmt.Errorf("missing required field 'payload'")
	}

	var payloadObj any
	if err := json.Unmarshal(rm.Payload, &payloadObj); err != nil {
		return fmt.Errorf("invalid JSON in payload: %w", err)
	}

	// Ensure payload is a JSON object
	switch payloadObj.(type) {
	case map[string]any:
		// Valid JSON object
	case nil:
		return fmt.Errorf("payload must be a JSON object")
	default:
		return fmt.Errorf("payload must be a JSON object")
	}

	// Validate payload structure based on type
	switch rm.Type {
	case "event":
		var eventPayload EventPayload
		if err := json.Unmarshal(rm.Payload, &eventPayload); err != nil {
			return fmt.Errorf("invalid event payload: %w", err)
		}
		if eventPayload.EventType == "" {
			return fmt.Errorf("event payload missing required field 'eventType'")
		}
	case "error":
		var errorPayload ErrorPayload
		if err := json.Unmarshal(rm.Payload, &errorPayload); err != nil {
			return fmt.Errorf("invalid error payload: %w", err)
		}
		if errorPayload.Message == "" {
			return fmt.Errorf("error payload missing required field 'message'")
		}
	}

	return nil
}
