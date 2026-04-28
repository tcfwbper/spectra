package race_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfwbper/spectra/storage"
)

// TestFileAccessor_ConcurrentAccessSameFile tests multiple goroutines call FileAccessor for same non-existent file
func TestFileAccessor_ConcurrentAccessSameFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "concurrent.txt")

	numGoroutines := 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			callback := func() error {
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					return os.WriteFile(testFile, []byte("created"), 0644)
				}
				return nil
			}

			result, err := storage.FileAccessor(testFile, callback)
			if err == nil && result == testFile {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	assert.Greater(t, successCount, 0, "at least one goroutine should succeed")
}

// TestFileAccessor_ConcurrentAccessDifferentFiles tests multiple goroutines call FileAccessor for different files
func TestFileAccessor_ConcurrentAccessDifferentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	numGoroutines := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			testFile := filepath.Join(tmpDir, "file"+string(rune('0'+idx))+".txt")
			callback := func() error {
				return os.WriteFile(testFile, []byte("created"), 0644)
			}

			_, err := storage.FileAccessor(testFile, callback)
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "goroutine %d should succeed", i)
	}

	for i := 0; i < numGoroutines; i++ {
		testFile := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		_, err := os.Stat(testFile)
		assert.NoError(t, err, "file %d should exist", i)
	}
}
