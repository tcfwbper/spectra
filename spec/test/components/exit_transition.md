# Test Specification: `exit_transition_test.go`

## Source File Under Test
`components/exit_transition.go`

## Test File
`components/exit_transition_test.go`

---

## `ExitTransition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewExitTransition_ValidInputs` | `unit` | Constructs an ExitTransition with all valid fields. | | `fromNode="Reviewer"`, `eventType="Approved"`, `toNode="HumanApproval"` | Returns no error; all getters return the provided values |
| `TestNewExitTransition_SameFromAndToNode` | `unit` | Accepts ExitTransition where FromNode equals ToNode (self-loop not prohibited at this level). | | `fromNode="HumanApproval"`, `eventType="Completed"`, `toNode="HumanApproval"` | Returns no error; all getters return the provided values |

### Validation Failures — FromNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewExitTransition_EmptyFromNode` | `unit` | Rejects empty string FromNode. | | `fromNode=""` with other fields valid | Returns error `"from_node cannot be empty"` |
| `TestNewExitTransition_FromNodeStartsLowercase` | `unit` | Rejects FromNode starting with a lowercase letter. | | `fromNode="reviewer"` with other fields valid | Returns error indicating PascalCase is required |
| `TestNewExitTransition_FromNodeContainsSpecialChar` | `unit` | Rejects FromNode with non-alphanumeric characters. | | `fromNode="Review-Node"` with other fields valid | Returns error indicating PascalCase is required |

### Validation Failures — EventType

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewExitTransition_EmptyEventType` | `unit` | Rejects empty string EventType. | | `eventType=""` with other fields valid | Returns error `"event_type cannot be empty"` |
| `TestNewExitTransition_EventTypeStartsLowercase` | `unit` | Rejects EventType starting with a lowercase letter. | | `eventType="approved"` with other fields valid | Returns error indicating PascalCase is required |
| `TestNewExitTransition_EventTypeContainsHyphen` | `unit` | Rejects EventType with non-alphanumeric characters. | | `eventType="Requirement-Approved"` with other fields valid | Returns error indicating PascalCase is required |

### Validation Failures — ToNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewExitTransition_EmptyToNode` | `unit` | Rejects empty string ToNode. | | `toNode=""` with other fields valid | Returns error `"to_node cannot be empty"` |
| `TestNewExitTransition_ToNodeStartsLowercase` | `unit` | Rejects ToNode starting with a lowercase letter. | | `toNode="humanApproval"` with other fields valid | Returns error indicating PascalCase is required |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_Immutability` | `unit` | All fields remain unchanged after construction; no exported setters exist. | Construct a valid ExitTransition with `fromNode="Reviewer"`, `eventType="Approved"`, `toNode="HumanApproval"` | Access all getters after construction | All getter values remain identical to construction inputs |
