package runtime

import (
	"encoding/json"
	"fmt"
	"os"
)

// SessionFinalizer orchestrates the cleanup flow when a session reaches a terminal state.
type SessionFinalizer struct {
	logger Logger
}

// Logger defines the interface for logging operations.
type Logger interface {
	Warning(msg string)
}

// SessionForFinalizer defines the Session interface required by SessionFinalizer.
type SessionForFinalizer interface {
	GetID() string
	GetStatusSafe() string
	GetWorkflowName() string
	GetErrorSafe() error
}

// AgentError interface for agent-reported errors.
type AgentError interface {
	error
	GetAgentRole() string
	GetFailingState() string
	GetDetail() map[string]any
	IsAgentError() bool
}

// RuntimeError interface for runtime component errors.
type RuntimeError interface {
	error
	GetIssuer() string
	GetFailingState() string
	GetDetail() map[string]any
	IsRuntimeError() bool
}

// NewSessionFinalizer creates a new SessionFinalizer.
func NewSessionFinalizer(logger Logger) (*SessionFinalizer, error) {
	return &SessionFinalizer{
		logger: logger,
	}, nil
}

// Finalize performs cleanup operations for a terminated session.
func (sf *SessionFinalizer) Finalize(session SessionForFinalizer) {
	status := session.GetStatusSafe()

	// Validate terminal status
	if status != "completed" && status != "failed" {
		sf.logger.Warning(fmt.Sprintf("SessionFinalizer called with non-terminal session status '%s'. This may indicate a programming error or signal interruption.", status))
	}

	// Print status to appropriate output stream
	switch status {
	case "completed":
		sf.printCompletedStatus(session)
	case "failed":
		sf.printFailedStatus(session)
	default:
		// Non-terminal status - print to stderr with "terminated with status" format
		_, _ = fmt.Fprintf(os.Stderr, "Session %s terminated with status '%s'. Workflow: %s\n",
			session.GetID(), status, session.GetWorkflowName())
	}
}

func (sf *SessionFinalizer) printCompletedStatus(session SessionForFinalizer) {
	_, _ = fmt.Fprintf(os.Stdout, "Session %s completed successfully. Workflow: %s\n",
		session.GetID(), session.GetWorkflowName())
}

func (sf *SessionFinalizer) printFailedStatus(session SessionForFinalizer) {
	sessionID := session.GetID()
	workflowName := session.GetWorkflowName()
	err := session.GetErrorSafe()

	_, _ = fmt.Fprintf(os.Stderr, "Session %s failed. Workflow: %s\n", sessionID, workflowName)

	if err == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: <unknown error>\n")
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())

	// Check if error is AgentError
	if agentErr, ok := err.(AgentError); ok && agentErr.IsAgentError() {
		_, _ = fmt.Fprintf(os.Stderr, "Agent: %s\n", agentErr.GetAgentRole())
		_, _ = fmt.Fprintf(os.Stderr, "State: %s\n", agentErr.GetFailingState())
		sf.printDetail(agentErr.GetDetail())
		return
	}

	// Check if error is RuntimeError
	if runtimeErr, ok := err.(RuntimeError); ok && runtimeErr.IsRuntimeError() {
		_, _ = fmt.Fprintf(os.Stderr, "Issuer: %s\n", runtimeErr.GetIssuer())
		_, _ = fmt.Fprintf(os.Stderr, "State: %s\n", runtimeErr.GetFailingState())
		sf.printDetail(runtimeErr.GetDetail())
		return
	}

	// Unknown error type - already printed error message above
}

func (sf *SessionFinalizer) printDetail(detail map[string]any) {
	if len(detail) == 0 {
		return
	}

	detailJSON, err := json.Marshal(detail)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Detail: <failed to serialize detail>\n")
		sf.logger.Warning(fmt.Sprintf("Failed to serialize error detail: %v", err))
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, "Detail: %s\n", string(detailJSON))
}
