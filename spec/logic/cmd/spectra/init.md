# spectra init Command

## Overview

The `spectra init` command initializes a new Spectra project in the current directory by creating the `.spectra/` directory structure and the `spec/` directory structure, then copying built-in workflow, agent definition, and specification template files from embedded resources to their corresponding locations. It performs safe copying: directories and files are only created if they do not already exist. If a directory already exists, the command skips creation silently. If a file already exists, the command skips creation and prints a warning.

## Behavior

### Initialization Flow

1. The `init` command is invoked as `spectra init`.
2. The command determines the project root as the current working directory.

#### Phase 0: Ensure .gitignore contains .spectra
3. The command checks if `.gitignore` exists in the current working directory. If `.gitignore` is a symbolic link, the command follows the symlink and operates on the target file.
4. If `.gitignore` does not exist:
   - The command creates `.gitignore` with permissions `0644` (rw-r--r--) containing exactly one line: `.spectra` (followed by a newline).
5. If `.gitignore` exists:
   - The command reads the file content using line-oriented reading (e.g., `bufio.Scanner`).
   - The command checks if any line contains exactly `.spectra` after trimming leading/trailing spaces (` `) and tabs (`\t`). Other Unicode whitespace characters are not trimmed.
   - If no such line is found, the command appends a new line: `.spectra` (preceded by a newline if the file does not end with a newline).
   - If `.spectra` is already present, the command skips modification (no output printed).
6. The command does NOT print any success message or warning when `.gitignore` is created or modified. Users can verify changes using `git status` or `git diff`.
7. If reading `.gitignore` fails (e.g., permission denied), the command prints an error: `"Error: failed to read '.gitignore': <error>"` and exits with code 1.
8. If writing or appending to `.gitignore` fails (e.g., permission denied, disk full), the command prints an error: `"Error: failed to update '.gitignore': <error>"` and exits with code 1.

#### Phase 1: Create .spectra/ directories
9. The command attempts to create the following `.spectra/` directory structure:
   - `.spectra/`
   - `.spectra/sessions/`
   - `.spectra/workflows/`
   - `.spectra/agents/`
10. For each directory in step 9:
   - If the directory does not exist, the command creates it with permissions `0755` (rwxr-xr-x).
   - If the directory already exists, the command skips creation (no warning printed).
11. If any directory creation fails (e.g., permission denied, disk full), the command prints an error: `"Error: failed to create directory '<path>': <error>"`, leaves the partial state as-is, and exits with code 1.

#### Phase 2: Create .spectra/ files
12. Built-in files are embedded in the binary using Go's `embed` directive and stored in a virtual filesystem accessible at compile time.
13. For each built-in workflow definition file (e.g., `SimpleSdd.yaml`):
   - The command composes the target path: `.spectra/workflows/<WorkflowName>.yaml`.
   - If the file does not exist at the target path, the command copies the embedded file content to the target path with permissions `0644` (rw-r--r--).
   - If the file already exists at the target path, the command prints a warning: `"Warning: workflow definition '<WorkflowName>.yaml' already exists, skipping"` and continues to the next file.
14. For each built-in agent definition file (e.g., `Architect.yaml`, `QaAnalyst.yaml`):
   - The command composes the target path: `.spectra/agents/<AgentRole>.yaml`.
   - If the file does not exist at the target path, the command copies the embedded file content to the target path with permissions `0644` (rw-r--r--).
   - If the file already exists at the target path, the command prints a warning: `"Warning: agent definition '<AgentRole>.yaml' already exists, skipping"` and continues to the next file.
15. If any file write operation fails (e.g., permission denied, disk full), the command prints an error: `"Error: failed to write built-in file '<path>': <error>"`, leaves the partial state as-is, and exits with code 1.

#### Phase 3: Create spec/ directories
16. The command attempts to create the following `spec/` directory structure:
    - `spec/`
    - `spec/logic/`
    - `spec/test/`
17. For each directory in step 16:
    - If the directory does not exist, the command creates it with permissions `0755` (rwxr-xr-x).
    - If the directory already exists, the command skips creation (no warning printed).
