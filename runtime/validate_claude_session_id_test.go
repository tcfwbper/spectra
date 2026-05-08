package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Happy Path — ValidateClaudeSessionID
// =============================================================================

func TestValidateClaudeSessionID_AgentNode_Matches(t *testing.T) {
	// Setup: mock PersistentSession with GetSessionDataSafe("myAgent.ClaudeSessionID") -> ("abc-123", true)
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = "abc-123"
	sess.getSessionDataResultOK = true
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Setup: mock Node with Type()="agent", Name()="myAgent"
	node := &mockNode{nodeType: "agent", nodeName: "myAgent"}

	// Act
	err := ValidateClaudeSessionID(ps, node, "abc-123")

	// Assert: returns nil
	require.NoError(t, err)
}

func TestValidateClaudeSessionID_HumanNode_EmptyID(t *testing.T) {
	// Setup: mock PersistentSession (not accessed for human node validation)
	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Setup: mock Node with Type()="human"
	node := &mockNode{nodeType: "human", nodeName: "humanNode"}

	// Act
	err := ValidateClaudeSessionID(ps, node, "")

	// Assert: returns nil
	require.NoError(t, err)
}

// =============================================================================
// Error Propagation — ValidateClaudeSessionID
// =============================================================================

func TestValidateClaudeSessionID_AgentNode_KeyNotFound(t *testing.T) {
	// Setup: mock PersistentSession with GetSessionDataSafe("myAgent.ClaudeSessionID") -> (nil, false)
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Setup: mock Node with Type()="agent", Name()="myAgent"
	node := &mockNode{nodeType: "agent", nodeName: "myAgent"}

	// Act
	err := ValidateClaudeSessionID(ps, node, "abc-123")

	// Assert: returns error with expected message
	require.Error(t, err)
	assert.Equal(t, "claude session ID not found for node 'myAgent'", err.Error())
}

func TestValidateClaudeSessionID_AgentNode_Mismatch(t *testing.T) {
	// Setup: mock PersistentSession with GetSessionDataSafe("myAgent.ClaudeSessionID") -> ("expected-id", true)
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = "expected-id"
	sess.getSessionDataResultOK = true
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Setup: mock Node with Type()="agent", Name()="myAgent"
	node := &mockNode{nodeType: "agent", nodeName: "myAgent"}

	// Act
	err := ValidateClaudeSessionID(ps, node, "wrong-id")

	// Assert: returns error with expected message
	require.Error(t, err)
	assert.Equal(t, "claude session ID mismatch: expected expected-id but got wrong-id", err.Error())
}

func TestValidateClaudeSessionID_HumanNode_NonEmptyID(t *testing.T) {
	// Setup: mock PersistentSession (not accessed for human node validation)
	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Setup: mock Node with Type()="human"
	node := &mockNode{nodeType: "human", nodeName: "humanNode"}

	// Act
	err := ValidateClaudeSessionID(ps, node, "some-id")

	// Assert: returns error with expected message
	require.Error(t, err)
	assert.Equal(t, "invalid claude session ID for human node: must be empty", err.Error())
}

func TestValidateClaudeSessionID_UnsupportedNodeType(t *testing.T) {
	// Setup: mock PersistentSession (not accessed for unknown node type)
	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	// Setup: mock Node with Type()="unknown"
	node := &mockNode{nodeType: "unknown", nodeName: "unknownNode"}

	// Act
	err := ValidateClaudeSessionID(ps, node, "any")

	// Assert: returns error with expected message
	require.Error(t, err)
	assert.Equal(t, "unsupported node type 'unknown'", err.Error())
}
