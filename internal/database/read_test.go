package database

import (
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/item"
)

func newTestDB(t *testing.T) *Database {
	t.Helper()

	db, err := Open(WithPath(t.TempDir()))
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Conn.Close() })

	return db
}

var baseNote = item.Item{
	Front: item.Frontmatter{
		ID:      "2026-08-12-0001-note",
		Title:   "tech sync",
		Type:    "note",
		Created: "2026-08-12T00:00:00-03:00",
		Tags:    []string{"tech", "meetings"},
		Related: []string{"tech-group"},
	},
	Body: "notes about the deploy pipeline",
	Path: "/vault/notes/2026/08/2026-08-12-0001-note.md",
}

var baseTask = item.Item{
	Front: item.Frontmatter{
		ID:      "2026-08-13-0001-task",
		Title:   "write tests",
		Type:    "task",
		Created: "2026-08-13T00:00:00-03:00",
		Due:     "2026-08-20",
		Status:  "open",
		Tags:    []string{"tech"},
	},
	Body: "cover the database package",
	Path: "/vault/tasks/2026/08/2026-08-13-0001-task.md",
}

var baseReminder = item.Item{
	Front: item.Frontmatter{
		ID:       "2026-08-14-0001-reminder",
		Title:    "call back",
		Type:     "reminder",
		Created:  "2026-08-14T00:00:00-03:00",
		RemindAt: "2026-08-20T10:00:00-03:00",
		Status:   "pending",
	},
	Body: "call the client back",
	Path: "/vault/reminders/2026/08/2026-08-14-0001-reminder.md",
}

// seedAll indexes the three base items, one of each type.
func seedAll(t *testing.T, db *Database) {
	t.Helper()

	if err := db.Rebuild([]item.Item{baseNote, baseTask, baseReminder}); err != nil {
		t.Fatalf("Rebuild() returned unexpected error: %v", err)
	}
}

