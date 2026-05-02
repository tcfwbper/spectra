package main

import (
	"os"

	spectra_agent "github.com/tcfwbper/spectra/cmd/spectra_agent"
)

func main() {
	cmd := spectra_agent.NewRootCommand()
	os.Exit(cmd.Execute())
}
