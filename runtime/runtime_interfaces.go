package runtime

import "github.com/tcfwbper/spectra/entities"

// SessionInitializerInterface is the interface Runtime uses for session initialization.
type SessionInitializerInterface interface {
	Initialize(workflowName string, projectRoot string, terminationNotifier chan<- struct{}) (SessionForInitializer, error)
}

// SessionFinalizerInterface is the interface Runtime uses for session finalization.
type SessionFinalizerInterface interface {
	Finalize(session SessionForFinalizer)
}

// RuntimeSocketManagerInterface is the interface Runtime uses for socket management.
type RuntimeSocketManagerInterface interface {
	Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error)
	DeleteSocket() error
}

// MessageHandler type alias for the socket manager callback.
type MessageHandler func(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse

// MessageRouterInterface is the interface Runtime uses for message routing.
type MessageRouterInterface interface {
	RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse
}
