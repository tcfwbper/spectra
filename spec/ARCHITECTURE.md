# Architecture for Spectra

## Purpose

Spectra is a framework for defining and executing flexible AI agent workflows for software development.

1. Provide abstract definitions for AI agent workflows and agent roles.
2. Ship a default SDD+TDD workflow as a built-in implementation.
3. Allow users and developers to define custom workflows and agents.
4. Allow users and developers define the collaboration points between humans and AI agents.

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Framework Core (Abstract Definitions)                          │
│                                                                 │
│  ┌──────────────────────────┐  ┌──────────────────────────────┐ │
│  │ Workflow Definition      │  │ Agent Definition             │ │
│  │  - nodes / transitions   │  │  - role                      │ │
│  │  - human interaction     │  │  - capabilities              │ │
│  │  - entry / exit          │  │  - instructions              │ │
│  └──────────────────────────┘  └──────────────────────────────┘ │
│                                                                 │
│  Built-in ────────────────────────────────────────────────────  │
│    Default Workflow: SDD + TDD                                  │
│    Default Agents: Architect, QaAnalyst, SwEngineer, …          │
│                                                                 │
│  Extensions ──────────────────────────────────────────────────  │
│    Custom Workflows                  Custom Agents              │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Manage definitions
┌───────────────────────────▼─────────────────────────────────────┐
│  Admin Layer                                                    │
│                                                                 │
│  [Human] ──CLI──► [spectra]                                     │
│                     ├─ spectra init                             │
│                     ├─ spectra run --workflow <WorkflowName>    │
│                     └─ spectra clear [--session-id <UUID>]      │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Instantiate & run selected workflow
┌───────────────────────────▼─────────────────────────────────────┐
│  Workflow Runtime  (event-driven state machine executor)        │
│                                                                 │
│  Session ── holds state, event history                          │
│    │                                                            │
│    │  emit Event{type, message, payload}                        │
│    ▼                                                            │
│  State Machine ── evaluates transitions defined by workflow     │
│    │              (at most one active node at a time)           │
│    │                                                            │
│    │  dispatch to Agent role / Human                            │
│    ▼                                                            │
│  Agent / Human ── Emits next Event or reports an AgentError     │
│                                                                 │
│  AgentError ── halts state machine, persists error record,      │
│                notifies human, marks session as failed          │
│                                                                 │
│  ── Example: Default SDD+TDD workflow ───────────────────────── │
│                                                                 │
│   Event types and transitions are defined by the workflow;      │
│   the runtime only drives the state machine and manages         │
│   session lifecycle.                                            │
│                                                                 │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Invoke primitives
┌───────────────────────────▼─────────────────────────────────────┐
│  Agent Toolchain Layer  (operational primitives)                │
│                                                                 │
│  [spectra-agent]                                                │
│   ├─ spectra-agent event emit <type> --session-id <UUID> \      │
│   │                   [--claude-session-id <UUID>] \            │
│   │                   [--message <message>] [--payload <json>]  │
│   └─ spectra-agent error <message> --session-id <UUID> \        │
│                     [--claude-session-id <UUID>] \              │
│                     [--detail <json>]                           │
└─────────────────────────────────────────────────────────────────┘
```

## Components

- Framework Core: Provides abstract schemas for defining workflows and agent roles, along with built-in default implementations and an extension mechanism for custom workflows and agents.
- Spectra CLI: Responsible for system initialization, configuration, managing workflow and agent definitions, and running workflows.
- Spectra agent CLI: Called by humans and AI agents to interact with the workflow runtime.
- AI Agents: Focused on specific tasks, assisting humans in software development. Default agents include Architect, Architect Reviewer, QA Analyst, QA Spec Reviewer, QA Engineer, QA Reviewer, and SW Engineer.

## Entities

The following are framework-level structural primitives. Workflow-specific semantics are defined within each workflow's own specification and mapped onto these primitives.

- **Session**: A single execution instance of a workflow. Owns state and event history. Created on `spectra run` and persists until the workflow exits.
- **Event**: A typed signal that drives state transitions. Carries an optional message string and a workflow-defined JSON payload. Emitted by the active node (agent or human) via `spectra-agent event emit`.
- **AgentError**: A failure signal raised by an agent via `spectra-agent error` when it cannot complete its task (e.g. unrecoverable model error, missing context, tool failure). Immediately halts the state machine, records the error in the session, and notifies the human. The session is marked as failed and must be manually resolved or retried.
