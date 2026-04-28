package race_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

// TestCreateSessionDirectory_ConcurrentSameUUID tests two goroutines attempt to create same session directory simultaneously
func TestCreateSessionDirectory_ConcurrentSameUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	var wg sync.WaitGroup
	wg.Add(2)

	errors := make([]error, 2)

	for i := 0; i < 2; i++ {
		go func(idx int) {
			defer wg.Done()
			err := manager.CreateSessionDirectory(sessionUUID)
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}

	assert.Equal(t, 1, successCount, "exactly one goroutine should succeed")
}

// TestCreateSessionDirectory_ConcurrentDifferentUUIDs tests multiple goroutines create different session directories simultaneously
func TestCreateSessionDirectory_ConcurrentDifferentUUIDs(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			sessionUUID := "123e4567-e89b-12d3-a456-42661417" + string(rune('0'+idx))
			err := manager.CreateSessionDirectory(sessionUUID)
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "goroutine %d should succeed", i)
	}

	for i := 0; i < numGoroutines; i++ {
		sessionUUID := "123e4567-e89b-12d3-a456-42661417" + string(rune('0'+i))
		sessionDir := filepath.Join(sessionsDir, sessionUUID)
		info, err := os.Stat(sessionDir)
		assert.NoError(t, err, "session directory %d should exist", i)
		if err == nil {
			assert.True(t, info.IsDir())
			assert.Equal(t, os.FileMode(0775), info.Mode().Perm())
		}
	}
}

// TestSessionDirectoryManager_ThreadSafe tests multiple goroutines use same manager instance concurrently
func TestSessionDirectoryManager_ThreadSafe(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			sessionUUID := "123e4567-e89b-12d3-a456-42661417" + string(rune('0'+idx))
			err := manager.CreateSessionDirectory(sessionUUID)
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "goroutine %d should succeed", i)
	}
}
