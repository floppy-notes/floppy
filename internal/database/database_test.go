package database

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/floppy-notes/floppy/internal/domain"
	"github.com/floppy-notes/floppy/internal/item"
)

func countRows(t *testing.T, db *Database, table string) int {
	t.Helper()

	var count int
	if err := db.Conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("counting %s returned unexpected error: %v", table, err)
	}

	return count
}

func TestOpen(t *testing.T) {
	t.Run("should create the database file", func(t *testing.T) {
		dir := t.TempDir()

		db, err := Open(WithPath(dir))
		if err != nil {
			t.Fatalf("Open() returned unexpected error: %v", err)
		}
		defer db.Conn.Close()

		if err := db.Conn.Ping(); err != nil {
			t.Errorf("Ping() returned unexpected error: %v", err)
		}

		if db.config.Path != dir {
			t.Errorf("Open() path = %q, want %q", db.config.Path, dir)
		}
	})

	t.Run("should create the folder when it does not exist", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "vault")

		db, err := Open(WithPath(dir))
		if err != nil {
			t.Fatalf("Open() returned unexpected error: %v", err)
		}
		defer db.Conn.Close()

		if _, err := db.Conn.Exec("SELECT 1"); err != nil {
			t.Errorf("Exec() returned unexpected error: %v", err)
		}
	})

	t.Run("should create every table of the schema", func(t *testing.T) {
		db := newTestDB(t)

		for _, table := range []string{"items", "items_fts", "tags", "edges"} {
			if _, err := db.Conn.Exec("SELECT * FROM " + table); err != nil {
				t.Errorf("table %q is missing: %v", table, err)
			}
		}
	})

	t.Run("should be idempotent on an existing database", func(t *testing.T) {
		dir := t.TempDir()

		first, err := Open(WithPath(dir))
		if err != nil {
			t.Fatalf("Open() returned unexpected error: %v", err)
		}
		first.Conn.Close()

		second, err := Open(WithPath(dir))
		if err != nil {
			t.Fatalf("Open() returned unexpected error: %v", err)
		}
		defer second.Conn.Close()

		wantName := filepath.Join(dir, domain.DbName)
		if _, err := second.Conn.Exec("SELECT 1"); err != nil {
			t.Errorf("reopening %q returned unexpected error: %v", wantName, err)
		}
	})
}

func TestRunInTransaction(t *testing.T) {
	t.Run("should commit when the function succeeds", func(t *testing.T) {
		db := newTestDB(t)

		err := db.RunInTransaction(func(tx *sql.Tx) error {
			return insertItem(tx, ItemRow{ID: "a", Title: "t", Type: "note", Created: "c", Path: "p"})
		})

		if err != nil {
			t.Fatalf("RunInTransaction() returned unexpected error: %v", err)
		}

		if got := countRows(t, db, "items"); got != 1 {
			t.Errorf("items = %d, want %d", got, 1)
		}
	})

	t.Run("should roll back when the function fails", func(t *testing.T) {
		db := newTestDB(t)

		wantErr := errors.New("boom")

		err := db.RunInTransaction(func(tx *sql.Tx) error {
			if err := insertItem(tx, ItemRow{ID: "a", Title: "t", Type: "note", Created: "c", Path: "p"}); err != nil {
				return err
			}
			return wantErr
		})

		if !errors.Is(err, wantErr) {
			t.Errorf("RunInTransaction() error = %v, want %v", err, wantErr)
		}

		if got := countRows(t, db, "items"); got != 0 {
			t.Errorf("items = %d, want %d", got, 0)
		}
	})
}

func TestRebuild(t *testing.T) {
	t.Run("should index every item", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		if got := countRows(t, db, "items"); got != 3 {
			t.Errorf("items = %d, want %d", got, 3)
		}

		if got := countRows(t, db, "items_fts"); got != 3 {
			t.Errorf("items_fts = %d, want %d", got, 3)
		}

		if got := countRows(t, db, "tags"); got != 3 {
			t.Errorf("tags = %d, want %d", got, 3)
		}

		if got := countRows(t, db, "edges"); got != 1 {
			t.Errorf("edges = %d, want %d", got, 1)
		}
	})

	t.Run("should clear the previous content", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		if err := db.Rebuild([]item.Item{baseNote}); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		if got := countRows(t, db, "items"); got != 1 {
			t.Errorf("items = %d, want %d", got, 1)
		}

		if got := countRows(t, db, "tags"); got != 2 {
			t.Errorf("tags = %d, want %d", got, 2)
		}
	})

	t.Run("should empty the database when there is no item", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		if err := db.Rebuild(nil); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		for _, table := range []string{"items", "items_fts", "tags", "edges"} {
			if got := countRows(t, db, table); got != 0 {
				t.Errorf("%s = %d, want %d", table, got, 0)
			}
		}
	})

	t.Run("should fail when two items share an id", func(t *testing.T) {
		db := newTestDB(t)

		err := db.Rebuild([]item.Item{baseNote, baseNote})

		if err == nil {
			t.Fatal("Rebuild() returned nil error, want an error")
		}

		if got := countRows(t, db, "items"); got != 0 {
			t.Errorf("items = %d, want %d", got, 0)
		}
	})
}

func TestReindex(t *testing.T) {
	t.Run("should replace the row of an existing item", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		updated := baseTask
		updated.Front.Status = "done"

		if err := db.Reindex(updated); err != nil {
			t.Fatalf("Reindex() returned unexpected error: %v", err)
		}

		if got := countRows(t, db, "items"); got != 3 {
			t.Errorf("items = %d, want %d", got, 3)
		}

		row, err := FindByIdAndType(db.Conn, updated.Front.ID, "task")
		if err != nil {
			t.Fatalf("FindByIdAndType() returned unexpected error: %v", err)
		}

		if row.Status.String != "done" {
			t.Errorf("status = %q, want %q", row.Status.String, "done")
		}
	})

	t.Run("should index an item that is not in the database yet", func(t *testing.T) {
		db := newTestDB(t)

		if err := db.Reindex(baseNote); err != nil {
			t.Fatalf("Reindex() returned unexpected error: %v", err)
		}

		if got := countRows(t, db, "items"); got != 1 {
			t.Errorf("items = %d, want %d", got, 1)
		}
	})

	t.Run("should not leave orphan tags behind", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		updated := baseNote
		updated.Front.Tags = []string{"only-one"}

		if err := db.Reindex(updated); err != nil {
			t.Fatalf("Reindex() returned unexpected error: %v", err)
		}

		if got := countRows(t, db, "tags"); got != 2 {
			t.Errorf("tags = %d, want %d", got, 2)
		}
	})

	t.Run("should not duplicate the fts row", func(t *testing.T) {
		db := newTestDB(t)
		seedAll(t, db)

		if err := db.Reindex(baseNote); err != nil {
			t.Fatalf("Reindex() returned unexpected error: %v", err)
		}

		if got := countRows(t, db, "items_fts"); got != 3 {
			t.Errorf("items_fts = %d, want %d", got, 3)
		}
	})
}
