package database

import (
	"database/sql"
	"fmt"
	"strings"
)

func Search(db *sql.DB, query string, limit int) ([]ItemRow, error) {
	rows, err := db.Query(
		`
		SELECT items.id, items.title, items.type, items.created, items.due, items.status, items.remind_at, items.path
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
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Created, &r.Due, &r.Status, &r.RemindAt, &r.Path); err != nil {
			return []ItemRow{}, fmt.Errorf("scanning item: %w", err)
		}
		results = append(results, *r)
	}

	if err := rows.Err(); err != nil {
		return []ItemRow{}, fmt.Errorf("iterating results: %w", err)
	}

	return results, nil
}

type ListFilter struct {
	Type   string
	Status string
	Tag    string
	Since  string
	Until  string
}

func List(db *sql.DB, filter ListFilter, limit int) ([]ItemRow, error) {
	query := `
		SELECT items.id, items.title, items.type, items.created, items.due, items.status, items.remind_at, items.path FROM items
	`

	var conditions []string
	var args []any

	if filter.Type != "" {
		conditions = append(conditions, "items.type = ?")
		args = append(args, filter.Type)
	}

	if filter.Status != "" {
		conditions = append(conditions, "items.status = ?")
		args = append(args, filter.Status)
	}

	if filter.Since != "" {
		conditions = append(conditions, "items.created >= ?")
		args = append(args, filter.Since)
	}

	if filter.Until != "" {
		conditions = append(conditions, "items.created <= ?")
		args = append(args, filter.Until)
	}

	if filter.Tag != "" {
		query += ` JOIN tags ON tags.item_id = items.id`
		conditions = append(conditions, "tags.tag = ?")
		args = append(args, filter.Tag)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY items.created DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)

	if err != nil {
		return []ItemRow{}, fmt.Errorf("listing items: %w", err)
	}

	defer rows.Close()

	var results []ItemRow
	for rows.Next() {
		r := new(ItemRow)
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Created, &r.Due, &r.Status, &r.RemindAt, &r.Path); err != nil {
			return []ItemRow{}, fmt.Errorf("scanning item: %w", err)
		}
		results = append(results, *r)
	}

	if err := rows.Err(); err != nil {
		return []ItemRow{}, fmt.Errorf("iterating results: %w", err)
	}
	return results, nil
}
