package database

import "database/sql"

type ItemRow struct {
	ID       string
	Title    string
	Type     string
	Created  string
	Due      sql.NullString
	Status   sql.NullString
	RemindAt sql.NullString
	Path     string
}

type ItemFTSRow struct {
	ID    string
	Title string
	Body  string
}

type TagRow struct {
	ItemID string
	Tag    string
}

type EdgeRow struct {
	SourceID string
	TargetID string
}
