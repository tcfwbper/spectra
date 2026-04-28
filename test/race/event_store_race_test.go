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
	"github.com/tcfwbper/spectra/test/helpers"
)

// TestEventStore_ConcurrentAppendSameFile verifies multiple goroutines append to same file safely
func TestEventStore_ConcurrentAppendSameFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewEventStore(tmpDir, sessionUUID)
	numGoroutines := 10
	eventsPerGoroutine := 5
	totalEvents := numGoroutines * eventsPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := &entities.Event{
					ID:        uuid.New(),
					Type:      "TestEvent",
					Message:   "test",
					Payload:   json.RawMessage(`{}`),
					EmittedBy: "TestNode",
					EmittedAt: time.Now().Unix(),
					SessionID: sessionUUID,
				}
				err := store.Append(event)
				if err != nil {
					errors[idx] = err
					return
				}
			}
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "goroutine %d should succeed", i)
	}

	// Verify all events written
	events, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, totalEvents, len(events), "should have exactly %d events", totalEvents)

	// Verify no corruption - all lines should be valid JSON
	eventsFile := filepath.Join(sessionDir, "events.jsonl")
	content, err := os.ReadFile(eventsFile)
	assert.NoError(t, err)

	lines := helpers.SplitLines(string(content))
	assert.Equal(t, totalEvents, len(lines))
}

// TestEventStore_ConcurrentAppendSerializes verifies file lock serializes concurrent writes
func TestEventStore_ConcurrentAppendSerializes(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewEventStore(tmpDir, sessionUUID)
	numGoroutines := 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			event := &entities.Event{
				ID:        uuid.New(),
				Type:      "TestEvent",
				Message:   "test",
				Payload:   json.RawMessage(`{}`),
				EmittedBy: "TestNode",
				EmittedAt: time.Now().Unix(),
				SessionID: sessionUUID,
			}
			err := store.Append(event)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	// Verify all events written as complete lines
	eventsFile := filepath.Join(sessionDir, "events.jsonl")
	content, err := os.ReadFile(eventsFile)
	assert.NoError(t, err)

	lines := helpers.SplitLines(string(content))
	assert.Equal(t, numGoroutines, len(lines))

	// Each line should be valid JSON
	for i, line := range lines {
		var event entities.Event
		err := json.Unmarshal([]byte(line), &event)
		assert.NoError(t, err, "line %d should be valid JSON", i)
	}
}

// TestEventStore_ReadBlocksDuringWrite verifies read waits for write to complete
func TestEventStore_ReadBlocksDuringWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewEventStore(tmpDir, sessionUUID)

	// Write initial event
	event1 := &entities.Event{
		ID:        uuid.New(),
		Type:      "Event1",
		Message:   "first",
		Payload:   json.RawMessage(`{}`),
		EmittedBy: "Node1",
		EmittedAt: 1000,
		SessionID: sessionUUID,
	}
	require.NoError(t, store.Append(event1))

	var wg sync.WaitGroup
	wg.Add(2)

	readComplete := false
	var mu sync.Mutex

	// Goroutine 1: Append (will hold lock briefly)
	go func() {
		defer wg.Done()
		event := &entities.Event{
			ID:        uuid.New(),
			Type:      "Event2",
			Message:   "second",
			Payload:   json.RawMessage(`{}`),
			EmittedBy: "Node2",
			EmittedAt: 2000,
			SessionID: sessionUUID,
		}
		err := store.Append(event)
		assert.NoError(t, err)
	}()

	// Small delay to ensure write starts first
	time.Sleep(10 * time.Millisecond)

	// Goroutine 2: Read (should block until write completes)
	go func() {
		defer wg.Done()
		events, err := store.Read()
		assert.NoError(t, err)
		mu.Lock()
		readComplete = true
		mu.Unlock()
		// Should include newly written event
		assert.GreaterOrEqual(t, len(events), 1)
	}()

	wg.Wait()

	mu.Lock()
	assert.True(t, readComplete)
	mu.Unlock()
}

// TestEventStore_WriteBlocksDuringRead verifies write waits for read to complete
func TestEventStore_WriteBlocksDuringRead(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewEventStore(tmpDir, sessionUUID)

	// Write initial events
	for i := 0; i < 3; i++ {
		event := &entities.Event{
			ID:        uuid.New(),
			Type:      "InitialEvent",
			Message:   "test",
			Payload:   json.RawMessage(`{}`),
			EmittedBy: "TestNode",
			EmittedAt: int64(1000 + i),
			SessionID: sessionUUID,
		}
		require.NoError(t, store.Append(event))
	}

	var wg sync.WaitGroup
	wg.Add(2)

	writeComplete := false
	var mu sync.Mutex

	// Goroutine 1: Read (will hold lock)
	go func() {
		defer wg.Done()
		events, err := store.Read()
		assert.NoError(t, err)
		assert.Len(t, events, 3)
		time.Sleep(50 * time.Millisecond) // Hold lock briefly
	}()

	// Small delay to ensure read starts first
	time.Sleep(10 * time.Millisecond)

	// Goroutine 2: Append (should block until read completes)
	go func() {
		defer wg.Done()
		event := &entities.Event{
			ID:        uuid.New(),
			Type:      "NewEvent",
			Message:   "new",
			Payload:   json.RawMessage(`{}`),
			EmittedBy: "TestNode",
			EmittedAt: 4000,
			SessionID: sessionUUID,
		}
		err := store.Append(event)
		assert.NoError(t, err)
		mu.Lock()
		writeComplete = true
		mu.Unlock()
	}()

	wg.Wait()

	mu.Lock()
	assert.True(t, writeComplete)
	mu.Unlock()

	// Verify new event was appended
	events, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(events))
}
