package runtime

// =============================================================================
// Compile Stubs — SessionFinalizer
// =============================================================================
//
// These stubs exist ONLY to allow session_finalizer_test.go to compile before
// the production surface (runtime/session_finalizer.go) is implemented.
// They will be REMOVED once the real implementation is in place.
//
// Production surface to be provided:
//   - type SessionFinalizer struct { ... }
//   - func NewSessionFinalizer(logger logger.Logger) *SessionFinalizer
//   - func (sf *SessionFinalizer) Finalize(session *PersistentSession) int
// =============================================================================

import "github.com/tcfwbper/spectra/logger"

// SessionFinalizer is a compile stub. Remove when production file exists.
type SessionFinalizer struct{}

// NewSessionFinalizer is a compile stub. Remove when production file exists.
func NewSessionFinalizer(log logger.Logger) *SessionFinalizer {
	if log == nil {
		panic("NewSessionFinalizer: logger must not be nil")
	}
	return &SessionFinalizer{}
}

// Finalize is a compile stub. Remove when production file exists.
func (sf *SessionFinalizer) Finalize(session *PersistentSession) int {
	return 0
}
