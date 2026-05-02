package components_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/components"
)

// TestWorkflowDefinition_ValidWorkflowAllFields creates WorkflowDefinition with all fields provided
func TestWorkflowDefinition_ValidWorkflowAllFields(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Name="DefaultLogicSpec", Description="Simple workflow", EntryNode="HumanRequirement", ExitTransitions, Nodes, Transitions
	nodes := []*components.Node{
		createNode(t, "HumanRequirement", "human", "", ""),
		createNode(t, "HumanApproval", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "HumanRequirement", "RequirementProvided", "HumanApproval"),
		createTransition(t, "HumanApproval", "RequirementApproved", "HumanRequirement"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "HumanApproval", "RequirementApproved", "HumanRequirement"),
	}

	workflow, err := components.NewWorkflowDefinition(
		"DefaultLogicSpec",
		"Simple workflow",
		"HumanRequirement",
		exitTransitions,
		nodes,
		transitions,
	)
	require.NoError(t, err)

	// Expected: Returns valid WorkflowDefinition; all fields match input
	require.Equal(t, "DefaultLogicSpec", workflow.GetName())
	require.Equal(t, "Simple workflow", workflow.GetDescription())
	require.Equal(t, "HumanRequirement", workflow.GetEntryNode())
	require.Len(t, workflow.GetExitTransitions(), 1)
	require.Len(t, workflow.GetNodes(), 2)
	require.Len(t, workflow.GetTransitions(), 2)

	// Expected: YAML file created at <test-dir>/.spectra/workflows/DefaultLogicSpec.yaml
	yamlPath := filepath.Join(workflowsDir, "DefaultLogicSpec.yaml")
	err = workflow.SaveToFile(yamlPath)
	require.NoError(t, err)
	require.FileExists(t, yamlPath)
}

// TestWorkflowDefinition_EmptyDescription creates WorkflowDefinition with empty description
func TestWorkflowDefinition_EmptyDescription(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Name="TestWorkflow", Description="", EntryNode="Start", valid ExitTransitions, Nodes, Transitions
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "End", "Done", "Start"),
	}

	workflow, err := components.NewWorkflowDefinition(
		"TestWorkflow",
		"",
		"Start",
		exitTransitions,
		nodes,
		transitions,
	)
	require.NoError(t, err)

	// Expected: Returns valid WorkflowDefinition; Description=""
	require.Equal(t, "", workflow.GetDescription())
}

// TestWorkflowDefinition_MultipleExitTransitions creates WorkflowDefinition with multiple exit transitions
func TestWorkflowDefinition_MultipleExitTransitions(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Name="MultiExit", ExitTransitions with multiple entries, corresponding nodes and transitions
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End1", "human", "", ""),
		createNode(t, "End2", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "GoToEnd1", "End1"),
		createTransition(t, "Start", "GoToEnd2", "End2"),
		createTransition(t, "End1", "Done1", "Start"),
		createTransition(t, "End2", "Done2", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "End1", "Done1", "Start"),
		createExitTransition(t, "End2", "Done2", "Start"),
	}

	workflow, err := components.NewWorkflowDefinition(
		"MultiExit",
		"",
		"Start",
		exitTransitions,
		nodes,
		transitions,
	)
	require.NoError(t, err)

	// Expected: Returns valid WorkflowDefinition; ExitTransitions contains both transitions
	require.Len(t, workflow.GetExitTransitions(), 2)
}

// TestWorkflowDefinition_LoadValidYAML loads WorkflowDefinition from valid YAML file
func TestWorkflowDefinition_LoadValidYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with all required fields
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	yamlContent := `name: "DefaultLogicSpec"
description: "Simple workflow"
entry_node: "HumanRequirement"
exit_transitions:
  - from_node: "HumanApproval"
    event_type: "RequirementApproved"
    to_node: "HumanRequirement"
nodes:
  - name: "HumanRequirement"
    type: "human"
  - name: "HumanApproval"
    type: "human"
transitions:
  - from_node: "HumanRequirement"
    event_type: "RequirementProvided"
    to_node: "HumanApproval"
  - from_node: "HumanApproval"
    event_type: "RequirementApproved"
    to_node: "HumanRequirement"
`
	yamlPath := filepath.Join(workflowsDir, "DefaultLogicSpec.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load workflow with Name="DefaultLogicSpec"
	workflow, err := components.LoadWorkflowDefinition(yamlPath)
	require.NoError(t, err)

	// Expected: Returns valid WorkflowDefinition; all fields match YAML content
	require.Equal(t, "DefaultLogicSpec", workflow.GetName())
	require.Equal(t, "Simple workflow", workflow.GetDescription())
	require.Equal(t, "HumanRequirement", workflow.GetEntryNode())
	require.Len(t, workflow.GetExitTransitions(), 1)
	require.Len(t, workflow.GetNodes(), 2)
	require.Len(t, workflow.GetTransitions(), 2)
}

