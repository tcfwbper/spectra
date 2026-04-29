package runtime

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// SessionForEvent defines the interface that EventProcessor needs from Session.
type SessionForEvent interface {
	GetStatusSafe() string
	GetCurrentStateSafe() string
	GetID() string
	GetWorkflowName() string
	GetSessionDataSafe(key string) (any, bool)
	UpdateEventHistorySafe(event session.Event) error
	Fail(err error, terminationNotifier chan<- struct{}) error
}

// WorkflowDefinitionLoaderForEvent defines the interface for loading workflow definitions.
type WorkflowDefinitionLoaderForEvent interface {
	Load(workflowName string) (*storage.WorkflowDefinition, error)
}

// TransitionToNodeInterface defines the interface for node transition logic.
type TransitionToNodeInterface interface {
	Transition(message string, targetNodeName string, isExitTransition bool) error
}

// EventProcessor handles RuntimeMessage with type="event".
type EventProcessor struct {
	session             SessionForEvent
	workflowLoader      WorkflowDefinitionLoaderForEvent
	transitioner        TransitionToNodeInterface
	terminationNotifier chan<- struct{}
}

// NewEventProcessor creates a new EventProcessor instance.
func NewEventProcessor(
	session SessionForEvent,
	workflowLoader WorkflowDefinitionLoaderForEvent,
	transitioner TransitionToNodeInterface,
	terminationNotifier chan<- struct{},
) (*EventProcessor, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}
	if workflowLoader == nil {
		return nil, fmt.Errorf("workflowLoader cannot be nil")
	}
	if transitioner == nil {
		return nil, fmt.Errorf("transitioner cannot be nil")
	}
	if terminationNotifier == nil {
		return nil, fmt.Errorf("terminationNotifier cannot be nil")
	}

	return &EventProcessor{
		session:             session,
		workflowLoader:      workflowLoader,
		transitioner:        transitioner,
		terminationNotifier: terminationNotifier,
	}, nil
}

// ProcessEvent processes an event message from an agent or human node.
func (ep *EventProcessor) ProcessEvent(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	// Step 3: Validate session status
	status := ep.session.GetStatusSafe()
	if status != "running" {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("session not ready: status is '%s'", status),
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

	// Step 6: Validate claudeSessionID
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

	// Parse event payload
	var eventPayload entities.EventPayload
	if err := json.Unmarshal(message.Payload, &eventPayload); err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("invalid event payload: %v", err),
		}
	}

	// Validate eventType field
	if eventPayload.EventType == "" {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: "invalid event payload: missing eventType",
		}
	}

	// Step 7: Construct Event entity
	eventID := uuid.New().String()
	now := time.Now().Unix()

	// Convert payload to map[string]any
	var payloadMap map[string]any
	if len(eventPayload.Payload) > 0 {
		if err := json.Unmarshal(eventPayload.Payload, &payloadMap); err != nil {
			payloadMap = make(map[string]any)
		}
	} else {
		payloadMap = make(map[string]any)
	}

	event := session.Event{
		ID:        eventID,
		Type:      eventPayload.EventType,
		Message:   eventPayload.Message,
		Payload:   payloadMap,
		EmittedBy: currentState,
		EmittedAt: now,
		SessionID: ep.session.GetID(),
	}

	// Step 8: Write event to EventStore
	if err := ep.session.UpdateEventHistorySafe(event); err != nil {
		// Construct RuntimeError (using session.RuntimeError)
		runtimeError := &session.RuntimeError{
			Issuer:  "EventProcessor",
			Message: fmt.Sprintf("failed to record event: %v", err),
		}

		// Call Session.Fail
		_ = ep.session.Fail(runtimeError, ep.terminationNotifier)

		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to record event: %v", err),
		}
	}

	// Step 10: Evaluate transition
	transition, isExitTransition := evaluateTransition(workflowDef, currentState, eventPayload.EventType)
	if transition == nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("no transition found for event '%s' from node '%s'", eventPayload.EventType, currentState),
		}
	}

	// Step 12: Execute transition
	if err := ep.transitioner.Transition(eventPayload.Message, transition.ToNode, isExitTransition); err != nil {
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("transition failed: %v", err),
		}
	}

	// Step 14: Return success response
	newStatus := ep.session.GetStatusSafe()
	newState := ep.session.GetCurrentStateSafe()

	return entities.RuntimeResponse{
		Status:  "success",
		Message: fmt.Sprintf("event '%s' processed successfully | session=%s | currentState=%s | sessionStatus=%s", eventPayload.EventType, ep.session.GetID(), newState, newStatus),
	}
}

// evaluateTransition finds a matching transition for the given event type from the current node.
// It returns the transition and whether it's an exit transition.
func evaluateTransition(workflowDef *storage.WorkflowDefinition, currentState string, eventType string) (*storage.Transition, bool) {
	// Find matching transition
	var matchedTransition *storage.Transition
	for i := range workflowDef.Transitions {
		t := &workflowDef.Transitions[i]
		if t.FromNode == currentState && t.EventType == eventType {
			matchedTransition = t
			break
		}
	}

	if matchedTransition == nil {
		return nil, false
	}

	// Check if it's an exit transition
	for i := range workflowDef.ExitTransitions {
		et := &workflowDef.ExitTransitions[i]
		if et.FromNode == matchedTransition.FromNode &&
			et.EventType == matchedTransition.EventType &&
			et.ToNode == matchedTransition.ToNode {
			return matchedTransition, true
		}
	}

	return matchedTransition, false
}
