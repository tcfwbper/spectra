# Test Specification: `confirm_prompt_test.go`

## Source File Under Test
`internal/cmdutil/confirm_prompt.go`

## Test File
`internal/cmdutil/confirm_prompt_test.go`

---

## `ConfirmPrompt`

### Happy Path — ConfirmPrompt

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestConfirmPrompt_LowercaseY` | `unit` | Returns true when user enters "y". | Create `strings.NewReader("y\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(true, nil)`; writer contains `"Confirm? [y/N]: "` |
| `TestConfirmPrompt_UppercaseY` | `unit` | Returns true when user enters "Y". | Create `strings.NewReader("Y\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(true, nil)` |
| `TestConfirmPrompt_YWithWhitespace` | `unit` | Returns true when input is "y" surrounded by whitespace. | Create `strings.NewReader("  y  \n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(true, nil)` |
| `TestConfirmPrompt_PromptWritten` | `unit` | Writes the prompt string to the writer before reading input. | Create `strings.NewReader("n\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Delete all? [y/N]: "` | Writer buffer contains exactly `"Delete all? [y/N]: "` |

### Happy Path — Rejection

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestConfirmPrompt_EmptyInput` | `unit` | Returns false when user presses Enter without typing. | Create `strings.NewReader("\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |
| `TestConfirmPrompt_LowercaseN` | `unit` | Returns false when user enters "n". | Create `strings.NewReader("n\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |
| `TestConfirmPrompt_UppercaseN` | `unit` | Returns false when user enters "N". | Create `strings.NewReader("N\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |
| `TestConfirmPrompt_Yes` | `unit` | Returns false when user enters "yes" (only single char accepted). | Create `strings.NewReader("yes\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |
| `TestConfirmPrompt_No` | `unit` | Returns false when user enters "no". | Create `strings.NewReader("no\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |
| `TestConfirmPrompt_ArbitraryText` | `unit` | Returns false for any text that is not "y" or "Y". | Create `strings.NewReader("maybe\n")` as reader; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestConfirmPrompt_EOF` | `unit` | Returns false when reader is at EOF (no data). | Create `strings.NewReader("")` as reader (immediate EOF); `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |
| `TestConfirmPrompt_ReaderClosed` | `unit` | Returns false when reader returns read error. | Create a reader that returns an I/O error on Read; `bytes.Buffer` as writer. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, nil)` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestConfirmPrompt_WriterError` | `unit` | Returns error when writing prompt to writer fails. | Create `strings.NewReader("y\n")` as reader; create a writer that returns an error on Write. | `reader`, `writer`, `prompt="Confirm? [y/N]: "` | Returns `(false, <non-nil error>)` |