// TestWorkflowDefinition_LoadWithEmptyDescription loads WorkflowDefinition with empty description from YAML
func TestWorkflowDefinition_LoadWithEmptyDescription(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with description: ""
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	yamlContent := `name: "Test"
description: ""
entry_node: "Start"
exit_transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
`
	yamlPath := filepath.Join(workflowsDir, "Test.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load workflow with Name="Test"
	workflow, err := components.LoadWorkflowDefinition(yamlPath)
	require.NoError(t, err)

	// Expected: Returns valid WorkflowDefinition; Description=""
	require.Equal(t, "", workflow.GetDescription())
}

// TestWorkflowDefinition_EmptyName rejects WorkflowDefinition with empty Name
func TestWorkflowDefinition_EmptyName(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="", valid other fields
	_, err = components.NewWorkflowDefinition("", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /workflow name.*non-empty/i
	assertErrorMatches(t, err, `(?i)workflow name.*non-empty`)
}

// TestWorkflowDefinition_NameWithSpaces rejects WorkflowDefinition with Name containing spaces
func TestWorkflowDefinition_NameWithSpaces(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="Default LogicSpec", valid other fields
	_, err = components.NewWorkflowDefinition("Default LogicSpec", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /workflow name.*PascalCase.*spaces.*special.*characters/i
	assertErrorMatches(t, err, `(?i)workflow name.*PascalCase.*spaces.*special.*characters`)
}

// TestWorkflowDefinition_NameWithUnderscores rejects WorkflowDefinition with Name containing underscores
func TestWorkflowDefinition_NameWithUnderscores(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="Default_LogicSpec", valid other fields
	_, err = components.NewWorkflowDefinition("Default_LogicSpec", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /workflow name.*PascalCase.*spaces.*special.*characters/i
	assertErrorMatches(t, err, `(?i)workflow name.*PascalCase.*spaces.*special.*characters`)
}

// TestWorkflowDefinition_NameWithHyphens rejects WorkflowDefinition with Name containing hyphens
func TestWorkflowDefinition_NameWithHyphens(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="Default-LogicSpec", valid other fields
	_, err = components.NewWorkflowDefinition("Default-LogicSpec", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /workflow name.*PascalCase.*spaces.*special.*characters/i
	assertErrorMatches(t, err, `(?i)workflow name.*PascalCase.*spaces.*special.*characters`)
}

// TestWorkflowDefinition_NameNotPascalCase rejects WorkflowDefinition with Name not in PascalCase
func TestWorkflowDefinition_NameNotPascalCase(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="defaultLogicSpec", valid other fields
	_, err = components.NewWorkflowDefinition("defaultLogicSpec", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /workflow name.*PascalCase/i
	assertErrorMatches(t, err, `(?i)workflow name.*PascalCase`)
}

// TestWorkflowDefinition_EntryNodeNonExistent rejects WorkflowDefinition with EntryNode referencing non-existent node
func TestWorkflowDefinition_EntryNodeNonExistent(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Existing", "human", "", ""),
		createNode(t, "Other", "human", "", ""),
	}
	transitions := []*components.Transition{createTransition(t, "Existing", "Done", "Other")}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "Existing", "Done", "Other")}

	// Input: Name="Test", EntryNode="NonExistent", Nodes=[{Name:"Existing", Type:"human"}, {Name:"Other", Type:"human"}]
	_, err = components.NewWorkflowDefinition("Test", "", "NonExistent", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /entry.*node.*NonExistent.*not found/i
	assertErrorMatches(t, err, `(?i)entry.*node.*NonExistent.*not found`)
}

// TestWorkflowDefinition_EntryNodeNotHuman rejects WorkflowDefinition with EntryNode referencing agent node
func TestWorkflowDefinition_EntryNodeNotHuman(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "AgentNode", "agent", "Architect", ""),
		createNode(t, "HumanNode", "human", "", ""),
	}
	transitions := []*components.Transition{createTransition(t, "AgentNode", "Done", "HumanNode")}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "AgentNode", "Done", "HumanNode")}

	// Input: Name="Test", EntryNode="AgentNode", Nodes=[{Name:"AgentNode", Type:"agent", AgentRole:"Architect"}, {Name:"HumanNode", Type:"human"}]
	_, err = components.NewWorkflowDefinition("Test", "", "AgentNode", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /entry node.*AgentNode.*type.*human/i
	assertErrorMatches(t, err, `(?i)entry node.*AgentNode.*type.*human`)
}

// TestWorkflowDefinition_EmptyExitTransitions rejects WorkflowDefinition with empty ExitTransitions array
func TestWorkflowDefinition_EmptyExitTransitions(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{}

	// Input: Name="Test", ExitTransitions=[], valid other fields
	_, err = components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /at least one exit transition required/i
	assertErrorMatches(t, err, `(?i)at least one exit transition required`)
}

// TestWorkflowDefinition_ExitTransitionNoMatch rejects WorkflowDefinition when exit transition does not match any defined transition
func TestWorkflowDefinition_ExitTransitionNoMatch(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "A", "human", "", ""),
		createNode(t, "B", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "A", "Start", "B"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "A", "Done", "B"), // mismatch: event type "Done" vs "Start"
	}

	// Input: Name="Test", ExitTransitions=[{FromNode:"A", EventType:"Done", ToNode:"B"}], Transitions=[{FromNode:"A", EventType:"Start", ToNode:"B"}] (mismatch)
	_, err = components.NewWorkflowDefinition("Test", "", "A", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /exit transition.*from_node.*A.*event_type.*Done.*to_node.*B.*no corresponding transition/i
	assertErrorMatches(t, err, `(?i)exit transition.*from_node.*A.*event_type.*Done.*to_node.*B.*no corresponding transition`)
}

