# Spectra

Spectra is a specification orchestrator that help you run automated, customized AI agent pipelines.

We provides SDD + TDD workflows by default.

## Quickstart

1. prepare claude CLI and environment variables for authentication.
2. `go build -o spectra ./cmd/build/spectra`
3. `go build -o spectra-agent ./cmd/build/spectra_agent`
4. move and export the CLI commands
5. create a folder and run `git init && spectra init`
6. `spectra run --workflow DefaultLogicSpec`
7. `spectra-agent event emit DraftSpecRequested --session-id <spectra session id> --message "<your requirement>"`
