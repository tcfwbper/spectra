package runtime

import (
	"fmt"
	"runtime/debug"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// SessionForRouter defines the interface that MessageRouter needs from Session.
type SessionForRouter interface {
	GetID() string
	GetCurrentStateSafe() string
	Fail(err error, terminationNotifier chan<- struct{}) error
}

// EventProcessorInterface defines the interface for processing events.
type EventProcessorInterface interface {
	ProcessEvent(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse
}

// ErrorProcessorInterface defines the interface for processing errors.
type ErrorProcessorInterface interface {
	ProcessError(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse
}

// MessageRouter is the concrete implementation of the MessageHandler callback interface.
type MessageRouter struct {
	session             SessionForRouter
	eventProcessor      EventProcessorInterface
	errorProcessor      ErrorProcessorInterface
	terminationNotifier chan<- struct{}
}

// NewMessageRouter creates a new MessageRouter instance.
func NewMessageRouter(
	sess SessionForRouter,
	eventProcessor EventProcessorInterface,
	errorProcessor ErrorProcessorInterface,
	terminationNotifier chan<- struct{},
) (*MessageRouter, error) {
	if sess == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}
	if eventProcessor == nil {
		return nil, fmt.Errorf("eventProcessor cannot be nil")
	}
	if errorProcessor == nil {
		return nil, fmt.Errorf("errorProcessor cannot be nil")
	}
	if terminationNotifier == nil {
		return nil, fmt.Errorf("terminationNotifier cannot be nil")
	}

	return &MessageRouter{
		session:             sess,
		eventProcessor:      eventProcessor,
		errorProcessor:      errorProcessor,
		terminationNotifier: terminationNotifier,
	}, nil
}

// RouteMessage implements the MessageHandler interface.
func (mr *MessageRouter) RouteMessage(sessionUUID string, message entities.RuntimeMessage) (resp entities.RuntimeResponse) {
	// Step 5: Panic recovery
	defer func() {
		if r := recover(); r != nil {
			// Log the panic with stack trace
			stackTrace := debug.Stack()
			fmt.Printf("PANIC in MessageRouter: %v\n%s\n", r, stackTrace)

			// Construct RuntimeError (using session.RuntimeError)
			runtimeError := &session.RuntimeError{
				Issuer:  "MessageRouter",
				Message: "panic during message processing",
			}

			// Call Session.Fail (best effort, ignore any panics here)
			func() {
				defer func() {
					_ = recover()
				}()
				_ = mr.session.Fail(runtimeError, mr.terminationNotifier)
			}()

			// Set the response
			resp = entities.RuntimeResponse{
				Status:  "error",
				Message: "internal server error",
			}
		}
	}()

	// Step 7: Examine message type
	switch message.Type {
	case "event":
		// Step 8: Route to EventProcessor
		return mr.eventProcessor.ProcessEvent(sessionUUID, message)
	case "error":
		// Step 9: Route to ErrorProcessor
		return mr.errorProcessor.ProcessError(sessionUUID, message)
	default:
		// Step 10: Unknown message type
		return entities.RuntimeResponse{
			Status:  "error",
			Message: fmt.Sprintf("unknown message type '%s'", message.Type),
		}
	}
}