// TestWorkflowDefinition_ExitTransitionTargetsAgent rejects WorkflowDefinition when exit transition targets agent node
func TestWorkflowDefinition_ExitTransitionTargetsAgent(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "A", "human", "", ""),
		createNode(t, "AgentNode", "agent", "Architect", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "A", "Done", "AgentNode"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "A", "Done", "AgentNode"),
	}

	// Input: Name="Test", ExitTransitions=[{FromNode:"A", EventType:"Done", ToNode:"AgentNode"}], Nodes includes {Name:"AgentNode", Type:"agent", AgentRole:"Architect"}, matching transition in Transitions
	_, err = components.NewWorkflowDefinition("Test", "", "A", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /exit transition.*to_node.*AgentNode.*must target.*human.*type.*agent/i
	assertErrorMatches(t, err, `(?i)exit transition.*to_node.*AgentNode.*must target.*human.*type.*agent`)
}

// TestWorkflowDefinition_EmptyNodes rejects WorkflowDefinition with empty Nodes array
func TestWorkflowDefinition_EmptyNodes(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="Test", Nodes=[], valid other fields
	_, err = components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /at least one node required/i
	assertErrorMatches(t, err, `(?i)at least one node required`)
}

// TestWorkflowDefinition_DuplicateNodeNames rejects WorkflowDefinition with duplicate node names
func TestWorkflowDefinition_DuplicateNodeNames(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Node1", "human", "", ""),
		createNode(t, "Node1", "agent", "Architect", ""),
	}
	transitions := []*components.Transition{createTransition(t, "Node1", "Done", "Node2")}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "Node1", "Done", "Node2")}

	// Input: Name="Test", Nodes=[{Name:"Node1", Type:"human"}, {Name:"Node1", Type:"agent", AgentRole:"Architect"}]
	_, err = components.NewWorkflowDefinition("Test", "", "Node1", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /duplicate.*node.*name.*Node1/i
	assertErrorMatches(t, err, `(?i)duplicate.*node.*name.*Node1`)
}

// TestWorkflowDefinition_EmptyTransitions rejects WorkflowDefinition with empty Transitions array
func TestWorkflowDefinition_EmptyTransitions(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: Name="Test", Transitions=[], valid other fields
	_, err = components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /at least one transition required/i
	assertErrorMatches(t, err, `(?i)at least one transition required`)
}

