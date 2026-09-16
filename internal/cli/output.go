package cli

import (
	"database/sql"
	"encoding/json"
	"io"

	"github.com/floppy-notes/floppy/internal/database"
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
		items = append(items, itemView{
			ID:       r.ID,
			Title:    r.Title,
			Type:     r.Type,
			Created:  r.Created,
			Due:      nullToPtr(r.Due),
			Status:   nullToPtr(r.Status),
			RemindAt: nullToPtr(r.RemindAt),
			Path:     r.Path,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(itemsResult{Items: items, Count: len(items)})
}

func nullToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
