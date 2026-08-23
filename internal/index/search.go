package index

import "github.com/floppy-notes/floppy/internal/database"

func Search(db *database.Database, query string, limit int) ([]database.ItemRow, error) {
	return database.Search(db.Conn, query, limit)
}