18. If any directory creation fails (e.g., permission denied, disk full), the command prints an error: `"Error: failed to create directory '<path>': <error>"`, leaves the partial state as-is (including already-created `.spectra/` directories and files), and exits with code 1.

#### Phase 4: Create spec/ files
19. For each built-in specification template file:
    - `spec/ARCHITECTURE.md`: Architecture documentation template
    - `spec/CONVENTIONS.md`: Code conventions template
    - `spec/logic/README.md`: Logic specification guide
    - `spec/test/README.md`: Test specification guide
20. For each file in step 19:
    - The command composes the target path relative to the current working directory.
    - If the file does not exist at the target path, the command copies the embedded file content to the target path with permissions `0644` (rw-r--r--).
    - If the file already exists at the target path, the command prints a warning: `"Warning: spec file '<filename>' already exists, skipping"` (where `<filename>` is the relative path from `spec/`, e.g., `ARCHITECTURE.md`, `logic/README.md`) and continues to the next file.
21. If any file write operation fails (e.g., permission denied, disk full), the command prints an error: `"Error: failed to write built-in file '<path>': <error>"`, leaves the partial state as-is, and exits with code 1.

#### Completion
22. If all operations succeed (directories created/skipped and files copied/skipped), the command prints a success message: `"Spectra project initialized successfully"` and exits with code 0.
23. The command does NOT validate the YAML or Markdown syntax or structure of built-in files. It only ensures that files are copied to the correct paths.
24. The command does NOT use SpectraFinder. It always initializes in the current working directory, regardless of whether a `.spectra/` directory already exists in a parent directory.

### Built-in Files

The `init` command embeds the following built-in files (exact filenames and count depend on the default workflow and agents defined in ARCHITECTURE.md):

**Built-in Workflows** (stored in `builtin/workflows/`):
- `SimpleSdd.yaml` (default SDD+TDD workflow)

**Built-in Agents** (stored in `builtin/agents/`):
- `Architect.yaml`
- `ArchitectReviewer.yaml`
- `QaAnalyst.yaml`
- `QaSpecReviewer.yaml`
- `QaEngineer.yaml`
- `QaReviewer.yaml`
- `SwEngineer.yaml`

**Built-in Specification Templates** (stored in `builtin/spec/`):
- `builtin/spec/ARCHITECTURE.md` → `spec/ARCHITECTURE.md`
- `builtin/spec/CONVENTIONS.md` → `spec/CONVENTIONS.md`
- `builtin/spec/logic/README.md` → `spec/logic/README.md`
- `builtin/spec/test/README.md` → `spec/test/README.md`

These files are embedded using Go's `embed` directive:

```go
//go:embed builtin/workflows/*.yaml
var builtinWorkflows embed.FS

//go:embed builtin/agents/*.yaml
var builtinAgents embed.FS

//go:embed builtin/spec/ARCHITECTURE.md
//go:embed builtin/spec/CONVENTIONS.md
//go:embed builtin/spec/logic/README.md
//go:embed builtin/spec/test/README.md
var builtinSpecFiles embed.FS
```

### Success Output

When initialization completes successfully (all directories created/existed and all files copied/skipped):

```
Spectra project initialized successfully
```

### Warning Output (printed to stdout)

When a file already exists and is skipped:

```
Warning: workflow definition 'SimpleSdd.yaml' already exists, skipping
Warning: agent definition 'Architect.yaml' already exists, skipping
Warning: spec file 'ARCHITECTURE.md' already exists, skipping
Warning: spec file 'logic/README.md' already exists, skipping
```

### Error Output (printed to stderr)

When .gitignore read fails:

```
Error: failed to read '.gitignore': permission denied
```

When .gitignore update fails:

```
Error: failed to update '.gitignore': permission denied
Error: failed to update '.gitignore': disk quota exceeded
```

When directory creation fails:

```
Error: failed to create directory '.spectra/sessions': permission denied
Error: failed to create directory 'spec/logic': permission denied
```

When file write fails:

```
Error: failed to write built-in file '.spectra/workflows/SimpleSdd.yaml': disk quota exceeded
Error: failed to write built-in file 'spec/ARCHITECTURE.md': disk quota exceeded
```

## Inputs

