package database

import (
	"database/sql"
	"fmt"
)

var schemaStatements = []string{
	`
	CREATE TABLE IF NOT EXISTS items (
		id        TEXT PRIMARY KEY,
		type      TEXT NOT NULL,
		created   TEXT NOT NULL,
		due       TEXT,
		status    TEXT,
		remind_at TEXT,
		tags      TEXT,
		path      TEXT NOT NULL
	)
	`,
	`
	CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
		id UNINDEXED,
		body
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS tags (
		item_id TEXT NOT NULL,
		tag 	TEXT NOT NULL,
		PRIMARY KEY (item_id, tag)
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS edges (
		source_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		PRIMARY KEY (source_id, target_id)
	)
	`,
}

func initSchema(db *sql.DB) error {
	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("running schema statement: %w", err)
		}
	}
	return nil
}
