package index

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/floppy-notes/floppy/internal/item"
)

var baseTime = time.Date(2026, 8, 12, 9, 30, 0, 0, time.UTC)

var baseRequest = CreateRequest{
	Type:    item.TypeNote,
	Title:   "tech sync",
	Body:    "notes from the sync",
	Tags:    []string{"tech", "meetings"},
	Related: []string{"tech-group"},
}

func TestFolderFor(t *testing.T) {
	tests := []struct {
		name string
		in   item.ItemType
		want string
	}{
		{"note", item.TypeNote, "notes"},
		{"task", item.TypeTask, "tasks"},
		{"reminder", item.TypeReminder, "reminders"},
		{"unknown type is lowercased", item.ItemType("Custom"), "custom"},
		{"empty type", item.ItemType(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := folderFor(tt.in)

			if got != tt.want {
				t.Errorf("folderFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Run("should create a note on disk", func(t *testing.T) {
		root := t.TempDir()

		got, err := Create(root, baseRequest, baseTime)

		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		wantPath := filepath.Join(root, "notes", "2026", "08", "2026-08-12-0001-tech-sync.md")
		if got.Path != wantPath {
			t.Errorf("Create() path = %q, want %q", got.Path, wantPath)
		}

		if _, err := os.Stat(got.Path); err != nil {
			t.Errorf("Stat() returned unexpected error: %v", err)
		}
	})

	t.Run("should fill the frontmatter of a note", func(t *testing.T) {
		root := t.TempDir()

		got, err := Create(root, baseRequest, baseTime)

		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		want := item.Frontmatter{
			ID:      "2026-08-12-0001-tech-sync",
			Title:   "tech sync",
			Type:    "note",
			Created: baseTime.Format(time.RFC3339),
			Tags:    []string{"tech", "meetings"},
			Related: []string{"tech-group"},
		}

		if !reflect.DeepEqual(want, got.Front) {
			t.Errorf("Create() frontmatter = %#v, want %#v", got.Front, want)
		}
	})

	t.Run("should default a task to the open status", func(t *testing.T) {
		root := t.TempDir()

		req := baseRequest
		req.Type = item.TypeTask
		req.Due = "2026-08-20"

		got, err := Create(root, req, baseTime)

		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		if got.Front.Status != string(item.TaskOpenStatus) {
			t.Errorf("Create() status = %q, want %q", got.Front.Status, item.TaskOpenStatus)
		}

		if got.Front.Due != "2026-08-20" {
			t.Errorf("Create() due = %q, want %q", got.Front.Due, "2026-08-20")
		}

		wantPath := filepath.Join(root, "tasks", "2026", "08", "2026-08-12-0001-tech-sync.md")
		if got.Path != wantPath {
			t.Errorf("Create() path = %q, want %q", got.Path, wantPath)
		}
	})

	t.Run("should default a reminder to the pending status", func(t *testing.T) {
		root := t.TempDir()

		req := baseRequest
		req.Type = item.TypeReminder
		req.RemindAt = "2026-08-20T10:00:00-03:00"

		got, err := Create(root, req, baseTime)

		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		if got.Front.Status != string(item.ReminderPendingStatus) {
			t.Errorf("Create() status = %q, want %q", got.Front.Status, item.ReminderPendingStatus)
		}

		wantPath := filepath.Join(root, "reminders", "2026", "08", "2026-08-12-0001-tech-sync.md")
		if got.Path != wantPath {
			t.Errorf("Create() path = %q, want %q", got.Path, wantPath)
		}
	})

	t.Run("should write a file that can be parsed back", func(t *testing.T) {
		root := t.TempDir()

		created, err := Create(root, baseRequest, baseTime)

		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		raw, err := os.ReadFile(created.Path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		parsed, err := item.Parse(raw)
		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if parsed.Front.ID != created.Front.ID {
			t.Errorf("parsed ID = %q, want %q", parsed.Front.ID, created.Front.ID)
		}

		if parsed.Body != created.Body {
			t.Errorf("parsed body = %q, want %q", parsed.Body, created.Body)
		}
	})

	t.Run("should increment the sequence for a second item of the same month", func(t *testing.T) {
		root := t.TempDir()

		if _, err := Create(root, baseRequest, baseTime); err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		req := baseRequest
		req.Title = "another sync"

		got, err := Create(root, req, baseTime)

		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		want := "2026-08-12-0002-another-sync"
		if got.Front.ID != want {
			t.Errorf("Create() ID = %q, want %q", got.Front.ID, want)
		}
	})

	t.Run("should fail when the item is invalid", func(t *testing.T) {
		root := t.TempDir()

		req := baseRequest
		req.Type = item.TypeTask

		_, err := Create(root, req, baseTime)

		if err == nil {
			t.Fatal("Create() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "invalid due date format") {
			t.Errorf("Create() error = %q, want it to contain %q", err.Error(), "invalid due date format")
		}
	})

	t.Run("should not leave a file behind when the item is invalid", func(t *testing.T) {
		root := t.TempDir()

		req := baseRequest
		req.Type = item.TypeTask

		if _, err := Create(root, req, baseTime); err == nil {
			t.Fatal("Create() returned nil error, want an error")
		}

		path := filepath.Join(root, "tasks", "2026", "08", "2026-08-12-0001-tech-sync.md")

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Stat() error = %v, want a not exist error", err)
		}
	})
}
