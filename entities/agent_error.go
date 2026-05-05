package entities

import (
	"encoding/json"
	"fmt"
)

// AgentError represents an unrecoverable error raised by an agent (or from a
// human node) during workflow execution.
type AgentError struct {
	agentRole    string
	sessionError *SessionError
}

// NewAgentError validates all fields and returns an immutable AgentError.
// AgentRole accepts any string including empty (empty represents human node origin).
func NewAgentError(agentRole string, message string, detail json.RawMessage, occurredAt int64, sessionID string, failingState string) (*AgentError, error) {
	se, err := NewSessionError(message, detail, occurredAt, sessionID, failingState)
	if err != nil {
		return nil, fmt.Errorf("creating agent error: %w", err)
	}

	return &AgentError{
		agentRole:    agentRole,
		sessionError: se,
	}, nil
}

// AgentRole returns the agent role that raised the error.
func (ae *AgentError) AgentRole() string { return ae.agentRole }

// Message returns the human-readable error description.
func (ae *AgentError) Message() string { return ae.sessionError.Message() }

// Detail returns the additional error context as a JSON object or nil.
func (ae *AgentError) Detail() json.RawMessage { return ae.sessionError.Detail() }

// OccurredAt returns the POSIX timestamp when the error occurred.
func (ae *AgentError) OccurredAt() int64 { return ae.sessionError.OccurredAt() }

// SessionID returns the associated session identifier.
func (ae *AgentError) SessionID() string { return ae.sessionError.SessionID() }

// FailingState returns the state machine node where the error occurred.
func (ae *AgentError) FailingState() string { return ae.sessionError.FailingState() }

// Error implements the error interface.
func (ae *AgentError) Error() string { return ae.sessionError.Message() }
