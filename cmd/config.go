package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/util"
)

var configCmd = &cobra.Command{
	Use:   "config <key> <value>",
	Short: "Set spectra configuration values",
	Long: fmt.Sprintf(
		"Set a configuration value under [core] in .spectra/config.\n\nValid keys: %s\n\n"+
			"spec, test, and src are path settings.\nlanguage sets the project language (e.g. golang, python, rust).",
		strings.Join(util.ValidConfigKeys(), ", "),
	),
	Args: cobra.ExactArgs(2),
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	if !util.IsValidConfigKey(key) {
		return fmt.Errorf("unknown config key %q, valid keys: %s", key, strings.Join(util.ValidConfigKeys(), ", "))
	}

	configPath := ".spectra/config"
	if err := util.WriteConfigValue(configPath, key, value); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("config %s = %s\n", key, value)
	return nil
}
