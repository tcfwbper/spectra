package session_race

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunAndFail_Concurrent(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr := newTestAgentError(t)

	var wg sync.WaitGroup
	wg.Add(2)

	var runErr, failErr error

	go func() {
		defer wg.Done()
		runErr = s.Run()
	}()

	go func() {
		defer wg.Done()
		failErr = s.Fail(agentErr, ch)
	}()

	wg.Wait()

	// Exactly one should succeed; the other should return a precondition error
	if runErr == nil {
		// Run succeeded => Fail should have returned an error
		assert.Error(t, failErr)
		assert.Equal(t, "running", s.GetStatusSafe())
	} else if failErr == nil {
		// Fail succeeded => Run should have returned an error
		assert.Error(t, runErr)
		assert.Equal(t, "failed", s.GetStatusSafe())
		assert.NotNil(t, s.GetErrorSafe())
	} else {
		// Both failed — should not happen with a valid session in "initializing"
		t.Fatalf("both Run and Fail returned errors: run=%v, fail=%v", runErr, failErr)
	}
}
