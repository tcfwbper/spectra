# Test Specification: `transition_to_node.go`

## Source File Under Test
`runtime/transition_to_node.go`

## Test File
`runtime/transition_to_node_test.go`

---

## `TransitionToNode`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionToNode_New` | `unit` | Constructs TransitionToNode with valid inputs. | Test fixture; mock Session, valid WorkflowDefinition, mock AgentDefinitionLoader, mock AgentInvoker, TerminationNotifier channel | `Session=<mock>`, `WorkflowDefinition=<valid>`, `AgentDefinitionLoader=<mock>`, `AgentInvoker=<mock>`, `TerminationNotifier=<channel>` | Returns TransitionToNode instance; no error |

### Happy Path — Transition (Human Node, Regular)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_HumanNode_PrintsMessage` | `unit` | Prints message to stdout for human node. | Redirect stdout to test buffer; mock Session with valid state; WorkflowDefinition contains human node "NodeA"; mock Session.UpdateCurrentStateSafe succeeds | `Message="Hello world"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Stdout contains `[Human Node: NodeA] Hello world\n`; Session.UpdateCurrentStateSafe called with "NodeA"; returns `nil` |
| `TestTransition_HumanNode_EmptyMessage` | `unit` | Prints placeholder when message is empty. | Redirect stdout to test buffer; mock Session; WorkflowDefinition contains human node "NodeB" | `Message=""`, `TargetNodeName="NodeB"`, `IsExitTransition=false` | Stdout contains `[Human Node: NodeB] (no message)\n`; Session.UpdateCurrentStateSafe called; returns `nil` |
| `TestTransition_HumanNode_MessageWithNewlines` | `unit` | Preserves newlines and special characters in message. | Redirect stdout to test buffer; mock Session; WorkflowDefinition contains human node "NodeC" | `Message="Line1\nLine2\tTab"`, `TargetNodeName="NodeC"`, `IsExitTransition=false` | Stdout contains `[Human Node: NodeC] Line1\nLine2\tTab\n` with exact formatting preserved; returns `nil` |
| `TestTransition_HumanNode_MessageWithQuotes` | `unit` | Preserves quotes in message. | Redirect stdout to test buffer; mock Session; WorkflowDefinition contains human node "NodeD" | `Message="He said \"hello\""`, `TargetNodeName="NodeD"`, `IsExitTransition=false` | Stdout contains `[Human Node: NodeD] He said "hello"\n`; returns `nil` |
| `TestTransition_HumanNode_LargeMessage` | `unit` | Handles very large message (1 MB). | Redirect stdout to test buffer; mock Session; WorkflowDefinition contains human node "NodeE" | `Message=<1MB-string>`, `TargetNodeName="NodeE"`, `IsExitTransition=false` | Stdout contains complete message with prefix; returns `nil` |
| `TestTransition_HumanNode_UpdatesCurrentState` | `unit` | Updates Session.CurrentState after printing. | Mock Session tracks CurrentState updates; WorkflowDefinition contains human node "NodeF"; redirect stdout | `Message="test"`, `TargetNodeName="NodeF"`, `IsExitTransition=false` | Message printed first, then Session.UpdateCurrentStateSafe called with "NodeF"; returns `nil` |

### Happy Path — Transition (Agent Node, Regular)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_AgentNode_InvokesAgent` | `unit` | Invokes AgentInvoker for agent node. | Mock Session; WorkflowDefinition contains agent node "AgentNode1" with `agent_role="reviewer"`; mock AgentDefinitionLoader returns valid definition; mock AgentInvoker succeeds | `Message="Review this"`, `TargetNodeName="AgentNode1"`, `IsExitTransition=false` | AgentDefinitionLoader.Load called with `role="reviewer"`; AgentInvoker.InvokeAgent called with `NodeName="AgentNode1"`, `Message="Review this"`, `AgentDefinition=<loaded>`; Session.UpdateCurrentStateSafe called; returns `nil` |
| `TestTransition_AgentNode_PassesMessageUnmodified` | `unit` | Passes message to AgentInvoker without modification. | Mock Session; agent node with role; mock AgentDefinitionLoader; mock AgentInvoker tracks arguments | `Message="Complex message with 🎉 unicode"`, `TargetNodeName="AgentNode2"`, `IsExitTransition=false` | AgentInvoker receives exact message: `"Complex message with 🎉 unicode"`; returns `nil` |
| `TestTransition_AgentNode_UpdatesCurrentStateAfterInvoke` | `unit` | Updates CurrentState after agent invocation. | Mock Session tracks call order; agent node; mock AgentDefinitionLoader; mock AgentInvoker succeeds | `Message="test"`, `TargetNodeName="AgentNode3"`, `IsExitTransition=false` | AgentInvoker.InvokeAgent called first, then Session.UpdateCurrentStateSafe called with "AgentNode3"; returns `nil` |

