# spectra init

## Overview

The `spectra init` command initializes a new Spectra project in the current directory. It orchestrates four phases in sequence: (0) ensure `.gitignore` contains `.spectra`, (1) create `.spectra/` and `spec/` directory structures, (2) copy built-in files to their target locations. Each phase is delegated to a dedicated module. The `init` command itself is a pure orchestrator that calls phases in order and propagates errors.

## Boundaries

- Owns: phase sequencing (gitignore → directories → files).
- Owns: success message output on completion.
- Owns: error propagation from any phase (fail-fast).
- Owns: determining project root as CWD.
- Delegates: `.gitignore` handling to GitignoreEnsurer.
- Delegates: directory creation to DirectoryCreator.
- Delegates: built-in file copying to BuiltinResourceCopier.
- Must not: use SpectraFinder (always initializes in CWD).
- Must not: validate YAML/Markdown content of built-in files.
- Must not: perform file I/O directly (delegated to phase modules).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `GitignoreEnsurer` | Phase 0 | `Ensure(projectRoot)` | — |
| `DirectoryCreator` | Phase 1 | `CreateAll(projectRoot)` | — |
| `BuiltinResourceCopier` | Phase 2 | `CopyWorkflows(projectRoot)`, `CopyAgents(projectRoot)`, `CopySpecFiles(projectRoot)` | — |
| `cmdutil.ErrorFormatter` | Output formatting | `FormatError(msg)`, `FormatWarning(msg)` | — |

Construction constraint: Registered as a Cobra subcommand of the root command.

## Behavior

1. Determines project root as the current working directory (`os.Getwd()`).
2. If `os.Getwd()` fails, prints `"Error: failed to determine working directory: <error>"` to stderr and exits with code 1.
3. **Phase 0**: Calls `GitignoreEnsurer.Ensure(projectRoot)`. If error, prints error to stderr and exits with code 1.
4. **Phase 1**: Calls `DirectoryCreator.CreateAll(projectRoot)`. If error, prints error to stderr and exits with code 1.
5. **Phase 2a**: Calls `BuiltinResourceCopier.CopyWorkflows(projectRoot)`. Prints any returned warnings to stdout. If error, prints error to stderr and exits with code 1.
6. **Phase 2b**: Calls `BuiltinResourceCopier.CopyAgents(projectRoot)`. Prints any returned warnings to stdout. If error, prints error to stderr and exits with code 1.
7. **Phase 2c**: Calls `BuiltinResourceCopier.CopySpecFiles(projectRoot)`. Prints any returned warnings to stdout. If error, prints error to stderr and exits with code 1.
8. If all phases succeed, prints `"Spectra project initialized successfully"` to stdout and exits with code 0.

## Inputs

No command-line arguments or flags required (beyond `--help`).

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Current Working Directory | string | Process environment (`os.Getwd()`) | Yes (implicit) |

## Outputs

### stdout

- Warning messages from BuiltinResourceCopier (skipped files).
- Success message: `"Spectra project initialized successfully"`.

### stderr

- Error messages for failed operations.

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | All phases completed |
| 1 | Error | Any phase failed |

## Invariants

1. **Phase Ordering**: Phases execute in order: gitignore → directories → files. Each must complete before the next begins.
2. **Fail-Fast**: If any phase returns an error, the command exits immediately. Subsequent phases are not executed.
3. **Partial State on Failure**: Does not rollback. Previously created directories/files remain on disk.
4. **No SpectraFinder**: Always initializes in CWD regardless of parent `.spectra/` directories.
5. **Idempotent**: Running `spectra init` multiple times is safe. Existing directories are skipped, existing files produce warnings.
6. **No Validation**: Does not validate content of built-in files.

## Edge Cases

- Condition: All directories and files already exist (re-initialization).
  Expected: Warnings printed for each skipped file. Exits with code 0.

- Condition: `os.Getwd()` fails.
  Expected: Prints error, exits with code 1.

- Condition: Phase 0 fails (gitignore permission denied).
  Expected: Prints error, exits with code 1. No directories or files created.

- Condition: Phase 1 fails (directory creation permission denied).
  Expected: Prints error, exits with code 1. Gitignore already modified.

- Condition: Phase 2 partially fails (first workflow copied, second fails).
  Expected: Prints error, exits with code 1. First workflow remains on disk.

## Related

- [root](./root.md) - Parent command
- [GitignoreEnsurer](./gitignore_ensurer.md) - Phase 0 implementation
- [DirectoryCreator](./directory_creator.md) - Phase 1 implementation
- [BuiltinResourceCopier](./builtin_resource_copier.md) - Phase 2 implementation
