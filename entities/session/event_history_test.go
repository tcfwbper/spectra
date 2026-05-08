package session

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — UpdateEventHistorySafe ---

func TestUpdateEventHistorySafe_ValidEvent(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEvent(t, testEventID)

	err := s.UpdateEventHistorySafe(*ev)

	require.NoError(t, err)
	assert.Len(t, s.EventHistory, 1)
	assert.Equal(t, testEventID, s.EventHistory[0].ID())
}

func TestUpdateEventHistorySafe_MultipleEvents(t *testing.T) {
	s := newTestSession(t)
	e1 := newTestEvent(t, testEventID)
	e2 := newTestEvent(t, testEventID2)

	err1 := s.UpdateEventHistorySafe(*e1)
	err2 := s.UpdateEventHistorySafe(*e2)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Len(t, s.EventHistory, 2)
	assert.Equal(t, testEventID, s.EventHistory[0].ID())
	assert.Equal(t, testEventID2, s.EventHistory[1].ID())
}

func TestUpdateEventHistorySafe_EmptyMessage(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventMinimal(t, testEventID, "")

	err := s.UpdateEventHistorySafe(*ev)

	require.NoError(t, err)
	assert.Len(t, s.EventHistory, 1)
}

func TestUpdateEventHistorySafe_EmptyPayload(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventEmptyPayload(t, testEventID)

	err := s.UpdateEventHistorySafe(*ev)

	require.NoError(t, err)
	assert.Len(t, s.EventHistory, 1)
}

func TestUpdateEventHistorySafe_UpdatesUpdatedAt(t *testing.T) {
	s := newTestSession(t)
	initialUpdatedAt := s.GetMetadataSnapshotSafe().UpdatedAt
	ev := newTestEvent(t, testEventID)

	err := s.UpdateEventHistorySafe(*ev)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, s.GetMetadataSnapshotSafe().UpdatedAt, initialUpdatedAt)
}

// --- Validation Failures ---

func TestUpdateEventHistorySafe_MissingID(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventUnchecked(t, "", testEventType, "msg", json.RawMessage(`{}`), testEmittedBy, testEmittedAt, testSessionID)

	err := s.UpdateEventHistorySafe(ev)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
	assert.Empty(t, s.EventHistory)
}

func TestUpdateEventHistorySafe_MissingType(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventUnchecked(t, testEventID, "", "msg", json.RawMessage(`{}`), testEmittedBy, testEmittedAt, testSessionID)

	err := s.UpdateEventHistorySafe(ev)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Type is required")
	assert.Empty(t, s.EventHistory)
}

func TestUpdateEventHistorySafe_MissingSessionID(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventUnchecked(t, testEventID, testEventType, "msg", json.RawMessage(`{}`), testEmittedBy, testEmittedAt, "")

	err := s.UpdateEventHistorySafe(ev)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SessionID is required")
	assert.Empty(t, s.EventHistory)
}

func TestUpdateEventHistorySafe_InvalidEmittedAt(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventUnchecked(t, testEventID, testEventType, "msg", json.RawMessage(`{}`), testEmittedBy, 0, testSessionID)

	err := s.UpdateEventHistorySafe(ev)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "EmittedAt is required")
	assert.Empty(t, s.EventHistory)
}

func TestUpdateEventHistorySafe_MissingEmittedBy(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEventUnchecked(t, testEventID, testEventType, "msg", json.RawMessage(`{}`), "", testEmittedAt, testSessionID)

	err := s.UpdateEventHistorySafe(ev)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "EmittedBy is required")
	assert.Empty(t, s.EventHistory)
}

func TestUpdateEventHistorySafe_ValidationOrder(t *testing.T) {
	s := newTestSession(t)
	// All required fields are invalid — ID should be checked first per spec.
	ev := newTestEventUnchecked(t, "", "", "msg", json.RawMessage(`{}`), "", 0, "")

	err := s.UpdateEventHistorySafe(ev)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
	assert.Empty(t, s.EventHistory)
}

// --- Idempotency ---

func TestUpdateEventHistorySafe_DuplicateEventID(t *testing.T) {
	s := newTestSession(t)
	ev := newTestEvent(t, testEventID)

	err1 := s.UpdateEventHistorySafe(*ev)
	err2 := s.UpdateEventHistorySafe(*ev)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Len(t, s.EventHistory, 2)
}
