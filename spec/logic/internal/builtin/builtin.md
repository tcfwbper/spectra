# builtin

## Overview

The `builtin` package embeds default workflow definitions, agent definitions, and specification template files at compile time using Go's `//go:embed` directive. It exposes three read-only `embed.FS` variables that other packages use as data sources for copying built-in resources to user projects. This package does not interpret, validate, or transform the embedded content in any way.

## Boundaries

- Owns: compile-time embedding of files from `workflows/`, `agents/`, and `spec/` subdirectories.
- Owns: exposing three package-level `embed.FS` variables (`Workflows`, `Agents`, `SpecFiles`).
- Must not: interpret, parse, or validate the content of embedded files.
- Must not: perform any filesystem I/O at runtime.
- Must not: export any functions or types beyond the `embed.FS` variables.
- Must not: manage which files exist in the embedded directories (content is determined by the repository, not this package).

## Dependencies

None. This package has no runtime dependencies. It relies solely on the Go compiler's `//go:embed` mechanism.

Construction constraint: The embedded filesystems are populated at compile time. No runtime construction or initialization is needed.

## Behavior

1. Embeds all `.yaml` files under the `workflows/` subdirectory into the `Workflows` variable.
2. Embeds all `.yaml` files under the `agents/` subdirectory into the `Agents` variable.
3. Embeds specific files from the `spec/` subdirectory into the `SpecFiles` variable using explicit path directives (preserving directory structure within the embedded FS).
4. All three variables are available as package-level exports for use by other packages at runtime.
5. The embedded FS preserves the relative directory structure (e.g., `SpecFiles` contains paths like `spec/logic/README.md`).

## Inputs

None. All inputs are determined at compile time by the files present in the source tree.

## Outputs

| Variable | Type | Content | Root Directory in FS |
|----------|------|---------|---------------------|
| `Workflows` | `embed.FS` | All `.yaml` files from `workflows/` | `workflows/` |
| `Agents` | `embed.FS` | All `.yaml` files from `agents/` | `agents/` |
| `SpecFiles` | `embed.FS` | Explicitly listed files from `spec/` | `spec/` |

## Invariants

1. **Read-Only**: The embedded filesystems are immutable at runtime.
2. **No Logic**: This package contains no functions, methods, or runtime logic.
3. **Glob Pattern for Workflows and Agents**: `Workflows` and `Agents` use glob patterns (`*.yaml`) to include all matching files automatically.
4. **Explicit Listing for SpecFiles**: `SpecFiles` uses explicit `//go:embed` directives for each file path rather than a glob, to precisely control which spec templates are included.
5. **Directory Structure Preserved**: Paths within each `embed.FS` retain their subdirectory structure relative to the embed root.
6. **Compile-Time Failure**: If any explicitly listed file in `SpecFiles` does not exist at compile time, the build fails.

## Edge Cases

- Condition: A new `.yaml` file is added to `workflows/` or `agents/`.
  Expected: Automatically included in the next build without modifying this package.

- Condition: A new spec template file is added to `spec/`.
  Expected: Not included until an explicit `//go:embed` directive is added to this package.

- Condition: `workflows/` or `agents/` directory is empty (no `.yaml` files).
  Expected: Compile error — Go's `//go:embed` with a glob pattern that matches no files is a build failure.

- Condition: A non-`.yaml` file is placed in `workflows/` or `agents/`.
  Expected: Not embedded (glob only matches `*.yaml`).

## Related

- [BuiltinResourceCopier](../cmd/spectra/builtin_resource_copier.md) — Consumer that copies embedded files to user projects
