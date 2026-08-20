package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/floppy-notes/floppy/internal/item"
	_ "modernc.org/sqlite"
)

const (
	DbName = "floppy.db"
)

type Database struct {
	config DatabaseConfig
	Conn   *sql.DB
}

func Open(options ...Option) (*Database, error) {
	dbc, err := newDBC(options...)

	if err != nil {
		return nil, fmt.Errorf("setting up dbc: %w", err)
	}

	dataSourceName := filepath.Join(dbc.Path, DbName)

	if err := os.MkdirAll(dbc.Path, 0o755); err != nil {
		return nil, fmt.Errorf("creating db folder: %w", err)
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connection to db: %w", err)
	}
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return &Database{
		config: *dbc,
		Conn:   db,
	}, nil
}

func (db *Database) RunInTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := db.Conn.Begin()

	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *Database) clearAll(tx *sql.Tx) error {
	tables := []string{"items", "items_fts", "tags", "edges"}
	for _, t := range tables {
		if _, err := tx.Exec("DELETE FROM " + t); err != nil {
			return fmt.Errorf("clearing %s: %w", t, err)
		}
	}
	return nil
}

func (d *Database) Rebuild(items []item.Item) error {
	return d.RunInTransaction(func(tx *sql.Tx) error {
		if err := d.clearAll(tx); err != nil {
			return err
		}
		for _, it := range items {
			if err := indexOne(tx, it); err != nil {
				return fmt.Errorf("indexing %s: %w", it.Front.ID, err)
			}
		}
		return nil
	})
}
