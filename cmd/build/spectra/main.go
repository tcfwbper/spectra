package main

import (
	"os"

	"github.com/tcfwbper/spectra/cmd/spectra"
)

func main() {
	cmd := spectra.NewRootCommand()
	os.Exit(cmd.Execute())
}
