package vault

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/floppy-notes/floppy/internal/item"
)

func Walk(vaultRootDir string) ([]item.Item, []error, error) {
	items := []item.Item{}
	var skipped []error

	err := filepath.WalkDir(vaultRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		rawData, err := os.ReadFile(path)
		if err != nil {
			skipped = append(skipped, fmt.Errorf("skipping %s: %w", path, err))
			return nil
		}

		parsedItem, err := item.Parse(rawData)
		if err != nil {
			skipped = append(skipped, fmt.Errorf("skipping %s: %w", path, err))
			return nil
		}
		parsedItem.Path = path
		items = append(items, parsedItem)
		return nil

	})
	if err != nil {
		return nil, skipped, fmt.Errorf("walking vault: %w", err)
	}
	return items, skipped, nil

}
