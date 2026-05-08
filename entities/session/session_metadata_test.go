package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestSessionMetadata_AccessViaSession(t *testing.T) {
	s := newTestSession(t)

	assert.Equal(t, testUUID, s.ID)
	assert.Equal(t, testWorkflow, s.WorkflowName)
	assert.Equal(t, "initializing", s.Status)
	assert.Equal(t, testCreatedAt, s.CreatedAt)
	assert.Equal(t, testCreatedAt, s.UpdatedAt)
	assert.Equal(t, testEntryNode, s.CurrentState)
	assert.NotNil(t, s.SessionData)
	assert.Empty(t, s.SessionData)
	assert.Nil(t, s.Error)
}

// --- Data Independence (Copy Semantics) ---

func TestSessionMetadata_SnapshotIsDetachedCopy(t *testing.T) {
	s := newTestSession(t)
	snapshot := s.GetMetadataSnapshotSafe()

	err := s.Run()
	require.NoError(t, err)

	// Snapshot should be detached from session — still shows "initializing"
	assert.Equal(t, "initializing", snapshot.Status)
	assert.Equal(t, "running", s.GetStatusSafe())
}
