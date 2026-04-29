package session

import (
	"sync"
)

// SessionMetadataStore defines the interface for persisting session metadata.
type SessionMetadataStore interface {
	Write(metadata SessionMetadata) error
}

// EventStore defines the interface for persisting events.
type EventStore interface {
	WriteEvent(event Event) error
}

// Logger defines the interface for logging warnings.
type Logger interface {
	Warning(msg string)
}

// Session represents a single execution instance of a workflow.
// It embeds SessionMetadata and adds runtime-only fields.
type Session struct {
	SessionMetadata
	EventHistory        []Event
	mu                  sync.RWMutex
	terminationNotifier chan<- struct{}
	metadataStore       SessionMetadataStore
	eventStore          EventStore
	logger              Logger
}

// Event represents a workflow event.
type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	SessionID string         `json:"session_id"`
	EmittedAt int64          `json:"emitted_at"`
	EmittedBy string         `json:"emitted_by"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload"`
}

// AgentError represents an error from an agent node.
type AgentError struct {
	NodeName string
	Message  string
}

func (e *AgentError) Error() string {
	return e.Message
}

// RuntimeError represents an error from the runtime.
type RuntimeError struct {
	Issuer  string
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Message
}
