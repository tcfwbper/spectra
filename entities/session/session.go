package session

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/tcfwbper/spectra/entities"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Session represents a single execution instance of a workflow.
// It composes SessionMetadata (the persistable subset) with runtime-only state.
type Session struct {
	SessionMetadata

	// EventHistory is the chronological log of emitted events.
	EventHistory []entities.Event

	// mu protects all Session state after construction.
	mu sync.RWMutex
}

// NewSession validates all inputs and returns an initialized Session.
func NewSession(id string, workflowName string, entryNode string, createdAt int64) (*Session, error) {
	if !uuidRegex.MatchString(id) {
		return nil, fmt.Errorf("invalid session ID: must be a valid UUID")
	}

	if workflowName == "" {
		return nil, fmt.Errorf("workflow name cannot be empty")
	}

	if entryNode == "" {
		return nil, fmt.Errorf("entry node cannot be empty")
	}

	if createdAt <= 0 {
		return nil, fmt.Errorf("createdAt must be a positive POSIX timestamp")
	}

	return &Session{
		SessionMetadata: SessionMetadata{
			ID:           id,
			WorkflowName: workflowName,
			Status:       "initializing",
			CreatedAt:    createdAt,
			UpdatedAt:    createdAt,
			CurrentState: entryNode,
			SessionData:  make(map[string]any),
			Error:        nil,
		},
		EventHistory: []entities.Event{},
	}, nil
}
