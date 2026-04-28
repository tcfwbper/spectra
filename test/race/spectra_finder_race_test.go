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

// TestSpectraFinder_ConcurrentSearches tests multiple goroutines search for .spectra simultaneously
func TestSpectraFinder_ConcurrentSearches(t *testing.T) {
	tmpDir := t.TempDir()
	rootDir := filepath.Join(tmpDir, "root")
	require.NoError(t, os.MkdirAll(rootDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(rootDir, ".spectra"), 0755))

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			startDir := rootDir
			if idx%2 == 0 {
				subDir := filepath.Join(rootDir, "sub"+string(rune('0'+idx)))
				os.MkdirAll(subDir, 0755)
				startDir = subDir
			}
			result, err := storage.SpectraFinder(startDir)
			results[idx] = result
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	absRoot, err := filepath.Abs(rootDir)
	require.NoError(t, err)

	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errors[i], "goroutine %d should succeed", i)
		assert.Equal(t, absRoot, results[i], "goroutine %d should return correct path", i)
	}
}
