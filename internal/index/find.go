package index

import (
	"fmt"
	"os"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
)

func FindByID(db *database.Database, id string) (item.Item, error) {
	row, err := database.FindByID(db.Conn, id)

	if err != nil {
		return item.Item{}, err
	}

	return loadItem(row)
}

func loadItem(row database.ItemRow) (item.Item, error) {
	raw, err := os.ReadFile(row.Path)

	if err != nil {
		return item.Item{}, fmt.Errorf("reading item file: %w", err)
	}

	it, err := item.Parse(raw)

	if err != nil {
		return item.Item{}, fmt.Errorf("parsing item file: %w", err)
	}

	it.Path = row.Path

	return it, nil
}