// TestWorkflowDefinition_TransitionFromNodeNonExistent rejects WorkflowDefinition when transition references non-existent FromNode
func TestWorkflowDefinition_TransitionFromNodeNonExistent(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{createNode(t, "Existing", "human", "", "")}
	transitions := []*components.Transition{
		createTransition(t, "NonExistent", "Event", "Existing"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "NonExistent", "Event", "Existing"),
	}

	// Input: Name="Test", Transitions=[{FromNode:"NonExistent", EventType:"Event", ToNode:"Existing"}], Nodes=[{Name:"Existing", Type:"human"}]
	_, err = components.NewWorkflowDefinition("Test", "", "Existing", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /transition.*undefined.*node.*NonExistent/i
	assertErrorMatches(t, err, `(?i)transition.*undefined.*node.*NonExistent`)
}

// TestWorkflowDefinition_TransitionToNodeNonExistent rejects WorkflowDefinition when transition references non-existent ToNode
func TestWorkflowDefinition_TransitionToNodeNonExistent(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{createNode(t, "Existing", "human", "", "")}
	transitions := []*components.Transition{
		createTransition(t, "Existing", "Event", "NonExistent"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "Existing", "Event", "NonExistent"),
	}

	// Input: Name="Test", Transitions=[{FromNode:"Existing", EventType:"Event", ToNode:"NonExistent"}], Nodes=[{Name:"Existing", Type:"human"}]
	_, err = components.NewWorkflowDefinition("Test", "", "Existing", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /transition.*undefined.*node.*NonExistent/i
	assertErrorMatches(t, err, `(?i)transition.*undefined.*node.*NonExistent`)
}

// TestWorkflowDefinition_NodeNoOutgoingTransition rejects WorkflowDefinition when non-exit-target node has no outgoing transitions
func TestWorkflowDefinition_NodeNoOutgoingTransition(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Isolated", "human", "", ""),
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "Isolated"),
		createTransition(t, "Start", "Finish", "End"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "Start", "Finish", "End"),
	}

	// Input: Name="Test", Nodes=[{Name:"Isolated", Type:"human"}, {Name:"Start", Type:"human"}, {Name:"End", Type:"human"}], Transitions=[{FromNode:"Start", EventType:"Go", ToNode:"Isolated"}, {FromNode:"Start", EventType:"Finish", ToNode:"End"}], ExitTransitions=[{FromNode:"Start", EventType:"Finish", ToNode:"End"}] (Isolated is not exit target and has no outgoing transitions)
	_, err = components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error; error message matches /node.*Isolated.*no outgoing transitions.*not.*exit target/i
	assertErrorMatches(t, err, `(?i)node.*Isolated.*no outgoing transitions.*not.*exit target`)
}

// TestWorkflowDefinition_FileDoesNotExist returns error when workflow YAML file does not exist
func TestWorkflowDefinition_FileDoesNotExist(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created but empty
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Load workflow with Name="NonExistent"
	yamlPath := filepath.Join(workflowsDir, "NonExistent.yaml")
	_, err = components.LoadWorkflowDefinition(yamlPath)

	// Expected: Returns error; error message matches /workflow.*not found/i
	assertErrorMatches(t, err, `(?i)workflow.*not found`)
}

// TestWorkflowDefinition_MalformedYAML rejects WorkflowDefinition with malformed YAML syntax
func TestWorkflowDefinition_MalformedYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML file with invalid YAML syntax (unclosed quote)
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	yamlContent := `name: "Broken
description: "test"
`
	yamlPath := filepath.Join(workflowsDir, "Broken.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load workflow with Name="Broken"
	_, err = components.LoadWorkflowDefinition(yamlPath)

	// Expected: Returns parse error; error message indicates YAML syntax issue
	require.Error(t, err)
}

