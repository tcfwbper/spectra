package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
	entitysession "github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/storage"
)

// WorkflowLoader defines the interface for loading workflow definitions.
type WorkflowLoader interface {
	Load(workflowName string) (*components.WorkflowDefinition, error)
}

// SessionDirManager defines the interface for creating session directories.
type SessionDirManager interface {
	CreateSessionDirectory(projectRoot, sessionUUID string) error
}

// InitResult holds the result of session initialization.
type InitResult struct {
	PersistentSession  *PersistentSession
	WorkflowDefinition *components.WorkflowDefinition
	Error              error
}

// SessionFactory defines the interface for constructing a Session entity.
// In production this calls entitysession.NewSession; tests can replace it.
type SessionFactory func(id, workflowName, entryNode string, createdAt int64) (Session, error)

// ContextFactory creates a context and cancel function for the initialization flow.
// In production this returns context.WithTimeout(context.Background(), 30*time.Second).
type ContextFactory func() (context.Context, context.CancelFunc)

// SessionInitializer orchestrates the initialization flow for creating a new session.
type SessionInitializer struct {
	projectRoot    string
	loader         WorkflowLoader
	dirMgr         SessionDirManager
	logger         logger.Logger
	uuidGen        UUIDGenerator
	nowFunc        func() int64
	sessionFactory SessionFactory
	contextFactory ContextFactory
}

// NewSessionInitializer validates dependencies and returns a SessionInitializer.
func NewSessionInitializer(projectRoot string, loader WorkflowLoader, dirMgr SessionDirManager, log logger.Logger) *SessionInitializer {
	return &SessionInitializer{
		projectRoot: projectRoot,
		loader:      loader,
		dirMgr:      dirMgr,
		logger:      log,
		uuidGen:     &defaultUUIDGenerator{},
		nowFunc:     func() int64 { return time.Now().Unix() },
		sessionFactory: func(id, workflowName, entryNode string, createdAt int64) (Session, error) {
			sess, err := entitysession.NewSession(id, workflowName, entryNode, createdAt)
			if err != nil {
				return nil, err
			}
			return &sessionAdapter{sess: sess}, nil
		},
		contextFactory: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), 30*time.Second)
		},
	}
}

// Initialize performs the session initialization sequence:
// validates termination notifier, generates UUID, loads workflow,
// creates directory, constructs session, and transitions to running.
func (si *SessionInitializer) Initialize(workflowName string, terminationNotifier chan<- struct{}) InitResult {
	// Step 2: Validate terminationNotifier capacity >= 2.
	chanCap := cap(terminationNotifier)
	if chanCap < 2 {
		return InitResult{
			Error: fmt.Errorf("terminationNotifier channel must have buffer capacity >= 2, got %d", chanCap),
		}
	}

	// Step 3: Create context with 30-second timeout.
	ctx, cancel := si.contextFactory()
	defer cancel()

	// Step 4: Generate session UUID.
	sessionUUID, err := si.uuidGen.Generate()
	if err != nil {
		return InitResult{
			Error: fmt.Errorf("failed to generate session UUID: %v", err),
		}
	}

	// Step 5: Log session UUID immediately.
	si.logger.Info("session created", "sessionID", sessionUUID)

	// Step 6: Check context.
	if ctx.Err() != nil {
		return InitResult{
			Error: fmt.Errorf("session initialization timed out"),
		}
	}

	// Step 7: Load workflow definition.
	wfDef, err := si.loader.Load(workflowName)
	if err != nil {
		return InitResult{
			Error: fmt.Errorf("failed to load workflow definition: %s", err.Error()),
		}
	}

	// Step 9: Check context.
	if ctx.Err() != nil {
		return InitResult{
			Error: fmt.Errorf("session initialization timed out"),
		}
	}

	// Step 10: Create session directory.
	if err := si.dirMgr.CreateSessionDirectory(si.projectRoot, sessionUUID); err != nil {
		return InitResult{
			Error: fmt.Errorf("failed to create session directory: %s", err.Error()),
		}
	}

	// Step 12: Check context.
	if ctx.Err() != nil {
		return InitResult{
			Error: fmt.Errorf("session initialization timed out"),
		}
	}

	// Step 13: Construct Session entity.
	now := si.nowFunc()
	sess, err := si.sessionFactory(sessionUUID, workflowName, wfDef.EntryNode(), now)
	if err != nil {
		return InitResult{
			Error: fmt.Errorf("failed to construct session: %s", err.Error()),
		}
	}

	// Step 15: Construct SessionMetadataStore.
	metadataStore := storage.NewSessionMetadataStore(si.projectRoot, sessionUUID)

	// Step 16: Construct EventStore.
	eventStore := storage.NewEventStore(si.projectRoot, sessionUUID, si.logger)

	// Step 17: Construct PersistentSession.
	ps := NewPersistentSession(sess, metadataStore, eventStore, si.logger)

	// Step 18: Check context.
	if ctx.Err() != nil {
		rtErr := si.buildTimeoutError(sessionUUID, wfDef.EntryNode(), now)
		_ = ps.Fail(rtErr, terminationNotifier)
		return InitResult{
			PersistentSession: ps,
			Error:             fmt.Errorf("session initialization timed out"),
		}
	}

	// Step 19: Transition to running.
	if err := ps.Run(); err != nil {
		rtErr := si.buildRuntimeError(sessionUUID, wfDef.EntryNode(), now,
			fmt.Sprintf("failed to transition session to running: %s", err.Error()))
		_ = ps.Fail(rtErr, terminationNotifier)
		return InitResult{
			PersistentSession: ps,
			Error:             fmt.Errorf("failed to transition session to running: %s", err.Error()),
		}
	}

	// Step 21: Check context after Run succeeded.
	if ctx.Err() != nil {
		rtErr := si.buildTimeoutError(sessionUUID, wfDef.EntryNode(), now)
		_ = ps.Fail(rtErr, terminationNotifier)
		return InitResult{
			PersistentSession: ps,
			Error:             fmt.Errorf("session initialization timed out"),
		}
	}

	// Step 22: Success.
	return InitResult{
		PersistentSession:  ps,
		WorkflowDefinition: wfDef,
	}
}

