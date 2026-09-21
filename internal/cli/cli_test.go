package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/floppy-notes/floppy/internal/domain"
)

func TestDefaultVaultPath(t *testing.T) {
	t.Run("should use FLOPPY_VAULT when it is set", func(t *testing.T) {
		t.Setenv("FLOPPY_VAULT", "/tmp/custom-vault")

		got := defaultVaultPath()

		if got != "/tmp/custom-vault" {
			t.Errorf("defaultVaultPath() = %q, want %q", got, "/tmp/custom-vault")
		}
	})

	t.Run("should fall back to the home folder when FLOPPY_VAULT is empty", func(t *testing.T) {
		t.Setenv("FLOPPY_VAULT", "")

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() returned unexpected error: %v", err)
		}

		want := filepath.Join(home, domain.DefaultDbFolder)

		got := defaultVaultPath()

		if got != want {
			t.Errorf("defaultVaultPath() = %q, want %q", got, want)
		}
	})
}