// TestWorkflowDefinition_MissingRequiredField rejects WorkflowDefinition with missing required field (entry_node)
func TestWorkflowDefinition_MissingRequiredField(t *testing.T) {
	// Setup: Temporary test directory created; YAML file missing entry_node field
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	yamlContent := `name: "Incomplete"
description: "test"
exit_transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
`
	yamlPath := filepath.Join(workflowsDir, "Incomplete.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load workflow with Name="Incomplete"
	_, err = components.LoadWorkflowDefinition(yamlPath)

	// Expected: Returns error; error message matches /entry.*node.*required/i
	assertErrorMatches(t, err, `(?i)entry.*node.*required`)
}

// TestWorkflowDefinition_DuplicateName rejects loading multiple WorkflowDefinitions with same Name
func TestWorkflowDefinition_DuplicateName(t *testing.T) {
	// Setup: Temporary test directory created; YAML file at DefaultLogicSpec.yaml; second YAML with same name
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	yamlContent := `name: "DefaultLogicSpec"
description: "test"
entry_node: "Start"
exit_transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
nodes:
  - name: "Start"
    type: "human"
  - name: "End"
    type: "human"
transitions:
  - from_node: "Start"
    event_type: "Done"
    to_node: "End"
`
	yamlPath1 := filepath.Join(workflowsDir, "DefaultLogicSpec.yaml")
	err = os.WriteFile(yamlPath1, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load first workflow
	registry := components.NewWorkflowRegistry()
	workflow1, err := components.LoadWorkflowDefinition(yamlPath1)
	require.NoError(t, err)
	err = registry.Register(workflow1)
	require.NoError(t, err)

	// Create second workflow with same name
	yamlPath2 := filepath.Join(workflowsDir, "DefaultLogicSpec2.yaml")
	err = os.WriteFile(yamlPath2, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Input: Load both workflows with Name="DefaultLogicSpec"
	workflow2, err := components.LoadWorkflowDefinition(yamlPath2)
	require.NoError(t, err)
	err = registry.Register(workflow2)

	// Expected: Second workflow load returns error; error message matches /workflow.*DefaultLogicSpec.*already exists/i
	assertErrorMatches(t, err, `(?i)workflow.*DefaultLogicSpec.*already exists`)
}

// TestWorkflowDefinition_ExitTargetNodeMayLackOutgoing Node targeted by exit transition is allowed to have no outgoing transitions
func TestWorkflowDefinition_ExitTargetNodeMayLackOutgoing(t *testing.T) {
	// Setup: Temporary test directory created; .spectra/workflows/ directory created
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "Start", "Exit", "End"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "Start", "Exit", "End"),
	}

	// Input: Name="Test", Nodes=[{Name:"Start", Type:"human"}, {Name:"End", Type:"human"}], Transitions=[{FromNode:"Start", EventType:"Go", ToNode:"End"}, {FromNode:"Start", EventType:"Exit", ToNode:"End"}], ExitTransitions=[{FromNode:"Start", EventType:"Exit", ToNode:"End"}] (End is exit target, has no outgoing)
	workflow, err := components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	// Expected: Workflow validation succeeds; no error
	require.NotNil(t, workflow)
}

// TestWorkflowDefinition_UnreachableNodeRejected Returns error for node with no incoming transitions (except entry node)
func TestWorkflowDefinition_UnreachableNodeRejected(t *testing.T) {
	// Setup: Temporary test directory created; workflow with node "Isolated" that has no incoming transitions and is not the entry node
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

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

	// Input: Validate workflow
	_, err = components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Returns error matching /unreachable.*node.*Isolated/i; workflow rejected
	assertErrorMatches(t, err, `(?i)unreachable.*node.*Isolated`)
}

// TestWorkflowDefinition_ToYAML WorkflowDefinition serializes to YAML correctly
func TestWorkflowDefinition_ToYAML(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", "Test start"),
		createNode(t, "End", "human", "", "Test end"),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "End", "Done", "Start"),
	}

	// Input: WorkflowDefinition with all fields populated
	workflow, err := components.NewWorkflowDefinition(
		"TestWorkflow",
		"Test description",
		"Start",
		exitTransitions,
		nodes,
		transitions,
	)
	require.NoError(t, err)

	yamlPath := filepath.Join(workflowsDir, "TestWorkflow.yaml")
	err = workflow.SaveToFile(yamlPath)
	require.NoError(t, err)

	// Expected: YAML contains name, description, entry_node, exit_transitions, nodes, transitions with correct values and structure
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	yamlStr := string(content)
	require.Contains(t, yamlStr, "name: TestWorkflow")
	require.Contains(t, yamlStr, "description:")
	require.Contains(t, yamlStr, "entry_node: Start")
	require.Contains(t, yamlStr, "exit_transitions:")
	require.Contains(t, yamlStr, "nodes:")
	require.Contains(t, yamlStr, "transitions:")
}

