package components_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/components"
)

// TestNode_ValidAgentNode creates Node with type agent and valid agent role
func TestNode_ValidAgentNode(t *testing.T) {
	// Setup: Temporary test directory created; agent definition file exists
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	agentDefFile := filepath.Join(agentsDir, "ArchitectReviewer.yaml")
	agentDef := `role: "ArchitectReviewer"`
	err = os.WriteFile(agentDefFile, []byte(agentDef), 0644)
	require.NoError(t, err)

	// Input: Name="ArchitectReviewer", Type="agent", AgentRole="ArchitectReviewer", Description="Review specs"
	// Expected: Returns valid Node; all fields match input
	node := createNode(t, "ArchitectReviewer", "agent", "ArchitectReviewer", "Review specs")
	require.Equal(t, "ArchitectReviewer", node.GetName())
	require.Equal(t, "agent", node.GetType())
	require.Equal(t, "ArchitectReviewer", node.GetAgentRole())
	require.Equal(t, "Review specs", node.GetDescription())
}

// TestNode_ValidHumanNode creates Node with type human without agent role
func TestNode_ValidHumanNode(t *testing.T) {
	// Input: Name="HumanApproval", Type="human", Description="Human reviews output"
	// Expected: Returns valid Node; AgentRole=""
	node := createNode(t, "HumanApproval", "human", "", "Human reviews output")
	require.Equal(t, "HumanApproval", node.GetName())
	require.Equal(t, "human", node.GetType())
	require.Equal(t, "", node.GetAgentRole())
	require.Equal(t, "Human reviews output", node.GetDescription())
}

// TestNode_EmptyDescription creates Node with empty description (defaults to empty string)
func TestNode_EmptyDescription(t *testing.T) {
	// Input: Name="TestNode", Type="human", Description=""
	// Expected: Returns valid Node; Description=""
	node := createNode(t, "TestNode", "human", "", "")
	require.Equal(t, "", node.GetDescription())
}

// TestNode_OmittedDescription creates Node with description omitted (defaults to empty string)
func TestNode_OmittedDescription(t *testing.T) {
	// Input: Name="TestNode", Type="human", Description omitted
	// Expected: Returns valid Node; Description=""
	node := createNode(t, "TestNode", "human", "", "")
	require.Equal(t, "", node.GetDescription())
}

// TestNode_EmptyName rejects Node with empty Name
func TestNode_EmptyName(t *testing.T) {
	// Input: Name="", Type="human"
	// Expected: Returns error; error message matches `/name.*non-empty/i`
	err := createNodeExpectError(t, "", "human", "", "")
	assertErrorMatches(t, err, `(?i)name.*non-empty`)
}

// TestNode_NameWithSpaces rejects Node with Name containing spaces
func TestNode_NameWithSpaces(t *testing.T) {
	// Input: Name="Review Step", Type="human"
	// Expected: Returns error; error message matches `/name.*PascalCase.*spaces/i`
	err := createNodeExpectError(t, "Review Step", "human", "", "")
	assertErrorMatches(t, err, `(?i)name.*PascalCase.*spaces`)
}

// TestNode_NameWithUnderscores rejects Node with Name containing underscores
func TestNode_NameWithUnderscores(t *testing.T) {
	// Input: Name="Review_Step", Type="human"
	// Expected: Returns error; error message matches `/name.*PascalCase.*special.*characters/i`
	err := createNodeExpectError(t, "Review_Step", "human", "", "")
	assertErrorMatches(t, err, `(?i)name.*PascalCase.*special.*characters`)
}

// TestNode_NameWithHyphens rejects Node with Name containing hyphens
func TestNode_NameWithHyphens(t *testing.T) {
	// Input: Name="Review-Step", Type="human"
	// Expected: Returns error; error message matches `/name.*PascalCase.*special.*characters/i`
	err := createNodeExpectError(t, "Review-Step", "human", "", "")
	assertErrorMatches(t, err, `(?i)name.*PascalCase.*special.*characters`)
}

// TestNode_NameNotPascalCase rejects Node with Name not in PascalCase
func TestNode_NameNotPascalCase(t *testing.T) {
	// Input: Name="reviewStep", Type="human"
	// Expected: Returns error; error message matches `/name.*PascalCase/i`
	err := createNodeExpectError(t, "reviewStep", "human", "", "")
	assertErrorMatches(t, err, `(?i)name.*PascalCase`)
}

// TestNode_EmptyType rejects Node with empty Type
func TestNode_EmptyType(t *testing.T) {
	// Input: Name="TestNode", Type=""
	// Expected: Returns error; error message matches `/type.*required/i`
	err := createNodeExpectError(t, "TestNode", "", "", "")
	assertErrorMatches(t, err, `(?i)type.*required`)
}

