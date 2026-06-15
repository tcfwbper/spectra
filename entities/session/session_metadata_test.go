package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestSessionMetadata_AccessViaSession(t *testing.T) {
	// Scaffolded: NewSession will accept pid parameter; newTestSession helper must be updated
	// to pass pid=42 (or testPid constant). Then assert session.Pid == 42.
	t.Skip("blocked: SessionMetadata.Pid field and NewSession pid parameter do not yet exist — awaiting production surface update")

	s := newTestSession(t)

	assert.Equal(t, testUUID, s.ID)
	assert.Equal(t, testWorkflow, s.WorkflowName)
	// assert.Equal(t, 42, s.Pid) // TODO: uncomment when Pid field exists and newTestSession passes pid=42
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
	// Note: This test is not affected by the pid spec change directly —
	// it tests snapshot detachment, not specific fields. It uses newTestSession
	// which currently compiles. Once pid is added to NewSession, newTestSession
	// will be updated and this test will still be valid.
	s := newTestSession(t)
	snapshot := s.GetMetadataSnapshotSafe()

	err := s.Run()
	require.NoError(t, err)

	// Snapshot should be detached from session — still shows "initializing"
	assert.Equal(t, "initializing", snapshot.Status)
	assert.Equal(t, "running", s.GetStatusSafe())
}
