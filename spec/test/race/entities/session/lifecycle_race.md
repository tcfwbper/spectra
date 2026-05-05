# Test Specification: `lifecycle_race_test.go`

## Source File Under Test
`entities/session/lifecycle.go`

## Test File
`test/race/entities/session/lifecycle_race_test.go`

---

## `Run` / `Fail`

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunAndFail_Concurrent` | `race` | Concurrent Run and Fail during initialization are serialized. | Construct session; create buffered channel `ch` with capacity 2 | Launch `Run()` and `Fail(agentErr, ch)` in separate goroutines | Exactly one succeeds; the other returns a precondition error; no data race (run with `-race`) |
