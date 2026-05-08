package entities

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// Event is a typed signal entity that drives state transitions in the workflow
// state machine.
type Event struct {
	id        string
	eventType string
	message   string
	payload   json.RawMessage
	emittedBy string
	emittedAt int64
	sessionID string
}

// NewEvent validates all fields and returns an immutable Event.
func NewEvent(id string, eventType string, message string, payload json.RawMessage, emittedBy string, emittedAt int64, sessionID string) (*Event, error) {
	if err := validateUUID(id); err != nil {
		return nil, fmt.Errorf("validating event: id %w", err)
	}

	if err := validateEventType(eventType); err != nil {
		return nil, fmt.Errorf("validating event: %w", err)
	}

	if err := validatePayload(payload); err != nil {
		return nil, fmt.Errorf("validating event: %w", err)
	}

	if emittedBy == "" {
		return nil, fmt.Errorf("validating event: emittedBy must be non-empty")
	}

	if emittedAt <= 0 {
		return nil, fmt.Errorf("validating event: emittedAt must be a positive integer")
	}

	if err := validateUUID(sessionID); err != nil {
		return nil, fmt.Errorf("validating event: sessionID %w", err)
	}

	return &Event{
		id:        id,
		eventType: eventType,
		message:   message,
		payload:   payload,
		emittedBy: emittedBy,
		emittedAt: emittedAt,
		sessionID: sessionID,
	}, nil
}

// ID returns the unique identifier for the event.
func (e *Event) ID() string { return e.id }

// Type returns the event type identifier.
func (e *Event) Type() string { return e.eventType }

// Message returns the optional message for the event recipient.
func (e *Event) Message() string { return e.message }

// Payload returns the event-specific data as a JSON object.
func (e *Event) Payload() json.RawMessage { return e.payload }

// EmittedBy returns the node name from which the event was emitted.
func (e *Event) EmittedBy() string { return e.emittedBy }

// EmittedAt returns the POSIX timestamp when the event was emitted.
func (e *Event) EmittedAt() int64 { return e.emittedAt }

// SessionID returns the associated session identifier.
func (e *Event) SessionID() string { return e.sessionID }

// validateEventType checks that eventType is non-empty and in PascalCase format
// (starts with uppercase letter, contains only alphanumeric characters).
func validateEventType(eventType string) error {
	if eventType == "" {
		return fmt.Errorf("type must be non-empty")
	}

	runes := []rune(eventType)
	if !unicode.IsUpper(runes[0]) {
		return fmt.Errorf("type must start with an uppercase letter (PascalCase)")
	}

	for _, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("type must contain only alphanumeric characters (PascalCase)")
		}
	}

	return nil
}

// validatePayload checks that payload is a valid JSON object (not nil, not
// primitive, not array).
func validatePayload(payload json.RawMessage) error {
	if payload == nil {
		return fmt.Errorf("payload must not be nil")
	}

	trimmed := strings.TrimSpace(string(payload))
	if len(trimmed) == 0 {
		return fmt.Errorf("payload must be a valid JSON object")
	}

	if !json.Valid(payload) {
		return fmt.Errorf("payload contains invalid JSON")
	}

	if trimmed[0] != '{' {
		return fmt.Errorf("payload must be a valid JSON object, got non-object JSON value")
	}

	return nil
}