### Happy Path — Transition (Exit Transition)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ExitTransition_SkipsHumanAction` | `unit` | Skips stdout print for exit transition to human node. | Redirect stdout to test buffer; mock Session; WorkflowDefinition contains human node "ExitNode"; mock Session.Done succeeds | `Message="final message"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Stdout is empty (no print); Session.UpdateCurrentStateSafe called with "ExitNode"; Session.Done called; returns `nil` |
| `TestTransition_ExitTransition_SkipsAgentAction` | `unit` | Skips agent invocation for exit transition to agent node. | Mock Session; WorkflowDefinition contains agent node "ExitAgent"; mock AgentInvoker tracks calls; mock Session.Done succeeds | `Message="final"`, `TargetNodeName="ExitAgent"`, `IsExitTransition=true` | AgentDefinitionLoader NOT called; AgentInvoker NOT called; Session.UpdateCurrentStateSafe called; Session.Done called; returns `nil` |
| `TestTransition_ExitTransition_UpdatesStateBeforeDone` | `unit` | Updates CurrentState before calling Session.Done. | Mock Session tracks call order; WorkflowDefinition contains human node "ExitNode"; Session.Done succeeds | `Message=""`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Session.UpdateCurrentStateSafe called with "ExitNode" before Session.Done; returns `nil` |
| `TestTransition_ExitTransition_CallsDone` | `unit` | Calls Session.Done after state update. | Mock Session; WorkflowDefinition contains human node "ExitNode"; Session.Done succeeds | `Message="done"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Session.UpdateCurrentStateSafe called, then Session.Done called with terminationNotifier; returns `nil` |

### Validation Failures — Target Node

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_TargetNodeNotFound` | `unit` | Returns error when target node does not exist in workflow. | Mock Session; WorkflowDefinition does NOT contain "NonExistentNode"; mock Session.Fail tracks calls | `Message="test"`, `TargetNodeName="NonExistentNode"`, `IsExitTransition=false` | RuntimeError constructed with `Issuer="TransitionToNode"`, `Message="target node not found: 'NonExistentNode'"`; Session.Fail called with RuntimeError; returns error matching `/target node 'NonExistentNode' not found in workflow/i` |

### Validation Failures — Agent Definition Loading

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_AgentDefinitionNotFound` | `unit` | Returns error when agent definition not found. | Mock Session; WorkflowDefinition contains agent node "AgentNode" with `agent_role="unknown"`; mock AgentDefinitionLoader returns error "agent file not found" | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | RuntimeError constructed with `Issuer="TransitionToNode"`, message matching `/failed to load agent definition for role 'unknown'/i`; Session.Fail called; returns error matching `/failed to load agent definition for role 'unknown':/i` |
| `TestTransition_AgentDefinitionLoadError` | `unit` | Returns error when agent definition has parse error. | Mock Session; WorkflowDefinition contains agent node with role; mock AgentDefinitionLoader returns error "invalid YAML" | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | RuntimeError with `Issuer="TransitionToNode"`; Session.Fail called with details "invalid YAML"; returns error matching `/failed to load agent definition for role.*invalid YAML/i` |
| `TestTransition_AgentNodeEmptyRole` | `unit` | Returns error when agent node has empty agent_role. | Mock Session; WorkflowDefinition contains agent node "AgentNode" with `agent_role=""`; mock AgentDefinitionLoader returns error | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | AgentDefinitionLoader called with empty role, returns error; Session.Fail called; returns error matching `/failed to load agent definition/i` |

### Validation Failures — Agent Invocation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_AgentInvokerFails` | `unit` | Returns error when AgentInvoker fails. | Mock Session; agent node with valid role; mock AgentDefinitionLoader succeeds; mock AgentInvoker returns error "claude command not found" | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | RuntimeError with `Issuer="TransitionToNode"`, message matching `/failed to invoke agent for node 'AgentNode'/i`; Session.Fail called with details; returns error matching `/failed to invoke agent for node 'AgentNode'.*claude command not found/i` |
| `TestTransition_AgentInvokerPermissionDenied` | `unit` | Returns error when agent invocation denied due to permissions. | Mock Session; agent node with valid role; mock AgentDefinitionLoader succeeds; mock AgentInvoker returns error "permission denied" | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | Session.Fail called with RuntimeError including "permission denied"; returns error |
| `TestTransition_AgentInvokerWorkingDirInvalid` | `unit` | Returns error when working directory is invalid. | Mock Session; agent node with valid role; mock AgentDefinitionLoader succeeds; mock AgentInvoker returns error "working directory not found" | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | Session.Fail called with RuntimeError; returns error matching `/failed to invoke agent.*working directory not found/i` |

