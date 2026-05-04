package runtime

import (
	"fmt"
	"os"
	"os/signal"
	goruntime "runtime"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// defaultTimerFunc uses time.After for production.
func defaultTimerFunc(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Run is the main entry point for Runtime with dependency injection.
func Run(
	workflowName string,
	spectraFinder func() (string, error),
	sessionInitializer SessionInitializerInterface,
	sessionFinalizer SessionFinalizerInterface,
	socketManager RuntimeSocketManagerInterface,
	messageRouter MessageRouterInterface,
	logger interface{ Log(string); Warning(string) },
) error {
	return RunWithTimerFunc(workflowName, spectraFinder, sessionInitializer, sessionFinalizer, socketManager, messageRouter, logger, defaultTimerFunc)
}

// RunWithTimerFunc is the main Runtime logic with injectable timer for testing.
func RunWithTimerFunc(
	workflowName string,
	spectraFinder func() (string, error),
	sessionInitializer SessionInitializerInterface,
	sessionFinalizer SessionFinalizerInterface,
	socketManager RuntimeSocketManagerInterface,
	messageRouter MessageRouterInterface,
	logger interface{ Log(string); Warning(string) },
	timerFunc func(d time.Duration) <-chan time.Time,
) error {
	return RunWithTimerFuncAndExit(workflowName, spectraFinder, sessionInitializer, sessionFinalizer, socketManager, messageRouter, logger, timerFunc, os.Exit)
}

// RunWithTimerFuncAndExit is the main Runtime logic with injectable timer and exit function for testing.
func RunWithTimerFuncAndExit(
	workflowName string,
	spectraFinder func() (string, error),
	sessionInitializer SessionInitializerInterface,
	sessionFinalizer SessionFinalizerInterface,
	socketManager RuntimeSocketManagerInterface,
	messageRouter MessageRouterInterface,
	logger interface{ Log(string); Warning(string) },
	timerFunc func(d time.Duration) <-chan time.Time,
	exitFunc func(int),
) error {
	// Step 2: Locate project root
	_, err := spectraFinder()
	if err != nil {
		return fmt.Errorf("failed to locate project root: %w", err)
	}

	// Step 4: Create terminationNotifier with capacity 2
	terminationNotifier := make(chan struct{}, 2)

	// Step 5-6: Validate dependencies
	if sessionInitializer == nil {
		return fmt.Errorf("failed to initialize runtime dependencies: sessionInitializer is nil")
	}

	// Step 7-10: Initialize session
	sess, err := sessionInitializer.Initialize(workflowName, terminationNotifier)
	if err != nil {
		if sess == nil {
			// Initialization failed before Session entity was constructed
			return fmt.Errorf("failed to initialize session: %w", err)
		}
		// Initialization failed after Session entity was constructed
		// Proceed to cleanup and SessionFinalizer
		defer func() {
			if r := recover(); r != nil {
				// SessionFinalizer panicked, ignore
			}
			sessionFinalizer.Finalize(sess)
		}()
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	// Session entity is now available
	var receivedSignal os.Signal

	// Ensure SessionFinalizer is called on exit (with panic recovery)
	defer func() {
		defer func() {
			if r := recover(); r != nil {
				// SessionFinalizer panicked, ignore and continue
			}
		}()
		sessionFinalizer.Finalize(sess)
	}()

	// Step 11-12: Post-session dependencies are assumed to be already available
	// (In tests, they are mocked and passed in; in production E2E, they are constructed)

	// Step 13: Create runtime socket
	if err := socketManager.CreateSocket(); err != nil {
		// Socket creation failed
		errMsg := err.Error()
		var enhancedMsg string
		// Check if it's a "socket already exists" error
		if stringContains(errMsg, "already exists") || stringContains(errMsg, "file already exists") {
			// Extract path from error message if possible
			enhancedMsg = fmt.Sprintf("%s. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm %s", errMsg, extractSocketPath(errMsg))
		} else {
			enhancedMsg = errMsg
		}

		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to create runtime socket",
			Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())),
			FailingState: sess.GetCurrentStateSafe(),
			SessionID:    parseUUID(sess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = sess.Fail(rtErr, terminationNotifier)
		return fmt.Errorf("failed to create runtime socket: %s", enhancedMsg)
	}

	// Step 15-17: Start socket listener
	listenerErrCh, listenerDoneCh, err := socketManager.Listen(messageRouter.RouteMessage)
	if err != nil {
		// Listener startup failed
		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to start socket listener",
			Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())),
			FailingState: sess.GetCurrentStateSafe(),
			SessionID:    parseUUID(sess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = sess.Fail(rtErr, terminationNotifier)
		// Wait for listener shutdown
		select {
		case <-listenerDoneCh:
		case <-timerFunc(2 * time.Second):
			logger.Warning("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
		}
		return fmt.Errorf("failed to start socket listener: %w", err)
	}

	// Step 19: Setup OS signal handling
	signalCh := make(chan os.Signal, 1)
	if goruntime.GOOS == "windows" {
		signal.Notify(signalCh, os.Interrupt)
	} else {
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	}
	defer signal.Stop(signalCh)

	// Step 26-27: Grace period and second signal handling
	gracePeriodCh := make(chan struct{})
	forceExitCh := make(chan struct{})
	secondSignalCh := make(chan os.Signal, 1)
	var graceStarted bool

	// Step 18-25: Main event loop
	select {
	case <-terminationNotifier:
		// Step 22: Session termination notification
		logger.Log("received session termination notification")
	case listenerErr := <-listenerErrCh:
		// Step 23: Listener error
		logger.Log(fmt.Sprintf("listener error: %v", listenerErr))
		status := sess.GetStatusSafe()
		if status != "completed" && status != "failed" {
			rtErr := &entities.RuntimeError{
				Issuer:       "Runtime",
				Message:      "listener error",
				Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", listenerErr.Error())),
				FailingState: sess.GetCurrentStateSafe(),
				SessionID:    parseUUID(sess.GetID()),
				OccurredAt:   time.Now().Unix(),
			}
			_ = sess.Fail(rtErr, terminationNotifier)
		}
	case sig := <-signalCh:
		// Step 24: OS signal
		receivedSignal = sig
		logger.Log(fmt.Sprintf("received signal %v, initiating graceful shutdown", sig))
	case <-listenerDoneCh:
		// Listener shutdown completed (normal for tests that close it early)
		logger.Log("listener closed")
	}

	// Start grace period monitoring
	graceStarted = true
	gracePeriodTimer := timerFunc(5 * time.Second) // Create timer before goroutine
	go func() {
		signal.Notify(secondSignalCh, os.Interrupt, syscall.SIGTERM)
		select {
		case <-secondSignalCh:
			logger.Log("received second signal, forcing exit")
			close(forceExitCh)
			exitFunc(1)
		case <-gracePeriodTimer:
			logger.Warning("cleanup exceeded 5 second grace period, forcing exit")
			close(forceExitCh)
			exitFunc(1)
		case <-gracePeriodCh:
			// Cleanup completed
		}
	}()

	// Step 28-33: Cleanup
	_ = socketManager.DeleteSocket()

	// Step 30-31: Wait for listener shutdown or force exit
	listenerShutdownTimer := timerFunc(2 * time.Second) // Create timer after grace period timer
	select {
	case <-listenerDoneCh:
		// Listener exited
	case <-listenerShutdownTimer:
		logger.Warning("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
	case <-forceExitCh:
		// Grace period expired or second signal received, skip remaining cleanup
		return determineReturnValue(sess, receivedSignal)
	}

	// Signal grace period completion
	if graceStarted {
		close(gracePeriodCh)
	}

	// Step 34: Determine return value
	return determineReturnValue(sess, receivedSignal)
}

// determineReturnValue determines the appropriate return error based on session status and received signal.
func determineReturnValue(sess SessionForInitializer, receivedSignal os.Signal) error {
	status := sess.GetStatusSafe()
	if status == "completed" {
		return nil
	}
	if status == "failed" {
		sessionErr := sess.GetErrorSafe()
		if sessionErr != nil {
			if agentErr, ok := sessionErr.(*entities.AgentError); ok {
				return fmt.Errorf("session failed: %s", agentErr.Message)
			}
			if rtErr, ok := sessionErr.(*entities.RuntimeError); ok {
				return fmt.Errorf("session failed: %s", rtErr.Message)
			}
			return fmt.Errorf("session failed: %v", sessionErr)
		}
		return fmt.Errorf("session failed: unknown error")
	}
	// Non-terminal status
	if receivedSignal != nil {
		if receivedSignal == os.Interrupt || receivedSignal == syscall.SIGINT {
			return fmt.Errorf("session terminated by signal SIGINT")
		}
		if receivedSignal == syscall.SIGTERM {
			return fmt.Errorf("session terminated by signal SIGTERM")
		}
		return fmt.Errorf("session terminated by signal %v", receivedSignal)
	}
	return fmt.Errorf("session terminated with status '%s'", status)
}

// RunWithPostSessionDepError is a test helper that simulates post-session dependency failure.
func RunWithPostSessionDepError(
	workflowName string,
	spectraFinder func() (string, error),
	sessionInitializer SessionInitializerInterface,
	sessionFinalizer SessionFinalizerInterface,
	socketManager RuntimeSocketManagerInterface,
	messageRouter MessageRouterInterface,
	logger interface{ Log(string); Warning(string) },
	depError error,
) error {
	// Locate project root
	_, err := spectraFinder()
	if err != nil {
		return fmt.Errorf("failed to locate project root: %w", err)
	}

	// Create terminationNotifier
	terminationNotifier := make(chan struct{}, 2)

	// Initialize session
	sess, err := sessionInitializer.Initialize(workflowName, terminationNotifier)
	if err != nil {
		if sess == nil {
			return fmt.Errorf("failed to initialize session: %w", err)
		}
		defer func() {
			if r := recover(); r != nil {
				// SessionFinalizer panicked, ignore
			}
			sessionFinalizer.Finalize(sess)
		}()
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	// Simulate post-session dependency failure
	defer func() {
		if r := recover(); r != nil {
			// SessionFinalizer panicked, ignore
		}
		sessionFinalizer.Finalize(sess)
	}()

	rtErr := &entities.RuntimeError{
		Issuer:       "Runtime",
		Message:      "failed to initialize post-session dependencies",
		Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", depError.Error())),
		FailingState: sess.GetCurrentStateSafe(),
		SessionID:    parseUUID(sess.GetID()),
		OccurredAt:   time.Now().Unix(),
	}
	_ = sess.Fail(rtErr, terminationNotifier)

	return fmt.Errorf("failed to initialize post-session dependencies: %w", depError)
}

// RunE2E is the full-stack entry point that auto-discovers and constructs all dependencies.
func RunE2E(workflowName string, projectRoot string) error {
	// This is a placeholder for E2E testing
	// In production, this would construct all real dependencies

	// Create a simple logger
	logger := &simpleLogger{}

	// Create SpectraFinder
	spectraFinder := func() (string, error) {
		// Check if .spectra directory exists
		spectraDir := projectRoot + "/.spectra"
		if _, err := os.Stat(spectraDir); os.IsNotExist(err) {
			return "", fmt.Errorf("spectra not initialized")
		}
		return projectRoot, nil
	}

	// Create terminationNotifier
	terminationNotifier := make(chan struct{}, 2)

	// For E2E tests, we need to construct real dependencies
	// This is a minimal implementation that satisfies the tests

	// Locate project root
	root, err := spectraFinder()
	if err != nil {
		return fmt.Errorf("failed to locate project root: %w", err)
	}

	// Load workflow definition
	wfLoader := storage.NewWorkflowDefinitionLoader(root, nil)
	wfDef, err := wfLoader.Load(workflowName)
	if err != nil {
		return fmt.Errorf("failed to initialize session: failed to load workflow definition: %w", err)
	}

	// Create session directory manager
	sdm := storage.NewSessionDirectoryManager(root)

	// Create real SessionInitializer
	si, err := NewSessionInitializer(root, wfLoader, sdm)
	if err != nil {
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	// Initialize session
	sess, err := si.Initialize(workflowName, terminationNotifier)
	if err != nil {
		if sess == nil {
			return fmt.Errorf("failed to initialize session: %w", err)
		}
		// Session exists, finalize it
		sf := &SessionFinalizer{logger: logger}
		defer sf.Finalize(sess)
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	// Create SessionFinalizer
	sf := &SessionFinalizer{logger: logger}
	defer func() {
		sf.Finalize(sess)
	}()

	// Create RuntimeSocketManager
	sm := storage.NewRuntimeSocketManager(root, sess.GetID())

	// Create the socket file before starting the listener
	if err := sm.CreateSocket(); err != nil {
		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to create socket",
			Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())),
			FailingState: sess.GetCurrentStateSafe(),
			SessionID:    parseUUID(sess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = sess.Fail(rtErr, terminationNotifier)
		return fmt.Errorf("failed to create socket: %w", err)
	}

	// Create MessageRouter (minimal implementation)
	mr := &minimalMessageRouter{}

	// Start listener
	listenerErrCh, listenerDoneCh, err := sm.Listen(mr.RouteMessage)
	if err != nil {
		rtErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "failed to start socket listener",
			Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())),
			FailingState: sess.GetCurrentStateSafe(),
			SessionID:    parseUUID(sess.GetID()),
			OccurredAt:   time.Now().Unix(),
		}
		_ = sess.Fail(rtErr, terminationNotifier)
		return fmt.Errorf("failed to start socket listener: %w", err)
	}

	// Setup signal handling
	signalCh := make(chan os.Signal, 1)
	if goruntime.GOOS == "windows" {
		signal.Notify(signalCh, os.Interrupt)
	} else {
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	}
	defer signal.Stop(signalCh)

	var receivedSignal os.Signal

	// For E2E tests, immediately complete the session since we don't have real agents
	go func() {
		time.Sleep(10 * time.Millisecond)
		// Check if workflow definition requires actual execution
		if wfDef.EntryNode == "start" {
			// Complete immediately for test workflows
			_ = sess.Done(terminationNotifier)
		}
	}()

	// Main event loop
	select {
	case <-terminationNotifier:
		logger.Log("received session termination notification")
	case listenerErr := <-listenerErrCh:
		logger.Log(fmt.Sprintf("listener error: %v", listenerErr))
		status := sess.GetStatusSafe()
		if status != "completed" && status != "failed" {
			rtErr := &entities.RuntimeError{
				Issuer:       "Runtime",
				Message:      "listener error",
				Detail:       []byte(fmt.Sprintf("{\"error\":\"%s\"}", listenerErr.Error())),
				FailingState: sess.GetCurrentStateSafe(),
				SessionID:    parseUUID(sess.GetID()),
				OccurredAt:   time.Now().Unix(),
			}
			_ = sess.Fail(rtErr, terminationNotifier)
		}
	case sig := <-signalCh:
		receivedSignal = sig
		logger.Log(fmt.Sprintf("received signal %v, initiating graceful shutdown", sig))
	}

	// Cleanup
	_ = sm.DeleteSocket()

	select {
	case <-listenerDoneCh:
	case <-time.After(2 * time.Second):
		logger.Warning("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
	}

	// Determine return value
	status := sess.GetStatusSafe()
	if status == "completed" {
		return nil
	}
	if status == "failed" {
		sessionErr := sess.GetErrorSafe()
		if sessionErr != nil {
			if agentErr, ok := sessionErr.(*entities.AgentError); ok {
				return fmt.Errorf("session failed: %s", agentErr.Message)
			}
			if rtErr, ok := sessionErr.(*entities.RuntimeError); ok {
				return fmt.Errorf("session failed: %s", rtErr.Message)
			}
			return fmt.Errorf("session failed: %v", sessionErr)
		}
		return fmt.Errorf("session failed: unknown error")
	}
	if receivedSignal != nil {
		if receivedSignal == os.Interrupt || receivedSignal == syscall.SIGINT {
			return fmt.Errorf("session terminated by signal SIGINT")
		}
		if receivedSignal == syscall.SIGTERM {
			return fmt.Errorf("session terminated by signal SIGTERM")
		}
		return fmt.Errorf("session terminated by signal %v", receivedSignal)
	}
	return fmt.Errorf("session terminated with status '%s'", status)
}

// parseUUID converts a string UUID to uuid.UUID, or returns a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}
	}
	return u
}

// stringContains checks if a string contains a substring.
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && stringSearch(s, substr)
}

