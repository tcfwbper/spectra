package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

// TestFileAccessor_FileExists returns path immediately when file exists without invoking callback
func TestFileAccessor_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))

	callbackInvoked := false
	callback := func() error {
		callbackInvoked = true
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)
	assert.False(t, callbackInvoked, "callback should not be invoked when file exists")
}

// TestFileAccessor_FileExistsCallbackNotCalled verifies callback is never invoked when file exists
func TestFileAccessor_FileExistsCallbackNotCalled(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))

	callCount := 0
	callback := func() error {
		callCount++
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)
	assert.Equal(t, 0, callCount, "callback invocation count should be 0")
}

// TestFileAccessor_CallbackCreatesFile invokes callback when file does not exist; returns path after successful creation
func TestFileAccessor_CallbackCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "new.txt")

	callbackInvoked := false
	callback := func() error {
		callbackInvoked = true
		return os.WriteFile(testFile, []byte("created"), 0644)
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)
	assert.True(t, callbackInvoked, "callback should be invoked")

	_, err = os.Stat(testFile)
	assert.NoError(t, err, "file should exist after callback")
}

// TestFileAccessor_CallbackCreatesParentDirs tests callback creates parent directories and file
func TestFileAccessor_CallbackCreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "a", "b", "c", "file.txt")

	callback := func() error {
		parentDir := filepath.Dir(testFile)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return err
		}
		return os.WriteFile(testFile, []byte("created"), 0644)
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)

	_, err = os.Stat(testFile)
	assert.NoError(t, err, "file should exist after callback")
}

// TestFileAccessor_CallbackReturnsError returns wrapped error when callback fails
func TestFileAccessor_CallbackReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fail.txt")

	callbackErr := errors.New("callback error")
	callback := func() error {
		return callbackErr
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)failed to prepare file.*fail\.txt`, err.Error())
	assert.Regexp(t, `(?i)callback error`, err.Error())

	_, statErr := os.Stat(testFile)
	assert.True(t, os.IsNotExist(statErr), "file should not be created")
}

// TestFileAccessor_CallbackErrorParentMissing tests callback returns error indicating parent directory missing
func TestFileAccessor_CallbackErrorParentMissing(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "missing", "file.txt")

	callback := func() error {
		parentDir := filepath.Dir(testFile)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			return errors.New("parent directory does not exist")
		}
		return os.WriteFile(testFile, []byte("created"), 0644)
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)failed to prepare file.*missing/file\.txt`, err.Error())
	assert.Regexp(t, `(?i)parent directory`, err.Error())
}

// TestFileAccessor_CallbackNilButFileNotCreated returns error when callback returns nil but file still does not exist
func TestFileAccessor_CallbackNilButFileNotCreated(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "buggy.txt")

	callback := func() error {
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)file preparation succeeded but file was not created.*buggy\.txt`, err.Error())
}

// TestFileAccessor_PermissionDeniedBeforeCallback returns permission error without invoking callback
func TestFileAccessor_PermissionDeniedBeforeCallback(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "restricted.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))
	require.NoError(t, os.Chmod(testFile, 0000))
	defer os.Chmod(testFile, 0644)

	callbackInvoked := false
	callback := func() error {
		callbackInvoked = true
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)permission denied.*restricted\.txt`, err.Error())
	assert.False(t, callbackInvoked, "callback should not be invoked")
}

// TestFileAccessor_PermissionDeniedAfterCallback returns permission error from post-callback stat
func TestFileAccessor_PermissionDeniedAfterCallback(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "denied.txt")

	callbackInvoked := false
	callback := func() error {
		callbackInvoked = true
		if err := os.WriteFile(testFile, []byte("created"), 0644); err != nil {
			return err
		}
		return os.Chmod(testFile, 0000)
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.True(t, callbackInvoked, "callback should be invoked")
	assert.Regexp(t, `(?i)permission denied.*denied\.txt`, err.Error())

	os.Chmod(testFile, 0644)
}

