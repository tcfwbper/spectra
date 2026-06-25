GO ?= go

.PHONY: build build-linux build-vscode clean

build:
	$(GO) build -o spectra ./cmd/spectra
	$(GO) build -o spectra-agent ./cmd/spectra_agent

build-vscode:
	. "$${NVM_DIR:-$$HOME/.nvm}/nvm.sh" && nvm use 20 && cd vscode && npm ci && npm run compile && npx @vscode/vsce package --out ../spectra-vscode.vsix

build-linux:
	mkdir -p dist/linux-amd64 dist/linux-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o dist/linux-amd64/spectra ./cmd/spectra
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o dist/linux-amd64/spectra-agent ./cmd/spectra_agent
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -o dist/linux-arm64/spectra ./cmd/spectra
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -o dist/linux-arm64/spectra-agent ./cmd/spectra_agent

clean:
	rm -rf dist spectra spectra-agent spectra-vscode.vsix