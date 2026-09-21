package index

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
)

func newTestDB(t *testing.T) (*database.Database, string) {
	t.Helper()

	dir := t.TempDir()

	db, err := database.Open(database.WithPath(dir))
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Conn.Close() })

	return db, dir
}

func seedItem(t *testing.T, db *database.Database, dir string, req CreateRequest) item.Item {
	t.Helper()

	it, err := Create(dir, req, baseTime)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	if err := db.Reindex(it); err != nil {
		t.Fatalf("Reindex() returned unexpected error: %v", err)
	}

	return it
}

func TestUpdateRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateRequest
		wantErr string
	}{
		{"note ignores status", UpdateRequest{Type: item.TypeNote, Status: "whatever"}, ""},
		{"task with valid status", UpdateRequest{Type: item.TypeTask, Status: "done"}, ""},
		{"task with empty status", UpdateRequest{Type: item.TypeTask}, ""},
		{"task with reminder status", UpdateRequest{Type: item.TypeTask, Status: "pending"}, `invalid task status: "pending" (should be: open, done, archived)`},
		{"reminder with valid status", UpdateRequest{Type: item.TypeReminder, Status: "fired"}, ""},
		{"reminder with task status", UpdateRequest{Type: item.TypeReminder, Status: "open"}, `invalid reminder status: "open" (should be: pending, fired, dismissed)`},
		{"unsupported type", UpdateRequest{Type: item.ItemType("event")}, "unsupported item type: event"},
		{"empty type", UpdateRequest{}, "unsupported item type: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validate() returned unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validate() returned nil error, want an error")
			}

			if err.Error() != tt.wantErr {
				t.Errorf("validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDefineCommon(t *testing.T) {
	t.Run("should keep the current values when the request is empty", func(t *testing.T) {
		got := item.Item{
			Front: item.Frontmatter{Title: "tech sync", Tags: []string{"tech"}, Related: []string{"group"}},
			Body:  "original",
		}

		defineCommon(&got, UpdateRequest{}, false)

		if got.Front.Title != "tech sync" {
			t.Errorf("title = %q, want %q", got.Front.Title, "tech sync")
		}

		if got.Body != "original" {
			t.Errorf("body = %q, want %q", got.Body, "original")
		}

		if !reflect.DeepEqual([]string{"tech"}, got.Front.Tags) {
			t.Errorf("tags = %#v, want %#v", got.Front.Tags, []string{"tech"})
		}
	})

	t.Run("should replace the title when informed", func(t *testing.T) {
		got := item.Item{Front: item.Frontmatter{Title: "tech sync"}}

		defineCommon(&got, UpdateRequest{Title: "new title"}, false)

		if got.Front.Title != "new title" {
			t.Errorf("title = %q, want %q", got.Front.Title, "new title")
		}
	})

	t.Run("should append the body when not replacing", func(t *testing.T) {
		got := item.Item{Body: "first line\n"}

		defineCommon(&got, UpdateRequest{Body: "second line"}, false)

		want := "first line\nsecond line"
		if got.Body != want {
			t.Errorf("body = %q, want %q", got.Body, want)
		}
	})

	t.Run("should replace the body when replacing", func(t *testing.T) {
		got := item.Item{Body: "first line\n"}

		defineCommon(&got, UpdateRequest{Body: "second line"}, true)

		if got.Body != "second line" {
			t.Errorf("body = %q, want %q", got.Body, "second line")
		}
	})

	t.Run("should append tags when not replacing", func(t *testing.T) {
		got := item.Item{Front: item.Frontmatter{Tags: []string{"tech"}}}

		defineCommon(&got, UpdateRequest{Tags: []string{"meetings"}}, false)

		want := []string{"tech", "meetings"}
		if !reflect.DeepEqual(want, got.Front.Tags) {
			t.Errorf("tags = %#v, want %#v", got.Front.Tags, want)
		}
	})

	t.Run("should replace tags when replacing", func(t *testing.T) {
		got := item.Item{Front: item.Frontmatter{Tags: []string{"tech"}}}

		defineCommon(&got, UpdateRequest{Tags: []string{"meetings"}}, true)

		want := []string{"meetings"}
		if !reflect.DeepEqual(want, got.Front.Tags) {
			t.Errorf("tags = %#v, want %#v", got.Front.Tags, want)
		}
	})

	t.Run("should not duplicate a tag the item already has", func(t *testing.T) {
		got := item.Item{Front: item.Frontmatter{Tags: []string{"tech"}}}

		defineCommon(&got, UpdateRequest{Tags: []string{"tech"}}, false)

		want := []string{"tech"}
		if !reflect.DeepEqual(want, got.Front.Tags) {
			t.Errorf("tags = %#v, want %#v", got.Front.Tags, want)
		}
	})

	t.Run("should append related when not replacing", func(t *testing.T) {
		got := item.Item{Front: item.Frontmatter{Related: []string{"group"}}}

		defineCommon(&got, UpdateRequest{Related: []string{"other"}}, false)

		want := []string{"group", "other"}
		if !reflect.DeepEqual(want, got.Front.Related) {
			t.Errorf("related = %#v, want %#v", got.Front.Related, want)
		}
	})

	t.Run("should not duplicate a related id the item already has", func(t *testing.T) {
		got := item.Item{Front: item.Frontmatter{Related: []string{"group"}}}

		defineCommon(&got, UpdateRequest{Related: []string{"group"}}, false)

		want := []string{"group"}
		if !reflect.DeepEqual(want, got.Front.Related) {
			t.Errorf("related = %#v, want %#v", got.Front.Related, want)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("should update the title of a note", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		got, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Title: "new title"}, false)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		if got.Front.Title != "new title" {
			t.Errorf("Update() title = %q, want %q", got.Front.Title, "new title")
		}

		raw, err := os.ReadFile(seeded.Path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		if !strings.Contains(string(raw), "title: new title") {
			t.Errorf("file content = %q, want it to contain %q", string(raw), "title: new title")
		}
	})

	t.Run("should update the status of a task", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeTask
		req.Due = "2026-08-20"
		seeded := seedItem(t, db, dir, req)

		got, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeTask, Status: "done"}, false)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		if got.Front.Status != "done" {
			t.Errorf("Update() status = %q, want %q", got.Front.Status, "done")
		}
	})

	t.Run("should reindex so the new value is queryable", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeTask
		req.Due = "2026-08-20"
		seeded := seedItem(t, db, dir, req)

		if _, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeTask, Status: "done"}, false); err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		rows, err := List(db, database.ListFilter{Status: "done"}, 10)
		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(rows) != 1 {
			t.Fatalf("List() returned %d rows, want %d", len(rows), 1)
		}

		if rows[0].ID != seeded.Front.ID {
			t.Errorf("List() ID = %q, want %q", rows[0].ID, seeded.Front.ID)
		}
	})

	t.Run("should fail when the status is invalid", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeTask
		req.Due = "2026-08-20"
		seeded := seedItem(t, db, dir, req)

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeTask, Status: "pending"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		wantMsg := `invalid task status: "pending" (should be: open, done, archived)`
		if err.Error() != wantMsg {
			t.Errorf("Update() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should fail when the id does not exist", func(t *testing.T) {
		db, _ := newTestDB(t)

		_, err := Update(db, UpdateRequest{ID: "missing", Type: item.TypeNote, Title: "x"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "no item found with ID missing") {
			t.Errorf("Update() error = %q, want it to contain %q", err.Error(), "no item found with ID missing")
		}
	})

	t.Run("should fail when the id exists with another type", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeTask, Status: "done"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "no item found with ID") {
			t.Errorf("Update() error = %q, want it to contain %q", err.Error(), "no item found with ID")
		}
	})

	t.Run("should fail when the file left the vault", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if err := os.Remove(seeded.Path); err != nil {
			t.Fatalf("Remove() returned unexpected error: %v", err)
		}

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Title: "new title"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "reading item file") {
			t.Errorf("Update() error = %q, want prefix %q", err.Error(), "reading item file")
		}
	})

	t.Run("should fail when the file is no longer parseable", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if err := os.WriteFile(seeded.Path, []byte("no frontmatter here"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Title: "new title"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "parsing item file") {
			t.Errorf("Update() error = %q, want prefix %q", err.Error(), "parsing item file")
		}
	})

	t.Run("should fail when the resulting item is invalid", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeTask
		req.Due = "2026-08-20"
		seeded := seedItem(t, db, dir, req)

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeTask, Due: "20/08/2026"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "invalid due date format") {
			t.Errorf("Update() error = %q, want it to contain %q", err.Error(), "invalid due date format")
		}
	})

	t.Run("should succeed when appending a tag the item already has", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		got, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Tags: []string{"tech"}}, false)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		want := []string{"tech", "meetings"}
		if !reflect.DeepEqual(want, got.Front.Tags) {
			t.Errorf("Update() tags = %#v, want %#v", got.Front.Tags, want)
		}
	})

	t.Run("should succeed when appending a related id the item already has", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		got, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Related: []string{"tech-group"}}, false)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		want := []string{"tech-group"}
		if !reflect.DeepEqual(want, got.Front.Related) {
			t.Errorf("Update() related = %#v, want %#v", got.Front.Related, want)
		}
	})
}
