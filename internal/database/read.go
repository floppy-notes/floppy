package database

import (
	"database/sql"
	"fmt"
)

func Search(db *sql.DB, query string, limit int) ([]ItemRow, error) {
	rows, err := db.Query(
		`
		SELECT items.id, items.type, items.created, items.due, items.status, items.remind_at, items.path
		FROM items_fts
		JOIN items on items.id = items_fts.id
		WHERE items_fts MATCH ?
		ORDER BY items_fts.rank
		LIMIT ?
		`, query, limit,
	)

	if err != nil {
		return []ItemRow{}, fmt.Errorf("searching items: %w", err)
	}

	defer rows.Close()

	var results []ItemRow
	for rows.Next() {
		r := new(ItemRow)
		if err := rows.Scan(&r.ID, &r.Type, &r.Created, &r.Due, &r.Status, &r.RemindAt, &r.Path); err != nil {
			return []ItemRow{}, fmt.Errorf("scanning item: %w", err)
		}
		results = append(results, *r)
	}

	if err := rows.Err(); err != nil {
		return []ItemRow{}, fmt.Errorf("iterating results: %w", err)
	}

	return results, nil
}
