package runtime

import (
	"fmt"
	"os"
	"os/signal"
	goruntime "runtime"
	"syscall"
	"time"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// productionSession is the super-interface that combines all sub-interfaces
// required by the post-session runtime components. The sessionWrapper returned
// by SessionInitializer satisfies this after the delegation methods are added.
type productionSession interface {
	SessionForInitializer
	// AgentInvoker, EventProcessor, ErrorProcessor
	GetSessionDataSafe(key string) (any, bool)
	// AgentInvoker
	UpdateSessionDataSafe(key string, value any) error
	// TransitionToNode
	UpdateCurrentStateSafe(newState string) error
	// EventProcessor
	UpdateEventHistorySafe(event session.Event) error
}

// ProductionRunner implements cmd/spectra.RunRuntime.
// It wires all runtime dependencies per the spec and drives the session lifecycle.
type ProductionRunner struct{}

// NewProductionRunner creates a new ProductionRunner.
func NewProductionRunner() *ProductionRunner {
	return &ProductionRunner{}
}

// Run is the production entry point called by `spectra run`.
func (r *ProductionRunner) Run(workflowName string) error {
	logger := &stdLogger{}

	// Step 2: Locate project root via SpectraFinder.
	projectRoot, err := storage.SpectraFinder("")
	if err != nil {
		return fmt.Errorf("failed to locate project root: %w", err)
	}

	// Step 4: Create terminationNotifier with capacity 2.
	terminationNotifier := make(chan struct{}, 2)

	// Step 5: Construct pre-session dependencies.
	agentDefLoader := storage.NewAgentDefinitionLoader(projectRoot)
	wfLoader := storage.NewWorkflowDefinitionLoader(projectRoot, agentDefLoader)
	dirManager := storage.NewSessionDirectoryManager(projectRoot)

	si, err := NewSessionInitializer(projectRoot, wfLoader, dirManager)
	if err != nil {
		return fmt.Errorf("failed to initialize runtime dependencies: %w", err)
	}

	// Step 7-10: Initialize session.
	sess, err := si.Initialize(workflowName, terminationNotifier)
	if err != nil {
		if sess == nil {
			return fmt.Errorf("failed to initialize session: %w", err)
		}
		sf, _ := NewSessionFinalizer(logger)
		defer func() {
			defer func() { recover() }() //nolint:errcheck
			if sf != nil {
				sf.Finalize(sess)
			}
		}()
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	// SessionFinalizer — always called on exit once session exists.
	sf, _ := NewSessionFinalizer(logger)
	defer func() {
		defer func() { recover() }() //nolint:errcheck
		if sf != nil {
			sf.Finalize(sess)
		}
	}()

	// Upcast to productionSession so we can pass it to all post-session components.
	fullSess, ok := sess.(productionSession)
	if !ok {
		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to initialize post-session dependencies",
			Detail:       []byte(`{"error":"unexpected session type"}`),
			FailingState: sess.GetCurrentStateSafe(),
			SessionID:    parseUUID(sess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = sess.Fail(rtErr, terminationNotifier)
		return fmt.Errorf("failed to initialize post-session dependencies: unexpected session type")
	}

	// Step 11: Construct post-session dependencies.
	sm := storage.NewRuntimeSocketManager(projectRoot, fullSess.GetID())

	agentInvoker, err := NewAgentInvoker(fullSess, projectRoot)
	if err != nil {
		return r.failPostSession(fullSess, terminationNotifier, err)
	}

	wfDef, err := wfLoader.Load(workflowName)
	if err != nil {
		return r.failPostSession(fullSess, terminationNotifier, err)
	}

	ttn, err := NewTransitionToNode(fullSess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	if err != nil {
		return r.failPostSession(fullSess, terminationNotifier, err)
	}

	ep, err := NewEventProcessor(fullSess, wfLoader, ttn, terminationNotifier)
	if err != nil {
		return r.failPostSession(fullSess, terminationNotifier, err)
	}

	errp, err := NewErrorProcessor(fullSess, wfLoader, terminationNotifier)
	if err != nil {
		return r.failPostSession(fullSess, terminationNotifier, err)
	}

	mr, err := NewMessageRouter(fullSess, ep, errp, terminationNotifier)
	if err != nil {
		return r.failPostSession(fullSess, terminationNotifier, err)
	}

	// Step 13: Create runtime socket.
	if err := sm.CreateSocket(); err != nil {
		errMsg := err.Error()
		var enhancedMsg string
		if stringContains(errMsg, "already exists") || stringContains(errMsg, "file already exists") {
			enhancedMsg = fmt.Sprintf("%s. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm %s", errMsg, extractSocketPath(errMsg))
		} else {
			enhancedMsg = errMsg
		}
		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to create runtime socket",
			Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())),
			FailingState: fullSess.GetCurrentStateSafe(),
			SessionID:    parseUUID(fullSess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = fullSess.Fail(rtErr, terminationNotifier)
		return fmt.Errorf("failed to create runtime socket: %s", enhancedMsg)
	}

	// Step 15-17: Start socket listener.
	listenerErrCh, listenerDoneCh, err := sm.Listen(mr.RouteMessage)
	if err != nil {
		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to start socket listener",
			Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())),
			FailingState: fullSess.GetCurrentStateSafe(),
			SessionID:    parseUUID(fullSess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = fullSess.Fail(rtErr, terminationNotifier)
		select {
		case <-listenerDoneCh:
		case <-time.After(2 * time.Second):
			logger.Warning("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
		}
		return fmt.Errorf("failed to start socket listener: %w", err)
	}

	// Step 19: Setup OS signal handling.
	signalCh := make(chan os.Signal, 1)
	if goruntime.GOOS == "windows" {
		signal.Notify(signalCh, os.Interrupt)
	} else {
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	}
	defer signal.Stop(signalCh)

	var receivedSignal os.Signal

	// Step 18-25: Main event loop.
	select {
	case <-terminationNotifier:
		logger.Log("received session termination notification")
	case listenerErr := <-listenerErrCh:
		logger.Log(fmt.Sprintf("listener error: %v", listenerErr))
		status := fullSess.GetStatusSafe()
		if status != "completed" && status != "failed" {
			rtErr := &entities.RuntimeError{
				Issuer:       "Runtime",
				Message:      "listener error",
				Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", listenerErr.Error())),
				FailingState: fullSess.GetCurrentStateSafe(),
				SessionID:    parseUUID(fullSess.GetID()),
				OccurredAt:   time.Now().Unix(),
			}
			_ = fullSess.Fail(rtErr, terminationNotifier)
		}
	case sig := <-signalCh:
		receivedSignal = sig
		logger.Log(fmt.Sprintf("received signal %v, initiating graceful shutdown", sig))
	case <-listenerDoneCh:
		logger.Log("listener closed")
	}

	// Step 26-27: Grace period with second-signal / timeout force-exit.
	gracePeriodCh := make(chan struct{})
	forceExitCh := make(chan struct{})
	secondSignalCh := make(chan os.Signal, 1)

	gracePeriodTimer := time.After(5 * time.Second)
	go func() {
		signal.Notify(secondSignalCh, os.Interrupt, syscall.SIGTERM)
		select {
		case <-secondSignalCh:
			logger.Log("received second signal, forcing exit")
			close(forceExitCh)
			os.Exit(1)
		case <-gracePeriodTimer:
			logger.Warning("cleanup exceeded 5 second grace period, forcing exit")
			close(forceExitCh)
			os.Exit(1)
		case <-gracePeriodCh:
		}
	}()

	// Step 28-29: Delete socket.
	_ = sm.DeleteSocket()

	// Step 30-31: Wait for listener shutdown.
	select {
	case <-listenerDoneCh:
	case <-time.After(2 * time.Second):
		logger.Warning("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
	case <-forceExitCh:
		return determineReturnValue(fullSess, receivedSignal)
	}

	close(gracePeriodCh)

	// Step 34: Return value.
	return determineReturnValue(fullSess, receivedSignal)
}

// failPostSession marks the session as failed with a RuntimeError and returns the formatted error.
func (r *ProductionRunner) failPostSession(sess SessionForInitializer, tn chan<- struct{}, cause error) error {
	rtErr := &entities.RuntimeError{
		Issuer:       "Runtime",
		Message:      "failed to initialize post-session dependencies",
		Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", cause.Error())),
		FailingState: sess.GetCurrentStateSafe(),
		SessionID:    parseUUID(sess.GetID()),
		OccurredAt:   time.Now().Unix(),
	}
	_ = sess.Fail(rtErr, tn)
	return fmt.Errorf("failed to initialize post-session dependencies: %w", cause)
}

// stdLogger writes to stdout (Log) and stderr (Warning).
type stdLogger struct{}

func (l *stdLogger) Log(msg string)     { fmt.Fprintln(os.Stdout, msg) }
func (l *stdLogger) Warning(msg string) { fmt.Fprintln(os.Stderr, msg) }
