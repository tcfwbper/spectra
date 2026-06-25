package runtime

import (
	"fmt"
	"strings"

	"github.com/tcfwbper/spectra/logger"
)

// ProcessInspector defines the interface for checking process state.
type ProcessInspector interface {
	IsRunning(pid int) bool
	Command(pid int) string
}

// SignalSender defines the interface for sending signals to processes.
type SignalSender interface {
	SendSIGTERM(pid int) error
	SendSIGKILL(pid int) error
}

// ProcessWaiter defines the interface for waiting on process exit.
type ProcessWaiter interface {
	WaitForExit(pid int) bool
}

// ClaudeProcessCleaner finds and terminates orphaned Claude CLI processes
// that were spawned by the current session.
type ClaudeProcessCleaner struct {
	ps        *PersistentSession
	log       logger.Logger
	inspector ProcessInspector
	sender    SignalSender
	waiter    ProcessWaiter
}

// ClaudeProcessCleanerOption is a functional option for ClaudeProcessCleaner.
type ClaudeProcessCleanerOption func(*ClaudeProcessCleaner)

// WithProcessInspector sets a custom process inspector.
func WithProcessInspector(inspector ProcessInspector) ClaudeProcessCleanerOption {
	return func(c *ClaudeProcessCleaner) {
		c.inspector = inspector
	}
}

// WithSignalSender sets a custom signal sender.
func WithSignalSender(sender SignalSender) ClaudeProcessCleanerOption {
	return func(c *ClaudeProcessCleaner) {
		c.sender = sender
	}
}

// WithProcessWaiter sets a custom process waiter.
func WithProcessWaiter(waiter ProcessWaiter) ClaudeProcessCleanerOption {
	return func(c *ClaudeProcessCleaner) {
		c.waiter = waiter
	}
}

// NewClaudeProcessCleaner constructs a ClaudeProcessCleaner with the given
// PersistentSession and Logger.
func NewClaudeProcessCleaner(ps *PersistentSession, log logger.Logger, opts ...ClaudeProcessCleanerOption) *ClaudeProcessCleaner {
	c := &ClaudeProcessCleaner{
		ps:  ps,
		log: log,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Clean finds and terminates orphaned Claude CLI processes spawned by this session.
// It reads PID values from session data, verifies each is running and belongs to
// a claude process, sends SIGTERM, waits up to 2 seconds, then escalates to SIGKILL.
// All failures are logged and absorbed.
func (c *ClaudeProcessCleaner) Clean() {
	// Step 1: Get a snapshot of session metadata.
	snapshot := c.ps.GetMetadataSnapshotSafe()

	// Step 2-5: Collect and deduplicate PID values from keys ending in ".PID".
	seen := make(map[int]bool)
	var candidates []int
	for key, value := range snapshot.SessionData {
		if !strings.HasSuffix(key, ".PID") {
			continue
		}
		pid, ok := value.(int)
		if !ok {
			c.log.Warn(fmt.Sprintf("skipping non-integer PID value for key %s: got %T", key, value))
			continue
		}
		if !seen[pid] {
			seen[pid] = true
			candidates = append(candidates, pid)
		}
	}

	// Step 6: If no candidates, return immediately.
	if len(candidates) == 0 {
		return
	}

	// Step 7-8: Verify each candidate PID.
	var killList []int
	for _, pid := range candidates {
		if c.inspector == nil {
			continue
		}
		if !c.inspector.IsRunning(pid) {
			continue
		}
		cmd := c.inspector.Command(pid)
		if !strings.Contains(strings.ToLower(cmd), "claude") {
			c.log.Warn(fmt.Sprintf("PID does not belong to a claude process, skipping: pid=%d cmd=%s", pid, cmd))
			continue
		}
		killList = append(killList, pid)
	}

	// Step 9: If kill list is empty, log and return.
	if len(killList) == 0 {
		c.log.Info("no active claude processes to terminate")
		return
	}

	// Step 10-12: Send SIGTERM to all processes in kill list.
	var waitList []int
	for _, pid := range killList {
		if c.sender == nil {
			continue
		}
		if err := c.sender.SendSIGTERM(pid); err != nil {
			c.log.Warn(fmt.Sprintf("failed to send SIGTERM to PID %d: %s", pid, err.Error()))
			continue
		}
		waitList = append(waitList, pid)
	}
	c.log.Info(fmt.Sprintf("sent SIGTERM to %d claude process(es)", len(waitList)))

	// Step 13-16: Wait for processes to exit, escalate to SIGKILL if needed.
	if c.waiter != nil && c.sender != nil {
		for _, pid := range waitList {
			if c.waiter.WaitForExit(pid) {
				continue
			}
			// Process did not exit within timeout — escalate to SIGKILL.
			c.log.Warn(fmt.Sprintf("escalating to SIGKILL for PID %d", pid))
			if err := c.sender.SendSIGKILL(pid); err != nil {
				c.log.Warn(fmt.Sprintf("failed to kill claude process: pid=%d error=%s", pid, err.Error()))
			}
		}
	}

	// Step 17: Log summary.
	c.log.Info("claude process cleanup complete")
}
