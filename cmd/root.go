package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spectra",
	Short: "Spectra CLI tool",
}

func Execute() error {
	return rootCmd.Execute()
}
