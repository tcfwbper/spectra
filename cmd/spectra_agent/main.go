package main

import (
	"os"

	spectraagent "github.com/tcfwbper/spectra/internal/cmd/spectra_agent"
)

var osExit = os.Exit
var execute = spectraagent.Execute

func main() {
	osExit(execute())
}
