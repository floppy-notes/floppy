package database

import (
	"database/sql"
	"fmt"

	"github.com/floppy-notes/floppy/internal/item"
)

func insertItem(tx *sql.Tx, r ItemRow) error {
	_, err := tx.Exec(
		`INSERT INTO items (id, title, type, created, due, status, remind_at, path)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Title, r.Type, r.Created, r.Due, r.Status, r.RemindAt, r.Path,
	)
	if err != nil {
		return fmt.Errorf("inserting item %q: %w", r.ID, err)
	}
	return nil
}

func insertTag(tx *sql.Tx, r TagRow) error {
	_, err := tx.Exec(
		`INSERT INTO tags (item_id, tag) VALUES (?, ?)`,
		r.ItemID, r.Tag,
	)
	if err != nil {
		return fmt.Errorf("inserting tag %q on %q: %w", r.Tag, r.ItemID, err)
	}
	return nil
}

func insertItemFTS(tx *sql.Tx, r ItemFTSRow) error {
	_, err := tx.Exec(
		`INSERT INTO items_fts (id, title, body) VALUES (?, ?, ?)`,
		r.ID, r.Title, r.Body,
	)
	if err != nil {
		return fmt.Errorf("inserting fts for item %q: %w", r.ID, err)
	}
	return nil
}

func insertEdge(tx *sql.Tx, r EdgeRow) error {
	_, err := tx.Exec(
		`INSERT INTO edges (source_id, target_id) VALUES (?, ?)`,
		r.SourceID, r.TargetID,
	)
	if err != nil {
		return fmt.Errorf("inserting edge %q->%q: %w", r.SourceID, r.TargetID, err)
	}
	return nil
}

func indexOne(tx *sql.Tx, it item.Item) error {
	if err := insertItem(tx, ItemRow{
		ID:       it.Front.ID,
		Title:    it.Front.Title,
		Type:     it.Front.Type,
		Created:  it.Front.Created,
		Due:      sql.NullString{String: it.Front.Due, Valid: it.Front.Due != ""},
		Status:   sql.NullString{String: it.Front.Status, Valid: it.Front.Status != ""},
		RemindAt: sql.NullString{String: it.Front.RemindAt, Valid: it.Front.RemindAt != ""},
		Path:     it.Path,
	}); err != nil {
		return err
	}
	if err := insertItemFTS(tx, ItemFTSRow{ID: it.Front.ID, Title: it.Front.Title, Body: it.Body}); err != nil {
		return err
	}
	for _, tag := range it.Front.Tags {
		if err := insertTag(tx, TagRow{ItemID: it.Front.ID, Tag: tag}); err != nil {
			return err
		}
	}
	for _, target := range it.Front.Related {
		if err := insertEdge(tx, EdgeRow{SourceID: it.Front.ID, TargetID: target}); err != nil {
			return err
		}
	}
	return nil
}
