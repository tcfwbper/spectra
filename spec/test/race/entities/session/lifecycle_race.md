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
| `TestRunAndFail_Concurrent` | `race` | Concurrent Run and Fail during initialization are serialized, and `Fail` remains authoritative because it is valid from both `initializing` and `running`. | Construct session; create buffered channel `ch` with capacity 2 | Launch `Run()` and `Fail(agentErr, ch)` in separate goroutines | `Fail()` succeeds, the final status is `failed`, the session error is recorded, exactly one termination notification is sent, and `Run()` either succeeds first or returns the precondition error for status `failed`; no data race (run with `-race`) |
