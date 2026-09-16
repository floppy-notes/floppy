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

func (req UpdateRequest) validate() error {
	switch req.Type {
	case item.TypeNote:
	case item.TypeTask:
		if req.Status != "" && !item.IsValidTaskStatus(req.Status) {
			return fmt.Errorf("invalid task status: %q (should be: open, done, archived)", req.Status)
		}
	case item.TypeReminder:
		if req.Status != "" && !item.IsValidReminderStatus(req.Status) {
			return fmt.Errorf("invalid reminder status: %q (should be: pending, fired, dismissed)", req.Status)
		}
	default:
		return fmt.Errorf("unsupported item type: %s", req.Type)
	}
	return nil
}

func Update(db *database.Database, req UpdateRequest, isReplace bool) (item.Item, error) {
	if err := req.validate(); err != nil {
		return item.Item{}, err
	}

	row, err := database.FindByIdAndType(db.Conn, req.ID, string(req.Type))

	if err != nil {
		return item.Item{}, err
	}

	raw, err := os.ReadFile(row.Path)

	if err != nil {
		return item.Item{}, fmt.Errorf("reading item file: %w", err)
	}

	it, err := item.Parse(raw)

	if err != nil {
		return item.Item{}, fmt.Errorf("parsing item file : %w", err)
	}

	it.Path = row.Path

	defineCommon(&it, req, isReplace)

	switch req.Type {
	case item.TypeTask:
		if req.Due != "" {
			it.Front.Due = req.Due
		}
		if req.Status != "" {
			it.Front.Status = req.Status
		}
	case item.TypeReminder:
		if req.RemindAt != "" {
			it.Front.RemindAt = req.RemindAt
		}
		if req.Status != "" {
			it.Front.Status = req.Status
		}
	}

	rawContent, err := item.Format(it)

	if err != nil {
		return item.Item{}, err
	}

	if err := vault.Overwrite(it.Path, rawContent); err != nil {
		return item.Item{}, err
	}

	if err := db.Reindex(it); err != nil {
		return item.Item{}, err
	}

	return it, nil
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
