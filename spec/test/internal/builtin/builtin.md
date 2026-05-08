# Test Specification: `builtin_test.go`

## Source File Under Test

`internal/builtin/builtin.go`

## Test File

`internal/builtin/builtin_test.go`

---

## `EmbeddedFilesystems`

### Happy Path — Workflows

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflows_ContainsYAMLFiles` | `unit` | Workflows embed.FS contains at least one .yaml file under workflows/ directory. | None (compile-time embedding). | `fs.Glob(builtin.Workflows, "workflows/*.yaml")` | Returns non-empty slice; no error |
| `TestWorkflows_FilesAreReadable` | `unit` | Each .yaml file in Workflows is readable and non-empty. | None (compile-time embedding). | Read each file matched by `fs.Glob(builtin.Workflows, "workflows/*.yaml")` | Each file content has length > 0; no read error |
| `TestWorkflows_PreservesDirectoryStructure` | `unit` | Files are accessible under the workflows/ prefix path. | None (compile-time embedding). | `fs.ReadDir(builtin.Workflows, "workflows")` | Returns non-empty directory entries; no error |

### Happy Path — Agents

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgents_ContainsYAMLFiles` | `unit` | Agents embed.FS contains at least one .yaml file under agents/ directory. | None (compile-time embedding). | `fs.Glob(builtin.Agents, "agents/*.yaml")` | Returns non-empty slice; no error |
| `TestAgents_FilesAreReadable` | `unit` | Each .yaml file in Agents is readable and non-empty. | None (compile-time embedding). | Read each file matched by `fs.Glob(builtin.Agents, "agents/*.yaml")` | Each file content has length > 0; no read error |
| `TestAgents_PreservesDirectoryStructure` | `unit` | Files are accessible under the agents/ prefix path. | None (compile-time embedding). | `fs.ReadDir(builtin.Agents, "agents")` | Returns non-empty directory entries; no error |

### Happy Path — SpecFiles

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSpecFiles_ContainsFiles` | `unit` | SpecFiles embed.FS contains at least one file under spec/ directory. | None (compile-time embedding). | `fs.ReadDir(builtin.SpecFiles, "spec")` | Returns non-empty directory entries; no error |
| `TestSpecFiles_FilesAreReadable` | `unit` | Each file in SpecFiles is readable and non-empty. | None (compile-time embedding). | Walk `builtin.SpecFiles` and read each regular file | Each file content has length > 0; no read error |
| `TestSpecFiles_PreservesNestedDirectoryStructure` | `unit` | SpecFiles preserves subdirectory structure (e.g., spec/logic/ or spec/test/ paths). | None (compile-time embedding). | Walk `builtin.SpecFiles` and collect file paths | At least one file path contains a nested subdirectory (more than one path separator after `spec/`) |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflows_NonYAMLFilesExcluded` | `unit` | Non-.yaml files in workflows/ directory are not embedded. | None (compile-time embedding). | `fs.Glob(builtin.Workflows, "workflows/*")` then filter entries not ending in `.yaml` | All matched entries end in `.yaml` |
| `TestAgents_NonYAMLFilesExcluded` | `unit` | Non-.yaml files in agents/ directory are not embedded. | None (compile-time embedding). | `fs.Glob(builtin.Agents, "agents/*")` then filter entries not ending in `.yaml` | All matched entries end in `.yaml` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflows_ReadOnly` | `unit` | Workflows embed.FS does not implement io.Writer or any write interface. | None. | Attempt to type-assert `builtin.Workflows` to a writable interface | Type assertion fails; embed.FS is read-only |