### Validation Failures — Session.Done

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_SessionDoneFails_StatusNotRunning` | `unit` | Returns error when Session.Done fails due to invalid status. | Mock Session with `Status="completed"`; WorkflowDefinition contains human node "ExitNode"; Session.Done returns error "cannot complete session: status is 'completed', expected 'running'" | `Message="test"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | RuntimeError constructed with `Issuer="TransitionToNode"`, `Message="failed to complete session after exit transition"`; Session.Fail called; returns error matching `/failed to complete session after exit transition:.*cannot complete session: status is 'completed'/i` |
| `TestTransition_SessionDoneFails_SessionAlreadyFailed` | `unit` | Returns error when Session.Done fails because session already failed. | Mock Session with `Status="failed"`; WorkflowDefinition contains human node "ExitNode"; Session.Done returns error "cannot complete session: status is 'failed', expected 'running'" | `Message="test"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Session.Fail called (returns error "session already failed"); returns error matching `/failed to complete session after exit transition/i` |

### Error Propagation — Session.Fail Always Called

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_SessionFailCalledOnNodeNotFound` | `unit` | Calls Session.Fail when target node not found. | Mock Session tracks Session.Fail calls; WorkflowDefinition missing target node | `Message="test"`, `TargetNodeName="Missing"`, `IsExitTransition=false` | Session.Fail called once with RuntimeError `Issuer="TransitionToNode"`, `Message="target node not found: 'Missing'"`; returns error |
| `TestTransition_SessionFailCalledOnAgentLoadError` | `unit` | Calls Session.Fail when agent definition load fails. | Mock Session tracks calls; agent node; AgentDefinitionLoader returns error | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | Session.Fail called once with RuntimeError including load error details; returns error |
| `TestTransition_SessionFailCalledOnAgentInvokeError` | `unit` | Calls Session.Fail when agent invocation fails. | Mock Session tracks calls; agent node; AgentDefinitionLoader succeeds; AgentInvoker returns error | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | Session.Fail called once with RuntimeError including invocation error details; returns error |
| `TestTransition_SessionFailCalledOnDoneError` | `unit` | Calls Session.Fail when Session.Done fails. | Mock Session tracks calls; human exit node; Session.Done returns error | `Message="test"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Session.Fail called once with RuntimeError about done failure; returns error |
| `TestTransition_SessionFailNotCalledOnSuccess` | `unit` | Does not call Session.Fail when transition succeeds. | Mock Session tracks calls; human node; Session.UpdateCurrentStateSafe succeeds | `Message="test"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Session.Fail NOT called; returns `nil` |

### Error Propagation — Caller Responsibility

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_CallerDoesNotCallFailAgain` | `unit` | Documents that caller (EventProcessor) must not call Session.Fail again. | Mock Session; target node not found; Session.Fail called by TransitionToNode | `Message="test"`, `TargetNodeName="Missing"`, `IsExitTransition=false` | TransitionToNode calls Session.Fail internally; returns error; caller (test simulates EventProcessor) observes error return and does NOT call Session.Fail again |

### State Transitions — Action-Before-State-Update

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_HumanNode_PrintBeforeStateUpdate` | `unit` | Prints message before updating CurrentState. | Mock Session tracks call order; redirect stdout; human node | `Message="test"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Stdout write occurs before Session.UpdateCurrentStateSafe call; verified via call order tracking |
| `TestTransition_AgentNode_InvokeBeforeStateUpdate` | `unit` | Invokes agent before updating CurrentState. | Mock Session tracks call order; agent node; AgentDefinitionLoader succeeds; AgentInvoker succeeds | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | AgentInvoker.InvokeAgent called before Session.UpdateCurrentStateSafe; verified via call order |
| `TestTransition_ExitTransition_SkipsActionNoOrdering` | `unit` | Exit transition skips action, no action-before-state ordering applies. | Mock Session tracks calls; human exit node | `Message="test"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | No stdout write; no agent invoke; Session.UpdateCurrentStateSafe called before Session.Done |

