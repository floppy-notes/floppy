package index

import (
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
)

func TestSearch(t *testing.T) {
	t.Run("should find an item by a word of its body", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Body = "notes about the deploy pipeline"
		seeded := seedItem(t, db, dir, req)

		got, err := Search(db, "pipeline", 10)

		if err != nil {
			t.Fatalf("Search() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("Search() returned %d rows, want %d", len(got), 1)
		}

		if got[0].ID != seeded.Front.ID {
			t.Errorf("Search() ID = %q, want %q", got[0].ID, seeded.Front.ID)
		}
	})

	t.Run("should return an empty list when nothing matches", func(t *testing.T) {
		db, dir := newTestDB(t)
		seedItem(t, db, dir, baseRequest)

		got, err := Search(db, "kubernetes", 10)

		if err != nil {
			t.Fatalf("Search() returned unexpected error: %v", err)
		}

		if len(got) != 0 {
			t.Errorf("Search() returned %d rows, want %d", len(got), 0)
		}
	})

	t.Run("should propagate a query syntax error", func(t *testing.T) {
		db, dir := newTestDB(t)
		seedItem(t, db, dir, baseRequest)

		_, err := Search(db, `"unbalanced`, 10)

		if err == nil {
			t.Fatal("Search() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "check for unbalanced quotes") {
			t.Errorf("Search() error = %q, want it to contain %q", err.Error(), "check for unbalanced quotes")
		}
	})
}

func TestListDelegation(t *testing.T) {
	t.Run("should return every indexed item", func(t *testing.T) {
		db, dir := newTestDB(t)
		seedItem(t, db, dir, baseRequest)

		got, err := List(db, database.ListFilter{}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Errorf("List() returned %d rows, want %d", len(got), 1)
		}
	})

	t.Run("should apply the filter", func(t *testing.T) {
		db, dir := newTestDB(t)
		seedItem(t, db, dir, baseRequest)

		got, err := List(db, database.ListFilter{Type: string(item.TypeTask)}, 10)

		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}

		if len(got) != 0 {
			t.Errorf("List() returned %d rows, want %d", len(got), 0)
		}
	})
}
