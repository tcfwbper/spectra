package entities_test

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities"
)

// NewAgentError is a test helper that wraps entities.NewAgentError.
// It returns error for validation failures, nil for success.
func NewAgentError(
	agentRole string,
	message string,
	detail json.RawMessage,
	sessionID uuid.UUID,
	failingState string,
	occurredAt int64,
) error {
	_, err := entities.NewAgentError(agentRole, message, detail, sessionID, failingState, occurredAt)
	return err
}

// NewEvent is a test helper that wraps entities.NewEvent.
// It returns the created event and any validation error.
func NewEvent(
	eventType string,
	message string,
	payload json.RawMessage,
	sessionID uuid.UUID,
) (*entities.Event, error) {
	return entities.NewEvent(eventType, message, payload, sessionID)
}

// NewRuntimeError is a test helper that wraps entities.NewRuntimeError.
// It returns error for validation failures, nil for success.
func NewRuntimeError(
	issuer string,
	message string,
	detail json.RawMessage,
	sessionID uuid.UUID,
	failingState string,
	occurredAt int64,
) error {
	_, err := entities.NewRuntimeError(issuer, message, detail, sessionID, failingState, occurredAt)
	return err
}

// RecoverSession is a test helper that simulates session recovery attempts.
// According to the spec, recovery is not supported, so this always returns an error.
func RecoverSession(sessionID uuid.UUID) error {
	// Recovery is not supported according to the spec
	// The spec states: "The runtime must reject the recovery request.
	// The human must create a new session to retry the workflow."
	return &RecoveryNotSupportedError{SessionID: sessionID}
}

// RecoveryNotSupportedError represents an error when session recovery is attempted
type RecoveryNotSupportedError struct {
	SessionID uuid.UUID
}

func (e *RecoveryNotSupportedError) Error() string {
	return "recovery not supported: cannot recover terminated session; create new session instead"
}
