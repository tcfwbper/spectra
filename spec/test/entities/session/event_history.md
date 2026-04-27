# Test Specification: `event_history.go`

## Source File Under Test
`entities/session/event_history.go`

## Test File
`entities/session/event_history_test.go`

---

**Fixture Isolation**: All tests create Session instances in memory using programmatic construction. No external files or directories are required unless explicitly stated in the Setup column. Mock dependencies (EventStore, loggers, etc.) are created within each test.

---

## `Session` Event History Method

### Happy Path — UpdateEventHistorySafe

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_AppendsEvent` | `unit` | Appends valid event to EventHistory. | Session with empty `EventHistory`, `UpdatedAt=T0` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Started", SessionID: "<session-uuid>", EmittedAt: 1000, EmittedBy: "node1", Message: "begin", Payload: {}})` | Returns `nil`; `EventHistory` length is 1; event stored; `UpdatedAt > T0` |
| `TestUpdateEventHistorySafe_AppendsToExisting` | `unit` | Appends event to existing EventHistory. | Session with `EventHistory` containing 2 events, `UpdatedAt=T0` | Call `UpdateEventHistorySafe(Event{ID: "evt-3", Type: "Progress", SessionID: "<session-uuid>", EmittedAt: 2000, EmittedBy: "node2", Message: "", Payload: {}})` | Returns `nil`; `EventHistory` length is 3; new event appended at end; `UpdatedAt > T0` |
| `TestUpdateEventHistorySafe_EmptyMessage` | `unit` | Accepts event with empty Message. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node", Message: "", Payload: {}})` | Returns `nil`; event appended; `Message=""` |
| `TestUpdateEventHistorySafe_NilPayload` | `unit` | Accepts event with nil Payload. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node", Message: "msg", Payload: nil})` | Returns `nil`; event appended; `Payload=nil` |
| `TestUpdateEventHistorySafe_EmptyMessageAndNilPayload` | `unit` | Accepts event with both empty Message and nil Payload. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node", Message: "", Payload: nil})` | Returns `nil`; event appended |
| `TestUpdateEventHistorySafe_PersistsToEventStore` | `unit` | Persists event to EventStore. | Mock EventStore; session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node", Message: "msg", Payload: {}})` | Returns `nil`; EventStore write called with event |

### Validation Failures — Required Fields

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_EmptyID` | `unit` | Rejects event with empty ID. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node"})` | Returns error matching `/invalid event.*ID is required/i`; `EventHistory` unchanged; lock not acquired |
| `TestUpdateEventHistorySafe_EmptyType` | `unit` | Rejects event with empty Type. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node"})` | Returns error matching `/invalid event.*Type is required/i`; `EventHistory` unchanged; lock not acquired |
| `TestUpdateEventHistorySafe_EmptySessionID` | `unit` | Rejects event with empty SessionID. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "", EmittedAt: 1000, EmittedBy: "node"})` | Returns error matching `/invalid event.*SessionID is required/i`; `EventHistory` unchanged; lock not acquired |
| `TestUpdateEventHistorySafe_ZeroEmittedAt` | `unit` | Rejects event with EmittedAt = 0. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 0, EmittedBy: "node"})` | Returns error matching `/invalid event.*EmittedAt is required/i`; `EventHistory` unchanged; lock not acquired |
| `TestUpdateEventHistorySafe_NegativeEmittedAt` | `unit` | Rejects event with negative EmittedAt. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: -100, EmittedBy: "node"})` | Returns error matching `/invalid event.*EmittedAt is required/i`; `EventHistory` unchanged; lock not acquired |
| `TestUpdateEventHistorySafe_EmptyEmittedBy` | `unit` | Rejects event with empty EmittedBy. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: ""})` | Returns error matching `/invalid event.*EmittedBy is required/i`; `EventHistory` unchanged; lock not acquired |

### Validation Failures — Multiple Fields

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_MultipleFieldsMissing` | `unit` | Returns error for first failing field when multiple fields invalid. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "", Type: "", SessionID: "", EmittedAt: 0, EmittedBy: ""})` | Returns error matching `/invalid event.*is required/i` (identifies first failing field); `EventHistory` unchanged |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_NotIdempotent` | `unit` | Same event appended multiple times (no deduplication). | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(event)` twice with same event | Both return `nil`; `EventHistory` length is 2; both copies stored |
| `TestUpdateEventHistorySafe_SameIDAppendedTwice` | `unit` | Events with same ID are both appended (no ID uniqueness check). | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", ...})` twice | Both return `nil`; `EventHistory` contains 2 events with same ID |

### Ordering — Chronological

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_MaintainsChronologicalOrder` | `unit` | Events appended in order; chronological order matches lock-acquisition order. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe()` with events having `EmittedAt=1000`, then `EmittedAt=2000`, then `EmittedAt=1500` | All succeed; `EventHistory` order is [1000, 2000, 1500] (append order, not sorted by timestamp) |
| `TestEventHistory_AppendOrder` | `unit` | EventHistory order matches append order, not EmittedAt sorting. | Session with empty `EventHistory` | Append events with `EmittedAt` in non-ascending order | Events stored in append order; no automatic sorting |