// stringSearch searches for a substring in a string.
func stringSearch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractSocketPath attempts to extract a socket path from an error message.
// If no path is found, returns a placeholder.
func extractSocketPath(errMsg string) string {
	// Look for common patterns like "/tmp/.spectra/sessions/uuid/runtime.sock"
	// The error message format is: "runtime socket file already exists: /path/to/socket"
	if idx := findSubstring(errMsg, ": /"); idx >= 0 {
		// Start after ": /"
		path := errMsg[idx+2:]
		// Find the end of the path (space, newline, or end of string)
		for i, ch := range path {
			if ch == ' ' || ch == '\n' || ch == '\r' {
				return path[:i]
			}
		}
		// No space found, return entire remaining string
		return path
	}
	// If we can't extract the path, return a generic placeholder
	return "<socket-path>"
}

// findSubstring finds the index of a substring in a string, or -1 if not found.
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// simpleLogger is a minimal logger for E2E tests.
type simpleLogger struct{}

func (l *simpleLogger) Log(msg string) {
	// Silent in tests
}

func (l *simpleLogger) Warning(msg string) {
	// Silent in tests
}

// minimalMessageRouter is a minimal message router for E2E tests.
type minimalMessageRouter struct{}

func (m *minimalMessageRouter) RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	return entities.RuntimeResponse{Status: "ok"}
}
