package storage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — FileAccessor ---

func TestFileAccessor_FileExists(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "existing.txt")
	makeTempFile(t, filePath)

	callbackCalled := false
	prepare := func() error {
		callbackCalled = true
		return nil
	}

	result, err := FileAccessor(filePath, prepare)

	require.NoError(t, err)
	assert.Equal(t, filePath, result)
	assert.False(t, callbackCalled, "callback should not be called when file exists")
}

func TestFileAccessor_FileNotExistsCallbackCreates(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "newfile.txt")

	prepare := func() error {
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		return f.Close()
	}

	result, err := FileAccessor(filePath, prepare)

	require.NoError(t, err)
	assert.Equal(t, filePath, result)
}

// --- Error Propagation ---

func TestFileAccessor_CallbackReturnsError(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "missing.txt")

	prepare := func() error {
		return errors.New("disk full")
	}

	_, err := FileAccessor(filePath, prepare)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare file")
	assert.Contains(t, err.Error(), filePath)
	assert.Contains(t, err.Error(), "disk full")
}

func TestFileAccessor_CallbackSucceedsButFileNotCreated(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "ghost.txt")

	prepare := func() error {
		return nil // no-op: does not create file
	}

	_, err := FileAccessor(filePath, prepare)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file preparation succeeded but file was not created")
	assert.Contains(t, err.Error(), filePath)
}

func TestFileAccessor_InitialStatPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}

	dir := makeTempDir(t)
	restrictedDir := filepath.Join(dir, "restricted")
	require.NoError(t, os.Mkdir(restrictedDir, 0000))
	t.Cleanup(func() { os.Chmod(restrictedDir, 0755) })

	filePath := filepath.Join(restrictedDir, "secret.txt")

	callbackCalled := false
	prepare := func() error {
		callbackCalled = true
		return nil
	}

	_, err := FileAccessor(filePath, prepare)

	require.Error(t, err)
	assert.True(t, os.IsPermission(err) || errors.Is(err, os.ErrPermission) || containsPermissionError(err),
		"expected permission error, got: %v", err)
	assert.False(t, callbackCalled, "callback should not be called on initial stat permission error")
}

func TestFileAccessor_PostCallbackStatPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "lockdown.txt")

	prepare := func() error {
		// Create the file first
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		f.Close()
		// Then restrict parent directory so post-callback stat fails
		return os.Chmod(dir, 0000)
	}

	t.Cleanup(func() { os.Chmod(dir, 0755) })

	_, err := FileAccessor(filePath, prepare)

	require.Error(t, err)
	// Should be a permission error from post-callback stat
	assert.True(t, os.IsPermission(err) || errors.Is(err, os.ErrPermission) || containsPermissionError(err),
		"expected permission error, got: %v", err)
}

func TestFileAccessor_CallbackPanics(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "panic.txt")

	prepare := func() error {
		panic("boom")
	}

	assert.PanicsWithValue(t, "boom", func() {
		FileAccessor(filePath, prepare)
	})
}

// --- Null / Empty Input ---

func TestFileAccessor_EmptyFilePath(t *testing.T) {

	prepare := func() error {
		return nil
	}

	_, err := FileAccessor("", prepare)

	require.Error(t, err)
}

func TestFileAccessor_NilCallbackFileNotExists(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "noexist.txt")

	assert.Panics(t, func() {
		FileAccessor(filePath, nil)
	})
}

// --- Boundary Values — filePath ---

func TestFileAccessor_FilePathIsDirectory(t *testing.T) {

	dir := makeTempDir(t)
	subDir := filepath.Join(dir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	callbackCalled := false
	prepare := func() error {
		callbackCalled = true
		return nil
	}

	result, err := FileAccessor(subDir, prepare)

	require.NoError(t, err)
	assert.Equal(t, subDir, result)
	assert.False(t, callbackCalled, "callback should not be called when stat succeeds on directory")
}

// --- Idempotency ---

func TestFileAccessor_IndependentInvocations(t *testing.T) {

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "idempotent.txt")

	callCount := 0
	prepare := func() error {
		callCount++
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		return f.Close()
	}

	// First call: file does not exist, callback creates it
	result1, err1 := FileAccessor(filePath, prepare)
	require.NoError(t, err1)
	assert.Equal(t, filePath, result1)
	assert.Equal(t, 1, callCount)

	// Second call: file now exists, callback should NOT be called
	result2, err2 := FileAccessor(filePath, prepare)
	require.NoError(t, err2)
	assert.Equal(t, filePath, result2)
	assert.Equal(t, 1, callCount, "callback should not be invoked on second call when file exists")
}

// --- Test Helpers ---

// containsPermissionError checks if the error chain contains a permission-related error.
func containsPermissionError(err error) bool {
	if err == nil {
		return false
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return os.IsPermission(pathErr)
	}
	return false
}
