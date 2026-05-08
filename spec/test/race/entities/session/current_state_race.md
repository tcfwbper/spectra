# Test Specification: `current_state_race_test.go`

## Source File Under Test
`entities/session/current_state.go`

## Test File
`test/race/entities/session/current_state_race_test.go`

---

## `UpdateCurrentStateSafe`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestUpdateCurrentStateSafe_ConcurrentUpdates` | `race` | Concurrent updates are serialized without data race. | Construct session | Launch multiple goroutines each calling `UpdateCurrentStateSafe` with distinct non-empty values | No data race (run with `-race`); `GetCurrentStateSafe()` returns one of the written values |
