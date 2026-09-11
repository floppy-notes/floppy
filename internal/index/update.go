package index

import (
	"fmt"
	"os"
	"strings"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
	"github.com/floppy-notes/floppy/internal/vault"
)

type UpdateRequest struct {
	ID       string
	Type     item.ItemType
	Title    string
	Body     string
	Tags     []string
	Related  []string
	Due      string
	RemindAt string
	Status   string
}

func Update(db *database.Database, req UpdateRequest, isReplace bool) error {
	row, err := database.FindByIdAndType(db.Conn, req.ID, string(req.Type))

	if err != nil {
		return err
	}

	raw, err := os.ReadFile(row.Path)

	if err != nil {
		return fmt.Errorf("reading item file: %w", err)
	}

	it, err := item.Parse(raw)

	if err != nil {
		return fmt.Errorf("parsing item file : %w", err)
	}

	it.Path = row.Path

	defineCommon(&it, req, isReplace)

	switch req.Type {
	case item.TypeNote:
	case item.TypeTask:
		if req.Due != "" {
			it.Front.Due = req.Due
		}
		if req.Status != "" && item.IsValidTaskStatus(req.Status) {
			it.Front.Status = req.Status
		}
	case item.TypeReminder:
		if req.RemindAt != "" {
			it.Front.RemindAt = req.RemindAt
		}
		if req.Status != "" && item.IsValidReminderStatus(req.Status) {
			it.Front.Status = req.Status
		}
	default:
		return fmt.Errorf("unsupported item type: %s", req.Type)
	}

	rawContent, err := item.Format(it)

	if err != nil {
		return err
	}

	if err := vault.Overwrite(it.Path, rawContent); err != nil {
		return err
	}

	return db.Reindex(it)
}

func defineCommon(i *item.Item, req UpdateRequest, isReplace bool) {
	if req.Title != "" {
		i.Front.Title = req.Title
	}

	if req.Body != "" {
		if isReplace {
			i.Body = req.Body
		} else {
			i.Body = strings.TrimRight(i.Body, "\n") + "\n" + req.Body
		}
	}

	if len(req.Tags) > 0 {
		if isReplace {
			i.Front.Tags = req.Tags
		} else {
			i.Front.Tags = append(i.Front.Tags, req.Tags...)
		}
	}

	if len(req.Related) > 0 {
		if isReplace {
			i.Front.Related = req.Related
		} else {
			i.Front.Related = append(i.Front.Related, req.Related...)
		}
	}
}
