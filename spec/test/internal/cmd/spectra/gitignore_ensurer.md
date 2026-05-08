# Test Specification: `gitignore_ensurer_test.go`

## Source File Under Test

`internal/cmd/spectra/gitignore_ensurer.go`

## Test File

`internal/cmd/spectra/gitignore_ensurer_test.go`

---

## `GitignoreEnsurer`

### Happy Path — Ensure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGitignoreEnsurer_Ensure_CreatesNewFile` | `unit` | Creates `.gitignore` with `.spectra` entry when file does not exist. | Create a temporary directory as `projectRoot` using `t.TempDir()`. No `.gitignore` file present. | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` exists with content `.spectra\n` and permissions `0644` |
| `TestGitignoreEnsurer_Ensure_AppendsWhenMissing_EndsWithNewline` | `unit` | Appends `.spectra` entry when `.gitignore` exists without it and ends with a newline. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"node_modules\n"`. | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` content is `"node_modules\n.spectra\n"` |
| `TestGitignoreEnsurer_Ensure_AppendsWhenMissing_NoTrailingNewline` | `unit` | Appends newline then `.spectra` entry when file does not end with newline. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"node_modules"` (no trailing newline). | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` content is `"node_modules\n.spectra\n"` |
| `TestGitignoreEnsurer_Ensure_AlreadyPresent` | `unit` | Returns nil without modification when `.spectra` already exists in file. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"node_modules\n.spectra\n"`. | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` content unchanged |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGitignoreEnsurer_Ensure_Idempotent` | `unit` | Calling Ensure twice does not duplicate the entry. | Create a temporary directory as `projectRoot` using `t.TempDir()`. No `.gitignore` file present. Call `Ensure` once. | `projectRoot` = temp dir path (second call) | Returns `nil`; `.gitignore` contains exactly one `.spectra` entry |

### Boundary Values — Line Matching

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGitignoreEnsurer_Ensure_MatchesWithLeadingTrailingSpaces` | `unit` | Matches `.spectra` when line has leading/trailing spaces. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"  .spectra  \n"`. | `projectRoot` = temp dir path | Returns `nil`; file content unchanged |
| `TestGitignoreEnsurer_Ensure_MatchesWithTabs` | `unit` | Matches `.spectra` when line has leading/trailing tabs. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"\t.spectra\t\n"`. | `projectRoot` = temp dir path | Returns `nil`; file content unchanged |
| `TestGitignoreEnsurer_Ensure_DoesNotMatchSpectraSlash` | `unit` | Does not match `.spectra/` as it is not an exact match. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `".spectra/\n"`. | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` content is `".spectra/\n.spectra\n"` |
| `TestGitignoreEnsurer_Ensure_DoesNotMatchCommented` | `unit` | Does not match `# .spectra` as it is a comment. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"# .spectra\n"`. | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` content is `"# .spectra\n.spectra\n"` |
| `TestGitignoreEnsurer_Ensure_DoesNotTrimUnicodeWhitespace` | `unit` | Does not trim non-breaking space (Unicode whitespace) when matching. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `" .spectra\n"` (NBSP prefix). | `projectRoot` = temp dir path | Returns `nil`; `.spectra\n` appended (NBSP-prefixed line does not match) |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGitignoreEnsurer_Ensure_EmptyFile` | `unit` | Appends `.spectra` entry to an empty `.gitignore` file. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create empty `.gitignore` file (0 bytes). | `projectRoot` = temp dir path | Returns `nil`; `.gitignore` content is `.spectra\n` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGitignoreEnsurer_Ensure_ReadPermissionDenied` | `unit` | Returns error when `.gitignore` cannot be read. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with permissions `0000`. Cleanup: restore permissions in `t.Cleanup`. | `projectRoot` = temp dir path | Returns error containing `"failed to read '.gitignore'"` |
| `TestGitignoreEnsurer_Ensure_WritePermissionDenied` | `unit` | Returns error when `.gitignore` exists but is read-only and append is needed. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `.gitignore` with content `"node_modules\n"` and permissions `0444`. Cleanup: restore permissions in `t.Cleanup`. | `projectRoot` = temp dir path | Returns error containing `"failed to update '.gitignore'"` |
| `TestGitignoreEnsurer_Ensure_BrokenSymlink` | `unit` | Returns error when `.gitignore` is a broken symlink. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create symlink `.gitignore` pointing to a non-existent target. | `projectRoot` = temp dir path | Returns error containing `"failed to read '.gitignore'"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGitignoreEnsurer_Ensure_FollowsSymlink` | `unit` | Operates on the symlink target file when `.gitignore` is a valid symlink. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create a real file `real_gitignore` with content `"node_modules\n"`. Create symlink `.gitignore` pointing to `real_gitignore`. | `projectRoot` = temp dir path | Returns `nil`; symlink target `real_gitignore` content is `"node_modules\n.spectra\n"` |
