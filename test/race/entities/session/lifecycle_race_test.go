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

	assert.NoError(t, failErr)
	assert.Equal(t, "failed", s.GetStatusSafe())
	assert.Same(t, agentErr, s.GetErrorSafe())

	if runErr != nil {
		assert.EqualError(t, runErr, "cannot run session: status is 'failed', expected 'initializing'")
	}

	select {
	case <-ch:
		// OK: Fail always sends exactly one termination notification.
	default:
		t.Fatal("expected notification on terminationNotifier channel")
	}

	select {
	case <-ch:
		t.Fatal("unexpected second notification on terminationNotifier channel")
	default:
	}
}
