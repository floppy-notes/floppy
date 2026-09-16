package cli

import (
	"database/sql"
	"encoding/json"
	"io"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
)

type itemView struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Type     string  `json:"type"`
	Created  string  `json:"created"`
	Due      *string `json:"due"`
	Status   *string `json:"status"`
	RemindAt *string `json:"remind_at"`
	Path     string  `json:"path"`
}

type itemsResult struct {
	Items []itemView `json:"items"`
	Count int        `json:"count"`
}

func writeItems(w io.Writer, rows []database.ItemRow) error {
	items := make([]itemView, 0, len(rows))
	for _, r := range rows {
		items = append(items, viewFromRow(r))
	}
	return encode(w, itemsResult{Items: items, Count: len(items)})
}

func writeItem(w io.Writer, it item.Item) error {
	return encode(w, viewFromItem(it))
}

func encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func viewFromRow(r database.ItemRow) itemView {
	return itemView{
		ID:       r.ID,
		Title:    r.Title,
		Type:     r.Type,
		Created:  r.Created,
		Due:      nullToPtr(r.Due),
		Status:   nullToPtr(r.Status),
		RemindAt: nullToPtr(r.RemindAt),
		Path:     r.Path,
	}
}

func viewFromItem(it item.Item) itemView {
	return itemView{
		ID:       it.Front.ID,
		Title:    it.Front.Title,
		Type:     it.Front.Type,
		Created:  it.Front.Created,
		Due:      emptyToPtr(it.Front.Due),
		Status:   emptyToPtr(it.Front.Status),
		RemindAt: emptyToPtr(it.Front.RemindAt),
		Path:     it.Path,
	}
}

func nullToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func emptyToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