// TestWorkflowDefinition_ToYAMLEmptyDescription WorkflowDefinition with empty description serializes to YAML correctly
func TestWorkflowDefinition_ToYAMLEmptyDescription(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	// Input: WorkflowDefinition with Description=""
	workflow, err := components.NewWorkflowDefinition("TestWorkflow", "", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	yamlPath := filepath.Join(workflowsDir, "TestWorkflow.yaml")
	err = workflow.SaveToFile(yamlPath)
	require.NoError(t, err)

	// Expected: YAML contains description: ""
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	yamlStr := string(content)
	require.Contains(t, yamlStr, `description: ""`)
}

// TestWorkflowDefinition_ToYAMLMultipleExitTransitions WorkflowDefinition with multiple exit transitions serializes to YAML correctly
func TestWorkflowDefinition_ToYAMLMultipleExitTransitions(t *testing.T) {
	// Setup: Temporary test directory created; YAML output written to test directory
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End1", "human", "", ""),
		createNode(t, "End2", "human", "", ""),
		createNode(t, "End3", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go1", "End1"),
		createTransition(t, "Start", "Go2", "End2"),
		createTransition(t, "Start", "Go3", "End3"),
		createTransition(t, "End1", "Done1", "Start"),
		createTransition(t, "End2", "Done2", "Start"),
		createTransition(t, "End3", "Done3", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "End1", "Done1", "Start"),
		createExitTransition(t, "End2", "Done2", "Start"),
		createExitTransition(t, "End3", "Done3", "Start"),
	}

	// Input: WorkflowDefinition with 3 exit transitions
	workflow, err := components.NewWorkflowDefinition("TestWorkflow", "", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	yamlPath := filepath.Join(workflowsDir, "TestWorkflow.yaml")
	err = workflow.SaveToFile(yamlPath)
	require.NoError(t, err)

	// Expected: YAML exit_transitions array contains all 3 transitions; order preserved
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	yamlStr := string(content)
	require.Contains(t, yamlStr, "exit_transitions:")
	// Count occurrences - should have 3 exit transitions
	require.Contains(t, yamlStr, "Done1")
	require.Contains(t, yamlStr, "Done2")
	require.Contains(t, yamlStr, "Done3")
}

// TestWorkflowDefinition_FieldsImmutable WorkflowDefinition fields cannot be modified after creation
func TestWorkflowDefinition_FieldsImmutable(t *testing.T) {
	// Setup: WorkflowDefinition instance created
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	workflow, err := components.NewWorkflowDefinition("TestWorkflow", "Test", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	// Expected: Field modification attempt fails or has no effect; original values remain
	// In Go, we enforce immutability through unexported fields and getter methods only
	require.Equal(t, "TestWorkflow", workflow.GetName())
	require.Equal(t, "Test", workflow.GetDescription())
	require.Equal(t, "Start", workflow.GetEntryNode())
	require.Len(t, workflow.GetExitTransitions(), 1)
	require.Len(t, workflow.GetNodes(), 2)
	require.Len(t, workflow.GetTransitions(), 2)
}

// TestWorkflowDefinition_ImplementsInterface WorkflowDefinition type implements expected interface
func TestWorkflowDefinition_ImplementsInterface(t *testing.T) {
	// Setup: WorkflowDefinition instance created
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "End", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "End"),
		createTransition(t, "End", "Done", "Start"),
	}
	exitTransitions := []*components.ExitTransition{createExitTransition(t, "End", "Done", "Start")}

	workflow, err := components.NewWorkflowDefinition("TestWorkflow", "", "Start", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	// Expected: WorkflowDefinition satisfies WorkflowDefinition interface contract
	// Verify all getter methods are available
	_ = workflow.GetName()
	_ = workflow.GetDescription()
	_ = workflow.GetEntryNode()
	_ = workflow.GetExitTransitions()
	_ = workflow.GetNodes()
	_ = workflow.GetTransitions()
}

// TestWorkflowDefinition_NodeOrderPreserved Nodes in workflow Nodes array preserve definition order
func TestWorkflowDefinition_NodeOrderPreserved(t *testing.T) {
	// Setup: Temporary test directory created; workflow definition in test directory
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Add 5 nodes in order: N1, N2, N3, N4, N5
	nodes := []*components.Node{
		createNode(t, "N1", "human", "", ""),
		createNode(t, "N2", "human", "", ""),
		createNode(t, "N3", "human", "", ""),
		createNode(t, "N4", "human", "", ""),
		createNode(t, "N5", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "N1", "E1", "N2"),
		createTransition(t, "N2", "E2", "N3"),
		createTransition(t, "N3", "E3", "N4"),
		createTransition(t, "N4", "E4", "N5"),
		createTransition(t, "N5", "Done", "N1"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "N5", "Done", "N1"),
	}

	workflow, err := components.NewWorkflowDefinition("Test", "", "N1", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	// Expected: Query workflow; nodes returned in order: N1, N2, N3, N4, N5
	returnedNodes := workflow.GetNodes()
	require.Len(t, returnedNodes, 5)
	require.Equal(t, "N1", returnedNodes[0].GetName())
	require.Equal(t, "N2", returnedNodes[1].GetName())
	require.Equal(t, "N3", returnedNodes[2].GetName())
	require.Equal(t, "N4", returnedNodes[3].GetName())
	require.Equal(t, "N5", returnedNodes[4].GetName())
}

// TestWorkflowDefinition_TransitionOrderPreserved Transitions in workflow Transitions array preserve definition order
func TestWorkflowDefinition_TransitionOrderPreserved(t *testing.T) {
	// Setup: Temporary test directory created; workflow definition in test directory
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Add 5 transitions in order: T1, T2, T3, T4, T5
	nodes := []*components.Node{
		createNode(t, "N1", "human", "", ""),
		createNode(t, "N2", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "N1", "T1", "N2"),
		createTransition(t, "N2", "T2", "N1"),
		createTransition(t, "N1", "T3", "N2"),
		createTransition(t, "N2", "T4", "N1"),
		createTransition(t, "N1", "T5", "N2"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "N2", "T2", "N1"),
	}

	workflow, err := components.NewWorkflowDefinition("Test", "", "N1", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	// Expected: Query workflow; transitions returned in order: T1, T2, T3, T4, T5
	returnedTransitions := workflow.GetTransitions()
	require.Len(t, returnedTransitions, 5)
	require.Equal(t, "T1", returnedTransitions[0].GetEventType())
	require.Equal(t, "T2", returnedTransitions[1].GetEventType())
	require.Equal(t, "T3", returnedTransitions[2].GetEventType())
	require.Equal(t, "T4", returnedTransitions[3].GetEventType())
	require.Equal(t, "T5", returnedTransitions[4].GetEventType())
}

// TestWorkflowDefinition_ExitTransitionOrderPreserved Exit transitions in workflow ExitTransitions array preserve definition order
func TestWorkflowDefinition_ExitTransitionOrderPreserved(t *testing.T) {
	// Setup: Temporary test directory created; workflow definition in test directory
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err)

	// Input: Add 3 exit transitions in order: E1, E2, E3
	nodes := []*components.Node{
		createNode(t, "N1", "human", "", ""),
		createNode(t, "N2", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "N1", "E1", "N2"),
		createTransition(t, "N1", "E2", "N2"),
		createTransition(t, "N1", "E3", "N2"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "N1", "E1", "N2"),
		createExitTransition(t, "N1", "E2", "N2"),
		createExitTransition(t, "N1", "E3", "N2"),
	}

	workflow, err := components.NewWorkflowDefinition("Test", "", "N1", exitTransitions, nodes, transitions)
	require.NoError(t, err)

	// Expected: Query workflow; exit transitions returned in order: E1, E2, E3
	returnedExitTransitions := workflow.GetExitTransitions()
	require.Len(t, returnedExitTransitions, 3)
	require.Equal(t, "E1", returnedExitTransitions[0].GetEventType())
	require.Equal(t, "E2", returnedExitTransitions[1].GetEventType())
	require.Equal(t, "E3", returnedExitTransitions[2].GetEventType())
}

// TestWorkflowDefinition_ExitTargetWithOutgoingWarning allows exit target node to have outgoing transitions
func TestWorkflowDefinition_ExitTargetWithOutgoingWarning(t *testing.T) {
	// Setup: Workflow with exit transition to "Final"; "Final" has outgoing transition to "Start"
	nodes := []*components.Node{
		createNode(t, "Start", "human", "", ""),
		createNode(t, "Final", "human", "", ""),
	}
	transitions := []*components.Transition{
		createTransition(t, "Start", "Go", "Final"),
		createTransition(t, "Final", "Back", "Start"),
	}
	exitTransitions := []*components.ExitTransition{
		createExitTransition(t, "Start", "Go", "Final"),
	}

	// Input: Validate workflow
	workflow, err := components.NewWorkflowDefinition("Test", "", "Start", exitTransitions, nodes, transitions)

	// Expected: Workflow creation succeeds; Runtime should handle warning about unreachable transitions
	require.NoError(t, err)
	require.NotNil(t, workflow)
}
