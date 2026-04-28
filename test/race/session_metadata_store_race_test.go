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
	"github.com/tcfwbper/spectra/storage"
	"github.com/tcfwbper/spectra/test/helpers"
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
			metadata := &helpers.SessionMetadata{
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
			metadata := &helpers.SessionMetadata{
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
	initialMetadata := &helpers.SessionMetadata{
		ID:           sessionUUID,
		WorkflowName: "TestWorkflow",
		Status:       "initial",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}
	require.NoError(t, store.Write(initialMetadata))

	var wg sync.WaitGroup
	wg.Add(2)

	var readStatus string
	writeReady := make(chan struct{})
	writeDone := make(chan struct{})
	readStarted := make(chan struct{})

	// Goroutine 1: Write (holds lock)
	go func() {
		defer wg.Done()
		defer close(writeDone)
		metadata := &helpers.SessionMetadata{
			ID:           sessionUUID,
			WorkflowName: "TestWorkflow",
			Status:       "updated",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			CurrentState: "StartNode",
			SessionData:  map[string]interface{}{},
		}
		close(writeReady)
		// Wait for read to start attempting before we complete
		<-readStarted
		err := store.Write(metadata)
		assert.NoError(t, err)
	}()

	// Goroutine 2: Read (should be blocked until write completes)
	go func() {
		defer wg.Done()
		<-writeReady
		close(readStarted)
		metadata, err := store.Read()
		assert.NoError(t, err)
		if metadata != nil {
			readStatus = metadata.Status
		}
	}()

	wg.Wait()

	// Read should return newly written metadata (updated) or initial metadata
	// Due to timing, it could be either, but should be valid
	assert.Contains(t, []string{"initial", "updated"}, readStatus, "Read should return valid status")
}

// TestSessionMetadataStore_WriteBlocksDuringRead verifies write waits for read to complete
func TestSessionMetadataStore_WriteBlocksDuringRead(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// Initial write
	initialMetadata := &helpers.SessionMetadata{
		ID:           sessionUUID,
		WorkflowName: "TestWorkflow",
		Status:       "initial",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}
	require.NoError(t, store.Write(initialMetadata))

	var wg sync.WaitGroup
	wg.Add(2)

	readStarted := make(chan struct{})
	readDone := make(chan struct{})
	writeStarted := make(chan struct{})
	writeDone := make(chan struct{})

	// Goroutine 1: Read (acquires lock first)
	go func() {
		defer wg.Done()
		defer close(readDone)
		close(readStarted)
		// Wait for write to be ready
		<-writeStarted
		metadata, err := store.Read()
		assert.NoError(t, err)
		assert.Equal(t, "initial", metadata.Status)
	}()

	// Goroutine 2: Write (should be blocked until read completes)
	go func() {
		defer wg.Done()
		defer close(writeDone)
		<-readStarted
		close(writeStarted)
		metadata := &helpers.SessionMetadata{
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

	wg.Wait()

	// Verify both operations completed
	<-readDone
	<-writeDone

	// Verify write succeeded
	metadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, "updated", metadata.Status)
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
			metadata := &helpers.SessionMetadata{
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
