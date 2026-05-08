package runtime

import (
	"encoding/json"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
)

// SessionFinalizer logs the final status of a session and returns an
// appropriate exit code. It is invoked by Runtime after socket cleanup.
type SessionFinalizer struct {
	logger logger.Logger
}

// NewSessionFinalizer validates the logger is non-nil and returns a
// SessionFinalizer instance. Panics if logger is nil.
func NewSessionFinalizer(log logger.Logger) *SessionFinalizer {
	if log == nil {
		panic("NewSessionFinalizer: logger must not be nil")
	}
	return &SessionFinalizer{logger: log}
}

// Finalize reads the session's final status, logs accordingly, and returns
// an exit code (0 for completed, 1 otherwise).
func (sf *SessionFinalizer) Finalize(session *PersistentSession) int {
	if session == nil {
		sf.logger.Error("SessionFinalizer called with nil session")
		return 1
	}

	status := session.GetStatusSafe()

	switch status {
	case "completed":
		sf.logger.Info("session completed", "sessionID", session.ID, "workflow", session.WorkflowName)
		return 0

	case "failed":
		sf.logFailure(session)
		return 1

	default:
		// Non-terminal status (initializing, running)
		sf.logger.Warn("session terminated with non-terminal status",
			"sessionID", session.ID, "workflow", session.WorkflowName, "status", status)
		// If there is an error, also log the failure details.
		sessionErr := session.GetErrorSafe()
		if sessionErr != nil {
			sf.logFailure(session)
		}
		return 1
	}
}

// logFailure logs error details based on the error type.
func (sf *SessionFinalizer) logFailure(session *PersistentSession) {
	sessionErr := session.GetErrorSafe()
	if sessionErr == nil {
		sf.logger.Error("session failed",
			"sessionID", session.ID, "workflow", session.WorkflowName, "error", "unknown error")
		return
	}

	switch e := sessionErr.(type) {
	case *entities.AgentError:
		args := []any{
			"sessionID", session.ID,
			"workflow", session.WorkflowName,
			"error", e.Message(),
			"agent", e.AgentRole(),
			"state", e.FailingState(),
		}
		if detailStr := sf.formatDetail(e.Detail()); detailStr != "" {
			args = append(args, "detail", detailStr)
		}
		sf.logger.Error("session failed", args...)

	case *entities.RuntimeError:
		args := []any{
			"sessionID", session.ID,
			"workflow", session.WorkflowName,
			"error", e.Message(),
			"issuer", e.Issuer(),
			"state", e.FailingState(),
		}
		if detailStr := sf.formatDetail(e.Detail()); detailStr != "" {
			args = append(args, "detail", detailStr)
		}
		sf.logger.Error("session failed", args...)

	default:
		sf.logger.Error("session failed",
			"sessionID", session.ID, "workflow", session.WorkflowName, "error", sessionErr.Error())
	}
}

// formatDetail serializes detail as compact JSON. Returns empty string if
// detail is nil or an empty JSON object ("{}"). Returns a fallback string
// if serialization fails.
func (sf *SessionFinalizer) formatDetail(detail json.RawMessage) string {
	if len(detail) == 0 {
		return ""
	}

	// Compact the JSON.
	var buf json.RawMessage
	if err := json.Unmarshal(detail, &buf); err != nil {
		return "<failed to serialize detail>"
	}

	compacted, err := compactJSON(detail)
	if err != nil {
		return "<failed to serialize detail>"
	}

	// Omit if empty object.
	if compacted == "{}" {
		return ""
	}

	return compacted
}

// compactJSON removes unnecessary whitespace from JSON bytes.
func compactJSON(data json.RawMessage) (string, error) {
	var buf []byte
	buffer := json.RawMessage{}
	if err := json.Unmarshal(data, &buffer); err != nil {
		return "", err
	}
	buf, err := json.Marshal(buffer)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
