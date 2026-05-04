package runtime

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// WorkflowDefinitionLoader defines the interface for loading workflow definitions.
type WorkflowDefinitionLoader interface {
	Load(workflowName string) (*storage.WorkflowDefinition, error)
}

// SessionDirectoryManager defines the interface for managing session directories.
type SessionDirectoryManager interface {
	CreateSessionDirectory(sessionUUID string) error
}

// SessionInitializer orchestrates the complete initialization flow for creating a new session.
type SessionInitializer struct {
	projectRoot        string
	wdLoader           WorkflowDefinitionLoader
	dirManager         SessionDirectoryManager
	timeoutDuration    time.Duration
	sessionConstructor func(wfDef *storage.WorkflowDefinition, sessionUUID string) (SessionForInitializer, error)

	// Test injection points
	spectraFinderFunc       func() (string, error)
	spectraFinderBlockCh    chan struct{}
	fileAccessorFunc        func() error
	metadataStoreFunc       func() (MetadataStore, error)
	sessionRunErrorFunc     func() error
	sessionRunBlockCh       chan struct{}
	sessionRunCallbackFunc  func(chan<- struct{})
	sessionFailCallbackFunc func(error, chan<- struct{})
	callOrderTracker        callOrderRecorder
	wasSessionFailCalled    atomic.Bool
	sessionFailErrorMu      sync.Mutex
	sessionFailError        error
}

// callOrderRecorder is an interface for recording call order in tests.
type callOrderRecorder interface {
	record(name string)
}

// MetadataStore interface for session metadata persistence.
type MetadataStore interface {
	Write(metadata interface{}) error
}

// NewSessionInitializer creates a new SessionInitializer.
func NewSessionInitializer(projectRoot string, wdLoader WorkflowDefinitionLoader, dirManager SessionDirectoryManager) (*SessionInitializer, error) {
	si := &SessionInitializer{
		projectRoot:     projectRoot,
		wdLoader:        wdLoader,
		dirManager:      dirManager,
		timeoutDuration: 30 * time.Second,
	}

	// Default session constructor
	si.sessionConstructor = si.defaultSessionConstructor

	return si, nil
}

func (si *SessionInitializer) defaultSessionConstructor(wfDef *storage.WorkflowDefinition, sessionUUID string) (SessionForInitializer, error) {
	now := time.Now().Unix()
	sess := &session.Session{
		SessionMetadata: session.SessionMetadata{
			ID:           sessionUUID,
			WorkflowName: wfDef.Name,
			Status:       "initializing",
			CreatedAt:    now,
			UpdatedAt:    now,
			CurrentState: wfDef.EntryNode,
			SessionData:  make(map[string]any),
			Error:        nil,
		},
		EventHistory: []session.Event{},
	}

	return &sessionWrapper{Session: sess}, nil
}

// Initialize performs the complete initialization flow for a new session.
func (si *SessionInitializer) Initialize(workflowName string, terminationNotifier chan<- struct{}) (SessionForInitializer, error) {
	// Validate terminationNotifier buffer capacity
	if cap(terminationNotifier) < 2 {
		return nil, fmt.Errorf("terminationNotifier channel must have buffer capacity >= 2, got %d", cap(terminationNotifier))
	}

	// Validate projectRoot
	info, err := os.Stat(si.projectRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project root does not exist: %s", si.projectRoot)
		}
		return nil, fmt.Errorf("failed to stat project root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project root is not a directory: %s", si.projectRoot)
	}

	// Shared state for timeout handler
	var sessionMu sync.Mutex
	var sess SessionForInitializer
	var initCompleted bool
	var timedOutEarly atomic.Bool

	// Start timeout timer
	timer := time.AfterFunc(si.timeoutDuration, func() {
		si.handleTimeout(&sessionMu, &sess, &initCompleted, &timedOutEarly, terminationNotifier)
	})

	// Ensure timer is stopped on return
	defer timer.Stop()

	// Check for early timeout
	if timedOutEarly.Load() {
		return nil, fmt.Errorf("session initialization timeout exceeded 30 seconds before session entity was constructed")
	}

	// Generate session UUID
	sessionUUID := uuid.New().String()

	// Load workflow definition
	wfDef, err := si.wdLoader.Load(workflowName)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow definition: %w", err)
	}

	// Check for early timeout
	if timedOutEarly.Load() {
		return nil, fmt.Errorf("session initialization timeout exceeded 30 seconds before session entity was constructed")
	}

	// Create session directory
	if si.callOrderTracker != nil {
		si.callOrderTracker.record("CreateSessionDirectory")
	}
	if err := si.dirManager.CreateSessionDirectory(sessionUUID); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Check for early timeout
	if timedOutEarly.Load() {
		return nil, fmt.Errorf("session initialization timeout exceeded 30 seconds before session entity was constructed")
	}

	// Construct Session entity
	constructedSess, err := si.sessionConstructor(wfDef, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to construct session entity: %w", err)
	}

	// Store session reference for timeout handler
	sessionMu.Lock()
	sess = constructedSess
	sessionMu.Unlock()

	// Initialize storage files (FileAccessor)
	if si.callOrderTracker != nil {
		si.callOrderTracker.record("FileAccessorPrepare")
	}
	if si.fileAccessorFunc != nil {
		if err := si.fileAccessorFunc(); err != nil {
			return sess, fmt.Errorf("failed to initialize storage files: %w", err)
		}
	}

	// Persist initial session metadata
	if si.callOrderTracker != nil {
		si.callOrderTracker.record("MetadataWrite")
	}
	if si.metadataStoreFunc != nil {
		metadataStore, err := si.metadataStoreFunc()
		if err != nil {
			return sess, fmt.Errorf("failed to persist initial session metadata: %w", err)
		}
		if err := metadataStore.Write(sess); err != nil {
			return sess, fmt.Errorf("failed to persist initial session metadata: %w", err)
		}
	}

	// Transition session to running
	if si.callOrderTracker != nil {
		si.callOrderTracker.record("SessionRun")
	}

	// Check for session run block (test injection)
	if si.sessionRunBlockCh != nil {
		<-si.sessionRunBlockCh
	}

	// Call session run callback if set (test injection)
	if si.sessionRunCallbackFunc != nil {
		si.sessionRunCallbackFunc(terminationNotifier)
	}

	// Check for injected error
	var runErr error
	if si.sessionRunErrorFunc != nil {
		runErr = si.sessionRunErrorFunc()
	} else {
		runErr = sess.Run(terminationNotifier)
	}

	if runErr != nil {
		// Create RuntimeError and call Session.Fail
		rtErr := fmt.Errorf("failed to transition session to running status: %w", runErr)
		si.wasSessionFailCalled.Store(true)
		si.sessionFailErrorMu.Lock()
		si.sessionFailError = rtErr
		si.sessionFailErrorMu.Unlock()

		if si.sessionFailCallbackFunc != nil {
			si.sessionFailCallbackFunc(rtErr, terminationNotifier)
		} else {
			_ = sess.Fail(rtErr, terminationNotifier)
		}

		return sess, fmt.Errorf("failed to transition session to running: %w", runErr)
	}

	// Mark initialization as completed
	sessionMu.Lock()
	initCompleted = true
	sessionMu.Unlock()

	return sess, nil
}

