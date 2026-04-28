package session

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Happy Path — UpdateEventHistorySafe

func TestUpdateEventHistorySafe_AppendsEvent(t *testing.T) {
	session := createTestSession(t, "running", "node1")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Started",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node1",
		Message:   "begin",
		Payload:   map[string]any{},
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
	assert.Equal(t, event, session.EventHistory[0])
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestUpdateEventHistorySafe_AppendsToExisting(t *testing.T) {
	events := []Event{
		{ID: "evt-1", Type: "Event", SessionID: uuid.New().String(), EmittedAt: 1000, EmittedBy: "node1"},
		{ID: "evt-2", Type: "Event", SessionID: uuid.New().String(), EmittedAt: 2000, EmittedBy: "node2"},
	}
	session := createTestSessionWithEvents(t, "running", "node2", events)
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event3 := Event{
		ID:        "evt-3",
		Type:      "Progress",
		SessionID: session.ID,
		EmittedAt: 3000,
		EmittedBy: "node2",
		Message:   "",
		Payload:   map[string]any{},
	}
	err := session.UpdateEventHistorySafe(event3)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(session.EventHistory))
	assert.Equal(t, event3, session.EventHistory[2])
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestUpdateEventHistorySafe_EmptyMessage(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
		Message:   "",
		Payload:   map[string]any{},
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
	assert.Equal(t, "", session.EventHistory[0].Message)
}

func TestUpdateEventHistorySafe_NilPayload(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
		Message:   "msg",
		Payload:   nil,
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
	assert.Nil(t, session.EventHistory[0].Payload)
}

func TestUpdateEventHistorySafe_PersistsToEventStore(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
		Message:   "msg",
		Payload:   map[string]any{},
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	session.eventStore.AssertCalled(t, "WriteEvent", event)
}

// Validation Failures — Required Fields

func TestUpdateEventHistorySafe_EmptyID(t *testing.T) {
	session := createTestSession(t, "running", "node")

	event := Event{
		ID:        "",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid event.*ID is required", err.Error())
	assert.Equal(t, 0, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_EmptyType(t *testing.T) {
	session := createTestSession(t, "running", "node")

	event := Event{
		ID:        "evt-1",
		Type:      "",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid event.*Type is required", err.Error())
	assert.Equal(t, 0, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_EmptySessionID(t *testing.T) {
	session := createTestSession(t, "running", "node")

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: "",
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid event.*SessionID is required", err.Error())
	assert.Equal(t, 0, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_ZeroEmittedAt(t *testing.T) {
	session := createTestSession(t, "running", "node")

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 0,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid event.*EmittedAt is required", err.Error())
	assert.Equal(t, 0, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_NegativeEmittedAt(t *testing.T) {
	session := createTestSession(t, "running", "node")

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: -100,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid event.*EmittedAt is required", err.Error())
	assert.Equal(t, 0, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_EmptyEmittedBy(t *testing.T) {
	session := createTestSession(t, "running", "node")

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid event.*EmittedBy is required", err.Error())
	assert.Equal(t, 0, len(session.EventHistory))
}

// Idempotency

func TestUpdateEventHistorySafe_NotIdempotent(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}

	err1 := session.UpdateEventHistorySafe(event)
	err2 := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, 2, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_SameIDAppendedTwice(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event1 := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	event2 := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 2000,
		EmittedBy: "node",
	}

	err1 := session.UpdateEventHistorySafe(event1)
	err2 := session.UpdateEventHistorySafe(event2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, 2, len(session.EventHistory))
	assert.Equal(t, "evt-1", session.EventHistory[0].ID)
	assert.Equal(t, "evt-1", session.EventHistory[1].ID)
}

// Ordering — Chronological

func TestUpdateEventHistorySafe_MaintainsChronologicalOrder(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event1 := Event{ID: "evt-1", Type: "Event", SessionID: session.ID, EmittedAt: 1000, EmittedBy: "node"}
	event2 := Event{ID: "evt-2", Type: "Event", SessionID: session.ID, EmittedAt: 2000, EmittedBy: "node"}
	event3 := Event{ID: "evt-3", Type: "Event", SessionID: session.ID, EmittedAt: 1500, EmittedBy: "node"}

	_ = session.UpdateEventHistorySafe(event1)
	_ = session.UpdateEventHistorySafe(event2)
	_ = session.UpdateEventHistorySafe(event3)

	assert.Equal(t, 3, len(session.EventHistory))
	assert.Equal(t, int64(1000), session.EventHistory[0].EmittedAt)
	assert.Equal(t, int64(2000), session.EventHistory[1].EmittedAt)
	assert.Equal(t, int64(1500), session.EventHistory[2].EmittedAt)
}

// Concurrent Behaviour

func TestEventHistory_ConcurrentAppends(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			event := Event{
				ID:        uuid.New().String(),
				Type:      "Event",
				SessionID: session.ID,
				EmittedAt: int64(index),
				EmittedBy: "node",
			}
			_ = session.UpdateEventHistorySafe(event)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 10, len(session.EventHistory))
}

// Error Propagation

func TestUpdateEventHistorySafe_EventStorePersistenceFailureLogged(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
	session.logger.AssertCalled(t, "Warning", mock.MatchedBy(func(msg string) bool {
		return strings.Contains(strings.ToLower(msg), "eventstore") ||
			strings.Contains(strings.ToLower(msg), "persistence failed")
	}))
}

func TestUpdateEventHistorySafe_PersistenceFailureDoesNotRevert(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
}

// Invariants — Append-Only

func TestEventHistory_AppendOnly(t *testing.T) {
	events := []Event{
		{ID: "evt-1", Type: "Event", SessionID: uuid.New().String(), EmittedAt: 1000, EmittedBy: "node"},
		{ID: "evt-2", Type: "Event", SessionID: uuid.New().String(), EmittedAt: 2000, EmittedBy: "node"},
	}
	session := createTestSessionWithEvents(t, "running", "node", events)

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event3 := Event{ID: "evt-3", Type: "Event", SessionID: session.ID, EmittedAt: 3000, EmittedBy: "node"}
	_ = session.UpdateEventHistorySafe(event3)

	assert.Equal(t, 3, len(session.EventHistory))
	assert.Equal(t, "evt-1", session.EventHistory[0].ID)
	assert.Equal(t, "evt-2", session.EventHistory[1].ID)
	assert.Equal(t, "evt-3", session.EventHistory[2].ID)
}

func TestEventHistory_NoMutation(t *testing.T) {
	events := []Event{
		{ID: "evt-1", Type: "Event", SessionID: uuid.New().String(), EmittedAt: 1000, EmittedBy: "node", Message: "original"},
	}
	session := createTestSessionWithEvents(t, "running", "node", events)

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event2 := Event{ID: "evt-2", Type: "Event", SessionID: session.ID, EmittedAt: 2000, EmittedBy: "node"}
	_ = session.UpdateEventHistorySafe(event2)

	assert.Equal(t, "original", session.EventHistory[0].Message)
}

// Invariants — UpdatedAt Refresh

func TestUpdateEventHistorySafe_RefreshesUpdatedAt(t *testing.T) {
	session := createTestSession(t, "running", "node")
	oldUpdatedAt := session.UpdatedAt

	time.Sleep(1 * time.Second)

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

// Edge Cases

func TestUpdateEventHistorySafe_EmittedAtInFuture(t *testing.T) {
	session := createTestSession(t, "running", "node")
	now := time.Now().Unix()

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: now + 1000000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_LargePayload(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	largePayload := make(map[string]any)
	for i := 0; i < 1000; i++ {
		largePayload[string(rune(i))] = strings.Repeat("x", 10*1024)
	}

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
		Payload:   largePayload,
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
}

func TestUpdateEventHistorySafe_UnicodeInMessage(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: session.ID,
		EmittedAt: 1000,
		EmittedBy: "node",
		Message:   "通知: 完成 🎉",
	}
	err := session.UpdateEventHistorySafe(event)

	assert.NoError(t, err)
	assert.Equal(t, "通知: 完成 🎉", session.EventHistory[0].Message)
}

func TestUpdateEventHistorySafe_SessionIDMismatch(t *testing.T) {
	session := createTestSession(t, "running", "node")

	session.eventStore.On("WriteEvent", mock.Anything).Return(nil)

	event := Event{
		ID:        "evt-1",
		Type:      "Event",
		SessionID: "session-B",
		EmittedAt: 1000,
		EmittedBy: "node",
	}
	err := session.UpdateEventHistorySafe(event)

	// Accepts event even with mismatched SessionID
	assert.NoError(t, err)
	assert.Equal(t, 1, len(session.EventHistory))
}
