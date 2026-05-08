# ConfirmPrompt

## Overview

ConfirmPrompt is a shared utility function that prompts the user for confirmation via stdin and returns whether the user confirmed. It prints a prompt message, reads a single line of input, and returns true only if the user enters exactly `y` or `Y`. All other input (including empty input, `n`, `yes`, `no`, or EOF) is treated as rejection.

## Boundaries

- Owns: printing the prompt to the provided writer, reading one line from the provided reader, interpreting the response.
- Must not: decide what action to take based on the response (caller's responsibility).
- Must not: retry or re-prompt on invalid input.
- Must not: perform any destructive operations.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `io.Reader` | Input source | Read one line | Must not close or seek |
| `io.Writer` | Output destination | Write prompt string | Must not close |

Construction constraint: Package-level function. Accepts `reader io.Reader` and `writer io.Writer` parameters for testability. In production, caller passes `os.Stdin` and `os.Stdout`.

## Behavior

1. `ConfirmPrompt(reader io.Reader, writer io.Writer, prompt string) (bool, error)`.
2. Writes `prompt` to `writer` (no trailing newline added — the prompt string should include formatting like `[y/N]: `).
3. Reads one line from `reader` using line-oriented reading (e.g., `bufio.Scanner`).
4. If reading fails (I/O error or EOF), returns `(false, nil)`. Treated as rejection, not an error.
5. Trims leading and trailing whitespace from the input.
6. If the trimmed input equals `y` or `Y`, returns `(true, nil)`.
7. For all other input (including empty, `n`, `N`, `yes`, `no`, any other text), returns `(false, nil)`.
8. If writing the prompt to `writer` fails, returns `(false, error)`.

## Inputs

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| reader | io.Reader | Non-nil, connected to user input source | Yes |
| writer | io.Writer | Non-nil, connected to user output destination | Yes |
| prompt | string | Non-empty, should end with a space or colon for readability | Yes |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| confirmed | bool | true if user entered `y` or `Y`, false otherwise |
| error | error | Non-nil only if writing prompt fails. Read failures return false without error. |

## Invariants

1. **Single Read**: Reads exactly one line from reader per invocation. Does not loop or retry.
2. **Strict Match**: Only `y` or `Y` (after trimming whitespace) is treated as confirmation. No other variations (yes, YES, etc.).
3. **EOF as Rejection**: If reader returns EOF (e.g., piped input with no data), treated as rejection (false), not an error.
4. **No Side Effects**: Does not perform any action beyond printing the prompt and reading input.
5. **Testable I/O**: Uses injected reader/writer for testability. Does not reference os.Stdin or os.Stdout directly.

## Edge Cases

- Condition: User enters `y` followed by Enter.
  Expected: Returns (true, nil).

- Condition: User enters `Y` followed by Enter.
  Expected: Returns (true, nil).

- Condition: User enters `  y  ` (with whitespace).
  Expected: Returns (true, nil) after trimming.

- Condition: User presses Enter without typing anything.
  Expected: Returns (false, nil). Empty input is rejection.

- Condition: User enters `yes`.
  Expected: Returns (false, nil). Only `y` or `Y` is accepted.

- Condition: User enters `n` or `N`.
  Expected: Returns (false, nil).

- Condition: stdin is EOF (piped empty input, e.g., `echo "" | spectra clear`).
  Expected: Returns (false, nil). Treated as rejection.

- Condition: stdin is closed before reading.
  Expected: Returns (false, nil). Read failure treated as rejection.

- Condition: Writing prompt to writer fails (e.g., broken pipe).
  Expected: Returns (false, error).

## Related

- [spectra clear](../cmd/spectra/clear.md) - Primary consumer of ConfirmPrompt
- [ErrorFormatter](./error_formatter.md) - Sibling cmdutil utility
