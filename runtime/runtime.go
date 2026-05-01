package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SpectraFinderInterface defines the interface for locating the project root.
type SpectraFinderInterface interface {
	Find() (string, error)
}

// Runtime is the top-level entry point and main loop for the workflow execution system.
type Runtime struct {
	spectraFinder      SpectraFinderInterface
	sessionInitializer SessionInitializerInterface
	sessionFinalizer   SessionFinalizerInterface
	socketManager      RuntimeSocketManagerInterface
}

// NewRuntime creates a new Runtime instance.
func NewRuntime(
	spectraFinder SpectraFinderInterface,
	sessionInitializer SessionInitializerInterface,
	sessionFinalizer SessionFinalizerInterface,
	socketManager RuntimeSocketManagerInterface,
) *Runtime {
	return &Runtime{
		spectraFinder:      spectraFinder,
		sessionInitializer: sessionInitializer,
		sessionFinalizer:   sessionFinalizer,
		socketManager:      socketManager,
	}
}

// Run executes the main runtime loop for the given workflow.
func (rt *Runtime) Run(workflowName string) int {
	// Locate project root
	projectRoot, err := rt.spectraFinder.Find()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to locate project root: %v. Run 'spectra init' to initialize the project.\n", err)
		return 1
	}

	// Create termination notifier channel with buffer size 2
	terminationNotifier := make(chan struct{}, 2)

	// Initialize session
	sess, err := rt.sessionInitializer.Initialize(workflowName, projectRoot, terminationNotifier)
	if err != nil {
		// Check if session entity exists
		if sess == nil {
			// No session entity - print to stderr and exit
			fmt.Fprintf(os.Stderr, "Failed to initialize session: %v\n", err)
			return 1
		}
		// Session entity exists - proceed to finalization
		rt.sessionFinalizer.Finalize(sess)
		return 1
	}

	// Check if session is already in terminal state
	status := sess.GetStatusSafe()
	if status == "completed" || status == "failed" {
		// Session already terminated - skip socket setup and proceed to finalization
		rt.sessionFinalizer.Finalize(sess)
		if status == "completed" {
			return 0
		}
		return 1
	}

	// Create MessageRouter
	eventProcessor := &EventProcessor{}
	errorProcessor := &ErrorProcessor{}
	messageRouter, err := NewMessageRouter(sess, eventProcessor, errorProcessor, terminationNotifier)
	if err != nil {
		// This should not happen with valid inputs
		fmt.Fprintf(os.Stderr, "Failed to create MessageRouter: %v\n", err)
		rt.sessionFinalizer.Finalize(sess)
		return 1
	}

	// Start socket listener
	listenerErrCh, listenerDoneCh, syncErr := rt.socketManager.Listen(messageRouter.RouteMessage)
	if syncErr != nil {
		// Synchronous setup failure - listener goroutine was never spawned
		_ = sess.Fail(syncErr, terminationNotifier)
		// Wait for listenerDoneCh to close (it should already be closed)
		<-listenerDoneCh
		rt.sessionFinalizer.Finalize(sess)
		return 1
	}

	// Setup OS signal handling
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Track whether we've received a signal
	signalReceived := false

	// Main monitoring loop
	terminated := false

	for !terminated {
		select {
		case <-terminationNotifier:
			// Session completed or failed
			terminated = true

		case err := <-listenerErrCh:
			// Asynchronous listener error
			if !terminated {
				_ = sess.Fail(err, terminationNotifier)
				// Continue loop to wait for termination notification
			}

		case sig := <-signalChan:
			if signalReceived {
				// Second signal - exit immediately with code 130
				return 130
			}
			signalReceived = true

			// Log graceful shutdown
			fmt.Fprintf(os.Stderr, "Received signal %v. Initiating graceful shutdown.\n", sig)

			// Stop socket listener
			_ = rt.socketManager.DeleteSocket()

			// Exit the loop
			terminated = true
		}
	}

	// Cleanup: Stop socket listener (idempotent)
	_ = rt.socketManager.DeleteSocket()

	// Wait for listener goroutine to exit, but also monitor for second signal
	select {
	case <-listenerDoneCh:
		// Listener exited normally
	case sig := <-signalChan:
		if signalReceived {
			// Second signal during cleanup - exit immediately with code 130
			return 130
		}
		// First signal received here (shouldn't happen in normal flow)
		signalReceived = true
		fmt.Fprintf(os.Stderr, "Received signal %v. Initiating graceful shutdown.\n", sig)
	}

	// Drain any remaining errors from listenerErrCh (non-blocking, best effort)
	for {
		select {
		case <-listenerErrCh:
			// Discard
		default:
			goto drained
		}
	}
drained:

	// Determine exit code based on session status
	var exitCode int
	finalStatus := sess.GetStatusSafe()
	if finalStatus == "completed" {
		exitCode = 0
	} else {
		exitCode = 1
	}

	// Monitor for second signal during finalization
	finalizeDone := make(chan struct{})
	go func() {
		rt.sessionFinalizer.Finalize(sess)
		close(finalizeDone)
	}()

	select {
	case <-finalizeDone:
		// Finalization completed normally
	case sig := <-signalChan:
		if signalReceived {
			// Second signal during finalization - exit immediately with code 130
			return 130
		}
		// First signal received here (shouldn't happen in normal flow)
		// Log and wait for finalization to complete
		fmt.Fprintf(os.Stderr, "Received signal %v. Initiating graceful shutdown.\n", sig)
		<-finalizeDone // Wait for finalization anyway
	}

	return exitCode
}
