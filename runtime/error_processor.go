package runtime

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// SessionForError defines the interface that ErrorProcessor needs from Session.
type SessionForError interface {
	GetStatusSafe() string
	GetCurrentStateSafe() string
	GetID() string
	GetWorkflowName() string
	GetSessionDataSafe(key string) (any, bool)
	Fail(err error, terminationNotifier chan<- struct{}) error
}

// WorkflowDefinitionLoaderForError defines the interface for loading workflow definitions.
type WorkflowDefinitionLoaderForError interface {
	Load(workflowName string) (*storage.WorkflowDefinition, error)
}

// ErrorProcessor handles RuntimeMessage with type="error".
type ErrorProcessor struct {
	session             SessionForError
	workflowLoader      WorkflowDefinitionLoaderForError
	terminationNotifier chan<- struct{}
}

// NewErrorProcessor creates a new ErrorProcessor instance.
func NewErrorProcessor(
	session SessionForError,
	workflowLoader WorkflowDefinitionLoaderForError,
	terminationNotifier chan<- struct{},
) (*ErrorProcessor, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}
	if workflowLoader == nil {
		return nil, fmt.Errorf("workflowLoader cannot be nil")
	}
	if terminationNotifier == nil {
		return nil, fmt.Errorf("terminationNotifier cannot be nil")
	}

	return &ErrorProcessor{
		session:             session,
		workflowLoader:      workflowLoader,
		terminationNotifier: terminationNotifier,
	}, nil
}

// ProcessError processes an error message from an agent or human node.
func (ep *ErrorProcessor) ProcessError(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	// Step 3: Validate session status
	status := ep.session.GetStatusSafe()
	if status != "initializing" && status != "running" {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("session terminated: status is '%s'", status),
		}
	}

	// Step 4: Load workflow definition
	workflowName := ep.session.GetWorkflowName()
	workflowDef, err := ep.workflowLoader.Load(workflowName)
	if err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to load workflow definition: %v", err),
		}
	}

	// Step 5: Get current node definition
	currentState := ep.session.GetCurrentStateSafe()
	var currentNode *storage.Node
	for i := range workflowDef.Nodes {
		if workflowDef.Nodes[i].Name == currentState {
			currentNode = &workflowDef.Nodes[i]
			break
		}
	}
	if currentNode == nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("current node '%s' not found in workflow definition", currentState),
		}
	}

	// Step 6: Derive agentRole
	agentRole := ""
	if currentNode.Type == "agent" {
		agentRole = currentNode.AgentRole
	}

	// Step 7: Validate claudeSessionID
	if currentNode.Type == "agent" {
		key := fmt.Sprintf("%s.ClaudeSessionID", currentState)
		storedValue, ok := ep.session.GetSessionDataSafe(key)
		if !ok {
			return entities.RuntimeResponse{
				Status:  "error",
				Message: fmt.Sprintf("claude session ID not found for node '%s'", currentState),
			}
		}
		storedUUID, ok := storedValue.(string)
		if !ok {
			return entities.RuntimeResponse{
				Status:  "error",
				Message: fmt.Sprintf("claude session ID not found for node '%s'", currentState),
			}
		}
		if storedUUID != message.ClaudeSessionID {
			return entities.RuntimeResponse{
				Status:  "error",
				Message: fmt.Sprintf("claude session ID mismatch: expected %s but got %s", storedUUID, message.ClaudeSessionID),
			}
		}
	} else if currentNode.Type == "human" {
		if message.ClaudeSessionID != "" {
			return entities.RuntimeResponse{
				Status:  "error",
				Message: "invalid claude session ID for human node: must be empty",
			}
		}
	}

	// Parse error payload
	var errorPayload entities.ErrorPayload
	if err := json.Unmarshal(message.Payload, &errorPayload); err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("invalid error payload: %v", err),
		}
	}

	// Validate message field
	if errorPayload.Message == "" {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: "invalid error payload: missing required field",
		}
	}

	// Step 8: Construct AgentError
	sessionID, err := uuid.Parse(ep.session.GetID())
	if err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("invalid session UUID: %v", err),
		}
	}

	// Handle nil or empty detail
	detail := errorPayload.Detail
	if len(detail) == 0 || string(detail) == "null" {
		detail = json.RawMessage(`{}`)
	}

	agentError, err := entities.NewAgentError(
		agentRole,
		errorPayload.Message,
		detail,
		sessionID,
		currentState,
		time.Now().Unix(),
	)
	if err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to construct agent error: %v", err),
		}
	}

	// Step 9: Call Session.Fail
	if err := ep.session.Fail(agentError, ep.terminationNotifier); err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to record error: %v", err),
		}
	}

	// Step 11: Return success response
	agentRoleStr := agentRole
	if agentRoleStr == "" {
		agentRoleStr = `""`
	}
	return entities.RuntimeResponse{
		Status:  "success",
		Message: fmt.Sprintf("error recorded | session=%s | failingState=%s | agentRole=%s | error=%s", sessionID, currentState, agentRoleStr, errorPayload.Message),
	}
}
