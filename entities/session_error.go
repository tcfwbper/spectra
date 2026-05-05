package entities

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// SessionError is a shared base structure embedded by both AgentError and
// RuntimeError. It holds common fields and validation logic for all
// session-halting error entities.
type SessionError struct {
	message      string
	detail       json.RawMessage
	occurredAt   int64
	sessionID    string
	failingState string
}

// NewSessionError validates all fields and returns an immutable SessionError.
func NewSessionError(message string, detail json.RawMessage, occurredAt int64, sessionID string, failingState string) (*SessionError, error) {
	if strings.TrimSpace(message) == "" {
		return nil, fmt.Errorf("validating session error: message must contain at least one non-whitespace character")
	}

	if err := validateDetail(detail); err != nil {
		return nil, fmt.Errorf("validating session error: %w", err)
	}

	if occurredAt <= 0 {
		return nil, fmt.Errorf("validating session error: occurredAt must be a positive integer")
	}

	if err := validateUUID(sessionID); err != nil {
		return nil, fmt.Errorf("validating session error: sessionID %w", err)
	}

	if failingState == "" {
		return nil, fmt.Errorf("validating session error: failingState must be non-empty")
	}

	return &SessionError{
		message:      message,
		detail:       detail,
		occurredAt:   occurredAt,
		sessionID:    sessionID,
		failingState: failingState,
	}, nil
}

// Message returns the human-readable error description.
func (se *SessionError) Message() string { return se.message }

// Detail returns the additional error context as a JSON object or nil.
func (se *SessionError) Detail() json.RawMessage { return se.detail }

// OccurredAt returns the POSIX timestamp when the error occurred.
func (se *SessionError) OccurredAt() int64 { return se.occurredAt }

// SessionID returns the associated session identifier.
func (se *SessionError) SessionID() string { return se.sessionID }

// FailingState returns the state machine node where the error occurred.
func (se *SessionError) FailingState() string { return se.failingState }

// validateUUID checks that s is a valid UUID format string.
func validateUUID(s string) error {
	if !uuidRegex.MatchString(s) {
		return fmt.Errorf("must be a valid UUID format")
	}
	return nil
}

// validateDetail checks that detail is either nil or a valid JSON object.
func validateDetail(detail json.RawMessage) error {
	if detail == nil {
		return nil
	}

	trimmed := strings.TrimSpace(string(detail))
	if len(trimmed) == 0 {
		return fmt.Errorf("detail must be nil or a valid JSON object")
	}

	if !json.Valid(detail) {
		return fmt.Errorf("detail contains invalid JSON")
	}

	if trimmed[0] != '{' {
		return fmt.Errorf("detail must be nil or a valid JSON object, got non-object JSON value")
	}

	return nil
}
