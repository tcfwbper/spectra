package entities

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RuntimeError represents an unrecoverable error raised by a runtime component
// during session execution.
type RuntimeError struct {
	issuer       string
	sessionError *SessionError
}

// NewRuntimeError validates all fields and returns an immutable RuntimeError.
// Issuer must contain at least one non-whitespace character.
func NewRuntimeError(issuer string, message string, detail json.RawMessage, occurredAt int64, sessionID string, failingState string) (*RuntimeError, error) {
	if strings.TrimSpace(issuer) == "" {
		return nil, fmt.Errorf("creating runtime error: issuer must contain at least one non-whitespace character")
	}

	se, err := NewSessionError(message, detail, occurredAt, sessionID, failingState)
	if err != nil {
		return nil, fmt.Errorf("creating runtime error: %w", err)
	}

	return &RuntimeError{
		issuer:       issuer,
		sessionError: se,
	}, nil
}

// Issuer returns the runtime component that raised the error.
func (re *RuntimeError) Issuer() string { return re.issuer }

// Message returns the human-readable error description.
func (re *RuntimeError) Message() string { return re.sessionError.Message() }

// Detail returns the additional error context as a JSON object or nil.
func (re *RuntimeError) Detail() json.RawMessage { return re.sessionError.Detail() }

// OccurredAt returns the POSIX timestamp when the error occurred.
func (re *RuntimeError) OccurredAt() int64 { return re.sessionError.OccurredAt() }

// SessionID returns the associated session identifier.
func (re *RuntimeError) SessionID() string { return re.sessionError.SessionID() }

// FailingState returns the state machine node where the error occurred.
func (re *RuntimeError) FailingState() string { return re.sessionError.FailingState() }
