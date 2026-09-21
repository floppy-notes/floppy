package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/item"
)

func closedDB(t *testing.T) *Database {
	t.Helper()

	db := newTestDB(t)

	if err := db.Conn.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	return db
}

func TestReadOnClosedConnection(t *testing.T) {
	t.Run("FindByIdAndType should return the query error", func(t *testing.T) {
		db := closedDB(t)

		_, err := FindByIdAndType(db.Conn, "any", "note")

		if err == nil {
			t.Fatal("FindByIdAndType() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "Query failed") {
			t.Errorf("FindByIdAndType() error = %q, want prefix %q", err.Error(), "Query failed")
		}
	})

	t.Run("Search should wrap the query error", func(t *testing.T) {
		db := closedDB(t)

		got, err := Search(db.Conn, "anything", 10)

		if err == nil {
			t.Fatal("Search() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "searching items") {
			t.Errorf("Search() error = %q, want prefix %q", err.Error(), "searching items")
		}

		if got == nil {
			t.Error("Search() returned a nil slice, want an empty one")
		}
	})

	t.Run("List should wrap the query error", func(t *testing.T) {
		db := closedDB(t)

		got, err := List(db.Conn, ListFilter{}, 10)

		if err == nil {
			t.Fatal("List() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "listing items") {
			t.Errorf("List() error = %q, want prefix %q", err.Error(), "listing items")
		}

		if got == nil {
			t.Error("List() returned a nil slice, want an empty one")
		}
	})
}

func TestWriteOnClosedConnection(t *testing.T) {
	t.Run("RunInTransaction should fail to begin", func(t *testing.T) {
		db := closedDB(t)

		err := db.RunInTransaction(func(_ *sql.Tx) error {
			t.Error("the transaction body should not run")
			return nil
		})

		if err == nil {
			t.Fatal("RunInTransaction() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "beginning transaction") {
			t.Errorf("RunInTransaction() error = %q, want prefix %q", err.Error(), "beginning transaction")
		}
	})

	t.Run("Rebuild should propagate the error", func(t *testing.T) {
		db := closedDB(t)

		err := db.Rebuild([]item.Item{baseNote})

		if err == nil {
			t.Fatal("Rebuild() returned nil error, want an error")
		}
	})

	t.Run("Reindex should propagate the error", func(t *testing.T) {
		db := closedDB(t)

		err := db.Reindex(baseNote)

		if err == nil {
			t.Fatal("Reindex() returned nil error, want an error")
		}
	})
}

func TestRebuildRollsBackOnBadItem(t *testing.T) {
	t.Run("should report which item failed", func(t *testing.T) {
		db := newTestDB(t)

		duplicatedTag := baseNote
		duplicatedTag.Front.Tags = []string{"tech", "tech"}

		err := db.Rebuild([]item.Item{duplicatedTag})

		if err == nil {
			t.Fatal("Rebuild() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), duplicatedTag.Front.ID) {
			t.Errorf("Rebuild() error = %q, want it to contain %q", err.Error(), duplicatedTag.Front.ID)
		}

		if !strings.Contains(err.Error(), "inserting tag") {
			t.Errorf("Rebuild() error = %q, want it to contain %q", err.Error(), "inserting tag")
		}
	})

	t.Run("should leave the database empty", func(t *testing.T) {
		db := newTestDB(t)

		duplicatedEdge := baseNote
		duplicatedEdge.Front.Related = []string{"group", "group"}

		if err := db.Rebuild([]item.Item{duplicatedEdge}); err == nil {
			t.Fatal("Rebuild() returned nil error, want an error")
		}

		if got := countRows(t, db, "items"); got != 0 {
			t.Errorf("items = %d, want %d", got, 0)
		}
	})
}

func TestSetDefaults(t *testing.T) {
	t.Run("should point at the home folder", func(t *testing.T) {
		dbc := new(DatabaseConfig)

		if err := dbc.setDefaults(); err != nil {
			t.Fatalf("setDefaults() returned unexpected error: %v", err)
		}

		if dbc.Path == "" {
			t.Error("setDefaults() path is empty, want the home folder")
		}
	})

	t.Run("should be overridden by WithPath", func(t *testing.T) {
		dbc, err := newDBC(WithPath("/tmp/custom"))

		if err != nil {
			t.Fatalf("newDBC() returned unexpected error: %v", err)
		}

		if dbc.Path != "/tmp/custom" {
			t.Errorf("newDBC() path = %q, want %q", dbc.Path, "/tmp/custom")
		}
	})

	t.Run("should keep the default when there is no option", func(t *testing.T) {
		withOption, err := newDBC(WithPath("/tmp/custom"))
		if err != nil {
			t.Fatalf("newDBC() returned unexpected error: %v", err)
		}

		withoutOption, err := newDBC()
		if err != nil {
			t.Fatalf("newDBC() returned unexpected error: %v", err)
		}

		if withoutOption.Path == withOption.Path {
			t.Errorf("newDBC() path = %q, want the default instead", withoutOption.Path)
		}
	})
}

func TestOpenOnAnInvalidPath(t *testing.T) {
	t.Run("should fail when the folder cannot be created", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "a-file")

		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := Open(WithPath(filepath.Join(blocker, "nested")))

		if err == nil {
			t.Fatal("Open() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "creating db folder") {
			t.Errorf("Open() error = %q, want prefix %q", err.Error(), "creating db folder")
		}
	})
}
