package session

// SessionMetadata represents the persistable subset of a Session's state,
// excluding the event history and runtime-only fields.
type SessionMetadata struct {
	ID           string         `json:"id"`
	WorkflowName string         `json:"workflowName"`
	Pid          int            `json:"pid"`
	Status       string         `json:"status"`
	CreatedAt    int64          `json:"createdAt"`
	UpdatedAt    int64          `json:"updatedAt"`
	CurrentState string         `json:"currentState"`
	SessionData  map[string]any `json:"sessionData"`
	Error        error          `json:"error,omitempty"`
}
