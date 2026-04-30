# Logic Specification Style Guide

This document defines the conventions for writing logic spec files and provides language-agnostic guidance. But you should follow the target programming conventions when you write logic spec files.

---

## Purpose

A logic specification describes **what** a unit of logic does — its contract with the outside world — without prescribing **how** it is implemented. It serves as the single source of truth for behavior expectations.

---

## Scope

Each specification file covers exactly one logical unit. A logical unit is any discrete, nameable piece of behavior:

- a function / procedure
- a module / service
- an algorithm
- a business rule or policy

Do **not** mix multiple unrelated units in one file.

---

## File Structure

Every specification file must contain the following sections, in order:

```
# <Unit Name>

## Overview
## Behavior
## Inputs
## Outputs
## Invariants
## Edge Cases
## Related
```

Omit a section only if it genuinely does not apply.

---

## Section Conventions

### Overview

One to three sentences. State:

1. What the unit does (its responsibility).
2. What it does **not** do (scope boundary, if non-obvious).

### Behavior

Describe the core logic as a numbered list of declarative statements.

- Use present tense: *"Returns the largest element."*
- One statement per observable behavior.
- Avoid implementation language (no pseudo-code unless unavoidable).
- Order from the happy path to exceptional paths.

### Inputs

A table or list. For each input, specify:

| Field | Description |
|-------|-------------|
| Name | Identifier used in the spec |
| Type | Logical type (e.g., integer, list, boolean) — not a language type |
| Constraints | Valid range, allowed values, nullability |
| Required? | Yes / No / Conditional |

### Outputs

Same format as Inputs. Include error / failure outputs.

### Invariants

Conditions that must hold **at all times**, regardless of input:

- State what the unit never violates.
- Use *"Must …"* / *"Must not …"* phrasing.
- Keep each invariant independently testable.

### Edge Cases

List inputs or states that require special handling. For each:

```
- Condition: <what triggers this case>
  Expected: <what the unit must do>
```

### Related

Links to related specs, external references, or upstream policies. Use relative paths for workspace files.
