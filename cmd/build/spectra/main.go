package main

import (
	"os"

	"github.com/tcfwbper/spectra/cmd/spectra"
	"github.com/tcfwbper/spectra/runtime"
)

func main() {
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithRunRuntime(runtime.NewProductionRunner()),
	)
	os.Exit(cmd.Execute())
}
