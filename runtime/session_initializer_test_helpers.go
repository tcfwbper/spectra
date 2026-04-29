package runtime

import (
	"sync/atomic"
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

// SetSocketCreateError configures the initializer to fail during socket creation.
func (si *SessionInitializer) SetSocketCreateError(err error) {
	si.socketManagerFunc = func() (RuntimeSocketManager, error) {
		return &mockSocketManagerForHelper{}, err
	}
}

// SetFileAccessorError configures the initializer to fail during file accessor preparation.
func (si *SessionInitializer) SetFileAccessorError(err error) {
	si.fileAccessorFunc = func() error {
		return err
	}
}

// SetSpectraFinderError configures the initializer to fail during spectra finder execution.
func (si *SessionInitializer) SetSpectraFinderError(err error) {
	si.spectraFinderFunc = func() (string, error) {
		return "", err
	}
}

// SetSpectraFinderBlock configures the initializer to block on spectra finder until channel is closed.
func (si *SessionInitializer) SetSpectraFinderBlock(blockCh chan struct{}) {
	si.spectraFinderBlockCh = blockCh
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

// WasDeleteSocketCalled returns whether DeleteSocket was called.
func (si *SessionInitializer) WasDeleteSocketCalled() bool {
	return si.wasDeleteSocketCalled.Load()
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

// mockSocketManagerForHelper is a simple mock for test helpers.
type mockSocketManagerForHelper struct {
	deleteSocketCalled atomic.Bool
}

func (m *mockSocketManagerForHelper) DeleteSocket() error {
	m.deleteSocketCalled.Store(true)
	return nil
}

func (m *mockSocketManagerForHelper) CreateSocket() error {
	return nil
}
