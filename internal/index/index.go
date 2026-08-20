package index

import (
	"fmt"
	"os"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/vault"
)

func Rebuild(db *database.Database, vaultRootDir string) error {
	items, skipped, err := vault.Walk(vaultRootDir)
	if err != nil {
		return fmt.Errorf("walking vault: %w", err)
	}
	for _, s := range skipped {
		fmt.Fprintln(os.Stderr, s)
	}
	return db.Rebuild(items)
}