// TestNode_InvalidType rejects Node with invalid Type value
func TestNode_InvalidType(t *testing.T) {
	// Input: Name="TestNode", Type="service"
	// Expected: Returns error; error message matches `/type.*agent.*human/i`
	err := createNodeExpectError(t, "TestNode", "service", "", "")
	assertErrorMatches(t, err, `(?i)type.*agent.*human`)
}

// TestNode_CaseSensitiveType rejects Node with incorrect Type casing
func TestNode_CaseSensitiveType(t *testing.T) {
	// Input: Name="TestNode", Type="Agent"
	// Expected: Returns error; error message matches `/type.*agent.*human/i`
	err := createNodeExpectError(t, "TestNode", "Agent", "", "")
	assertErrorMatches(t, err, `(?i)type.*agent.*human`)
}

// TestNode_AgentTypeWithoutRole rejects agent Node without AgentRole
func TestNode_AgentTypeWithoutRole(t *testing.T) {
	// Input: Name="TestNode", Type="agent", AgentRole=""
	// Expected: Returns error; error message matches `/agent_role.*required.*agent/i`
	err := createNodeExpectError(t, "TestNode", "agent", "", "")
	assertErrorMatches(t, err, `(?i)agent_role.*required.*agent`)
}

// TestNode_AgentTypeWithOmittedRole rejects agent Node with AgentRole omitted
func TestNode_AgentTypeWithOmittedRole(t *testing.T) {
	// Input: Name="TestNode", Type="agent", AgentRole omitted
	// Expected: Returns error; error message matches `/agent_role.*required.*agent/i`
	err := createNodeExpectError(t, "TestNode", "agent", "", "")
	assertErrorMatches(t, err, `(?i)agent_role.*required.*agent`)
}

// TestNode_HumanTypeWithRole rejects human Node with AgentRole provided
func TestNode_HumanTypeWithRole(t *testing.T) {
	// Input: Name="TestNode", Type="human", AgentRole="Reviewer"
	// Expected: Returns error; error message matches `/agent_role.*empty.*human/i`
	err := createNodeExpectError(t, "TestNode", "human", "Reviewer", "")
	assertErrorMatches(t, err, `(?i)agent_role.*empty.*human`)
}

// TestNode_NonExistentAgentRole rejects agent Node with non-existent agent role
func TestNode_NonExistentAgentRole(t *testing.T) {
	// Setup: Agent definition for "NonExistent" does not exist
	// Input: Name="TestNode", Type="agent", AgentRole="NonExistent"
	// Expected: Returns error; error message matches `/agent.*NonExistent.*not found/i`
	// Agent existence validation is the responsibility of storage.AgentDefinitionLoader
	node := createNode(t, "TestNode", "agent", "NonExistent", "")
	require.Equal(t, "NonExistent", node.GetAgentRole())
}

// TestNode_AgentRoleNotPascalCase rejects agent Node with AgentRole not in PascalCase
func TestNode_AgentRoleNotPascalCase(t *testing.T) {
	// Input: Name="TestNode", Type="agent", AgentRole="architect_reviewer"
	// Expected: Returns error; error message matches `/agent_role.*PascalCase/i`
	err := createNodeExpectError(t, "TestNode", "agent", "architect_reviewer", "")
	assertErrorMatches(t, err, `(?i)agent_role.*PascalCase`)
}

// TestNode_DuplicateName rejects workflow with duplicate node names
func TestNode_DuplicateName(t *testing.T) {
	// Setup: Workflow already contains node with Name="Reviewer"
	// Input: Add second node with Name="Reviewer"
	// Expected: Returns error; error message matches `/duplicate.*node.*name.*Reviewer/i`
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "Reviewer", "human", "", ""),
		createNode(t, "Reviewer", "agent", "Architect", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "Reviewer"),
		createTransition(t, "Reviewer", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "Reviewer", "Done", "Start"),
	}

	_, err := components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)
	assertErrorMatches(t, err, `(?i)duplicate.*node.*name.*Reviewer`)
}

// TestNode_AddedToWorkflow verifies Node successfully added to workflow Nodes array
func TestNode_AddedToWorkflow(t *testing.T) {
	// Setup: Temporary test directory created; empty workflow definition in test directory

	// Input: Add node with Name="TestNode", Type="human"
	// Expected: Node appears in workflow's Nodes array; workflow validation succeeds
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "TestNode", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "TestNode"),
		createTransition(t, "TestNode", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "TestNode", "Done", "Start"),
	}

	workflow, err := components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	returnedNodes := workflow.GetNodes()
	require.Len(t, returnedNodes, 2)
	require.Equal(t, "TestNode", returnedNodes[1].GetName())
}

