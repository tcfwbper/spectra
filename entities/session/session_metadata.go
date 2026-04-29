package session

import (
	"encoding/json"
)

// SessionMetadata represents the persistable state of a session, excluding the event history.
// It is JSON-serializable for persistence to session.json by SessionMetadataStore.
// SessionMetadata is a plain data structure with no methods of its own.
type SessionMetadata struct {
	ID           string         `json:"id"`
	WorkflowName string         `json:"workflowName"`
	Status       string         `json:"status"`
	CreatedAt    int64          `json:"createdAt"`
	UpdatedAt    int64          `json:"updatedAt"`
	CurrentState string         `json:"currentState"`
	SessionData  map[string]any `json:"sessionData"`
	Error        error          `json:"error,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for SessionMetadata
func (s *SessionMetadata) UnmarshalJSON(data []byte) error {
	type Alias SessionMetadata
	aux := &struct {
		Error json.RawMessage `json:"error,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try to unmarshal error field if present
	if len(aux.Error) > 0 {
		// Try AgentError first
		var agentErr AgentError
		if err := json.Unmarshal(aux.Error, &agentErr); err == nil {
			s.Error = &agentErr
			return nil
		}

		// Try RuntimeError
		var runtimeErr RuntimeError
		if err := json.Unmarshal(aux.Error, &runtimeErr); err == nil {
			s.Error = &runtimeErr
			return nil
		}
	}

	return nil
}
