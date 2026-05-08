package runtime

import "fmt"

// Node defines the interface for node definition objects consumed by
// ValidateClaudeSessionID. It provides read-only access to node type and name.
type Node interface {
	Type() string
	Name() string
}

// ValidateClaudeSessionID validates the claudeSessionID from a RuntimeMessage
// against the current node's requirements. It is a stateless package-level
// helper that does not modify session state, record events, or perform any
// lifecycle transitions.
func ValidateClaudeSessionID(ps *PersistentSession, node Node, claudeSessionID string) error {
	switch node.Type() {
	case "agent":
		key := node.Name() + ".ClaudeSessionID"
		stored, ok := ps.GetSessionDataSafe(key)
		if !ok {
			return fmt.Errorf("claude session ID not found for node '%s'", node.Name())
		}
		storedStr, _ := stored.(string)
		if storedStr != claudeSessionID {
			return fmt.Errorf("claude session ID mismatch: expected %s but got %s", storedStr, claudeSessionID)
		}
		return nil
	case "human":
		if claudeSessionID != "" {
			return fmt.Errorf("invalid claude session ID for human node: must be empty")
		}
		return nil
	default:
		return fmt.Errorf("unsupported node type '%s'", node.Type())
	}
}
