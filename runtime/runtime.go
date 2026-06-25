package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/storage"
)

// --- Injectable seams for testing ---

// spectraFinderFunc is a package-level seam for SpectraFinder.Find().
// In production it calls storage.FindSpectraRoot("").
var spectraFinderFunc = func() (string, error) {
	return storage.FindSpectraRoot("")
}

// preSessionDepsConstructor constructs pre-session dependencies.
// Returns (WorkflowLoader, SessionDirManager, error).
var preSessionDepsConstructor = func(projectRoot string) (WorkflowLoader, SessionDirManager, error) {
	agentLoader := storage.NewAgentDefinitionLoader(projectRoot)
	wfLoader := storage.NewWorkflowDefinitionLoader(projectRoot, agentLoader)
	dirMgr := &sessionDirManagerAdapter{projectRoot: projectRoot}
	return wfLoader, dirMgr, nil
}

// signalNotifyFunc is a seam for signal.Notify.
var signalNotifyFunc = func(c chan<- os.Signal, sig ...os.Signal) {
	signal.Notify(c, sig...)
}

// signalStopFunc is a seam for signal.Stop.
var signalStopFunc = func(c chan<- os.Signal) {
	signal.Stop(c)
}

// newGraceTimerFunc creates a timer channel that fires after the given duration.
// Returns a channel and a stop function.
var newGraceTimerFunc = func(d time.Duration) (<-chan struct{}, func()) {
	t := time.NewTimer(d)
	ch := make(chan struct{}, 1)
	go func() {
		<-t.C
		ch <- struct{}{}
	}()
	return ch, func() { t.Stop() }
}

// newSubTimerFunc creates a sub-timeout timer (for listener shutdown wait).
var newSubTimerFunc = func(d time.Duration) (<-chan struct{}, func()) {
	t := time.NewTimer(d)
	ch := make(chan struct{}, 1)
	go func() {
		<-t.C
		ch <- struct{}{}
	}()
	return ch, func() { t.Stop() }
}

// sessionInitializeFunc is a seam for constructing and invoking SessionInitializer.
// In production it constructs a real SessionInitializer and calls Initialize.
var sessionInitializeFunc = func(projectRoot string, wfLoader WorkflowLoader, dirMgr SessionDirManager, log logger.Logger, workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
	si := NewSessionInitializer(projectRoot, wfLoader, dirMgr, log)
	return si.Initialize(workflowName, sessionID, terminationNotifier)
}

// constructPostSessionDepsFunc is a seam for constructing post-session dependencies.
// In production it calls the internal constructPostSessionDeps function.
var constructPostSessionDepsFunc = func(projectRoot string, ps *PersistentSession, wfDef *components.WorkflowDefinition, terminationNotifier chan<- struct{}, log logger.Logger) (*runtimePostSessionDeps, error) {
	return constructPostSessionDeps(projectRoot, ps, wfDef, terminationNotifier, log)
}

// --- Interfaces for runtime orchestration ---

// SocketManager defines the interface for socket lifecycle management consumed
// by the Run function. In production this is implemented by storage.RuntimeSocketManager.
type SocketManager interface {
	CreateSocket() error
	Listen(handler storage.MessageHandler) (<-chan error, <-chan struct{}, error)
	DeleteSocket()
}

// --- Adapters ---

// sessionDirManagerAdapter adapts the package-level storage.CreateSessionDirectory
// function to the SessionDirManager interface expected by SessionInitializer.
type sessionDirManagerAdapter struct {
	projectRoot string
}

func (a *sessionDirManagerAdapter) CreateSessionDirectory(projectRoot, sessionUUID string) error {
	return storage.CreateSessionDirectory(projectRoot, sessionUUID)
}

// agentDefLoaderAdapter adapts storage.AgentDefinitionLoader to the
// TransitionAgentDefLoader interface expected by TransitionToNode.
type agentDefLoaderAdapter struct {
	loader *storage.AgentDefinitionLoader
}

func (a *agentDefLoaderAdapter) Load(agentRole string) (AgentDef, error) {
	return a.loader.Load(agentRole)
}

