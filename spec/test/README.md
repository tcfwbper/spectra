# Test Spec Style Guide

This document defines the conventions for writing test spec files and provides language-agnostic guidance. But you should follow the target programming conventions when you write test spec files.

---

## File Structure

Each spec file must contain the following sections in order:

```
# Test Specification: `<filename>`

## Source File Under Test
## Test File
---
## `<ClassName>`
### <Subsection Title>
...
```

**No numbered headings** — do not prefix section or subsection titles with numeric indexes or identifiers of any kind (e.g. `1.`, `1.1`, `TC-001:`, `TC001`). Subsection titles must come only from the list in the **Subsection Types** section below.

---

## Subsection Types

Use only these subsection titles (include only those that are applicable):

**Construction**

| Title | Purpose |
|---|---|
| `Happy Path — Construction` | Valid inputs; verify fields are stored correctly |
| `Happy Path — Default Construction` | Zero-argument construction; verify defaults |
| `Happy Path — Explicit Construction` | Named arguments; verify non-default values |

**Method Behaviour**

| Title | Purpose |
|---|---|
| `Happy Path — <method>` | Successful invocation of a named method or operation; verify return value |
| `Idempotency` | Repeated calls produce the same result with no additional side effects |
| `State Transitions` | Object moves through lifecycle states in the expected sequence |
| `Error Propagation` | Exceptions raised by a dependency surface correctly at the call site |
| `Ordering — <criterion>` | Output sequence satisfies a named ordering guarantee |

**Input Validity**

| Title | Purpose |
|---|---|
| `Boundary Values — <field>` | Inputs at or just outside valid range boundaries |
| `Null / Empty Input` | A null/nil/None value or an empty collection supplied in place of a value; verify acceptance or rejection |
| `Validation Failures` | Invalid inputs that must raise an exception |
| `Validation Failures — <field>` | Narrow to a specific field when multiple fields are tested |

**Object Characteristics**

| Title | Purpose |
|---|---|
| `Immutability` | Field assignment on an immutable instance must raise |
| `Not Immutable` | Mutable instance; field reassignment must not raise |
| `Read-Only Convention` | Mutation is not enforced; verify it does not raise |
| `Atomic Replacement` | Constructing a new instance does not mutate the original |
| `Data Independence (Copy Semantics)` | Mutation of the source array does not affect the stored value |

**Type and Exception**

| Title | Purpose |
|---|---|
| `Type Hierarchy` | Type checking and abstract base class / interface enforcement |
| `Catch Behaviour` | Exception caught by its parent class |

**Resource and Concurrency**

| Title | Purpose |
|---|---|
| `Resource Cleanup` | `close()`, context teardown, or equivalent finaliser releases resources and leaves the object inert |
| `Concurrent Behaviour` | Correct behaviour under concurrent or interleaved access |
| `Asynchronous Flow` | Async operation resolves or rejects with the expected value or error |

**External Dependencies**

| Title | Purpose |
|---|---|
| `Mock / Dependency Interaction` | Verify that the correct dependency methods are called with the correct arguments |

---

## Table Columns

Each subsection contains **exactly one Markdown table**. There is no other format. Do **not** use prose-style test cases with headings, bullet Pre-conditions, numbered Test Steps, or Expected Results sections.

Use these columns:

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|

- **Test ID** — the exact function name that will be implemented, in backticks.
- **Category** — one of `unit`, `e2e`, or `race`. Determines the test file's placement (see `CONVENTIONS.md` Code Location). All rows within a single test file must share the same category; if a spec contains rows of different categories, the QA engineer must create separate test files for each category.
- **Description** — one short sentence describing the scenario.
- **Setup** — pre-conditions to establish before the test runs, such as environment state, files, or permissions.
- **Input** — the construct or value passed as function arguments; use inline code. If no meaningful input, keep this empty.
- **Expected** — the assertion or exception, using inline code.

### Example

```markdown
### Happy Path — Write

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSafeWrite_NewFile` | `unit` | Writes content to a path that does not exist. | | `path="f.txt"`, `content=[]byte("hi")`, `perm=0644` | Returns `nil`; file exists with correct content and permissions |
| `TestSafeWrite_CreatesParentDirs` | `unit` | Creates missing parent directories before writing. | | `path="a/b/c/f.txt"`, dirs absent | Returns `nil`; all parents created with `0755`; file written |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSafeWrite_FileExists` | `unit` | Returns error without modifying the file when target already exists. | Target file exists at `path` | existing file at `path` | Returns error wrapping `ErrFileExists`; file content unchanged |
```

---

## Test ID Naming

Each Test ID is the **exact test function name** as it will appear in source code, following the naming conventions of the target language.

Examples:
- Go: `` `TestYamlUpdater_EmptyFile` ``
- Python: `` `test_yaml_updater_empty_file` ``
- Java: `` `yamlUpdater_emptyFile_throwsException` ``

Do **not** use placeholder identifiers such as `TC1`, `TC2`, or any numeric prefix. There is no separate numbering scheme — the function name is the only identifier.

Each Test ID maps to exactly **one** test function. Do not merge rows.
