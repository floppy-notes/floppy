package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/floppy-notes/floppy/internal/domain"
	"github.com/spf13/cobra"
)

var vaultPath string

var rootCmd = &cobra.Command{
	Use:           "floppy",
	Short:         "local-first notes, tasks & reminders",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "floppy:", err)
	}
	return err
}

func init() {
	rootCmd.Version = versionString()
	rootCmd.SetVersionTemplate("floppy {{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&vaultPath, "vault", defaultVaultPath(), "vault path")
}

func defaultVaultPath() string {
	if v := os.Getenv("FLOPPY_VAULT"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, domain.DefaultDbFolder)
}
