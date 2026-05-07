package runtime

// =============================================================================
// Compile Stubs — SessionInitializer
// =============================================================================
//
// These stubs exist ONLY to allow session_initializer_test.go to compile before
// the production surface (runtime/session_initializer.go) is implemented.
// They will be REMOVED once the real implementation is in place.
//
// Production surface to be provided:
//   - type SessionInitializer struct { ... }
//   - func NewSessionInitializer(projectRoot string, loader WorkflowLoader,
//       dirMgr SessionDirManager, logger logger.Logger) *SessionInitializer
//   - func (si *SessionInitializer) Initialize(workflowName string,
//       terminationNotifier chan<- struct{}) InitResult
//   - type InitResult struct {
//       PersistentSession  *PersistentSession
//       WorkflowDefinition *components.WorkflowDefinition
//       Error              error
//     }
//   - type WorkflowLoader interface { Load(workflowName string) (*components.WorkflowDefinition, error) }
//   - type SessionDirManager interface { CreateSessionDirectory(projectRoot, sessionUUID string) error }
// =============================================================================

import (
	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/logger"
)

// WorkflowLoader is a compile stub interface. Remove when production file exists.
type WorkflowLoader interface {
	Load(workflowName string) (*components.WorkflowDefinition, error)
}

// SessionDirManager is a compile stub interface. Remove when production file exists.
type SessionDirManager interface {
	CreateSessionDirectory(projectRoot, sessionUUID string) error
}

// InitResult is a compile stub. Remove when production file exists.
type InitResult struct {
	PersistentSession  *PersistentSession
	WorkflowDefinition *components.WorkflowDefinition
	Error              error
}

// SessionInitializer is a compile stub. Remove when production file exists.
type SessionInitializer struct{}

// NewSessionInitializer is a compile stub. Remove when production file exists.
func NewSessionInitializer(projectRoot string, loader WorkflowLoader, dirMgr SessionDirManager, log logger.Logger) *SessionInitializer {
	return &SessionInitializer{}
}

// Initialize is a compile stub. Remove when production file exists.
func (si *SessionInitializer) Initialize(workflowName string, terminationNotifier chan<- struct{}) InitResult {
	return InitResult{}
}
