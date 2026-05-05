# Test Specification: `event_history_race_test.go`

## Source File Under Test
`entities/session/event_history.go`

## Test File
`test/race/entities/session/event_history_race_test.go`

---

## `UpdateEventHistorySafe`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateEventHistorySafe_ConcurrentAppends` | `race` | Concurrent appends are serialized without data race. | Construct session; create N valid events with distinct IDs | Launch N goroutines each appending one event | No data race (run with `-race`); EventHistory length is N; all events present |