### State Transitions — State-Before-Done

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ExitTransition_StateUpdateBeforeDone` | `unit` | Updates CurrentState before calling Session.Done. | Mock Session tracks call order; human exit node | `Message="test"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Session.UpdateCurrentStateSafe called with "ExitNode" before Session.Done called; verified via mock call order |

### Idempotency — UpdateCurrentStateSafe Always Succeeds

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_UpdateCurrentStateSafeAlwaysNil` | `unit` | Session.UpdateCurrentStateSafe returns nil for valid node name. | Mock Session.UpdateCurrentStateSafe returns nil (best-effort contract); human node | `Message="test"`, `TargetNodeName="ValidNode"`, `IsExitTransition=false` | Session.UpdateCurrentStateSafe called; return value not checked; transition continues successfully |
| `TestTransition_UpdateCurrentStateSafePersistenceFails` | `unit` | Continues when persistence fails during UpdateCurrentStateSafe. | Mock Session.UpdateCurrentStateSafe simulates persistence failure (logs warning, returns nil); human node | `Message="test"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Session.UpdateCurrentStateSafe returns nil; TransitionToNode does not check return value; in-memory state authoritative; returns `nil` |

### Boundary Values — Message Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_MessageVeryLarge` | `unit` | Handles very large message (5 MB). | Redirect stdout; mock Session; human node | `Message=<5MB-string>`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Complete message printed to stdout with prefix; returns `nil` (unless stdout write fails at system level) |
| `TestTransition_MessageUnicodeCharacters` | `unit` | Preserves Unicode characters in message. | Redirect stdout; mock Session; human node | `Message="测试🎉emoji"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Stdout contains `[Human Node: NodeA] 测试🎉emoji\n` with exact Unicode preserved; returns `nil` |
| `TestTransition_MessageWithEmbeddedNulls` | `unit` | Handles message with embedded null bytes (if supported by language). | Redirect stdout; mock Session; human node | `Message="before\x00after"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Stdout contains message as-is (Go strings support null bytes); returns `nil` |

### Boundary Values — Target Node Names

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_TargetNodePascalCase` | `unit` | Accepts PascalCase target node name. | Mock Session; WorkflowDefinition contains "MyNodeName" | `TargetNodeName="MyNodeName"`, `IsExitTransition=false` | Node found; transition succeeds |
| `TestTransition_TargetNodeSingleCharacter` | `unit` | Accepts single-character node name. | Mock Session; WorkflowDefinition contains "A" | `TargetNodeName="A"`, `IsExitTransition=false` | Node found; transition succeeds |
| `TestTransition_TargetNodeLongName` | `unit` | Accepts very long node name. | Mock Session; WorkflowDefinition contains node with 256-character name | `TargetNodeName=<256-char-PascalCase>`, `IsExitTransition=false` | Node found; transition succeeds |

### Mock / Dependency Interaction — AgentDefinitionLoader

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_AgentDefinitionLoaderCalledWithCorrectRole` | `unit` | Passes correct agent_role to AgentDefinitionLoader. | Mock AgentDefinitionLoader tracks arguments; agent node with `agent_role="code_reviewer"` | `TargetNodeName="AgentNode"`, `IsExitTransition=false` | AgentDefinitionLoader.Load called with `role="code_reviewer"`; AgentInvoker receives loaded definition |
| `TestTransition_AgentDefinitionLoaderNotCalledForHuman` | `unit` | Does not call AgentDefinitionLoader for human nodes. | Mock AgentDefinitionLoader tracks calls; human node | `TargetNodeName="HumanNode"`, `IsExitTransition=false` | AgentDefinitionLoader NOT called; only stdout print occurs |
| `TestTransition_AgentDefinitionLoaderNotCalledForExit` | `unit` | Does not call AgentDefinitionLoader for exit transitions. | Mock AgentDefinitionLoader tracks calls; agent exit node | `TargetNodeName="ExitAgent"`, `IsExitTransition=true` | AgentDefinitionLoader NOT called; no agent invocation |