No command-line arguments or flags are required.

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Current Working Directory | string | Process environment | Yes (implicit) |

## Outputs

### stdout

- Success message: `"Spectra project initialized successfully"`
- Warning messages for skipped files (format: `"Warning: <type> definition '<filename>' already exists, skipping"`)

### stderr

- Error messages for failed operations (format: `"Error: failed to <operation> '<path>': <error>"`)

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | All directories created/existed and all files copied/skipped |
| 1 | Error | Directory creation failed, or file write failed |

### Filesystem Changes

On success or partial success:

- Files created or modified:
  - `.gitignore` (created if it did not exist, or modified to include `.spectra` if it was missing)
- Directories created (if they did not exist):
  - `.spectra/`
  - `.spectra/sessions/`
  - `.spectra/workflows/`
  - `.spectra/agents/`
  - `spec/`
  - `spec/logic/`
  - `spec/test/`
- Files created (if they did not exist):
  - `.spectra/workflows/<WorkflowName>.yaml` (for each built-in workflow)
  - `.spectra/agents/<AgentRole>.yaml` (for each built-in agent)
  - `spec/ARCHITECTURE.md`
  - `spec/CONVENTIONS.md`
  - `spec/logic/README.md`
  - `spec/test/README.md`

## Invariants

1. **Idempotent Directory Creation**: If a directory already exists, the command must not fail or print a warning. It simply skips creation.

2. **Safe File Copying**: The command must never overwrite existing files. If a file exists, it must print a warning and skip the file.

3. **No Validation**: The command must not validate the YAML or Markdown syntax or structure of built-in files. It only checks that files are successfully written to disk.

4. **Partial State on Failure**: If an operation fails (directory creation or file write), the command must leave the partial state as-is and exit immediately. It must not attempt to rollback.

5. **No SpectraFinder**: The command must not use SpectraFinder. It always initializes in the current working directory.

6. **Embedded Resources**: Built-in files must be embedded in the binary at compile time using Go's `embed` package. They are not loaded from external files.

7. **File Permissions**: Created directories must have permissions `0755` (rwxr-xr-x). Created files must have permissions `0644` (rw-r--r--).

8. **Phase Ordering**: The command must execute operations in five phases: (0) ensure `.gitignore` contains `.spectra`, (1) create `.spectra/` directories, (2) create `.spectra/` files, (3) create `spec/` directories, (4) create `spec/` files. Each phase must complete before the next phase begins. Within a phase, the order of operations does not matter. If one operation fails, the command exits immediately (fail-on-first-error).

9. **Gitignore Entry Format**: When checking or adding the `.spectra` entry to `.gitignore`, the command must match exactly `.spectra` after trimming leading/trailing spaces (` `) and tabs (`\t`). The command must not match `.spectra/` or other variations. Other Unicode whitespace characters (e.g., non-breaking space, zero-width space) are not trimmed and will cause a line to be considered non-matching.

10. **Cross-Platform Newline Handling**: The command must use Go's standard file I/O (e.g., `os.OpenFile`, `bufio.Scanner`) which automatically handles platform-specific newline conventions (`\n` on Unix/Linux/macOS, `\r\n` on Windows). When appending to `.gitignore`, the command writes `\n` and relies on Go's text mode conversion to handle platform differences. When reading lines, the command uses `bufio.Scanner` or similar utilities that strip line endings automatically.

11. **Symlink Following**: If `.gitignore` is a symbolic link, the command must follow the symlink and operate on the target file. If the symlink is broken (target does not exist), the operation fails with an error (e.g., `"Error: failed to read '.gitignore': no such file or directory"`).

12. **Current Directory as Project Root**: The command treats the current working directory as the project root, regardless of whether a parent directory contains `.spectra/`.

## Edge Cases

- **Condition**: `.gitignore` does not exist.
  **Expected**: The command creates `.gitignore` with content `.spectra\n` (where `\n` is a newline character) and proceeds to Phase 1.

- **Condition**: `.gitignore` exists and already contains a line with exactly `.spectra`.
  **Expected**: The command skips modification (no warning) and proceeds to Phase 1.

- **Condition**: `.gitignore` exists but does not contain `.spectra`. The file ends with a newline.
  **Expected**: The command appends `.spectra\n` to the file and proceeds to Phase 1.

