# Spectra VS Code Extension

A VS Code extension that provides a graphical sidebar interface for managing Spectra workflow sessions.

## Features

- **Session List** — View all active and past workflow sessions, launch new sessions, and terminate running ones.
- **Session Detail** — Inspect a session's event history, current workflow node, and available transitions.
- **Event Dispatch** — Emit events into a running session directly from the editor.
- **File Watchers** — Automatically refreshes when session or workflow files change on disk.

## Prerequisites

- VS Code 1.85+
- Node.js 20+
- `spectra` and `spectra-agent` binaries installed and available on PATH (or configured via settings)

## Development

It is recommended to use Spectra to develop this extension for specification maintenance. If you need to verify the development results locally, please refer to the following commands.

```bash
cd vscode
npm install
npm run compile
npm test
```

## Packaging

```bash
npm run package
```

This produces a `.vsix` file that can be installed with:

```bash
code --install-extension spectra-vscode-*.vsix
```

## Extension Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `spectra.binaryPath` | Path to the `spectra` binary | `spectra` on PATH |
| `spectra.agentBinaryPath` | Path to the `spectra-agent` binary | `spectra-agent` on PATH |
| `spectra.projectRoot` | Root directory of the project | Workspace folder |

## License

Apache-2.0