### Mock / Dependency Interaction — AgentInvoker

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_AgentInvokerCalledWithCorrectArguments` | `unit` | Passes correct arguments to AgentInvoker. | Mock AgentInvoker tracks arguments; agent node "AgentNode"; AgentDefinitionLoader returns definition | `Message="Do this task"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | AgentInvoker.InvokeAgent called with `NodeName="AgentNode"`, `Message="Do this task"`, `AgentDefinition=<loaded-def>` |
| `TestTransition_AgentInvokerNotCalledForHuman` | `unit` | Does not call AgentInvoker for human nodes. | Mock AgentInvoker tracks calls; human node | `TargetNodeName="HumanNode"`, `IsExitTransition=false` | AgentInvoker NOT called; only stdout print occurs |
| `TestTransition_AgentInvokerNotCalledForExit` | `unit` | Does not call AgentInvoker for exit transitions. | Mock AgentInvoker tracks calls; agent exit node | `TargetNodeName="ExitAgent"`, `IsExitTransition=true` | AgentInvoker NOT called; action skipped for exit transitions |

### Mock / Dependency Interaction — Session Methods

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_UpdateCurrentStateSafeCalledForAll` | `unit` | Calls Session.UpdateCurrentStateSafe for all transition types. | Mock Session tracks calls; test regular and exit transitions | Regular transition and exit transition | Session.UpdateCurrentStateSafe called for both cases with correct target node name |
| `TestTransition_DoneOnlyCalledForExit` | `unit` | Calls Session.Done only for exit transitions. | Mock Session tracks calls; test regular and exit transitions | Regular transition with `IsExitTransition=false`, then exit with `IsExitTransition=true` | Session.Done called only for exit transition; not called for regular transition |
| `TestTransition_SessionMethodsNotCalledOnEarlyError` | `unit` | Does not call Session methods when early validation fails. | Mock Session tracks all method calls; WorkflowDefinition missing target node | `TargetNodeName="Missing"`, `IsExitTransition=false` | Only Session.Fail called; Session.UpdateCurrentStateSafe and Session.Done NOT called |
| `TestTransition_SessionUpdateCalledEvenIfPersistenceFails` | `unit` | Calls Session.UpdateCurrentStateSafe even when persistence fails. | Mock Session.UpdateCurrentStateSafe simulates persistence failure (returns nil); human node | `Message="test"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Session.UpdateCurrentStateSafe called; returns nil; TransitionToNode proceeds successfully |

### Concurrent Behaviour — Thread Safety

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ConcurrentCallsSameSession` | `race` | Handles concurrent transitions on the same session safely. | TransitionToNode instance shared; mock Session with thread-safe methods; 5 goroutines call Transition simultaneously | 5 concurrent calls with different target nodes | All calls complete; Session methods serialize via write lock; last UpdateCurrentStateSafe wins; no data races detected |
| `TestTransition_ConcurrentFailures` | `race` | Handles concurrent failures safely. | TransitionToNode shared; mock Session; 3 goroutines with invalid target nodes | 3 concurrent calls with missing target nodes | Session.Fail called multiple times (first error wins in Session); all goroutines return errors; no data races |
| `TestTransition_ConcurrentExitTransitions` | `race` | Handles concurrent exit transitions safely. | TransitionToNode shared; mock Session; 2 goroutines call with exit transitions | 2 concurrent exit transitions | One Session.Done succeeds, other may fail with status error; first completion wins; no data races |

### Resource Cleanup — No Rollback

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_NoRollbackOnDoneFailure` | `unit` | Does not roll back agent invocation if Session.Done fails. | Mock Session; agent exit node; AgentInvoker succeeds; Session.UpdateCurrentStateSafe succeeds; Session.Done returns error | `Message="test"`, `TargetNodeName="ExitAgent"`, `IsExitTransition=true` | AgentInvoker called (process started); Session.UpdateCurrentStateSafe called (state updated); Session.Done fails; no rollback of previous steps; Session.Fail called; returns error |
| `TestTransition_NoRollbackOnAgentInvokeFailure` | `unit` | Does not roll back state changes if agent invocation fails mid-way. | Mock Session; agent node; AgentDefinitionLoader succeeds; AgentInvoker returns error after partial execution | `Message="test"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | AgentInvoker error returned; Session.Fail called; no attempt to undo previous successful steps; returns error |

### Panic Recovery — Not Handled

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_PanicNotRecovered` | `unit` | Does not recover from panics (delegates to MessageRouter). | Mock Session; WorkflowDefinition lookup panics (simulated programming error) | `Message="test"`, `TargetNodeName="NodeA"`, `IsExitTransition=false` | Panic propagates to caller (not caught by TransitionToNode); caller (MessageRouter) handles panic recovery |

