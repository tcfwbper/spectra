package entities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AgentError represents a failure signal raised by an agent when it cannot complete
// its task due to an unrecoverable error (e.g., model failure, missing context, tool failure).
type AgentError struct {
	AgentRole    string
	Message      string
	Detail       json.RawMessage
	SessionID    uuid.UUID
	FailingState string
	OccurredAt   int64
}

// NewAgentError creates a new AgentError with validation.
// It validates that:
// - Message is non-empty and not just whitespace
// - Detail is valid JSON if provided
//
// Note: SessionID existence and session status are validated by the runtime layer
// (ErrorProcessor), not here.
func NewAgentError(
	agentRole string,
	message string,
	detail json.RawMessage,
	sessionID uuid.UUID,
	failingState string,
	occurredAt int64,
) (*AgentError, error) {
	// Validate Message
	if message == "" {
		return nil, fmt.Errorf("message must be non-empty")
	}
	if strings.TrimSpace(message) == "" {
		return nil, fmt.Errorf("message cannot be only whitespace")
	}

	// Validate Detail JSON if provided
	if len(detail) > 0 {
		var temp any
		if err := json.Unmarshal(detail, &temp); err != nil {
			return nil, fmt.Errorf("invalid JSON in detail: failed to parse: %w", err)
		}
	}

	// SessionID existence and session status are validated by the runtime layer
	// (ErrorProcessor), not the entity constructor.

	return &AgentError{
		AgentRole:    agentRole,
		Message:      message,
		Detail:       detail,
		SessionID:    sessionID,
		FailingState: failingState,
		OccurredAt:   occurredAt,
	}, nil
}

// Error implements the error interface for AgentError.
func (e *AgentError) Error() string {
	return e.Message
}
