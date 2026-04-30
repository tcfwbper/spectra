package entities_test

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities"
)

func NewAgentError(
	agentRole string,
	message string,
	detail json.RawMessage,
	sessionID uuid.UUID,
	failingState string,
	occurredAt int64,
) (*entities.AgentError, error) {
	return entities.NewAgentError(agentRole, message, detail, sessionID, failingState, occurredAt)
}

func NewEvent(
	eventType string,
	message string,
	payload json.RawMessage,
	sessionID uuid.UUID,
) (*entities.Event, error) {
	return entities.NewEvent(eventType, message, payload, sessionID)
}

func NewRuntimeError(
	issuer string,
	message string,
	detail json.RawMessage,
	sessionID uuid.UUID,
	failingState string,
	occurredAt int64,
) (*entities.RuntimeError, error) {
	return entities.NewRuntimeError(issuer, message, detail, sessionID, failingState, occurredAt)
}

func RecoverSession(sessionID uuid.UUID) error {
	return &RecoveryNotSupportedError{SessionID: sessionID}
}

type RecoveryNotSupportedError struct {
	SessionID uuid.UUID
}

func (e *RecoveryNotSupportedError) Error() string {
	return "recovery not supported: cannot recover terminated session; create new session instead"
}
