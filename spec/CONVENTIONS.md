# Conventions

## Language & Toolchain

**Go (spectra core)**
- Go 1.22+
- `gofmt` for formatting (enforced)
- `golangci-lint` for linting

**TypeScript (spectra-vscode extension)**
- TypeScript with strict mode enabled
- Node.js LTS
- `eslint` + `prettier` for linting and formatting (enforced)
- `vsce` for packaging and publishing

## Naming

**Go**

| Kind | Style | Example |
|---|---|---|
| Package | lowercase, no underscore | `agentrunner` |
| File | snake_case | `agent_runner.go` |
| Exported | PascalCase | `RunAgent` |
| Unexported | camelCase | `runAgent` |
| Constants | PascalCase or ALL_CAPS only for universe-block consts | `MaxRetries` |
| Interface | noun or `-er` suffix | `Runner`, `MessageSender` |
| Test file | `<file>_test.go` | `agent_runner_test.go` |

**TypeScript (spectra-vscode extension)**

| Kind | Style | Example |
|---|---|---|
| File | camelCase | `sessionTree.ts` |
| Class / Interface | PascalCase | `SessionTreeProvider` |
| Function / variable | camelCase | `startSession` |
| VS Code command ID | `spectra.<action>` | `spectra.startSession` |
| Configuration key | `spectra.<key>` | `spectra.socketPath` |
| Event emitter | `on` prefix | `onSessionChanged` |
| Test file | `*.test.ts` | `sessionTree.test.ts` |

## Code Location

Define and follow these placement rules:

**Go**

- Production source code goes under `./`. e.g. `./main.go`, `./cmd/root.go`
- Unit tests live next to source files in the same package using `_test.go` suffix.
- End-to-end tests go under `./test/e2e`.
- Race condition tests go under `./test/race`.
- Shared test helpers and fixtures for non-unit tests should also go somewhere under `./test/`.

**TypeScript (spectra-vscode extension)**

```
vscode/
  src/
    extension.ts          # activation entry point
    commands/             # one file per registered command
    providers/            # TreeDataProvider, WebviewProvider, etc.
    services/             # business logic; communicate with spectra CLI/socket
    utils/                # pure utility helpers
  test/
    suite/                # Mocha unit test suites (*.test.ts)
    e2e/                  # VS Code integration tests via @vscode/test-electron
  package.json
  tsconfig.json
  .eslintrc.json
```

- Declare `activationEvents` explicitly in `package.json`; avoid `onStartupFinished` unless truly necessary.
- Register all disposables via `context.subscriptions.push(...)`.
- Interact with the spectra runtime exclusively through its Unix socket or CLI subprocess — never by importing Go packages directly.
- Wrap all socket/process I/O in a `SpectraClient` service class; keep I/O concerns out of providers and commands.
- Use VS Code's `OutputChannel` for surfacing raw CLI output; do not write to `console.log` in production code.
- Use a Content Security Policy on every webview: `default-src 'none'`.
- Pass data from extension to webview only via `postMessage`; never inject raw user data into HTML strings.

## Error Handling

**Go**

- Never ignore errors
- Wrap with context: `fmt.Errorf("doing X: %w", err)`
- Define sentinel errors at package level: `var ErrNotFound = errors.New("not found")`
- Return errors, do not panic in library code

**TypeScript (spectra-vscode extension)**

- Surface errors to the user via `vscode.window.showErrorMessage`
- Log diagnostic detail to a dedicated `OutputChannel`
- Never swallow errors silently; always `await` promises or attach `.catch`

## Testing

**Go**

- Use `testing` stdlib + `github.com/stretchr/testify`
- Test function: `TestFoo_scenario` (e.g. `TestRunAgent_returnsErrOnTimeout`)
- Table-driven tests preferred for multiple cases
- Mock interfaces, not concrete types

**TypeScript (spectra-vscode extension)**

- Unit tests: Mocha + `sinon` for stubs/spies; run without a VS Code host when possible
- Integration tests: `@vscode/test-electron` for tests that require the extension host
- Mock the `SpectraClient` interface in unit tests; never spawn a real spectra process

## Imports

**Go**

Group in this order (separated by blank lines):

```go
import (
    "stdlib"

    "external/module"
    "internal/package"
)
```

**TypeScript (spectra-vscode extension)**

Group in this order (separated by blank lines):

```ts
import * as vscode from 'vscode';

import { externalLib } from 'external-lib';

import { localModule } from './localModule';
```

## Code Style

**Go**

- Prefer short variable names in small scopes (`i`, `err`, `ok`)
- Return early on error (guard clauses)
- No naked returns
- Context (`context.Context`) is always the first parameter when present
- Unexported struct fields by default; export only what is needed
- Interfaces belong in the package that *uses* them, not the one that implements them

**TypeScript (spectra-vscode extension)**

- Prefer `const` over `let`; never use `var`
- Return early on error (guard clauses)
- Use `async`/`await` over raw Promise chains
- Export only what is needed; keep implementation details unexported
- Interfaces belong in the file that *uses* them, not the one that implements them
