package runtime

import (
	"time"
)

// Test helper methods for SessionInitializer

// SetTimeoutDuration sets a custom timeout duration for testing.
func (si *SessionInitializer) SetTimeoutDuration(duration time.Duration) {
	si.timeoutDuration = duration
}

// SetSessionRunError configures the initializer to return an error when Session.Run is called.
func (si *SessionInitializer) SetSessionRunError(err error) {
	si.sessionRunErrorFunc = func() error {
		return err
	}
}

// SetMetadataWriteError configures the initializer to fail during metadata write.
func (si *SessionInitializer) SetMetadataWriteError(err error) {
	si.metadataStoreFunc = func() (MetadataStore, error) {
		return nil, err
	}
}

// SetFileAccessorError configures the initializer to fail during file accessor preparation.
func (si *SessionInitializer) SetFileAccessorError(err error) {
	si.fileAccessorFunc = func() error {
		return err
	}
}

// SetSessionRunBlock configures the initializer to block on Session.Run until channel is closed.
func (si *SessionInitializer) SetSessionRunBlock(blockCh chan struct{}) {
	si.sessionRunBlockCh = blockCh
}

// SetSessionRunCallback sets a callback to be invoked when Session.Run is called.
func (si *SessionInitializer) SetSessionRunCallback(callback func(chan<- struct{})) {
	si.sessionRunCallbackFunc = callback
}

// SetSessionFailCallback sets a callback to be invoked when Session.Fail is called.
func (si *SessionInitializer) SetSessionFailCallback(callback func(error, chan<- struct{})) {
	si.sessionFailCallbackFunc = callback
}

// SetCallOrderTracker sets a tracker for recording call order.
func (si *SessionInitializer) SetCallOrderTracker(tracker callOrderRecorder) {
	si.callOrderTracker = tracker
}

// WasSessionFailCalled returns whether Session.Fail was called.
func (si *SessionInitializer) WasSessionFailCalled() bool {
	return si.wasSessionFailCalled.Load()
}

// GetSessionFailError returns the error passed to Session.Fail.
func (si *SessionInitializer) GetSessionFailError() error {
	si.sessionFailErrorMu.Lock()
	defer si.sessionFailErrorMu.Unlock()
	return si.sessionFailError
}
