package runtime

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
)

// EventProcessorWorkflowDef defines the read-only interface for workflow
// definitions consumed by EventProcessor. It extends the TransitionEvaluator's
// WorkflowDef with Nodes() for node lookup.
type EventProcessorWorkflowDef interface {
	Nodes() []*components.Node
	Transitions() []*components.Transition
	ExitTransitions() []*components.ExitTransition
}

// TransitionToNodeExecutor defines the interface for TransitionToNode consumed
// by EventProcessor.
type TransitionToNodeExecutor interface {
	Execute(targetNodeName, message string) error
}

// EventProcessor handles RuntimeMessage with type="event". It validates session
// status, validates the Claude session ID, constructs an Event entity, records
// it to the session's event history, evaluates the transition, and invokes
// TransitionToNode for dispatch.
type EventProcessor struct {
	ps                  *PersistentSession
	wfDef               EventProcessorWorkflowDef
	transitionToNode    TransitionToNodeExecutor
	terminationNotifier chan<- struct{}
}

// NewEventProcessor constructs an EventProcessor with the given dependencies.
func NewEventProcessor(ps *PersistentSession, wfDef EventProcessorWorkflowDef, transitionToNode TransitionToNodeExecutor, terminationNotifier chan<- struct{}) *EventProcessor {
	return &EventProcessor{
		ps:                  ps,
		wfDef:               wfDef,
		transitionToNode:    transitionToNode,
		terminationNotifier: terminationNotifier,
	}
}

// ProcessEvent validates the session state, validates the Claude session ID,
// constructs an Event, records it, evaluates the transition, and dispatches.
func (ep *EventProcessor) ProcessEvent(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse {
	// Step 1: Validate session status (must be "running").
	status := ep.ps.GetStatusSafe()
	if status != "running" {
		return entities.ErrorResponse(fmt.Sprintf("session not ready: status is '%s'", status))
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

	// Step 5: Parse event payload.
	var payload struct {
		EventType string          `json:"eventType"`
		Message   string          `json:"message"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		return entities.ErrorResponse(fmt.Sprintf("invalid event payload: %s", err.Error()))
	}
	if payload.EventType == "" {
		return entities.ErrorResponse("invalid event payload: missing eventType")
	}

	// Step 6: Construct Event entity.
	eventID, err := generateUUID()
	if err != nil {
		return entities.ErrorResponse(fmt.Sprintf("invalid event payload: failed to generate event ID: %s", err.Error()))
	}
	now := time.Now().Unix()

	event, err := entities.NewEvent(
		eventID,
		payload.EventType,
		payload.Message,
		payload.Payload,
		currentNodeName,
		now,
		sessionUUID,
	)
	if err != nil {
		return entities.ErrorResponse(fmt.Sprintf("invalid event payload: %s", err.Error()))
	}

	// Step 7: Record event to session.
	if err := ep.ps.UpdateEventHistorySafe(*event); err != nil {
		// Construct RuntimeError and Fail.
		rtErr, rtErrBuildErr := entities.NewRuntimeError(
			"EventProcessor",
			"failed to record event",
			nil,
			now,
			sessionUUID,
			currentNodeName,
		)
		if rtErrBuildErr == nil {
			ep.ps.Fail(rtErr, ep.terminationNotifier)
		}
		return entities.ErrorResponse(fmt.Sprintf("failed to record event: %s", err.Error()))
	}

	// Step 8: Evaluate transition.
	transition, isExit := EvaluateTransition(ep.wfDef, currentNodeName, payload.EventType)
	if transition == nil {
		return entities.ErrorResponse(fmt.Sprintf("no transition found for event '%s' from node '%s'", payload.EventType, currentNodeName))
	}

	// Step 9: Execute transition.
	if err := ep.transitionToNode.Execute(transition.ToNode(), payload.Message); err != nil {
		// Construct RuntimeError and Fail.
		rtErr, rtErrBuildErr := entities.NewRuntimeError(
			"EventProcessor",
			"transition failed",
			nil,
			time.Now().Unix(),
			sessionUUID,
			currentNodeName,
		)
		if rtErrBuildErr == nil {
			ep.ps.Fail(rtErr, ep.terminationNotifier)
		}
		return entities.ErrorResponse(fmt.Sprintf("transition failed: %s", err.Error()))
	}

	// Step 10: Handle exit transition.
	if isExit {
		if err := ep.ps.Done(ep.terminationNotifier); err != nil {
			// Construct RuntimeError and Fail.
			rtErr, rtErrBuildErr := entities.NewRuntimeError(
				"EventProcessor",
				"failed to complete session",
				nil,
				time.Now().Unix(),
				sessionUUID,
				transition.ToNode(),
			)
			if rtErrBuildErr == nil {
				ep.ps.Fail(rtErr, ep.terminationNotifier)
			}
			return entities.ErrorResponse(fmt.Sprintf("failed to complete session: %s", err.Error()))
		}
		return entities.SuccessResponse(fmt.Sprintf(
			"event '%s' processed successfully | session=%s | currentState=%s | sessionStatus=completed",
			payload.EventType, sessionUUID, ep.ps.GetCurrentStateSafe(),
		))
	}

	// Step 11: Return success for regular transition.
	return entities.SuccessResponse(fmt.Sprintf(
		"event '%s' processed successfully | session=%s | currentState=%s | sessionStatus=running",
		payload.EventType, sessionUUID, ep.ps.GetCurrentStateSafe(),
	))
}

// generateUUID generates a UUID v4 string using crypto/rand.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Set version (4) and variant (RFC 4122).
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
