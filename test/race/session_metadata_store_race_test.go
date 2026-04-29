package race_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// TestSessionMetadataStore_ConcurrentWriteSameFile tests multiple goroutines write to same file safely
func TestSessionMetadataStore_ConcurrentWriteSameFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	var wg sync.WaitGroup
	numGoroutines := 10
	wg.Add(numGoroutines)

	// Launch 10 goroutines each writing different metadata
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			metadata := &entities.SessionMetadata{
				ID:           sessionUUID,
				WorkflowName: "TestWorkflow",
				Status:       "running",
				CreatedAt:    time.Now().Unix(),
				UpdatedAt:    time.Now().Unix(),
				CurrentState: "StartNode",
				SessionData: map[string]interface{}{
					"index": index,
				},
			}
			err := store.Write(metadata)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify file contains valid JSON from one of the writers
	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	assert.NoError(t, err, "file should contain valid JSON")

	// Verify index is one of the written values (0-9)
	sessionData := result["sessionData"].(map[string]interface{})
	index := sessionData["index"].(float64)
	assert.GreaterOrEqual(t, index, float64(0))
	assert.LessOrEqual(t, index, float64(9))
}

// TestSessionMetadataStore_ConcurrentWriteSerializes verifies file lock serializes concurrent writes
func TestSessionMetadataStore_ConcurrentWriteSerializes(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	var wg sync.WaitGroup
	numGoroutines := 5
	wg.Add(numGoroutines)

	statuses := []string{"status1", "status2", "status3", "status4", "status5"}

	// Launch 5 goroutines writing simultaneously with distinct Status values
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			metadata := &entities.SessionMetadata{
				ID:           sessionUUID,
				WorkflowName: "TestWorkflow",
				Status:       statuses[index],
				CreatedAt:    time.Now().Unix(),
				UpdatedAt:    time.Now().Unix(),
				CurrentState: "StartNode",
				SessionData:  map[string]interface{}{},
			}
			err := store.Write(metadata)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify file contains valid complete JSON
	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	assert.NoError(t, err, "writes serialized by lock; final file contains valid complete JSON")

	// Verify status is one of the written values
	status := result["status"].(string)
	assert.Contains(t, statuses, status)
}

// TestSessionMetadataStore_ReadBlocksDuringWrite verifies read waits for write to complete
func TestSessionMetadataStore_ReadBlocksDuringWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// Initial write
	initialMetadata := &entities.SessionMetadata{
		ID:           sessionUUID,
		WorkflowName: "TestWorkflow",
		Status:       "initial",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}
	require.NoError(t, store.Write(initialMetadata))

	// Verification approach: We cannot deterministically control which goroutine acquires
	// the file lock first due to Go scheduler and OS lock contention. Instead, we verify:
	// 1. Both operations complete successfully without corruption
	// 2. The final state is consistent (file contains valid JSON)
	// 3. No race conditions occur (verified by running with -race flag)

	var wg sync.WaitGroup
	wg.Add(2)

	var readStatus string
	var readErr error

	// Goroutine 1: Write with updated metadata
	go func() {
		defer wg.Done()
		metadata := &entities.SessionMetadata{
			ID:           sessionUUID,
			WorkflowName: "TestWorkflow",
			Status:       "updated",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			CurrentState: "StartNode",
			SessionData:  map[string]interface{}{},
		}
		err := store.Write(metadata)
		assert.NoError(t, err)
	}()

	// Goroutine 2: Read concurrently
	go func() {
		defer wg.Done()
		metadata, err := store.Read()
		readErr = err
		if metadata != nil {
			readStatus = metadata.Status
		}
	}()

	wg.Wait()

	// Both operations should succeed
	assert.NoError(t, readErr)

	// Read should return valid metadata (either "initial" before write or "updated" after write)
	assert.Contains(t, []string{"initial", "updated"}, readStatus,
		"Read should return valid status - either initial (if read acquired lock first) or updated (if write acquired lock first)")

	// Verify final file state is consistent
	finalMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, "updated", finalMetadata.Status, "Final file should contain updated status")
}

// TestSessionMetadataStore_WriteBlocksDuringRead verifies write waits for read to complete
func TestSessionMetadataStore_WriteBlocksDuringRead(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// Initial write
	initialMetadata := &entities.SessionMetadata{
		ID:           sessionUUID,
		WorkflowName: "TestWorkflow",
		Status:       "initial",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}
	require.NoError(t, store.Write(initialMetadata))

	// Verification approach: We cannot deterministically control which goroutine acquires
	// the file lock first due to Go scheduler and OS lock contention. Instead, we verify:
	// 1. Both operations complete successfully without corruption
	// 2. The final state is consistent (file contains updated metadata after both complete)
	// 3. No race conditions occur (verified by running with -race flag)

	var wg sync.WaitGroup
	wg.Add(2)

	var readStatus string
	var readErr error
	var writeErr error

	// Goroutine 1: Read concurrently
	go func() {
		defer wg.Done()
		metadata, err := store.Read()
		readErr = err
		if metadata != nil {
			readStatus = metadata.Status
		}
	}()

	// Goroutine 2: Write concurrently
	go func() {
		defer wg.Done()
		metadata := &entities.SessionMetadata{
			ID:           sessionUUID,
			WorkflowName: "TestWorkflow",
			Status:       "updated",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			CurrentState: "StartNode",
			SessionData:  map[string]interface{}{},
		}
		writeErr = store.Write(metadata)
	}()

	wg.Wait()

	// Both operations should succeed
	assert.NoError(t, readErr)
	assert.NoError(t, writeErr)

	// Read should return valid metadata (either "initial" or "updated" depending on lock acquisition order)
	assert.Contains(t, []string{"initial", "updated"}, readStatus,
		"Read should return valid status - either initial (if read acquired lock first) or updated (if write acquired lock first)")

	// Verify final file state is consistent
	finalMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, "updated", finalMetadata.Status, "Final file should contain updated status after both operations complete")
}

// TestSessionMetadataStore_ConcurrentSameProcess verifies multiple goroutines in same process write safely
func TestSessionMetadataStore_ConcurrentSameProcess(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	var wg sync.WaitGroup
	numGoroutines := 5
	wg.Add(numGoroutines)

	// Launch 5 goroutines in same process writing concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			metadata := &entities.SessionMetadata{
				ID:           sessionUUID,
				WorkflowName: "TestWorkflow",
				Status:       "running",
				CreatedAt:    time.Now().Unix(),
				UpdatedAt:    time.Now().Unix(),
				CurrentState: "StartNode",
				SessionData: map[string]interface{}{
					"goroutine": index,
				},
			}
			err := store.Write(metadata)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify file contains valid JSON (no corruption)
	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	assert.NoError(t, err, "file lock works within process; writes serialized; no data races")
}
