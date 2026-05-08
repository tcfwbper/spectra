package runtime

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
)

// MessageRouterEventProcessor defines the interface for event processing
// consumed by MessageRouter.
type MessageRouterEventProcessor interface {
	ProcessEvent(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse
}

// MessageRouterErrorProcessor defines the interface for error processing
// consumed by MessageRouter.
type MessageRouterErrorProcessor interface {
	ProcessError(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse
}

// MessageRouter dispatches RuntimeMessages to EventProcessor or ErrorProcessor
// based on message type. It implements panic recovery to prevent failures from
// crashing the runtime process.
type MessageRouter struct {
	ps                  *PersistentSession
	eventProcessor      MessageRouterEventProcessor
	errorProcessor      MessageRouterErrorProcessor
	terminationNotifier chan<- struct{}
	logger              logger.Logger
}

// NewMessageRouter constructs a MessageRouter with the given dependencies.
func NewMessageRouter(ps *PersistentSession, eventProcessor MessageRouterEventProcessor, errorProcessor MessageRouterErrorProcessor, terminationNotifier chan<- struct{}, log logger.Logger) *MessageRouter {
	return &MessageRouter{
		ps:                  ps,
		eventProcessor:      eventProcessor,
		errorProcessor:      errorProcessor,
		terminationNotifier: terminationNotifier,
		logger:              log,
	}
}

// Handle dispatches the RuntimeMessage to the appropriate processor based on
// message type. It wraps the dispatch in panic recovery.
func (mr *MessageRouter) Handle(sessionUUID string, msg *entities.RuntimeMessage) (resp *entities.RuntimeResponse) {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			mr.logger.Error("panic during message processing",
				"panic", r,
				"stack", stackTrace,
				"sessionID", mr.ps.ID,
				"currentState", mr.ps.GetCurrentStateSafe(),
			)

			// Construct RuntimeError.
			rtErr, err := entities.NewRuntimeError(
				"MessageRouter",
				"panic during message processing",
				nil,
				time.Now().Unix(),
				sessionUUID,
				mr.ps.GetCurrentStateSafe(),
			)
			if err != nil {
				// RuntimeError construction failed — log and skip Fail.
				mr.logger.Error("failed to construct RuntimeError during panic recovery",
					"error", err,
					"sessionID", mr.ps.ID,
				)
			} else {
				// Call PersistentSession.Fail.
				if failErr := mr.ps.Fail(rtErr, mr.terminationNotifier); failErr != nil {
					mr.logger.Error("failed to fail session during panic recovery",
						"error", failErr,
						"sessionID", mr.ps.ID,
					)
				}
			}

			resp = entities.ErrorResponse("internal server error")
		}
	}()

	switch msg.Type() {
	case "event":
		return mr.eventProcessor.ProcessEvent(sessionUUID, msg)
	case "error":
		return mr.errorProcessor.ProcessError(sessionUUID, msg)
	default:
		return entities.ErrorResponse(fmt.Sprintf("unknown message type '%s'", msg.Type()))
	}
}
