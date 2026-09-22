package index

import (
	"os"
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/item"
)

func TestFindByID(t *testing.T) {
	t.Run("should return the item with its body", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Body = "notes about the deploy pipeline"
		seeded := seedItem(t, db, dir, req)

		got, err := FindByID(db, seeded.Front.ID)

		if err != nil {
			t.Fatalf("FindByID() returned unexpected error: %v", err)
		}

		if got.Front.ID != seeded.Front.ID {
			t.Errorf("id = %q, want %q", got.Front.ID, seeded.Front.ID)
		}

		if got.Body != "notes about the deploy pipeline" {
			t.Errorf("body = %q, want %q", got.Body, "notes about the deploy pipeline")
		}

		if got.Path != seeded.Path {
			t.Errorf("path = %q, want %q", got.Path, seeded.Path)
		}
	})

	t.Run("should find an item of any type", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeTask
		req.Due = "2026-08-20"
		seeded := seedItem(t, db, dir, req)

		got, err := FindByID(db, seeded.Front.ID)

		if err != nil {
			t.Fatalf("FindByID() returned unexpected error: %v", err)
		}

		if got.Front.Type != "task" {
			t.Errorf("type = %q, want %q", got.Front.Type, "task")
		}
	})

	t.Run("should fail when the id does not exist", func(t *testing.T) {
		db, _ := newTestDB(t)

		_, err := FindByID(db, "missing")

		if err == nil {
			t.Fatal("FindByID() returned nil error, want an error")
		}

		wantMsg := "no item found with ID missing"
		if err.Error() != wantMsg {
			t.Errorf("FindByID() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should fail when the file left the vault", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if err := os.Remove(seeded.Path); err != nil {
			t.Fatalf("Remove() returned unexpected error: %v", err)
		}

		_, err := FindByID(db, seeded.Front.ID)

		if err == nil {
			t.Fatal("FindByID() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "reading item file") {
			t.Errorf("FindByID() error = %q, want prefix %q", err.Error(), "reading item file")
		}
	})

	t.Run("should fail when the file is no longer parseable", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if err := os.WriteFile(seeded.Path, []byte("no frontmatter here"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := FindByID(db, seeded.Front.ID)

		if err == nil {
			t.Fatal("FindByID() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "parsing item file") {
			t.Errorf("FindByID() error = %q, want prefix %q", err.Error(), "parsing item file")
		}
	})
}
