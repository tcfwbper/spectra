package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — UpdateCurrentStateSafe ---

func TestUpdateCurrentStateSafe_ValidState(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateCurrentStateSafe("processing")

	require.NoError(t, err)
	assert.Equal(t, "processing", s.GetCurrentStateSafe())
}

func TestUpdateCurrentStateSafe_SelfTransition(t *testing.T) {
	s := newTestSession(t)
	initialUpdatedAt := s.UpdatedAt

	err := s.UpdateCurrentStateSafe("start")

	require.NoError(t, err)
	assert.Equal(t, "start", s.GetCurrentStateSafe())
	assert.GreaterOrEqual(t, s.GetMetadataSnapshotSafe().UpdatedAt, initialUpdatedAt)
}

func TestUpdateCurrentStateSafe_UpdatesUpdatedAt(t *testing.T) {
	s := newTestSession(t)
	initialUpdatedAt := s.GetMetadataSnapshotSafe().UpdatedAt

	err := s.UpdateCurrentStateSafe("next")

	require.NoError(t, err)
	assert.GreaterOrEqual(t, s.GetMetadataSnapshotSafe().UpdatedAt, initialUpdatedAt)
}

// --- Validation Failures ---

func TestUpdateCurrentStateSafe_EmptyState(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateCurrentStateSafe("")

	require.Error(t, err)
	assert.Equal(t, "current state cannot be empty", err.Error())
	assert.Equal(t, "start", s.GetCurrentStateSafe())
}