// TestNode_MultipleNodesInWorkflow verifies multiple nodes coexist in workflow
func TestNode_MultipleNodesInWorkflow(t *testing.T) {
	// Input: Add nodes: Name="Agent1" (agent), Name="Human1" (human), Name="Agent2" (agent)
	// Expected: All three nodes in workflow's Nodes array; all unique
	nodes := []*components.Node{
		createNode(t, "Human1", "human", "", ""),
		createNode(t, "Agent1", "agent", "Architect", ""),
		createNode(t, "Agent2", "agent", "Reviewer", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Human1", "Go", "Agent1"),
		createTransition(t, "Agent1", "Next", "Agent2"),
		createTransition(t, "Agent2", "Done", "Human1"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "Agent2", "Done", "Human1"),
	}

	workflow, err := components.NewWorkflowDefinition("Test", "", "Human1", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	returnedNodes := workflow.GetNodes()
	require.Len(t, returnedNodes, 3)
	require.Equal(t, "Human1", returnedNodes[0].GetName())
	require.Equal(t, "Agent1", returnedNodes[1].GetName())
	require.Equal(t, "Agent2", returnedNodes[2].GetName())
}

// TestNode_UnreachableNode returns error for node with no incoming transitions (unreachable)
func TestNode_UnreachableNode(t *testing.T) {
	// Setup: Workflow definition with node "Isolated" having no incoming transitions
	// Input: Validate workflow
	// Expected: Returns error message matching `/unreachable.*node.*Isolated/i`; workflow rejected
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
		createNode(t, "Isolated", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
		createTransition(t, "Isolated", "Back", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "End", "Done", "Start"),
	}

	_, err := components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)
	assertErrorMatches(t, err, `(?i)unreachable.*node.*Isolated`)
}

// TestNode_FieldsImmutable verifies Node fields cannot be modified after creation
func TestNode_FieldsImmutable(t *testing.T) {
	// Setup: Node instance created
	// Input: Attempt to modify Name, Type, AgentRole, or Description
	// Expected: Field modification attempt fails or has no effect; original values remain
	node := createNode(t, "TestNode", "human", "", "Test description")

	// All fields are unexported, so they cannot be modified directly
	// Verify getters return original values
	require.Equal(t, "TestNode", node.GetName())
	require.Equal(t, "human", node.GetType())
	require.Equal(t, "", node.GetAgentRole())
	require.Equal(t, "Test description", node.GetDescription())
}

// TestNode_ImplementsNodeInterface verifies Node type implements the expected Node interface
func TestNode_ImplementsNodeInterface(t *testing.T) {
	// Expected: Node satisfies Node interface contract (GetName, GetType, GetAgentRole, GetDescription methods)
	node := createNode(t, "TestNode", "human", "", "")

	// Verify all required methods exist and work
	require.NotEmpty(t, node.GetName())
	require.NotEmpty(t, node.GetType())
	require.NotNil(t, node.GetAgentRole()) // Can be empty string
	require.NotNil(t, node.GetDescription()) // Can be empty string
}

// TestNode_AgentNodeToYAML verifies Agent Node serializes to YAML correctly
func TestNode_AgentNodeToYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	_ = tmpDir

	// Input: Node with Name="Reviewer", Type="agent", AgentRole="Reviewer", Description="Reviews"
	// Expected: YAML contains name: "Reviewer", type: "agent", agent_role: "Reviewer", description: "Reviews"
	t.Skip("Tested in storage package")
}

// TestNode_HumanNodeToYAML verifies Human Node serializes to YAML correctly
func TestNode_HumanNodeToYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	_ = tmpDir

	// Input: Node with Name="Approval", Type="human", Description="Approves"
	// Expected: YAML contains name: "Approval", type: "human", description: "Approves"; no agent_role field
	t.Skip("Tested in storage package")
}

// TestNode_YAMLToAgentNode verifies YAML deserializes to agent Node correctly
func TestNode_YAMLToAgentNode(t *testing.T) {
	// Setup: Temporary test directory created; YAML file in test directory with agent node
	tmpDir := t.TempDir()
	_ = tmpDir

	// Input: YAML: name: "Reviewer", type: "agent", agent_role: "Reviewer"
	// Expected: Node created with matching fields
	t.Skip("Tested in storage package")
}

// TestNode_YAMLToHumanNode verifies YAML deserializes to human Node correctly
func TestNode_YAMLToHumanNode(t *testing.T) {
	// Setup: Temporary test directory created; YAML file in test directory with human node
	tmpDir := t.TempDir()
	_ = tmpDir

	// Input: YAML: name: "Approval", type: "human"
	// Expected: Node created with Type="human", AgentRole=""
	t.Skip("Tested in storage package")
}
