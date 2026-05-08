package main

import (
	"os"

	spectra "github.com/tcfwbper/spectra/internal/cmd/spectra"
)

var osExit = os.Exit
var execute = spectra.Execute

func main() {
	osExit(execute())
}
