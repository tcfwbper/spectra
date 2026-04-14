package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/embedded"
	"github.com/tcfwbper/spectra/util"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize spectra in the current directory",
	Long:  "Create .spectra/ directory with roles, skills, and README.md, copy templates to spec/, and update .gitignore.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := "."

	// 1. Create .spectra/ with roles, skills, README.md
	if err := util.CopyEmbeddedDir(embedded.RolesFS, "roles", dir+"/.spectra/roles"); err != nil {
		return fmt.Errorf("copying roles: %w", err)
	}
	if err := util.CopyEmbeddedDir(embedded.SkillsFS, "skills", dir+"/.spectra/skills"); err != nil {
		return fmt.Errorf("copying skills: %w", err)
	}

	// 2. Copy templates/ to spec/
	if err := util.CopyEmbeddedDir(embedded.TemplatesFS, "templates", dir+"/spec"); err != nil {
		return fmt.Errorf("copying templates to spec: %w", err)
	}

	// 3. Ensure .gitignore contains ".spectra/vfs"
	if err := util.EnsureGitignoreLine(dir+"/.gitignore", ".spectra/vfs"); err != nil {
		return fmt.Errorf("updating .gitignore: %w", err)
	}

	fmt.Println("spectra initialized successfully.")
	return nil
}