// buildTimeoutError constructs a RuntimeError for timeout scenarios.
func (si *SessionInitializer) buildTimeoutError(sessionUUID, entryNode string, now int64) *entities.RuntimeError {
	rtErr, _ := entities.NewRuntimeError(
		"SessionInitializer",
		"session initialization timed out",
		nil,
		now,
		sessionUUID,
		entryNode,
	)
	return rtErr
}

// buildRuntimeError constructs a RuntimeError with a custom message.
func (si *SessionInitializer) buildRuntimeError(sessionUUID, entryNode string, now int64, message string) *entities.RuntimeError {
	rtErr, _ := entities.NewRuntimeError(
		"SessionInitializer",
		message,
		nil,
		now,
		sessionUUID,
		entryNode,
	)
	return rtErr
}

// sessionAdapter adapts *session.Session (which uses bidirectional channels)
// to the runtime.Session interface (which uses send-only channels).
type sessionAdapter struct {
	sess *entitysession.Session
}

func (a *sessionAdapter) Run() error {
	return a.sess.Run()
}

func (a *sessionAdapter) Done(notifier chan<- struct{}) error {
	// session.Session.Done takes chan struct{} (bidirectional).
	// We create a bidirectional channel, forward the notification.
	ch := make(chan struct{}, cap(notifier)+1)
	err := a.sess.Done(ch)
	if err == nil {
		// Forward notification to the send-only channel.
		select {
		case v := <-ch:
			notifier <- v
		default:
		}
	}
	return err
}

func (a *sessionAdapter) Fail(err error, notifier chan<- struct{}) error {
	ch := make(chan struct{}, cap(notifier)+1)
	failErr := a.sess.Fail(err, ch)
	if failErr == nil {
		select {
		case v := <-ch:
			notifier <- v
		default:
		}
	}
	return failErr
}

func (a *sessionAdapter) UpdateCurrentStateSafe(newState string) error {
	return a.sess.UpdateCurrentStateSafe(newState)
}

func (a *sessionAdapter) UpdateSessionDataSafe(key string, value any) error {
	return a.sess.UpdateSessionDataSafe(key, value)
}

func (a *sessionAdapter) UpdateEventHistorySafe(event entities.Event) error {
	return a.sess.UpdateEventHistorySafe(event)
}

func (a *sessionAdapter) GetStatusSafe() string {
	return a.sess.GetStatusSafe()
}

func (a *sessionAdapter) GetCurrentStateSafe() string {
	return a.sess.GetCurrentStateSafe()
}

func (a *sessionAdapter) GetErrorSafe() error {
	return a.sess.GetErrorSafe()
}

func (a *sessionAdapter) GetMetadataSnapshotSafe() entitysession.SessionMetadata {
	return a.sess.GetMetadataSnapshotSafe()
}

func (a *sessionAdapter) GetSessionDataSafe(key string) (any, bool) {
	return a.sess.GetSessionDataSafe(key)
}