func TestFindByIdAndType(t *testing.T) {
	t.Run("should return the row when id and type match", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := FindByIdAndType(db.Conn, baseTask.Front.ID, "task")

		if err != nil {
			t.Fatalf("FindByIdAndType() returned unexpected error: %v", err)
		}

		if got.ID != baseTask.Front.ID {
			t.Errorf("FindByIdAndType() ID = %q, want %q", got.ID, baseTask.Front.ID)
		}

		if got.Path != baseTask.Path {
			t.Errorf("FindByIdAndType() path = %q, want %q", got.Path, baseTask.Path)
		}
	})

	t.Run("should return null fields for a note", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := FindByIdAndType(db.Conn, baseNote.Front.ID, "note")

		if err != nil {
			t.Fatalf("FindByIdAndType() returned unexpected error: %v", err)
		}

		if got.Due.Valid {
			t.Errorf("FindByIdAndType() due = %v, want an invalid null string", got.Due)
		}

		if got.Status.Valid {
			t.Errorf("FindByIdAndType() status = %v, want an invalid null string", got.Status)
		}
	})

	t.Run("should return error when the id does not exist", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		_, err := FindByIdAndType(db.Conn, "missing", "note")

		if err == nil {
			t.Fatal("FindByIdAndType() returned nil error, want an error")
		}

		wantMsg := "no item found with ID missing"
		if err.Error() != wantMsg {
			t.Errorf("FindByIdAndType() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should return error when the type does not match", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		_, err := FindByIdAndType(db.Conn, baseNote.Front.ID, "task")

		if err == nil {
			t.Fatal("FindByIdAndType() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "no item found with ID") {
			t.Errorf("FindByIdAndType() error = %q, want it to contain %q", err.Error(), "no item found with ID")
		}
	})
}

func TestSearch(t *testing.T) {
	t.Run("should find an item by a word of its body", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := Search(db.Conn, "pipeline", 10)

		if err != nil {
			t.Fatalf("Search() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("Search() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != baseNote.Front.ID {
			t.Errorf("Search() ID = %q, want %q", got[0].ID, baseNote.Front.ID)
		}
	})

	t.Run("should find an item by a word of its title", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := Search(db.Conn, "tests", 10)

		if err != nil {
			t.Fatalf("Search() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("Search() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != baseTask.Front.ID {
			t.Errorf("Search() ID = %q, want %q", got[0].ID, baseTask.Front.ID)
		}
	})

	t.Run("should return an empty list when nothing matches", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := Search(db.Conn, "kubernetes", 10)

		if err != nil {
			t.Fatalf("Search() returned unexpected error: %v", err)
		}

		if len(got) != 0 {
			t.Errorf("Search() returned %d rows, want %d", len(got), 0)
		}
	})

	t.Run("should respect the limit", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := Search(db.Conn, "the", 1)

		if err != nil {
			t.Fatalf("Search() returned unexpected error: %v", err)
		}

		if len(got) > 1 {
			t.Errorf("Search() returned %d rows, want at most %d", len(got), 1)
		}
	})

	t.Run("should return a friendly error when the query syntax is invalid", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		_, err := Search(db.Conn, `"unbalanced`, 10)

		if err == nil {
			t.Fatal("Search() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "check for unbalanced quotes") {
			t.Errorf("Search() error = %q, want it to contain %q", err.Error(), "check for unbalanced quotes")
		}
	})
}

func TestList(t *testing.T) {
	t.Run("should return every item when there is no filter", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 3 {
			t.Errorf("List() returned %d rows, want %d", len(got), 3)
		}
	})

	t.Run("should order by created descending", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		want := []string{baseReminder.Front.ID, baseTask.Front.ID, baseNote.Front.ID}
		for i, id := range want {
			if got[i].ID != id {
				t.Errorf("List() row %d = %q, want %q", i, got[i].ID, id)
			}
		}
	})

	t.Run("should filter by type", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Type: "task"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("List() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != baseTask.Front.ID {
			t.Errorf("List() ID = %q, want %q", got[0].ID, baseTask.Front.ID)
		}
	})

	t.Run("should filter by status", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Status: "pending"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("List() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != baseReminder.Front.ID {
			t.Errorf("List() ID = %q, want %q", got[0].ID, baseReminder.Front.ID)
		}
	})

	t.Run("should filter by tag", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Tag: "meetings"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("List() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != baseNote.Front.ID {
			t.Errorf("List() ID = %q, want %q", got[0].ID, baseNote.Front.ID)
		}
	})

	t.Run("should return every item that shares a tag", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Tag: "tech"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 2 {
			t.Errorf("List() returned %d rows, want %d", len(got), 2)
		}
	})

	t.Run("should filter by since", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Since: "2026-08-13T00:00:00-03:00"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 2 {
			t.Errorf("List() returned %d rows, want %d", len(got), 2)
		}
	})

	t.Run("should filter by until", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Until: "2026-08-13T00:00:00-03:00"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 2 {
			t.Errorf("List() returned %d rows, want %d", len(got), 2)
		}
	})

	t.Run("should combine filters", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Type: "task", Status: "open", Tag: "tech"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("List() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != baseTask.Front.ID {
			t.Errorf("List() ID = %q, want %q", got[0].ID, baseTask.Front.ID)
		}
	})

	t.Run("should return an empty list when no item matches", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{Type: "task", Status: "archived"}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 0 {
			t.Errorf("List() returned %d rows, want %d", len(got), 0)
		}
	})

	t.Run("should respect the limit", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		got, err := List(db.Conn, ListFilter{}, 2)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 2 {
			t.Errorf("List() returned %d rows, want %d", len(got), 2)
		}
	})
}

func TestFindByID(t *testing.T) {
	t.Run("should return the row regardless of the type", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		for _, want := range []ItemRow{
			{ID: baseNote.Front.ID, Type: "note"},
			{ID: baseTask.Front.ID, Type: "task"},
			{ID: baseReminder.Front.ID, Type: "reminder"},
		} {
			got, err := FindByID(db.Conn, want.ID)

			if err != nil {
				t.Fatalf("FindByID() returned unexpected error: %v", err)
			}

			if got.Type != want.Type {
				t.Errorf("FindByID(%q) type = %q, want %q", want.ID, got.Type, want.Type)
			}
		}
	})

	t.Run("should return error when the id does not exist", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		_, err := FindByID(db.Conn, "missing")

		if err == nil {
			t.Fatal("FindByID() returned nil error, want an error")
		}

		wantMsg := "no item found with ID missing"
		if err.Error() != wantMsg {
			t.Errorf("FindByID() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should wrap the query error", func(t *testing.T) {
		db := closedDB(t)

		_, err := FindByID(db.Conn, "any")

		if err == nil {
			t.Fatal("FindByID() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "querying item") {
			t.Errorf("FindByID() error = %q, want prefix %q", err.Error(), "querying item")
		}
	})
}