// agentInvokerAdapter adapts *AgentInvoker.Invoke to the TransitionAgentInvoker
// interface which expects InvokeAgent.
type agentInvokerAdapter struct {
	invoker *AgentInvoker
}

func (a *agentInvokerAdapter) InvokeAgent(nodeName, message string, agentDef AgentDef) error {
	return a.invoker.Invoke(nodeName, message, agentDef)
}

// messageHandlerAdapter adapts *MessageRouter (pointer-based Handle) to
// storage.MessageHandler (value-based Handle) interface.
type messageHandlerAdapter struct {
	router *MessageRouter
}

func (a *messageHandlerAdapter) Handle(sessionUUID string, msg entities.RuntimeMessage) entities.RuntimeResponse {
	resp := a.router.Handle(sessionUUID, &msg)
	if resp == nil {
		return *entities.ErrorResponse("internal server error")
	}
	return *resp
}

// --- Run function ---

// Run is the top-level runtime orchestrator. It bootstraps all dependencies,
// initializes a session, creates the runtime socket, performs the initial
// dispatch of the entry node, runs the main event loop, handles termination
// signals, and returns an exit code with an optional error.
func Run(workflowName string, sessionID string, log logger.Logger) (int, error) {
	// Step 2: Locate project root.
	projectRoot, err := spectraFinderFunc()
	if err != nil {
		return 1, fmt.Errorf("failed to locate project root: %w", err)
	}

	// Step 4: Create terminationNotifier with capacity 2.
	terminationNotifier := make(chan struct{}, 2)

	// Step 5-6: Construct pre-session dependencies.
	wfLoader, dirMgr, err := preSessionDepsConstructor(projectRoot)
	if err != nil {
		return 1, fmt.Errorf("failed to initialize runtime dependencies: %w", err)
	}

	// Step 7-8: Construct SessionInitializer and initialize session.
	initResult := sessionInitializeFunc(projectRoot, wfLoader, dirMgr, log, workflowName, sessionID, terminationNotifier)

	// Step 9-10: Handle initialization failure.
	if initResult.Error != nil {
		if initResult.PersistentSession == nil {
			// Step 9: Failure before session entity construction.
			return 1, fmt.Errorf("failed to initialize session: %w", initResult.Error)
		}
		// Step 10: Failure after session entity construction — proceed to cleanup.
		finalizer := NewSessionFinalizer(log)
		exitCode := finalizer.Finalize(initResult.PersistentSession)
		return exitCode, fmt.Errorf("failed to initialize session: %w", initResult.Error)
	}

	ps := initResult.PersistentSession
	wfDef := initResult.WorkflowDefinition

	// Step 11-12: Construct post-session dependencies.
	deps, err := constructPostSessionDepsFunc(projectRoot, ps, wfDef, terminationNotifier, log)
	if err != nil {
		rtErr := buildRuntimeError("Runtime", "failed to initialize post-session dependencies", err, ps)
		if failErr := ps.Fail(rtErr, terminationNotifier); failErr != nil {
			log.Warn(fmt.Sprintf("attempted to fail session but session already in terminal state: %s", failErr.Error()))
		}
		finalizer := NewSessionFinalizer(log)
		exitCode := finalizer.Finalize(ps)
		return exitCode, fmt.Errorf("failed to initialize post-session dependencies: %w", err)
	}

	socketMgr := deps.socketManager
	transitionToNode := deps.transitionNode
	messageRouter := deps.messageRouter
	finalizer := deps.finalizer

	// Step 13-14: Create socket.
	if err := socketMgr.CreateSocket(); err != nil {
		rtErr := buildRuntimeError("Runtime", "failed to create runtime socket", err, ps)
		if failErr := ps.Fail(rtErr, terminationNotifier); failErr != nil {
			log.Warn(fmt.Sprintf("attempted to fail session but session already in terminal state: %s", failErr.Error()))
		}
		exitCode := finalizer.Finalize(ps)
		return exitCode, fmt.Errorf("failed to create runtime socket: %w", err)
	}

	// Step 15-17: Start listener.
	handler := &messageHandlerAdapter{router: messageRouter}
	listenerErrCh, listenerDoneCh, err := socketMgr.Listen(handler)
	if err != nil {
		rtErr := buildRuntimeError("Runtime", "failed to start socket listener", err, ps)
		if failErr := ps.Fail(rtErr, terminationNotifier); failErr != nil {
			log.Warn(fmt.Sprintf("attempted to fail session but session already in terminal state: %s", failErr.Error()))
		}
		socketMgr.DeleteSocket()
		exitCode := finalizer.Finalize(ps)
		return exitCode, fmt.Errorf("failed to start socket listener: %w", err)
	}

	// Step 18-22: Initial dispatch.
	entryNodeName := wfDef.EntryNode()
	defaultMsg := fmt.Sprintf(
		"Workflow started. You are the first node and may begin your work. To transition, run: spectra-agent event emit <type> --session-id %s [--message <message>] [--claude-session-id <UUID>] [--payload <json>]",
		ps.ID,
	)
	if err := transitionToNode.Execute(entryNodeName, defaultMsg); err != nil {
		rtErr := buildRuntimeErrorWithState("Runtime", "failed to dispatch entry node", err, ps, entryNodeName)
		if failErr := ps.Fail(rtErr, terminationNotifier); failErr != nil {
			log.Warn(fmt.Sprintf("attempted to fail session but session already in terminal state: %s", failErr.Error()))
		}
		// Cleanup: delete socket, wait for listener, then finalize.
		socketMgr.DeleteSocket()
		subTimerCh, subTimerStop := newSubTimerFunc(2 * time.Second)
		select {
		case <-listenerDoneCh:
		case <-subTimerCh:
			log.Warn("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
		}
		subTimerStop()
		exitCode := finalizer.Finalize(ps)
		return exitCode, fmt.Errorf("failed to dispatch entry node: %w", err)
	}

	// Step 23-24: Register OS signal handling.
	signalCh := make(chan os.Signal, 2)
	signalNotifyFunc(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Step 26: Main event loop.
	var receivedSignal os.Signal

	select {
	case <-terminationNotifier:
		// Session reached terminal status.
		log.Info("received session termination notification")

	case err := <-listenerErrCh:
		// Fatal listener error.
		log.Info(fmt.Sprintf("listener error: %s", err.Error()))
		status := ps.GetStatusSafe()
		if status != "completed" && status != "failed" {
			rtErr := buildRuntimeError("Runtime", "listener error", err, ps)
			if failErr := ps.Fail(rtErr, terminationNotifier); failErr != nil {
				log.Warn(fmt.Sprintf("attempted to fail session but session already in terminal state: %s", failErr.Error()))
			}
		}

	case sig := <-signalCh:
		// OS signal received.
		receivedSignal = sig
		log.Info(fmt.Sprintf("received signal %s, initiating graceful shutdown", sig.String()))
		// Step 26 (case signalCh): If session is not terminal, fail it with RuntimeError.
		status := ps.GetStatusSafe()
		if status != "completed" && status != "failed" {
			rtErr := buildRuntimeError("Runtime", fmt.Sprintf("terminated by signal %s", sig.String()), nil, ps)
			if failErr := ps.Fail(rtErr, terminationNotifier); failErr != nil {
				log.Warn(fmt.Sprintf("attempted to fail session on signal but session already in terminal state: %s", failErr.Error()))
			}
		}
	}

	// Step 28-31: Grace period and cleanup.
	graceTimerCh, graceTimerStop := newGraceTimerFunc(5 * time.Second)
	defer graceTimerStop()

	// Cleanup goroutine result channel.
	cleanupDone := make(chan struct{})

	// Capture seam references before spawning goroutine to avoid races
	// with test cleanup that restores package-level vars.
	localSignalStop := signalStopFunc
	localNewSubTimer := newSubTimerFunc

	go func() {
		// Step 32: Stop OS signal notification.
		localSignalStop(signalCh)

		// Step 33: ClaudeProcessCleaner.Clean() — terminate orphaned claude processes.
		deps.claudeProcessCleaner.Clean()

		// Step 34: Delete socket.
		socketMgr.DeleteSocket()

		// Step 35-36: Wait for listenerDoneCh with 2-second sub-timeout.
		subTimerCh, subTimerStop := localNewSubTimer(2 * time.Second)
		defer subTimerStop()

		select {
		case <-listenerDoneCh:
			// Listener shut down cleanly.
		case <-subTimerCh:
			log.Warn("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
		}

		close(cleanupDone)
	}()

	// Wait for cleanup to complete, second signal, or grace timeout.
	select {
	case <-cleanupDone:
		// Cleanup completed normally.

	case <-graceTimerCh:
		// Step 30: Grace period exceeded.
		log.Warn("cleanup exceeded 5 second grace period, forcing exit")
		return 1, fmt.Errorf("cleanup timeout")

	case <-signalCh:
		// Step 31: Second signal forces exit.
		log.Warn("received second signal, forcing exit")
		return 1, fmt.Errorf("forced exit by second signal")
	}

	// Step 37-39: Invoke SessionFinalizer and determine return value.
	exitCode := finalizer.Finalize(ps)

	if receivedSignal != nil {
		// Step 39: Exit code is always 1 when signal received, regardless of SessionFinalizer.
		return 1, fmt.Errorf("session terminated by signal %s", receivedSignal.String())
	}

	if exitCode == 0 {
		return 0, nil
	}

	// exitCode == 1, no signal
	sessionErr := ps.GetErrorSafe()
	if sessionErr != nil {
		return exitCode, fmt.Errorf("session failed: %s", sessionErr.Error())
	}

	return exitCode, fmt.Errorf("session terminated with non-terminal status")
}

// --- Internal types ---

// runtimePostSessionDeps holds all post-session dependencies.
type runtimePostSessionDeps struct {
	socketManager        SocketManager
	transitionNode       TransitionToNodeExecutor
	messageRouter        *MessageRouter
	finalizer            *SessionFinalizer
	claudeProcessCleaner *ClaudeProcessCleaner
}

// constructPostSessionDeps creates all post-session dependencies.
func constructPostSessionDeps(projectRoot string, ps *PersistentSession, wfDef *components.WorkflowDefinition, terminationNotifier chan<- struct{}, log logger.Logger) (*runtimePostSessionDeps, error) {
	agentDefLoader := storage.NewAgentDefinitionLoader(projectRoot)
	socketMgr := storage.NewRuntimeSocketManager(projectRoot, ps.ID, log)
	agentInvoker := NewAgentInvoker(ps, projectRoot, WithLogger(log))
	invokerAdapter := &agentInvokerAdapter{invoker: agentInvoker}
	transitionToNode := NewTransitionToNode(ps, wfDef, &agentDefLoaderAdapter{loader: agentDefLoader}, invokerAdapter)
	eventProcessor := NewEventProcessor(ps, wfDef, transitionToNode, terminationNotifier)
	errorProcessor := NewErrorProcessor(ps, wfDef, terminationNotifier)
	messageRouter := NewMessageRouter(ps, eventProcessor, errorProcessor, terminationNotifier, log)
	claudeProcessCleaner := NewClaudeProcessCleaner(ps, log)
	finalizer := NewSessionFinalizer(log)

	return &runtimePostSessionDeps{
		socketManager:        socketMgr,
		transitionNode:       transitionToNode,
		messageRouter:        messageRouter,
		finalizer:            finalizer,
		claudeProcessCleaner: claudeProcessCleaner,
	}, nil
}

// --- Helper functions ---

// buildRuntimeError constructs a RuntimeError for post-session failures.
func buildRuntimeError(issuer, message string, _ error, ps *PersistentSession) *entities.RuntimeError {
	rtErr, _ := entities.NewRuntimeError(
		issuer,
		message,
		nil,
		time.Now().Unix(),
		ps.ID,
		ps.GetCurrentStateSafe(),
	)
	return rtErr
}

// buildRuntimeErrorWithState constructs a RuntimeError with a specific failing state.
func buildRuntimeErrorWithState(issuer, message string, _ error, ps *PersistentSession, failingState string) *entities.RuntimeError {
	rtErr, _ := entities.NewRuntimeError(
		issuer,
		message,
		nil,
		time.Now().Unix(),
		ps.ID,
		failingState,
	)
	return rtErr
}
