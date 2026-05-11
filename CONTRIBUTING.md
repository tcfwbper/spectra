# Contribution Guide

This project does not need a heavy open source process. If you want to contribute, follow the rules below and keep changes small, reviewable, and consistent with the existing specs.

## Core Rule

Contributors should use `spectra` for development whenever practical.

- `spectra` is the preferred way to plan, implement, and prepare changes for review.
- If you use another workflow or toolchain, you must still follow the same spec-first process.
- Do not edit implementation first when the behavior, rules, or flow are changing.
- Start from the relevant logic spec, workflow spec, or agent definition.

## Expected Flow

1. Identify the problem or improvement you want to make.
2. Update the relevant logic spec first when the behavior, rules, or flow changes.
3. After the logic spec is updated, notify a reviewer for an initial review.
4. Do not continue to implementation or follow-up code changes until that initial review passes.
5. After approval, make the required code, test, and document changes.
6. Run the relevant validation for the area you changed before asking for final review.

## What To Include In A Contribution

- A clear reason for the change.
- Updated specs when behavior changes.
- Tests for new behavior or regression coverage for bug fixes.
- Small, focused diffs whenever possible.

## Review Expectations

When you ask for review, include:

- what changed
- which spec was updated
- whether the initial spec review already passed
- what validation you ran

## Practical Notes

- Follow the conventions in `spec/CONVENTIONS.md`.
- Prefer incremental changes over large refactors.
- If a change is unclear, stop and align on the spec before touching more code.

This guide is intentionally short. The main requirement is simple: prefer developing through `spectra`, but regardless of workflow, update the spec first, get reviewer approval on that logic, then proceed with the rest of the change.