// TestFileAccessor_PathIsDirectory tests stat succeeds for directory; returns path
func TestFileAccessor_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "dir")
	require.NoError(t, os.Mkdir(testDir, 0755))

	callbackInvoked := false
	callback := func() error {
		callbackInvoked = true
		return nil
	}

	result, err := storage.FileAccessor(testDir, callback)
	assert.NoError(t, err)
	assert.Equal(t, testDir, result)
	assert.False(t, callbackInvoked, "callback should not be invoked when directory exists")
}

// TestFileAccessor_EmptyFilePath returns stat error for empty file path
func TestFileAccessor_EmptyFilePath(t *testing.T) {
	callback := func() error {
		return nil
	}

	result, err := storage.FileAccessor("", callback)
	assert.Error(t, err)
	assert.Empty(t, result)
}

// TestFileAccessor_NilCallback panics when callback is nil and file does not exist
func TestFileAccessor_NilCallback(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	assert.Panics(t, func() {
		storage.FileAccessor(testFile, nil)
	})
}

// TestFileAccessor_FileCreatedBetweenStatAndCallback tests file created by another process between stat checks
func TestFileAccessor_FileCreatedBetweenStatAndCallback(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "race.txt")

	callbackInvoked := false
	callback := func() error {
		callbackInvoked = true
		os.WriteFile(testFile, []byte("created by external process"), 0644)
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)
	assert.True(t, callbackInvoked)
}

// TestFileAccessor_FileDeletedBetweenCallbackAndPostStat tests file created by callback but deleted before post-callback stat
func TestFileAccessor_FileDeletedBetweenCallbackAndPostStat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "deleted.txt")

	callback := func() error {
		if err := os.WriteFile(testFile, []byte("created"), 0644); err != nil {
			return err
		}
		return os.Remove(testFile)
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)file preparation succeeded but file was not created`, err.Error())
}

// TestFileAccessor_CallbackIdempotent tests callback can be invoked multiple times safely
func TestFileAccessor_CallbackIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "idempotent.txt")

	callback := func() error {
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			return os.WriteFile(testFile, []byte("created"), 0644)
		}
		return nil
	}

	result1, err1 := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err1)
	assert.Equal(t, testFile, result1)

	result2, err2 := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err2)
	assert.Equal(t, testFile, result2)
}

// TestFileAccessor_NeverReadsFileContent tests FileAccessor only stats the file, never opens or reads it
func TestFileAccessor_NeverReadsFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "content.txt")
	originalContent := "secret data"
	require.NoError(t, os.WriteFile(testFile, []byte(originalContent), 0644))

	callback := func() error {
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)

	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, string(content))
}

// TestFileAccessor_NeverWritesFileContent tests FileAccessor never writes to the file
func TestFileAccessor_NeverWritesFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "readonly.txt")
	originalContent := "original"
	require.NoError(t, os.WriteFile(testFile, []byte(originalContent), 0644))

	callback := func() error {
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)

	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, string(content))
}

// TestFileAccessor_WrapsCallbackError tests callback errors are wrapped with file path context
func TestFileAccessor_WrapsCallbackError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fail.txt")

	callback := func() error {
		return errors.New("disk full")
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)failed to prepare file.*fail\.txt.*disk full`, err.Error())
}

// TestFileAccessor_CallbackPanics tests panic from callback propagates to caller
func TestFileAccessor_CallbackPanics(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "panic.txt")

	callback := func() error {
		panic("callback panic")
	}

	assert.Panics(t, func() {
		storage.FileAccessor(testFile, callback)
	})
}

// TestFileAccessor_AbsolutePath handles absolute paths correctly
func TestFileAccessor_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))

	callback := func() error {
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.True(t, filepath.IsAbs(result))
	assert.Equal(t, testFile, result)
}

// TestFileAccessor_RelativePath passes relative paths to stat as-is; no validation or conversion to absolute
func TestFileAccessor_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tmpDir))
	testFile := "./relative.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))

	callback := func() error {
		return nil
	}

	result, err := storage.FileAccessor(testFile, callback)
	assert.NoError(t, err)
	assert.Equal(t, testFile, result)
}
