package components_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/components"
)

// assertErrorMatches checks that an error is not nil and its message matches the given regex pattern
func assertErrorMatches(t *testing.T, err error, pattern string) {
	t.Helper()
	require.Error(t, err)
	re := regexp.MustCompile(pattern)
	require.True(t, re.MatchString(err.Error()), "error message %q does not match pattern %q", err.Error(), pattern)
}

// createNode is a helper to create a node for testing
func createNode(t *testing.T, name, nodeType, agentRole, description string) *components.Node {
	t.Helper()
	node, err := components.NewNode(name, nodeType, agentRole, description)
	require.NoError(t, err)
	return node
}

// createNodeExpectError is a helper to create a node expecting an error
func createNodeExpectError(t *testing.T, name, nodeType, agentRole, description string) error {
	t.Helper()
	_, err := components.NewNode(name, nodeType, agentRole, description)
	require.Error(t, err)
	return err
}

// createTransition is a helper to create a transition for testing
func createTransition(t *testing.T, fromNode, eventType, toNode string) *components.Transition {
	t.Helper()
	transition, err := components.NewTransition(fromNode, eventType, toNode)
	require.NoError(t, err)
	return transition
}

// createTransitionExpectError is a helper to create a transition expecting an error
func createTransitionExpectError(t *testing.T, fromNode, eventType, toNode string) error {
	t.Helper()
	_, err := components.NewTransition(fromNode, eventType, toNode)
	require.Error(t, err)
	return err
}

// createExitTransition is a helper to create an exit transition for testing
func createExitTransition(t *testing.T, fromNode, eventType, toNode string) *components.ExitTransition {
	t.Helper()
	exitTransition, err := components.NewExitTransition(fromNode, eventType, toNode)
	require.NoError(t, err)
	return exitTransition
}

// createExitTransitionExpectError is a helper to create an exit transition expecting an error
func createExitTransitionExpectError(t *testing.T, fromNode, eventType, toNode string) error {
	t.Helper()
	_, err := components.NewExitTransition(fromNode, eventType, toNode)
	require.Error(t, err)
	return err
}
