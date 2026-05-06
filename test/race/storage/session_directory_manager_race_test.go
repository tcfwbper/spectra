package storage_race

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spectra-ai/spectra/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSessionDirectory_ConcurrentSameUUID(t *testing.T) {

	// Setup: Create temp directory with `.spectra/sessions/`
	dir := t.TempDir()
	spectraDir := filepath.Join(dir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))
	sessionsDir := filepath.Join(spectraDir, "sessions")
	require.NoError(t, os.Mkdir(sessionsDir, 0755))

	uuid := "550e8400-e29b-41d4-a716-446655440000"
	const goroutines = 2

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = storage.CreateSessionDirectory(dir, uuid)
		}(i)
	}

	wg.Wait()

	// One should succeed (nil), the other should get ErrSessionDirExists or a wrapped "file exists" error
	nilCount := 0
	errCount := 0
	for _, err := range errs {
		if err == nil {
			nilCount++
		} else {
			errCount++
			// Verify it's either ErrSessionDirExists or contains "file exists"
			isExpectedErr := assert.ErrorIs(t, err, storage.ErrSessionDirExists) ||
				assert.Contains(t, err.Error(), "file exists")
			_ = isExpectedErr
		}
	}

	assert.Equal(t, 1, nilCount, "exactly one goroutine should succeed")
	assert.Equal(t, 1, errCount, "exactly one goroutine should get an error")
}
