package runtime

import (
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// SessionForInitializer is the interface Runtime receives from SessionInitializer.
type SessionForInitializer interface {
	Run(terminationNotifier chan<- struct{}) error
	Done(terminationNotifier chan<- struct{}) error
	Fail(err error, terminationNotifier chan<- struct{}) error
	GetStatusSafe() string
	GetCurrentStateSafe() string
	GetID() string
	GetWorkflowName() string
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetEventHistory() []session.Event
	GetSessionData() map[string]any
	GetErrorSafe() error
}

// SessionInitializerInterface is the interface Runtime uses for session initialization.
type SessionInitializerInterface interface {
	Initialize(workflowName string, terminationNotifier chan<- struct{}) (SessionForInitializer, error)
}

// SessionFinalizerInterface is the interface Runtime uses for session finalization.
type SessionFinalizerInterface interface {
	Finalize(session SessionForFinalizer)
}

// RuntimeSocketManagerInterface is the interface Runtime uses for socket management.
// CreateSocket() is part of this interface so its error can be injected in tests.
type RuntimeSocketManagerInterface interface {
	CreateSocket() error
	Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error)
	DeleteSocket() error
}

// MessageHandler type alias for the socket manager callback.
type MessageHandler func(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse

// MessageRouterInterface is the interface Runtime uses for message routing.
type MessageRouterInterface interface {
	RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse
}
