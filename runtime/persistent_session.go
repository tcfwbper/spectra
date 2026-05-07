package runtime

import (
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/logger"
)

// Session defines the interface for the in-memory session entity consumed by
// PersistentSession. It declares all thread-safe mutation and getter methods.
type Session interface {
	Run() error
	Done(notifier chan<- struct{}) error
	Fail(err error, notifier chan<- struct{}) error
	UpdateCurrentStateSafe(newState string) error
	UpdateSessionDataSafe(key string, value any) error
	UpdateEventHistorySafe(event entities.Event) error
	GetStatusSafe() string
	GetCurrentStateSafe() string
	GetErrorSafe() error
	GetMetadataSnapshotSafe() session.SessionMetadata
	GetSessionDataSafe(key string) (any, bool)
}

// SessionMetadataStore defines the interface for persisting session metadata.
type SessionMetadataStore interface {
	Write(meta session.SessionMetadata) error
}

// EventStore defines the interface for persisting events.
type EventStore interface {
	Append(event *entities.Event) error
}

// PersistentSession wraps the Session entity and automatically triggers
// persistence after every successful in-memory mutation. Persistence failures
// are logged but never propagate as errors to the caller.
type PersistentSession struct {
	ID           string
	WorkflowName string

	session       Session
	metadataStore SessionMetadataStore
	eventStore    EventStore
	logger        logger.Logger
}

// NewPersistentSession validates all inputs are non-nil and returns a
// PersistentSession instance. Panics if any dependency is nil.
func NewPersistentSession(sess Session, metadataStore SessionMetadataStore, eventStore EventStore, log logger.Logger) *PersistentSession {
	if sess == nil {
		panic("NewPersistentSession: session must not be nil")
	}
	if metadataStore == nil {
		panic("NewPersistentSession: metadataStore must not be nil")
	}
	if eventStore == nil {
		panic("NewPersistentSession: eventStore must not be nil")
	}
	if log == nil {
		panic("NewPersistentSession: logger must not be nil")
	}

	meta := sess.GetMetadataSnapshotSafe()
	return &PersistentSession{
		ID:            meta.ID,
		WorkflowName:  meta.WorkflowName,
		session:       sess,
		metadataStore: metadataStore,
		eventStore:    eventStore,
		logger:        log,
	}
}

// Run transitions the session to "running" and persists metadata.
func (ps *PersistentSession) Run() error {
	if err := ps.session.Run(); err != nil {
		return err
	}
	if err := ps.metadataStore.Write(ps.session.GetMetadataSnapshotSafe()); err != nil {
		ps.logger.Error("failed to persist session metadata after Run", "error", err, "sessionID", ps.ID)
	}
	return nil
}

// Done transitions the session to "completed" and persists metadata.
func (ps *PersistentSession) Done(terminationNotifier chan<- struct{}) error {
	if err := ps.session.Done(terminationNotifier); err != nil {
		return err
	}
	if err := ps.metadataStore.Write(ps.session.GetMetadataSnapshotSafe()); err != nil {
		ps.logger.Error("failed to persist session metadata after Done", "error", err, "sessionID", ps.ID)
	}
	return nil
}

// Fail transitions the session to "failed" and persists metadata.
func (ps *PersistentSession) Fail(err error, terminationNotifier chan<- struct{}) error {
	if failErr := ps.session.Fail(err, terminationNotifier); failErr != nil {
		return failErr
	}
	if writeErr := ps.metadataStore.Write(ps.session.GetMetadataSnapshotSafe()); writeErr != nil {
		ps.logger.Error("failed to persist session metadata after Fail", "error", writeErr, "sessionID", ps.ID)
	}
	return nil
}

// UpdateCurrentStateSafe updates the current state and persists metadata.
func (ps *PersistentSession) UpdateCurrentStateSafe(newState string) error {
	if err := ps.session.UpdateCurrentStateSafe(newState); err != nil {
		return err
	}
	if err := ps.metadataStore.Write(ps.session.GetMetadataSnapshotSafe()); err != nil {
		ps.logger.Error("failed to persist session metadata after UpdateCurrentStateSafe", "error", err, "sessionID", ps.ID)
	}
	return nil
}

// UpdateSessionDataSafe updates a session data key-value pair and persists metadata.
func (ps *PersistentSession) UpdateSessionDataSafe(key string, value any) error {
	if err := ps.session.UpdateSessionDataSafe(key, value); err != nil {
		return err
	}
	if err := ps.metadataStore.Write(ps.session.GetMetadataSnapshotSafe()); err != nil {
		ps.logger.Error("failed to persist session metadata after UpdateSessionDataSafe", "error", err, "sessionID", ps.ID, "key", key)
	}
	return nil
}

// UpdateEventHistorySafe appends an event to the history, persists the event,
// and persists updated metadata. Persistence failures are logged independently.
func (ps *PersistentSession) UpdateEventHistorySafe(event entities.Event) error {
	if err := ps.session.UpdateEventHistorySafe(event); err != nil {
		return err
	}
	if err := ps.eventStore.Append(&event); err != nil {
		ps.logger.Error("failed to persist event", "error", err, "sessionID", ps.ID, "eventID", event.ID())
	}
	if err := ps.metadataStore.Write(ps.session.GetMetadataSnapshotSafe()); err != nil {
		ps.logger.Error("failed to persist session metadata after UpdateEventHistorySafe", "error", err, "sessionID", ps.ID)
	}
	return nil
}

// GetStatusSafe returns the current session status.
func (ps *PersistentSession) GetStatusSafe() string {
	return ps.session.GetStatusSafe()
}

// GetCurrentStateSafe returns the current state.
func (ps *PersistentSession) GetCurrentStateSafe() string {
	return ps.session.GetCurrentStateSafe()
}

// GetErrorSafe returns the session error.
func (ps *PersistentSession) GetErrorSafe() error {
	return ps.session.GetErrorSafe()
}

// GetMetadataSnapshotSafe returns a snapshot of the session metadata.
func (ps *PersistentSession) GetMetadataSnapshotSafe() session.SessionMetadata {
	return ps.session.GetMetadataSnapshotSafe()
}

// GetSessionDataSafe retrieves a value from the session data.
func (ps *PersistentSession) GetSessionDataSafe(key string) (any, bool) {
	return ps.session.GetSessionDataSafe(key)
}
