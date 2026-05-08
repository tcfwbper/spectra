# Test Specification: `constants_test.go`

## Source File Under Test
`storage/constants.go`

## Test File
`storage/constants_test.go`

---

## `MaxPayloadSize`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMaxPayloadSize_Value` | `unit` | MaxPayloadSize equals 10 MB (10 * 1024 * 1024). | | | `MaxPayloadSize == 10485760` |
| `TestMaxPayloadSize_IsConst` | `unit` | MaxPayloadSize is usable as a compile-time constant (assignable to a const-compatible expression). | | Assign `MaxPayloadSize` to a variable and use in a const expression context | Compiles and equals `10485760` |
