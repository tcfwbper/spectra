# Test Specification: `transition_test.go`

## Source File Under Test
`components/transition.go`

## Test File
`components/transition_test.go`

---

## `Transition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewTransition_ValidInputs` | `unit` | Constructs a Transition with all valid and distinct fields. | | `fromNode="Architect"`, `eventType="DraftCompleted"`, `toNode="Reviewer"` | Returns no error; all getters return the provided values |

### Validation Failures — FromNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewTransition_EmptyFromNode` | `unit` | Rejects empty string FromNode. | | `fromNode=""` with other fields valid | Returns error `"from_node cannot be empty"` |
| `TestNewTransition_FromNodeStartsLowercase` | `unit` | Rejects FromNode starting with a lowercase letter. | | `fromNode="architect"` with other fields valid | Returns error indicating PascalCase is required |
| `TestNewTransition_FromNodeContainsHyphen` | `unit` | Rejects FromNode with non-alphanumeric characters. | | `fromNode="Archi-Tect"` with other fields valid | Returns error indicating PascalCase is required |

### Validation Failures — EventType

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewTransition_EmptyEventType` | `unit` | Rejects empty string EventType. | | `eventType=""` with other fields valid | Returns error `"event_type cannot be empty"` |
| `TestNewTransition_EventTypeStartsLowercase` | `unit` | Rejects EventType starting with a lowercase letter. | | `eventType="draftCompleted"` with other fields valid | Returns error indicating PascalCase is required |
| `TestNewTransition_EventTypeContainsUnderscore` | `unit` | Rejects EventType with non-alphanumeric characters. | | `eventType="Draft_Completed"` with other fields valid | Returns error indicating PascalCase is required |

### Validation Failures — ToNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewTransition_EmptyToNode` | `unit` | Rejects empty string ToNode. | | `toNode=""` with other fields valid | Returns error `"to_node cannot be empty"` |
| `TestNewTransition_ToNodeStartsLowercase` | `unit` | Rejects ToNode starting with a lowercase letter. | | `toNode="reviewer"` with other fields valid | Returns error indicating PascalCase is required |

### Validation Failures — Self-Loop

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewTransition_SelfLoop` | `unit` | Rejects Transition where FromNode equals ToNode. | | `fromNode="Architect"`, `eventType="DraftCompleted"`, `toNode="Architect"` | Returns error `"from_node and to_node must be different"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_Immutability` | `unit` | All fields remain unchanged after construction; no exported setters exist. | Construct a valid Transition with `fromNode="Architect"`, `eventType="DraftCompleted"`, `toNode="Reviewer"` | Access all getters after construction | All getter values remain identical to construction inputs |
