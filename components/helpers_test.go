package components

// --- Test Constants ---

const (
	// Valid PascalCase identifiers for reuse across component tests.
	validNodeName1  = "ReviewStep"
	validNodeName2  = "HumanApproval"
	validNodeName3  = "Draft"
	validAgentRole  = "Architect"
	validAgentRole2 = "Writer"
	validEventType  = "DraftCompleted"
	validEventType2 = "Approved"
	validEventType3 = "Completed"
	validDesc       = "Reviews code"

	// Valid agent definition defaults for reuse.
	validModel        = "claude-sonnet-4-20250514"
	validEffort       = "high"
	validSystemPrompt = "You are an architect."
	validAgentRoot    = "spec"
)