- **Condition**: `.gitignore` exists but does not contain `.spectra`. The file does not end with a newline.
  **Expected**: The command appends `\n.spectra\n` to the file (ensuring the previous last line remains valid) and proceeds to Phase 1.

- **Condition**: `.gitignore` contains `.spectra/` (with trailing slash) but not `.spectra`.
  **Expected**: The command appends `.spectra\n` because `.spectra/` does not match exactly `.spectra`.

- **Condition**: `.gitignore` contains `# .spectra` (commented) but not `.spectra` as an uncommented line.
  **Expected**: The command appends `.spectra\n` because the commented line does not match exactly `.spectra`.

- **Condition**: `.gitignore` contains a line `  .spectra  ` (with leading/trailing spaces and tabs).
  **Expected**: The command considers this as matching `.spectra` (after trimming spaces and tabs) and skips modification.

- **Condition**: `.gitignore` contains a line with `.spectra` but also includes other Unicode whitespace (e.g., non-breaking space U+00A0).
  **Expected**: The command considers this as non-matching (because only spaces and tabs are trimmed) and appends `.spectra\n`.

- **Condition**: `.gitignore` is a symbolic link to another file (e.g., `.gitignore` → `../shared-gitignore`).
  **Expected**: The command follows the symlink and reads/modifies the target file (`../shared-gitignore`).

- **Condition**: `.gitignore` is a broken symbolic link (target file does not exist).
  **Expected**: The command prints `"Error: failed to read '.gitignore': no such file or directory"` and exits with code 1. No subsequent operations are performed.

- **Condition**: `.gitignore` exists but is read-only (permission `0444`).
  **Expected**: Reading succeeds. If modification is needed, appending fails with permission denied. The command prints `"Error: failed to update '.gitignore': permission denied"` and exits with code 1. No `.spectra/` directories or files are created.

- **Condition**: `.gitignore` exists and reading fails due to insufficient permissions.
  **Expected**: The command prints `"Error: failed to read '.gitignore': permission denied"` and exits with code 1. No subsequent operations are performed.

- **Condition**: Disk is full when creating `.gitignore` or appending to it.
  **Expected**: The command prints `"Error: failed to update '.gitignore': no space left on device"` and exits with code 1. No `.spectra/` directories or files are created.

- **Condition**: Running on Windows where line endings are `\r\n`, and `.gitignore` already contains `.spectra` with Windows line ending.
  **Expected**: The command correctly detects `.spectra` (because `bufio.Scanner` strips both `\n` and `\r\n` automatically) and skips modification.

- **Condition**: Running on Windows and appending `.spectra` to `.gitignore`.
  **Expected**: The command writes `\n` (LF only). Git on Windows accepts both LF and CRLF line endings in `.gitignore`, so this is compatible. The command does not enforce CRLF on Windows.

- **Condition**: `.spectra/` directory already exists.
  **Expected**: The command skips directory creation (no warning) and proceeds to copy built-in files.

- **Condition**: `.spectra/workflows/` directory already exists but `.spectra/agents/` does not.
  **Expected**: The command skips `.spectra/workflows/` creation and creates `.spectra/agents/`.

- **Condition**: All `.spectra/` and `spec/` directories and all built-in files already exist.
  **Expected**: The command prints warnings for each skipped file and exits with code 0, printing `"Spectra project initialized successfully"`.

- **Condition**: A built-in workflow file `SimpleSdd.yaml` already exists but has different content.
  **Expected**: The command prints a warning and skips the file. The existing file is not overwritten or compared.

- **Condition**: Creating `.spectra/` directory fails due to permission denied.
  **Expected**: The command prints `"Error: failed to create directory '.spectra': permission denied"` and exits with code 1. No subsequent operations are performed. `.gitignore` remains modified (if it was modified in Phase 0).

- **Condition**: Creating `.spectra/sessions/` fails after `.spectra/` is successfully created.
  **Expected**: The command prints an error and exits with code 1. `.spectra/` remains on disk (partial state).

