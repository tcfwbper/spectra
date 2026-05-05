package entities

import (
	"encoding/json"
	"fmt"
)

// RuntimeMessage is the structured message entity used for communication
// between spectra-agent (client) and the runtime (server) over the runtime
// socket. It is a pure data entity — it does not perform serialization,
// transmission, size checking, or payload semantic validation.
type RuntimeMessage struct {
	msgType        string
	payload        json.RawMessage
	claudeSessionID string
}

// NewRuntimeMessage validates all fields and returns an immutable RuntimeMessage.
func NewRuntimeMessage(msgType string, payload json.RawMessage, claudeSessionID string) (*RuntimeMessage, error) {
	if err := validateMessageType(msgType); err != nil {
		return nil, err
	}

	if err := validatePayload(payload); err != nil {
		return nil, fmt.Errorf("validating runtime message: %w", err)
	}

	// Defensive copy of payload to guarantee immutability.
	copied := make(json.RawMessage, len(payload))
	copy(copied, payload)

	return &RuntimeMessage{
		msgType:        msgType,
		payload:        copied,
		claudeSessionID: claudeSessionID,
	}, nil
}

// Type returns the message type identifier.
func (m *RuntimeMessage) Type() string { return m.msgType }

// Payload returns a copy of the message payload as a JSON object.
func (m *RuntimeMessage) Payload() json.RawMessage {
	cp := make(json.RawMessage, len(m.payload))
	copy(cp, m.payload)
	return cp
}

// ClaudeSessionID returns the Claude session identifier.
func (m *RuntimeMessage) ClaudeSessionID() string { return m.claudeSessionID }

// validateMessageType checks that msgType is one of the recognized message types.
func validateMessageType(msgType string) error {
	if msgType == "" {
		return fmt.Errorf("validating runtime message: type must be non-empty")
	}
	if msgType != "event" && msgType != "error" {
		return fmt.Errorf("validating runtime message: type %q is not recognized (must be \"event\" or \"error\")", msgType)
	}
	return nil
}
