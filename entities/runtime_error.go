package entities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// RuntimeError represents a failure signal raised by runtime components when they
// encounter an unrecoverable error during session execution (e.g., socket creation failure,
// state transition failure, panic in message processing).
type RuntimeError struct {
	Issuer       string
	Message      string
	Detail       json.RawMessage
	SessionID    uuid.UUID
	FailingState string
	OccurredAt   int64
}

// NewRuntimeError creates a new RuntimeError with validation.
// It validates that:
// - Issuer is non-empty and not just whitespace
// - Message is non-empty and not just whitespace
// - Detail is valid JSON if provided
//
// Note: SessionID existence and session status are validated by the runtime layer
// (ErrorProcessor), not here.
func NewRuntimeError(
	issuer string,
	message string,
	detail json.RawMessage,
	sessionID uuid.UUID,
	failingState string,
	occurredAt int64,
) (*RuntimeError, error) {
	// Validate Issuer
	if issuer == "" {
		return nil, fmt.Errorf("issuer must be non-empty")
	}
	if strings.TrimSpace(issuer) == "" {
		return nil, fmt.Errorf("issuer cannot be only whitespace")
	}

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

	return &RuntimeError{
		Issuer:       issuer,
		Message:      message,
		Detail:       detail,
		SessionID:    sessionID,
		FailingState: failingState,
		OccurredAt:   occurredAt,
	}, nil
}

// Error implements the error interface for RuntimeError.
func (e *RuntimeError) Error() string {
	return e.Message
}
