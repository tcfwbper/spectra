package main

import (
	"os"

	"github.com/tcfwbper/spectra/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
