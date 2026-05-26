# Spectra Overview

Spectra is a workflow-centric development framework for coordinating humans and AI agents within the same execution flow. It provides two primary CLIs:

- `spectra`: initializes projects, loads workflows, and starts the runtime.
- `spectra-agent`: allows humans or AI agents to interact with the runtime, emit events, and advance the workflow.

In Spectra, a workflow is a cyclic graph, and only one node can be active at a time.

- Each node in the graph represents either a human or an AI agent.
- Each edge in the graph represents a Transition triggered by a specific event type.
- A workflow definition must guarantee that execution starts at a Human node and ends at a Human node.
- `EntryNode` defines where the workflow starts.
- `ExitTransitions` defines terminal events for the workflow.

# Prerequisites

- OS: Linux
- Go
- Claude CLI

# Installation

Users can get `spectra` and `spectra-agent` in two ways.

## Option 1: Download prebuilt binaries from GitHub Releases

For end users, this is the recommended approach. Each GitHub release can publish Linux tarballs that contain both binaries.

1. Open the Releases page for this repository.
2. Download the archive that matches your Linux architecture:
	- `spectra_<version>_linux_amd64.tar.gz`
	- `spectra_<version>_linux_arm64.tar.gz`
3. Extract the archive.
4. Move both binaries into a directory on your `PATH`, for example:

```bash
tar -xzf spectra_<version>_linux_amd64.tar.gz
sudo mv spectra /usr/local/bin/spectra
sudo mv spectra-agent /usr/local/bin/spectra-agent
```

After that, users can run:

```bash
spectra --help
spectra-agent --help
```

## Option 2: Build from source

If users already have Go installed, they can build from source:

```bash
go build -o spectra ./cmd/spectra
go build -o spectra-agent ./cmd/spectra_agent
```

This repository also provides a `Makefile` shortcut:

```bash
make build
```

# Setup

1. Set up Claude authentication.

	Make sure Claude CLI is installed and usable, then complete login or any required authentication setup so agent nodes can invoke Claude successfully.

2. Create and move into your project directory.

	```bash
	mkdir my-project
	cd my-project
	```

3. Run `spectra init`.

	```bash
	spectra init
	```

	This initializes the Spectra project structure, including the `spec/` and `.spectra/` directories and built-in resources.

4. Write `spec/ARCHITECTURE.md`.

	Use this file to describe the system architecture, major components, boundaries, and the key context for how the workflow operates.

5. Write `spec/CONVENTIONS.md`.

	Use this file to define shared team conventions for specs, process, naming, and interaction patterns.

6. Customize `.spectra/agents` and `.spectra/workflows` if needed.

	You can add or modify agent and workflow definitions to match your team's process.

# Usage

Start a workflow runtime:

```bash
spectra run --workflow <WorkflowName>
```

This command starts a Spectra runtime, creates a session, and begins execution from the workflow's `EntryNode`.

Interact with the runtime, provide requirements, and trigger a transition:

```bash
spectra-agent event emit <EventType> --session-id <SpectraSessionUUID> --message <UserRequirements>
```

This command sends an event into the current session so the runtime can evaluate whether a transition should occur based on the workflow definition.

If needed, you can also provide `--payload` or `--claude-session-id`:

```bash
spectra-agent event emit <EventType> \
  --session-id <SpectraSessionUUID> \
  --message <UserRequirements> \
  --payload '{"key":"value"}' \
  --claude-session-id <ClaudeSessionUUID>
```

# Contribution

This project prefers a spec-first contribution workflow.

1. When behavior, flow, or rules change, update the relevant spec first.
2. After the spec passes initial review, proceed with implementation and tests.
3. Keep changes small and focused, and include the relevant validation.
4. When practical, prefer using `spectra` to support the development workflow.

For more details, see [CONTRIBUTING.md](/CONTRIBUTING.md).

# Maintainer Release Flow

If you are maintaining this repository, you can publish downloadable binaries directly through GitHub Releases.

1. Commit and push your changes to `main`.
2. Create and push a version tag:

```bash
git tag 1.0.0
git push origin 1.0.0
```

3. GitHub Actions will build release archives for Linux `amd64` and `arm64`.
4. The workflow will attach the generated tarballs and `checksums.txt` to the GitHub release.

This is the simplest way to let users download `spectra` and `spectra-agent` without asking them to install Go first.