- **Condition**: Writing a built-in workflow file fails due to disk full.
  **Expected**: The command prints `"Error: failed to write built-in file '.spectra/workflows/SimpleSdd.yaml': no space left on device"` and exits with code 1. Any previously written files remain on disk (partial state). `spec/` directories and files are not created.

- **Condition**: Creating `spec/` directory fails after all `.spectra/` directories and files are successfully created.
  **Expected**: The command prints `"Error: failed to create directory 'spec': permission denied"` and exits with code 1. All `.spectra/` directories and files remain on disk (partial state).

- **Condition**: Creating `spec/logic/` directory fails after `spec/` is successfully created.
  **Expected**: The command prints an error and exits with code 1. `spec/` and all `.spectra/` directories and files remain on disk (partial state).

- **Condition**: Writing `spec/ARCHITECTURE.md` fails after all `spec/` directories are successfully created.
  **Expected**: The command prints `"Error: failed to write built-in file 'spec/ARCHITECTURE.md': disk quota exceeded"` and exits with code 1. All directories (`.spectra/` and `spec/`) and all `.spectra/` files remain on disk (partial state).

- **Condition**: `spec/ARCHITECTURE.md` already exists but `spec/CONVENTIONS.md` does not.
  **Expected**: The command prints `"Warning: spec file 'ARCHITECTURE.md' already exists, skipping"` and creates `spec/CONVENTIONS.md`.

- **Condition**: `spec/` exists as a file (not a directory).
  **Expected**: Directory creation fails with an error: `"Error: failed to create directory 'spec': file exists"`. The command exits with code 1. All `.spectra/` directories and files remain on disk.

- **Condition**: Current working directory is the filesystem root (`/` or `C:\`).
  **Expected**: The command attempts to create `/.spectra/` or `C:\.spectra\`. This may fail due to permissions, in which case the command exits with an error.

- **Condition**: Current working directory is read-only (e.g., a CD-ROM mount point).
  **Expected**: Directory creation fails with permission denied. The command exits with code 1.

- **Condition**: `.spectra` exists as a file (not a directory).
  **Expected**: Directory creation fails with an error: `"Error: failed to create directory '.spectra': file exists"`. The command exits with code 1.

- **Condition**: A built-in agent definition file is missing from the embedded resources (programming error).
  **Expected**: The command prints an error: `"Error: failed to read embedded file 'builtin/agents/<AgentRole>.yaml': file does not exist"` and exits with code 1.

- **Condition**: User invokes `spectra init` multiple times in the same directory.
  **Expected**: First invocation creates directories and files. Subsequent invocations print warnings for existing files and exit successfully.

- **Condition**: User invokes `spectra init` in a subdirectory of an existing Spectra project.
  **Expected**: The command initializes new `.spectra/` and `spec/` directories in the current directory (nested project). It does not use SpectraFinder.

- **Condition**: Embedded built-in files contain invalid YAML or Markdown syntax (programming error).
  **Expected**: The command copies the files as-is without validation. YAML files will fail validation when loaded by WorkflowDefinitionLoader or AgentDefinitionLoader later. Markdown files are not validated by the system.

- **Condition**: A built-in file path contains special characters or spaces (e.g., `My Workflow.yaml` or `My README.md`).
  **Expected**: The command composes the target path using the filename as-is. The file is created with the special characters/spaces in its name. This may cause issues if the naming convention is violated, but the `init` command does not validate filenames.

- **Condition**: A built-in spec template file is missing from the embedded resources (programming error).
  **Expected**: The command prints an error: `"Error: failed to read embedded file 'builtin/spec/ARCHITECTURE.md': file does not exist"` and exits with code 1.

## Related

- [run Subcommand](./run.md) - Run a workflow after initialization
- [clear Subcommand](./clear.md) - Clear session data
- [SpectraFinder](../../storage/spectra_finder.md) - Used by other commands to locate `.spectra/`, but not by `init`
- [StorageLayout](../../storage/storage_layout.md) - Defines the `.spectra/` directory structure
- [WorkflowDefinitionLoader](../../storage/workflow_definition_loader.md) - Loads and validates workflow files after initialization
- [AgentDefinitionLoader](../../storage/agent_definition_loader.md) - Loads and validates agent files after initialization
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - Framework architecture overview
