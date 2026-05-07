package runtime

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
)

// ErrorProcessorWorkflowDef defines the read-only interface for workflow
// definitions consumed by ErrorProcessor.
type ErrorProcessorWorkflowDef interface {
	Nodes() []*components.Node
}

// ErrorProcessor handles RuntimeMessage with type="error". It validates session
// status, validates the Claude session ID, derives the agent role from the
// current node definition, constructs an AgentError entity, and calls
// PersistentSession.Fail to transition the session to "failed" status.
type ErrorProcessor struct {
	ps                  *PersistentSession
	wfDef               ErrorProcessorWorkflowDef
	terminationNotifier chan<- struct{}
}

// NewErrorProcessor constructs an ErrorProcessor with the given dependencies.
func NewErrorProcessor(ps *PersistentSession, wfDef ErrorProcessorWorkflowDef, terminationNotifier chan<- struct{}) *ErrorProcessor {
	return &ErrorProcessor{
		ps:                  ps,
		wfDef:               wfDef,
		terminationNotifier: terminationNotifier,
	}
}

// ProcessError validates the session state, validates the Claude session ID,
// constructs an AgentError, and calls PersistentSession.Fail.
func (ep *ErrorProcessor) ProcessError(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse {
	// Step 1: Validate session status.
	status := ep.ps.GetStatusSafe()
	if status != "initializing" && status != "running" {
		return entities.ErrorResponse(fmt.Sprintf("session terminated: status is '%s'", status))
	}

	// Step 2: Retrieve current node name.
	currentNodeName := ep.ps.GetCurrentStateSafe()

	// Step 3: Find current node in workflow definition.
	var currentNode *components.Node
	for _, n := range ep.wfDef.Nodes() {
		if n.Name() == currentNodeName {
			currentNode = n
			break
		}
	}
	if currentNode == nil {
		return entities.ErrorResponse(fmt.Sprintf("current node '%s' not found in workflow definition", currentNodeName))
	}

	// Step 4: Validate Claude session ID.
	if err := ValidateClaudeSessionID(ep.ps, currentNode, msg.ClaudeSessionID()); err != nil {
		return entities.ErrorResponse(err.Error())
	}

	// Step 5: Parse error payload.
	var payload struct {
		Message string          `json:"message"`
		Detail  json.RawMessage `json:"detail"`
	}
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		return entities.ErrorResponse(fmt.Sprintf("invalid error payload: %s", err.Error()))
	}
	if payload.Message == "" {
		return entities.ErrorResponse("invalid error payload: missing required field 'message'")
	}

	// Step 6: Derive agent role from current node.
	var agentRole string
	if currentNode.Type() == "agent" {
		agentRole = currentNode.AgentRole()
	}

	// Step 7: Normalize detail — must be nil or a valid JSON object for AgentError.
	var detail json.RawMessage
	if len(payload.Detail) > 0 && string(payload.Detail) != "null" {
		trimmed := string(payload.Detail)
		if len(trimmed) > 0 && trimmed[0] == '{' {
			detail = payload.Detail
		}
		// Non-object detail values are treated as nil.
	}

	// Step 8: Construct AgentError.
	agentErr, err := entities.NewAgentError(
		agentRole,
		payload.Message,
		detail,
		time.Now().Unix(),
		sessionUUID,
		currentNodeName,
	)
	if err != nil {
		return entities.ErrorResponse(fmt.Sprintf("invalid error payload: %s", err.Error()))
	}

	// Step 9: Call PersistentSession.Fail.
	if err := ep.ps.Fail(agentErr, ep.terminationNotifier); err != nil {
		return entities.ErrorResponse(fmt.Sprintf("failed to record error: %s", err.Error()))
	}

	// Step 10: Return success response.
	return entities.SuccessResponse(fmt.Sprintf(
		"error recorded | session=%s | failingState=%s | agentRole=%s | error=%s",
		sessionUUID, currentNodeName, agentRole, payload.Message,
	))
}
