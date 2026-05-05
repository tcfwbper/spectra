# Test Specification: `event_history_test.go`

## Source File Under Test
`entities/session/event_history.go`

## Test File
`entities/session/event_history_test.go`

---

## `UpdateEventHistorySafe`

### Happy Path — UpdateEventHistorySafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_ValidEvent` | `unit` | Appends a valid event to EventHistory. | Construct session; create Event with all required fields populated (`ID`, `Type`, `SessionID`, `EmittedAt > 0`, `EmittedBy` non-empty) | Call `UpdateEventHistorySafe(event)` | Returns `nil`; EventHistory length is 1; the appended event matches the input |
| `TestUpdateEventHistorySafe_MultipleEvents` | `unit` | Appends multiple events preserving order. | Construct session; create two valid events `e1`, `e2` | Call `UpdateEventHistorySafe(e1)` then `UpdateEventHistorySafe(e2)` | Returns `nil` for both; EventHistory length is 2; `EventHistory[0]` is `e1`, `EventHistory[1]` is `e2` |
| `TestUpdateEventHistorySafe_EmptyMessage` | `unit` | Accepts event with empty Message field. | Construct session; create valid event with `Message=""` | Call `UpdateEventHistorySafe(event)` | Returns `nil`; event is appended |
| `TestUpdateEventHistorySafe_EmptyPayload` | `unit` | Accepts event with empty/nil Payload. | Construct session; create valid event with empty Payload | Call `UpdateEventHistorySafe(event)` | Returns `nil`; event is appended |
| `TestUpdateEventHistorySafe_UpdatesUpdatedAt` | `unit` | Successful append advances UpdatedAt. | Construct session; record initial `UpdatedAt`; create valid event | Call `UpdateEventHistorySafe(event)` | Returns `nil`; `GetMetadataSnapshotSafe().UpdatedAt >= initial UpdatedAt` |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_MissingID` | `unit` | Rejects event with empty ID. | Construct session; create event with `ID=""`, other fields valid | Call `UpdateEventHistorySafe(event)` | Returns error with message `"invalid event: ID is required"` |
| `TestUpdateEventHistorySafe_MissingType` | `unit` | Rejects event with empty Type. | Construct session; create event with `Type=""`, other fields valid | Call `UpdateEventHistorySafe(event)` | Returns error with message `"invalid event: Type is required"` |
| `TestUpdateEventHistorySafe_MissingSessionID` | `unit` | Rejects event with empty SessionID. | Construct session; create event with `SessionID=""`, other fields valid | Call `UpdateEventHistorySafe(event)` | Returns error with message `"invalid event: SessionID is required"` |
| `TestUpdateEventHistorySafe_InvalidEmittedAt` | `unit` | Rejects event with EmittedAt <= 0. | Construct session; create event with `EmittedAt=0`, other fields valid | Call `UpdateEventHistorySafe(event)` | Returns error with message `"invalid event: EmittedAt is required"` |
| `TestUpdateEventHistorySafe_MissingEmittedBy` | `unit` | Rejects event with empty EmittedBy. | Construct session; create event with `EmittedBy=""`, other fields valid | Call `UpdateEventHistorySafe(event)` | Returns error with message `"invalid event: EmittedBy is required"` |
| `TestUpdateEventHistorySafe_ValidationOrder` | `unit` | Returns error for the first failing field in validation order. | Construct session; create event with `ID=""` and `Type=""` | Call `UpdateEventHistorySafe(event)` | Returns error with message `"invalid event: ID is required"` (ID checked first) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_DuplicateEventID` | `unit` | Appending same event ID twice stores both copies (no deduplication). | Construct session; create valid event `e` | Call `UpdateEventHistorySafe(e)` twice | Returns `nil` for both; EventHistory length is 2 |