### Atomic Replacement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_AtomicAppend` | `unit` | EventHistory and UpdatedAt updated atomically. | Session with `EventHistory=[evt1]`, `UpdatedAt=T0`; concurrent goroutine reading EventHistory length | Call `UpdateEventHistorySafe(evt2)` | Concurrent reader never observes intermediate state; both EventHistory and UpdatedAt updated together |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventHistory_ConcurrentAppends` | `race` | Concurrent appends are serialized by write lock. | Session with empty `EventHistory` | 10 goroutines call `UpdateEventHistorySafe()` with different events simultaneously | All succeed; `EventHistory` length is 10; events ordered by lock acquisition; no race conditions |
| `TestEventHistory_ConcurrentReadWrite` | `race` | Concurrent reads of EventHistory during writes. | Session with `EventHistory=[evt1]` | 50 goroutines read `EventHistory` via session reference while 1 goroutine appends | No race conditions; readers see consistent state |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_ReleasesLock` | `unit` | Write lock released after append. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(event)`; then read EventHistory via getter | Both succeed without deadlock |
| `TestUpdateEventHistorySafe_ValidationBeforeLock` | `unit` | Validation occurs before lock acquisition. | Session with empty `EventHistory`; concurrent goroutine holding write lock | Call `UpdateEventHistorySafe(Event{ID: ""})` (empty ID) | Returns error immediately without blocking on lock |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_EventStorePersistenceFailureLogged` | `unit` | EventStore persistence failures logged but not returned. | Mock EventStore that returns error on write; session with empty `EventHistory`; mock logger | Call `UpdateEventHistorySafe(valid event)` | Returns `nil`; in-memory `EventHistory` updated; warning logged matching `/EventStore.*persistence failed/i` or error message |
| `TestUpdateEventHistorySafe_PersistenceFailureDoesNotRevert` | `unit` | In-memory EventHistory authoritative even when persistence fails. | Mock EventStore that fails; session with empty `EventHistory` | Call `UpdateEventHistorySafe(event)` | Returns `nil`; event appended to in-memory `EventHistory` |

### Invariants — Append-Only

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventHistory_AppendOnly` | `unit` | Events never removed or reordered after append. | Session with `EventHistory=[evt1, evt2]` | Append `evt3` | `EventHistory=[evt1, evt2, evt3]`; existing events unchanged and in same order |
| `TestEventHistory_NoMutation` | `unit` | Previously appended events are not modified. | Session with `EventHistory=[evt1{Message: "original"}]` | Append `evt2` | `evt1.Message` remains "original"; no mutations to existing events |

### Invariants — UpdatedAt Refresh

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_RefreshesUpdatedAt` | `unit` | UpdateEventHistorySafe refreshes UpdatedAt. | Session with empty `EventHistory`, `UpdatedAt=T0` | Wait 1 second; call `UpdateEventHistorySafe(event)` | Returns `nil`; `UpdatedAt > T0` |
| `TestUpdateEventHistorySafe_UpdatedAtInSameCriticalSection` | `unit` | UpdatedAt refreshed in same critical section as append. | Session with empty `EventHistory`, `UpdatedAt=T0`; concurrent goroutine reading EventHistory length | Call `UpdateEventHistorySafe(event)` | Concurrent reader observes consistent snapshot; never sees new event with old `UpdatedAt` |

### Invariants — Validation Strictness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_StrictRequiredFieldValidation` | `unit` | All required fields validated; invalid events never enter memory. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe()` with event missing each required field separately | Each returns error; `EventHistory` remains empty in all cases |
| `TestUpdateEventHistorySafe_MessageNotValidated` | `unit` | Message field is not validated (empty accepted). | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node", Message: ""})` | Returns `nil`; event appended |
| `TestUpdateEventHistorySafe_PayloadNotValidated` | `unit` | Payload field is not validated (nil accepted). | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{ID: "evt-1", Type: "Event", SessionID: "<uuid>", EmittedAt: 1000, EmittedBy: "node", Payload: nil})` | Returns `nil`; event appended |

### Invariants — Memory Authority

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventHistory_InMemoryAuthoritative` | `unit` | In-memory EventHistory is source of truth for transition evaluation. | Mock EventStore that fails; session with empty `EventHistory` | Call `UpdateEventHistorySafe(event)`; read EventHistory | Event present in in-memory EventHistory |

### Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_EmittedAtInFuture` | `unit` | Accepts event with EmittedAt timestamp in the future (not validated). | Session with empty `EventHistory`; current time T0 | Call `UpdateEventHistorySafe(Event{EmittedAt: T0 + 1000000})` (far future) | Returns `nil`; event appended |
| `TestUpdateEventHistorySafe_EmittedAtBeforeCreatedAt` | `unit` | Accepts event with EmittedAt before Session.CreatedAt (not validated). | Session with `CreatedAt=2000`, empty `EventHistory` | Call `UpdateEventHistorySafe(Event{EmittedAt: 1000})` (before session creation) | Returns `nil`; event appended (time consistency not enforced at this layer) |
| `TestUpdateEventHistorySafe_LargePayload` | `unit` | Accepts event with large Payload. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{..., Payload: <10MB JSON object>})` | Returns `nil`; event appended |
| `TestUpdateEventHistorySafe_UnicodeInMessage` | `unit` | Accepts event with Unicode characters in Message. | Session with empty `EventHistory` | Call `UpdateEventHistorySafe(Event{..., Message: "通知: 完成 🎉"})` | Returns `nil`; event appended; Unicode preserved |
| `TestUpdateEventHistorySafe_SessionIDMismatch` | `unit` | Accepts event even if SessionID does not match Session.ID (caller responsibility). | Session with `ID="session-A"`; empty `EventHistory` | Call `UpdateEventHistorySafe(Event{SessionID: "session-B", ...})` (mismatched) | Returns `nil`; event appended (SessionID validated for non-empty, but not checked against Session.ID) |
