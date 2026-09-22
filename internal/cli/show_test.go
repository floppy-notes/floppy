package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetShowFlags(t *testing.T) {
	t.Helper()

	original := showFlags
	t.Cleanup(func() { showFlags = original })
}

func TestShowCmd(t *testing.T) {
	t.Run("should render the item of the given id", func(t *testing.T) {
		useTempVault(t)
		resetShowFlags(t)

		id := seedNote(t, "tech sync")

		out, err := runRoot(t, "show", id)

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.ID != id {
			t.Errorf("id = %q, want %q", got.ID, id)
		}

		if got.Title != "tech sync" {
			t.Errorf("title = %q, want %q", got.Title, "tech sync")
		}

		if got.Type != "note" {
			t.Errorf("type = %q, want %q", got.Type, "note")
		}
	})

	t.Run("should find an item of any type", func(t *testing.T) {
		useTempVault(t)
		resetShowFlags(t)

		id := seedTask(t, "write tests")

		out, err := runRoot(t, "show", id)

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Type != "task" {
			t.Errorf("type = %q, want %q", got.Type, "task")
		}

		if got.Due == nil || *got.Due != "2026-08-20" {
			t.Errorf("due = %v, want %q", got.Due, "2026-08-20")
		}
	})

	t.Run("should print only the body when body-only is set", func(t *testing.T) {
		useTempVault(t)
		resetShowFlags(t)

		id := seedNote(t, "tech sync")

		out, err := runRoot(t, "show", id, "--body-only")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if strings.TrimSpace(out) != "original body" {
			t.Errorf("output = %q, want %q", strings.TrimSpace(out), "original body")
		}

		if strings.Contains(out, `"id"`) {
			t.Errorf("output = %q, want it to not contain %q", out, `"id"`)
		}
	})

	t.Run("should fail when the id does not exist", func(t *testing.T) {
		useTempVault(t)
		resetShowFlags(t)

		_, err := runRoot(t, "show", "missing")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		wantMsg := "no item found with ID missing"
		if err.Error() != wantMsg {
			t.Errorf("Execute() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should require exactly one id", func(t *testing.T) {
		useTempVault(t)
		resetShowFlags(t)

		if _, err := runRoot(t, "show"); err == nil {
			t.Error("Execute() with no id returned nil error, want an error")
		}

		if _, err := runRoot(t, "show", "a", "b"); err == nil {
			t.Error("Execute() with two ids returned nil error, want an error")
		}
	})

	t.Run("should fail when the file left the vault", func(t *testing.T) {
		useTempVault(t)
		resetShowFlags(t)

		id := seedNote(t, "tech sync")

		if err := os.RemoveAll(filepath.Join(vaultPath, "notes")); err != nil {
			t.Fatalf("RemoveAll() returned unexpected error: %v", err)
		}

		_, err := runRoot(t, "show", id)

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "reading item file") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "reading item file")
		}
	})

	t.Run("should fail when the database cannot be opened", func(t *testing.T) {
		resetShowFlags(t)

		original := vaultPath
		t.Cleanup(func() { vaultPath = original })

		blocker := filepath.Join(t.TempDir(), "a-file")
		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}
		vaultPath = filepath.Join(blocker, "nested")

		_, err := runRoot(t, "show", "any")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "opening database") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "opening database")
		}
	})
}
