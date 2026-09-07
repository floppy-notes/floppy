package index

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/floppy-notes/floppy/internal/item"
	"github.com/floppy-notes/floppy/internal/vault"
)

type CreateRequest struct {
	Type     item.ItemType
	Title    string
	Body     string
	Tags     []string
	Related  []string
	Due      string
	RemindAt string
}

func Create(vaultRootDir string, req CreateRequest, now time.Time) (item.Item, error) {
	folder := folderFor(req.Type)

	id, err := vault.BuildId(filepath.Join(vaultRootDir, folder), req.Title, now)

	if err != nil {
		return item.Item{}, fmt.Errorf("building id: %w", err)
	}

	path := filepath.Join(vaultRootDir, folder, now.Format("2006"), now.Format("01"), id+".md")

	front := item.Frontmatter{
		ID:      id,
		Title:   req.Title,
		Type:    string(req.Type),
		Created: now.Format(time.RFC3339),
		Tags:    req.Tags,
		Related: req.Related,
	}

	switch req.Type {
	case item.TypeTask:
		front.Due = req.Due
		front.Status = string(item.TaskOpenStatus)
	case item.TypeReminder:
		front.RemindAt = req.RemindAt
		front.Status = string(item.ReminderPendingStatus)
	}

	i := item.Item{Front: front, Body: req.Body, Path: path}

	rawContent, err := item.Format(i)

	if err != nil {
		return item.Item{}, err
	}

	if err := vault.Write(path, rawContent); err != nil {
		return item.Item{}, err
	}

	return i, nil
}

func folderFor(t item.ItemType) string {
	switch t {
	case item.TypeNote:
		return "notes"
	case item.TypeReminder:
		return "reminders"
	case item.TypeTask:
		return "tasks"
	default:
		return strings.ToLower(string(t))
	}
}
