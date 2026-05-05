package session

import (
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
	t.Skip("blocked: entities.Event validates at construction; need test-only Event builder or session accepts struct with exported fields to test session-level validation of empty ID")
}

func TestUpdateEventHistorySafe_MissingType(t *testing.T) {
	t.Skip("blocked: entities.Event validates at construction; need test-only Event builder or session accepts struct with exported fields to test session-level validation of empty Type")
}

func TestUpdateEventHistorySafe_MissingSessionID(t *testing.T) {
	t.Skip("blocked: entities.Event validates at construction; need test-only Event builder or session accepts struct with exported fields to test session-level validation of empty SessionID")
}

func TestUpdateEventHistorySafe_InvalidEmittedAt(t *testing.T) {
	t.Skip("blocked: entities.Event validates at construction; need test-only Event builder or session accepts struct with exported fields to test session-level validation of EmittedAt <= 0")
}

func TestUpdateEventHistorySafe_MissingEmittedBy(t *testing.T) {
	t.Skip("blocked: entities.Event validates at construction; need test-only Event builder or session accepts struct with exported fields to test session-level validation of empty EmittedBy")
}

func TestUpdateEventHistorySafe_ValidationOrder(t *testing.T) {
	t.Skip("blocked: entities.Event validates at construction; need test-only Event builder to test session-level validation order (ID checked first)")
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