func (si *SessionInitializer) handleTimeout(sessionMu *sync.Mutex, sess *SessionForInitializer, initCompleted *bool, timedOutEarly *atomic.Bool, terminationNotifier chan<- struct{}) {
	sessionMu.Lock()
	completed := *initCompleted
	currentSess := *sess
	sessionMu.Unlock()

	// If initialization already completed, exit silently
	if completed {
		return
	}

	// If session not yet constructed, set early timeout flag
	if currentSess == nil {
		timedOutEarly.Store(true)
		select {
		case terminationNotifier <- struct{}{}:
		default:
		}
		return
	}

	// If session is still initializing, call Session.Fail
	if currentSess.GetStatusSafe() == "initializing" {
		rtErr := fmt.Errorf("session initialization timeout exceeded 30 seconds")
		si.sessionFailErrorMu.Lock()
		si.sessionFailError = rtErr
		si.sessionFailErrorMu.Unlock()
		if si.sessionFailCallbackFunc != nil {
			si.sessionFailCallbackFunc(rtErr, terminationNotifier)
		} else {
			_ = currentSess.Fail(rtErr, terminationNotifier)
		}
	}
}

// sessionWrapper wraps session.Session to implement SessionForInitializer.
type sessionWrapper struct {
	Session *session.Session
	mu      sync.RWMutex
}

func (sw *sessionWrapper) Run(terminationNotifier chan<- struct{}) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if sw.Session.Status != "initializing" {
		return fmt.Errorf("cannot run session: status is '%s', expected 'initializing'", sw.Session.Status)
	}

	sw.Session.Status = "running"
	sw.Session.UpdatedAt = time.Now().Unix()
	return nil
}

func (sw *sessionWrapper) Done(terminationNotifier chan<- struct{}) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.Session.Status = "completed"
	sw.Session.UpdatedAt = time.Now().Unix()

	select {
	case terminationNotifier <- struct{}{}:
	default:
	}

	return nil
}

func (sw *sessionWrapper) Fail(err error, terminationNotifier chan<- struct{}) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.Session.Status = "failed"
	sw.Session.Error = err
	sw.Session.UpdatedAt = time.Now().Unix()

	select {
	case terminationNotifier <- struct{}{}:
	default:
	}

	return nil
}

func (sw *sessionWrapper) GetStatusSafe() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.Status
}

func (sw *sessionWrapper) GetID() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.ID
}

func (sw *sessionWrapper) GetWorkflowName() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.WorkflowName
}

func (sw *sessionWrapper) GetCurrentStateSafe() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.CurrentState
}

func (sw *sessionWrapper) GetCreatedAt() int64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.CreatedAt
}

func (sw *sessionWrapper) GetUpdatedAt() int64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.UpdatedAt
}

func (sw *sessionWrapper) GetEventHistory() []session.Event {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return append([]session.Event(nil), sw.Session.EventHistory...)
}

func (sw *sessionWrapper) GetSessionData() map[string]any {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	result := make(map[string]any)
	for k, v := range sw.Session.SessionData {
		result[k] = v
	}
	return result
}

func (sw *sessionWrapper) GetErrorSafe() error {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.Session.Error
}