### Boundary Values — IsExitTransition Flag

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_IsExitTransitionTrue_HumanNode` | `unit` | Exit transition to human node skips action. | Redirect stdout; mock Session; human node | `Message="exit"`, `TargetNodeName="HumanExit"`, `IsExitTransition=true` | No stdout print; Session.UpdateCurrentStateSafe called; Session.Done called; returns `nil` |
| `TestTransition_IsExitTransitionFalse_HumanNode` | `unit` | Regular transition to human node performs action. | Redirect stdout; mock Session; human node | `Message="regular"`, `TargetNodeName="HumanNode"`, `IsExitTransition=false` | Stdout print occurs; Session.UpdateCurrentStateSafe called; Session.Done NOT called; returns `nil` |
| `TestTransition_IsExitTransitionTrue_AgentNode` | `unit` | Exit transition to agent node skips action (despite validation forbidding this). | Mock Session; agent node; AgentInvoker tracks calls | `Message="exit"`, `TargetNodeName="AgentExit"`, `IsExitTransition=true` | AgentDefinitionLoader NOT called; AgentInvoker NOT called; Session.UpdateCurrentStateSafe called; Session.Done called; returns `nil` (bypass of workflow validation) |
| `TestTransition_IsExitTransitionFalse_AgentNode` | `unit` | Regular transition to agent node performs action. | Mock Session; agent node; AgentDefinitionLoader succeeds; AgentInvoker succeeds | `Message="regular"`, `TargetNodeName="AgentNode"`, `IsExitTransition=false` | AgentInvoker called; Session.UpdateCurrentStateSafe called; Session.Done NOT called; returns `nil` |

### Edge Cases — TerminationNotifier Channel

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_TerminationNotifierPassedToDone` | `unit` | Passes terminationNotifier to Session.Done. | Mock Session.Done tracks arguments; human exit node | `Message="test"`, `TargetNodeName="ExitNode"`, `IsExitTransition=true` | Session.Done called with terminationNotifier channel passed at initialization; verified via mock |
| `TestTransition_TerminationNotifierPassedToFail` | `unit` | Passes terminationNotifier to Session.Fail. | Mock Session.Fail tracks arguments; target node not found | `Message="test"`, `TargetNodeName="Missing"`, `IsExitTransition=false` | Session.Fail called with RuntimeError and terminationNotifier channel; verified via mock |

### Edge Cases — RuntimeError Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_RuntimeErrorIssuerAlwaysTransitionToNode` | `unit` | RuntimeError always has Issuer="TransitionToNode". | Mock Session tracks Session.Fail arguments; test multiple failure scenarios | Various error conditions (node not found, agent load fail, invoke fail, done fail) | All RuntimeErrors have `Issuer="TransitionToNode"`; verified via mock inspection of Session.Fail calls |
| `TestTransition_RuntimeErrorMessageDescriptive` | `unit` | RuntimeError messages describe the failure. | Mock Session tracks Session.Fail arguments; target node not found | `TargetNodeName="Missing"` | RuntimeError has `Message="target node not found: 'Missing'"`; descriptive and includes node name |
| `TestTransition_RuntimeErrorIncludesDetails` | `unit` | RuntimeError includes error details from dependencies. | Mock Session; agent node; AgentInvoker returns error "command not found" | `TargetNodeName="AgentNode"` | RuntimeError passed to Session.Fail includes details about "command not found"; verified via mock |

### Happy Path — Message Format Verification

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_HumanNodeMessageFormat` | `unit` | Verifies exact stdout format for human node. | Redirect stdout to test buffer; mock Session; human node "TestNode" | `Message="Test message"`, `TargetNodeName="TestNode"`, `IsExitTransition=false` | Stdout exactly matches `[Human Node: TestNode] Test message\n` with single newline at end |
| `TestTransition_HumanNodeEmptyMessageFormat` | `unit` | Verifies exact format for empty message. | Redirect stdout; mock Session; human node "EmptyNode" | `Message=""`, `TargetNodeName="EmptyNode"`, `IsExitTransition=false` | Stdout exactly matches `[Human Node: EmptyNode] (no message)\n` |
