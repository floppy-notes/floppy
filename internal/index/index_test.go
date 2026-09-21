package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/database"
)

const validNote = `---
id: 2026-08-12-0001-valid
title: tech sync
type: note
created: "2026-08-12T00:00:00-03:00"
tags:
    - tech
---

body
`

func writeRaw(t *testing.T, root string, name string, content string) {
	t.Helper()

	path := filepath.Join(root, name)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned unexpected error: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() returned unexpected error: %v", err)
	}
}

func countItems(t *testing.T, db *database.Database) int {
	t.Helper()

	rows, err := List(db, database.ListFilter{}, 100)
	if err != nil {
		t.Fatalf("List() returned unexpected error: %v", err)
	}

	return len(rows)
}

func TestRebuild(t *testing.T) {
	t.Run("should index every item of the vault", func(t *testing.T) {
		db, dir := newTestDB(t)

		writeRaw(t, dir, filepath.Join("notes", "2026", "08", "a.md"), validNote)
		writeRaw(t, dir, filepath.Join("notes", "2026", "08", "b.md"),
			strings.Replace(validNote, "0001-valid", "0002-valid", 1))

		if err := Rebuild(db, dir); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		if got := countItems(t, db); got != 2 {
			t.Errorf("indexed items = %d, want %d", got, 2)
		}
	})

	t.Run("should skip the files it cannot parse", func(t *testing.T) {
		db, dir := newTestDB(t)

		writeRaw(t, dir, "good.md", validNote)
		writeRaw(t, dir, "broken.md", "no frontmatter here")

		if err := Rebuild(db, dir); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		if got := countItems(t, db); got != 1 {
			t.Errorf("indexed items = %d, want %d", got, 1)
		}
	})

	t.Run("should index an empty vault without error", func(t *testing.T) {
		db, dir := newTestDB(t)

		if err := Rebuild(db, dir); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		if got := countItems(t, db); got != 0 {
			t.Errorf("indexed items = %d, want %d", got, 0)
		}
	})

	t.Run("should drop the items that left the vault", func(t *testing.T) {
		db, dir := newTestDB(t)
		seedItem(t, db, dir, baseRequest)

		if err := Rebuild(db, dir); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		if got := countItems(t, db); got != 1 {
			t.Fatalf("indexed items = %d, want %d", got, 1)
		}

		if err := os.RemoveAll(filepath.Join(dir, "notes")); err != nil {
			t.Fatalf("RemoveAll() returned unexpected error: %v", err)
		}

		if err := Rebuild(db, dir); err != nil {
			t.Fatalf("Rebuild() returned unexpected error: %v", err)
		}

		if got := countItems(t, db); got != 0 {
			t.Errorf("indexed items = %d, want %d", got, 0)
		}
	})

	t.Run("should fail when the vault does not exist", func(t *testing.T) {
		db, dir := newTestDB(t)

		err := Rebuild(db, filepath.Join(dir, "missing"))

		if err == nil {
			t.Fatal("Rebuild() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "walking vault") {
			t.Errorf("Rebuild() error = %q, want prefix %q", err.Error(), "walking vault")
		}
	})
}
