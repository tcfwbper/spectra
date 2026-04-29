package entities

import (
	"github.com/google/uuid"
)

// SessionMetadata represents the persistable state of a session, excluding the event history.
// It is JSON-serializable for persistence to session.json by SessionMetadataStore.
type SessionMetadata struct {
	ID           uuid.UUID              `json:"id"`
	WorkflowName string                 `json:"workflowName"`
	Status       string                 `json:"status"`
	CreatedAt    int64                  `json:"createdAt"`
	UpdatedAt    int64                  `json:"updatedAt"`
	CurrentState string                 `json:"currentState"`
	SessionData  map[string]interface{} `json:"sessionData"`
	Error        *AgentError            `json:"error,omitempty"`
	EventHistory []Event                `json:"-"`
}
