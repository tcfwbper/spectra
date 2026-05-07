package runtime_race

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/runtime"
)

// =============================================================================
// Concurrent Behaviour — MessageRouter
// =============================================================================
//
// Production surface required:
//   - runtime.NewMessageRouter(ps, eventProcessor, errorProcessor, terminationNotifier, logger)
//   - runtime.MessageRouter.Handle(sessionUUID, msg) RuntimeResponse
//   - runtime.MessageRouterEventProcessor interface
//   - runtime.MessageRouterErrorProcessor interface
//
// This test verifies that concurrent Handle calls are independent and do not
// race on shared state.
// =============================================================================

func TestMessageRouter_Handle_ConcurrentMessages(t *testing.T) {
	// Setup:
	// - Mock EventProcessor returns success.
	// - Mock ErrorProcessor returns success.
	// - Create one RuntimeMessage with Type()="event" and one with Type()="error".
	sess := &raceSafeMockSession{
		statusResult:       "running",
		currentStateResult: "NodeA",
		sessionDataVal:     "cs-123",
		sessionDataOK:      true,
	}

	ps := runtime.NewPersistentSession(
		sess,
		&raceSafeMetadataStore{},
		&raceSafeEventStore{},
		logger.NewNopLogger(),
	)

	evProcessor := &raceSafeEventProcessor{
		resp: entities.SuccessResponse("ev-ok"),
	}
	errProcessor := &raceSafeErrorProcessor{
		resp: entities.SuccessResponse("err-ok"),
	}
	terminationNotifier := newTerminationChannel()

	mr := runtime.NewMessageRouter(ps, evProcessor, errProcessor, terminationNotifier, logger.NewNopLogger())

	evPayload := json.RawMessage(`{"eventType":"MsgSent","message":"hi","payload":{}}`)
	errPayload := json.RawMessage(`{"message":"something went wrong"}`)
	evMsg := mustNewEventRuntimeMessage(t, "cs-123", evPayload)
	errMsg := mustNewErrorRuntimeMessage(t, "cs-123", errPayload)

	// Act:
	// - Launch two goroutines: one calls mr.Handle with event message, the other with error message.
	var wg sync.WaitGroup
	var resp1, resp2 *entities.RuntimeResponse
	wg.Add(2)
	go func() { defer wg.Done(); resp1 = mr.Handle(testSessionID, evMsg) }()
	go func() { defer wg.Done(); resp2 = mr.Handle(testSessionID, errMsg) }()
	wg.Wait()

	// Assert:
	// - Both calls complete without data race.
	// - Each returns appropriate response.
	assert.NotNil(t, resp1)
	assert.NotNil(t, resp2)
}
