package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRaw(t *testing.T, root string, name string, content string) string {
	t.Helper()

	path := filepath.Join(root, name)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned unexpected error: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() returned unexpected error: %v", err)
	}

	return path
}

const validNote = `---
id: 2026-08-12-0001-valid
title: tech sync
type: note
created: "2026-08-12T00:00:00-03:00"
---

body
`

func TestWalk(t *testing.T) {
	t.Run("should return the parsed items of the vault", func(t *testing.T) {
		root := t.TempDir()
		path := writeRaw(t, root, filepath.Join("notes", "2026", "08", "valid.md"), validNote)

		items, skipped, err := Walk(root)

		if err != nil {
			t.Fatalf("Walk() returned unexpected error: %v", err)
		}

		if len(skipped) != 0 {
			t.Errorf("Walk() skipped = %d, want %d", len(skipped), 0)
		}

		if len(items) != 1 {
			t.Fatalf("Walk() items = %d, want %d", len(items), 1)
		}

		if items[0].Front.ID != "2026-08-12-0001-valid" {
			t.Errorf("item ID = %q, want %q", items[0].Front.ID, "2026-08-12-0001-valid")
		}

		if items[0].Path != path {
			t.Errorf("item Path = %q, want %q", items[0].Path, path)
		}
	})

	t.Run("should return an empty list when the vault has no items", func(t *testing.T) {
		items, skipped, err := Walk(t.TempDir())

		if err != nil {
			t.Fatalf("Walk() returned unexpected error: %v", err)
		}

		if len(skipped) != 0 {
			t.Errorf("Walk() skipped = %d, want %d", len(skipped), 0)
		}

		if len(items) != 0 {
			t.Errorf("Walk() items = %d, want %d", len(items), 0)
		}
	})

	t.Run("should ignore files that are not markdown", func(t *testing.T) {
		root := t.TempDir()
		writeRaw(t, root, "floppy.db", "binary junk")
		writeRaw(t, root, "notes.txt", validNote)

		items, skipped, err := Walk(root)

		if err != nil {
			t.Fatalf("Walk() returned unexpected error: %v", err)
		}

		if len(skipped) != 0 {
			t.Errorf("Walk() skipped = %d, want %d", len(skipped), 0)
		}

		if len(items) != 0 {
			t.Errorf("Walk() items = %d, want %d", len(items), 0)
		}
	})

	t.Run("should skip a file that cannot be parsed", func(t *testing.T) {
		root := t.TempDir()
		writeRaw(t, root, "valid.md", validNote)
		writeRaw(t, root, "broken.md", "no frontmatter here")

		items, skipped, err := Walk(root)

		if err != nil {
			t.Fatalf("Walk() returned unexpected error: %v", err)
		}

		if len(items) != 1 {
			t.Errorf("Walk() items = %d, want %d", len(items), 1)
		}

		if len(skipped) != 1 {
			t.Fatalf("Walk() skipped = %d, want %d", len(skipped), 1)
		}

		if !strings.Contains(skipped[0].Error(), "broken.md") {
			t.Errorf("skipped error = %q, want it to contain %q", skipped[0].Error(), "broken.md")
		}
	})

	t.Run("should skip a file whose frontmatter is invalid", func(t *testing.T) {
		root := t.TempDir()
		writeRaw(t, root, "valid.md", validNote)
		writeRaw(t, root, "invalid.md", strings.Replace(validNote, "type: note", "type: unknown", 1))

		items, skipped, err := Walk(root)

		if err != nil {
			t.Fatalf("Walk() returned unexpected error: %v", err)
		}

		if len(items) != 1 {
			t.Errorf("Walk() items = %d, want %d", len(items), 1)
		}

		if len(skipped) != 1 {
			t.Fatalf("Walk() skipped = %d, want %d", len(skipped), 1)
		}

		if !strings.Contains(skipped[0].Error(), "invalid type") {
			t.Errorf("skipped error = %q, want it to contain %q", skipped[0].Error(), "invalid type")
		}
	})

	t.Run("should fail when the vault folder does not exist", func(t *testing.T) {
		missingRoot := filepath.Join(t.TempDir(), "missing")

		_, _, err := Walk(missingRoot)

		if err == nil {
			t.Fatal("Walk() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "walking vault") {
			t.Errorf("Walk() error = %q, want prefix %q", err.Error(), "walking vault")
		}
	})
}
