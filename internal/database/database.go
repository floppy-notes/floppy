package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	DB_NAME = "floppy.db"
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

	dataSourceName := filepath.Join(dbc.Path, DB_NAME)

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
